package builder

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
)

// NewSchemaBuilder creates a new Builder for the openapi3.SchemaRef object
func NewSchemaBuilder(module, packageName string, parentID string, schemaRef *openapi3.SchemaRef, addTags bool) Builder {
	return &schemaBuilder{
		Module:      module,
		PackageName: packageName,
		ParentID:    parentID,
		SchemaRef:   schemaRef,
		AddTags:     addTags,
	}
}

type schemaBuilder struct {
	AddTags     bool
	Module      string
	PackageName string
	ParentID    string
	SchemaRef   *openapi3.SchemaRef
}

func (builder *schemaBuilder) AsStruct(structName string) (jen.Code, []jen.Code) {
	schema := builder.SchemaRef.Value
	if schema == nil {
		panic("SchemaBuilder.AsStruct does not function without SchemaRef.Value")
	}

	extraStruct := make([]jen.Code, 0)

	schemaName := fmt.Sprintf("%s", strings.Title(structName))

	params := make([]jen.Code, 0)
	for parameterName, property := range schema.Properties {
		builder := NewSchemaBuilder(builder.Module, builder.PackageName, structName, property, true)
		main, extra := builder.AsField(parameterName)
		params = append(params, main)
		extraStruct = append(extraStruct, extra...)
	}

	if allOf := schema.AllOf; len(allOf) > 0 {
		for _, schemaRef := range allOf {
			builder := NewSchemaBuilder(builder.Module, builder.PackageName, structName, schemaRef, true)
			main, extra := builder.AsField("")
			params = append(params, main)
			extraStruct = append(extraStruct, extra...)
		}
	}
	if oneOf := schema.OneOf; len(oneOf) > 0 {
		for _, schemaRef := range oneOf {
			builder := NewSchemaBuilder(builder.Module, builder.PackageName, structName, schemaRef, true)
			main, extra := builder.AsField("")
			params = append(params, main)
			extraStruct = append(extraStruct, extra...)
		}
	}

	return jen.Commentf("%s: %s", schemaName, schema.Description).Line().Type().Id(schemaName).Struct(params...).Line(), extraStruct
}

func (builder *schemaBuilder) AsField(fieldName string) (jen.Code, []jen.Code) {
	var param *jen.Statement
	title := strings.Title(fieldName)

	extraStruct := make([]jen.Code, 0)

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
					builder := NewSchemaBuilder(builder.Module, builder.PackageName, fieldName, property, true)
					main, extra := builder.AsField(name)
					fields = append(fields, main)
					extraStruct = append(extraStruct, extra...)
				}
				return jen.Add(fields...), extraStruct
			default:
				panic("using oneOf or allOf with primitive or array instead of object")
			}
		} else {
			param = jen.Comment(schema.Description).Line().Id(title)

			switch schema.Type {
			case "object":
				// imbedded struct
				fields := make([]jen.Code, 0)
				for name, property := range schema.Properties {
					builder := NewSchemaBuilder(builder.Module, builder.PackageName, fieldName, property, true)
					main, extra := builder.AsField(name)
					fields = append(fields, main)
					extraStruct = append(extraStruct, extra...)
				}

				subID := fmt.Sprintf("%s%s", strings.Title(builder.ParentID), title)
				extraStruct = append(
					extraStruct,
					jen.Type().Id(subID).Struct(fields...),
				)

				param = param.Op("*").Qual("", subID)
			case "array":
				if schema.Items.Ref != "" {
					itemBuilder := NewSchemaBuilder(builder.Module, builder.PackageName, fieldName, schema.Items, false)
					main, extra := itemBuilder.AsField("")
					param = param.Op("[]").Op("*").Add(main)
					extraStruct = append(extraStruct, extra...)
				} else if itemValue := schema.Items.Value; itemValue != nil {
					switch itemValue.Type {
					case "object":
						fallthrough
					case "array":
						itemBuilder := NewSchemaBuilder(builder.Module, builder.PackageName, fieldName, schema.Items, false)
						main, extra := itemBuilder.AsField("")
						param = param.Op("[]").Op("*").Add(main)
						extraStruct = append(extraStruct, extra...)
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

	return param, extraStruct
}
