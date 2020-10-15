package components

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
)

const mimeApplicationJSON string = "application/json"

// NewResponseBuilder creates a new Builder for the openapi3.ResponseRef object
func NewResponseBuilder(module, packageName, parentName string, response *openapi3.ResponseRef) ComponentBuilder {
	return &responseBuilder{
		Module:      module,
		PackageName: packageName,
		ParentName:  parentName,
		ResponseRef: response,
	}
}

type responseBuilder struct {
	Module      string
	PackageName string
	ParentName  string
	ResponseRef *openapi3.ResponseRef
}

func (builder *responseBuilder) AsStruct(structName string) (jen.Code, []jen.Code, error) {
	if builder.ResponseRef.Ref != "" {
		return nil,
			nil,
			fmt.Errorf("cannot create a struct from a ResponseRef.Ref")
	} else if response := builder.ResponseRef.Value; response != nil {
		extras := make([]jen.Code, 0)

		schemaName := fmt.Sprintf("%s", strings.Title(structName))
		params := make([]jen.Code, 0)

		content := response.Content.Get(mimeApplicationJSON)
		if content == nil || content.Schema == nil {
			// TODO : non-application JSON, just allow anything and return
			return nil, nil, fmt.Errorf("no response type given, need to allow any")
		}

		schema := content.Schema
		if schema.Ref != "" {
			split := strings.Split(schema.Ref, "/")
			parameterName := split[len(split)-1]
			packageName := split[len(split)-2]
			params = append(params, jen.Commentf("%s %s", packageName, parameterName).Line())
			params = append(
				params,
				jen.Qual(fmt.Sprintf("%s/%s", builder.Module, packageName), parameterName).Tag(map[string]string{"bson": ",inline"}),
			)
		} else if schema := schema.Value; schema != nil {
			for parameterName, property := range schema.Properties {
				builder := NewSchemaBuilder(builder.Module, builder.PackageName, builder.ParentName, property, true)
				main, extra, err := builder.AsField(parameterName)
				if err != nil {
					return nil, nil, err
				}
				extras = append(extras, extra...)
				params = append(params, main)
			}
		} else {
			return nil, nil, fmt.Errorf("no ref or value")
		}

		return jen.Commentf("%s: %s", schemaName, *response.Description).Line().Type().Id(schemaName).Struct(params...).Line(),
			extras,
			nil
	}

	return nil, nil, fmt.Errorf("openapi3.ResponseRef must have non-empty Ref or non-nil Value")
}
func (builder *responseBuilder) AsField(fieldName string) (jen.Code, []jen.Code, error) {
	if builder.ResponseRef.Ref != "" {
		refSplit := strings.Split(builder.ResponseRef.Ref, "/")
		structName := strings.Title(refSplit[len(refSplit)-1])
		packageName := refSplit[len(refSplit)-2]
		return jen.Id(fmt.Sprintf("Response%s", fieldName)).Op("*").Qual(fmt.Sprintf("%s/%s", builder.Module, packageName), structName),
			nil,
			nil
	} else if value := builder.ResponseRef.Value; value != nil {
		extras := make([]jen.Code, 0)
		params := make([]jen.Code, 0)

		content := builder.ResponseRef.Value.Content.Get(mimeApplicationJSON)
		if content == nil || content.Schema == nil {
			// TODO : non-application JSON, just allow anything and return
			return nil, nil, fmt.Errorf("no response type given, need to allow any")
		}

		schemaBuilder := NewSchemaBuilder(builder.Module, builder.PackageName, builder.ParentName, content.Schema, false)
		main, extra, err := schemaBuilder.AsField(fmt.Sprintf("Response%s", fieldName))
		if err != nil {
			return nil, nil, err
		}
		params = append(params, main)
		extras = append(extras, extra...)

		return jen.Add(params...),
			extras,
			nil
	}

	return nil, nil, fmt.Errorf("openapi3.ResponseRef must have non-empty Ref or non-nil Value")
}
