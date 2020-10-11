package builder

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/evilmonkeyinc/openapi-go-gen/pkg/parser"
	"github.com/getkin/kin-openapi/openapi3"
)

type Builder interface {
	AsStruct(structName string) jen.Code
	AsField(fieldName string) jen.Code
}

func BuildSchemaFile(module, componentName string, schema *openapi3.Schema) string {
	file := jen.NewFile("schemas")
	file.HeaderComment("// Code generated by github.com/evilmonkeyinc/openapi-go-gen. DO NOT EDIT.")

	schemaBuilder := NewSchemaBuilder(module, "schemas", &openapi3.SchemaRef{
		Value: schema,
	}, true)
	file.Add(schemaBuilder.AsStruct(componentName))

	return file.GoString()
}

func BuildResponseFile(module, componentName string, response *openapi3.Response) string {
	schemaName := fmt.Sprintf("%s", strings.Title(componentName))

	file := jen.NewFile("responses")
	file.HeaderComment("// Code generated by github.com/evilmonkeyinc/openapi-go-gen. DO NOT EDIT.")

	params := make([]jen.Code, 0)
	for _, content := range response.Content {
		schema := content.Schema

		if schema.Ref != "" {
			split := strings.Split(schema.Ref, "/")
			parameterName := split[len(split)-1]
			packageName := split[len(split)-2]
			params = append(
				params,
				jen.Qual(fmt.Sprintf("%s/%s", module, packageName), parameterName).Tag(map[string]string{"bson": ",inline"}),
			)
		} else if schema := schema.Value; schema != nil {
			for parameterName, property := range schema.Properties {
				builder := NewSchemaBuilder(module, "responses", property, true)
				params = append(
					params,
					builder.AsField(parameterName),
				)
			}
		} else {
			panic("no ref or value")
		}
	}

	file.Commentf("%s: %s", schemaName, response.Description)
	file.Type().Id(schemaName).Struct(params...)
	file.Line()

	return file.GoString()
}

func BuildAPIFile(module string, api *parser.APIWrapper) string {
	apiName := fmt.Sprintf("%sAPI", strings.Title(api.Tag.Name))

	moduleSplit := strings.Split(module, "/")

	file := jen.NewFile(moduleSplit[len(moduleSplit)-1])
	file.HeaderComment("// Code generated by github.com/evilmonkeyinc/openapi-go-gen. DO NOT EDIT.")

	methods := make([]jen.Code, 0)

	for _, wrappers := range api.Functions {
		for _, wrapper := range wrappers {
			operationID := strings.Title(wrapper.Operation.OperationID)
			requestID := fmt.Sprintf("%sRequest", operationID)
			responseID := fmt.Sprintf("%sResponse", operationID)
			methods = append(
				methods,
				jen.Commentf("%s %s", operationID, wrapper.Operation.Description).Line().Id(operationID).Params(
					jen.Id("ctx").Qual("context", "Context"),
					jen.Id("request").Op("*").Qual("", requestID),
				).Params(jen.Op("*").Qual("", responseID), jen.Error()),
			)

			requestParams := make([]jen.Code, 0)
			for _, parameter := range wrapper.PathParameters {
				requestParams = append(
					requestParams,
					buildParameter(module, parameter),
				)
			}
			for _, parameter := range wrapper.Operation.Parameters {
				requestParams = append(
					requestParams,
					buildParameter(module, parameter),
				)
			}

			file.Commentf("%s encapsulates the expected request for %s()", requestID, operationID)
			file.Type().Id(requestID).Struct(requestParams...)
			file.Line()

			responseParams := make([]jen.Code, 0)
			for _, response := range wrapper.Operation.Responses {
				if response.Ref != "" {
					refSplit := strings.Split(response.Ref, "/")
					objType := strings.Title(refSplit[len(refSplit)-1])
					packageType := refSplit[len(refSplit)-2]
					responseParams = append(
						responseParams,
						jen.Qual(fmt.Sprintf("%s/%s", module, packageType), objType).Tag(map[string]string{"bson": ",inline"}),
					)

				} else if value := response.Value; value != nil {

				}
			}

			file.Commentf("%s encapsulates the expected response for %s()", responseID, operationID)
			file.Type().Id(responseID).Struct(responseParams...)
			file.Line()
		}
	}
	file.Commentf("%s functions linked to %s tag", apiName, api.Tag.Name)
	file.Commentf("%s: %s", api.Tag.Name, api.Tag.Description)
	if api.Tag.ExternalDocs != nil {
		file.Commentf("%s", api.Tag.ExternalDocs.Description)
		file.Commentf("%s", api.Tag.ExternalDocs.URL)
	}
	file.Type().Id(apiName).Interface(methods...)
	file.Line()

	return file.GoString()
}

func buildParameter(module string, parameter *openapi3.ParameterRef) jen.Code {
	if parameter.Value != nil {
		value := parameter.Value
		param := jen.Comment(value.Description).Line().Id(strings.Title(value.Name))

		if schema := value.Schema; schema != nil {
			if schema.Ref != "" {
				panic("not supporting schema.Ref for buildParameter")
			} else if schema.Value != nil {
				param = addPrimitiveTypeFromSchema(param, schema.Value)
			} else {
				panic("schema supplied with no ref or value")
			}

		} else {
			param = param.Interface()
		}

		required := ",omitempty"
		if value.Required {
			required = ""
		}
		param.Tag(map[string]string{
			"json": fmt.Sprintf("%s%s", value.Name, required),
			"yaml": fmt.Sprintf("%s%s", value.Name, required),
		})

		return param
	} else if parameter.Ref != "" {
		split := strings.Split(parameter.Ref, "/")
		paramType := split[len(split)-1]
		packageName := split[len(split)-2]
		return jen.Id(paramType).Qual(fmt.Sprintf("%s/%s", module, packageName), paramType)
	}

	return nil
}
