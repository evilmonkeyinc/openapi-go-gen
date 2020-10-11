package builder

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
)

func NewSchemaBuilder(module, packageName string, schemaRef *openapi3.SchemaRef, addTags bool) Builder {
	return &schemaBuilder{
		Module:      module,
		PackageName: packageName,
		SchemaRef:   schemaRef,
		AddTags:     addTags,
	}
}

type schemaBuilder struct {
	AddTags     bool
	Module      string
	PackageName string
	SchemaRef   *openapi3.SchemaRef
}

func (builder *schemaBuilder) AsStruct(structName string) jen.Code {
	schema := builder.SchemaRef.Value
	if schema == nil {
		panic("SchemaBuilder.AsStruct does not function without SchemaRef.Value")
	}

	schemaName := fmt.Sprintf("%s", strings.Title(structName))

	params := make([]jen.Code, 0)
	for parameterName, property := range schema.Properties {
		builder := NewSchemaBuilder(builder.Module, builder.PackageName, property, true)
		params = append(
			params,
			builder.AsField(parameterName),
		)
	}

	if allOf := schema.AllOf; len(allOf) > 0 {
		for _, schemaRef := range allOf {
			builder := NewSchemaBuilder(builder.Module, builder.PackageName, schemaRef, true)
			params = append(
				params,
				builder.AsField(""),
			)
		}
	}
	if oneOf := schema.OneOf; len(oneOf) > 0 {
		for _, schemaRef := range oneOf {
			builder := NewSchemaBuilder(builder.Module, builder.PackageName, schemaRef, true)
			params = append(
				params,
				builder.AsField(""),
			)
		}
	}

	return jen.Commentf("%s: %s", schemaName, schema.Description).Line().Type().Id(schemaName).Struct(params...).Line()
}

func (builder *schemaBuilder) AsField(fieldName string) jen.Code {
	var param *jen.Statement
	title := strings.Title(fieldName)

	if builder.SchemaRef.Ref != "" {
		split := strings.Split(builder.SchemaRef.Ref, "/")
		typeName := split[len(split)-1]
		packageName := split[len(split)-2]
		path := fmt.Sprintf("%s/%s", builder.Module, packageName)
		if packageName == builder.PackageName {
			path = ""
		}
		if title == "" {
			param = jen.Qual(path, typeName)
		} else {
			param = jen.Id(title).Op("*").Qual(path, typeName)
		}
	} else if schema := builder.SchemaRef.Value; schema != nil {
		if title == "" {
			switch schema.Type {
			case "object":
				// allOf or oneOf, add fields to existing struct
				fields := make([]jen.Code, 0)
				for name, property := range schema.Properties {
					builder := NewSchemaBuilder(builder.Module, builder.PackageName, property, true)
					fields = append(
						fields,
						builder.AsField(name),
					)
				}
				return jen.Add(fields...)
			default:
				panic("using oneof or allof with primitive or array instead of object")
			}
		} else {
			param = jen.Comment(schema.Description).Line().Id(title)

			switch schema.Type {
			case "object":
				// imbedded struct
				fields := make([]jen.Code, 0)
				for name, property := range schema.Properties {
					builder := NewSchemaBuilder(builder.Module, builder.PackageName, property, true)
					fields = append(
						fields,
						builder.AsField(name),
					)
				}

				subID := fmt.Sprintf("%s%s", "Embeded", title)
				// subStruct := jen.Type().Id(subID).Struct(fields...)

				param = param.Qual("", subID)
			case "array":
				if schema.Items.Ref != "" {
					itemBuilder := NewSchemaBuilder(builder.Module, builder.PackageName, schema.Items, false)
					param = param.Op("[]").Add(itemBuilder.AsField(""))
				} else if itemValue := schema.Items.Value; itemValue != nil {
					switch itemValue.Type {
					case "object":
						fallthrough
					case "array":
						itemBuilder := NewSchemaBuilder(builder.Module, builder.PackageName, schema.Items, false)
						param = param.Op("[]").Add(itemBuilder.AsField(""))
					default:
						param = param.Op("[]")
						param = addPrimitiveTypeFromSchema(param, itemValue)
					}
				} else {
					panic("invalid array item, must be Ref or Value")
				}
			default:
				param = addPrimitiveTypeFromSchema(param, schema)
			}
		}
	} else {
		panic("SchemaBuilder.AsField does only supports Ref and Value")
	}

	if builder.AddTags {
		if title == "" {
			param.Tag(map[string]string{
				"bson": ",inline",
			})
		} else {
			param.Tag(map[string]string{
				"json": fmt.Sprintf("%s,omitempty", fieldName),
				"yaml": fmt.Sprintf("%s,omitempty", fieldName),
			})
		}
	}

	return param
}
