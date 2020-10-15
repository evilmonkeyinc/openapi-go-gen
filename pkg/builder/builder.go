package builder

import (
	"fmt"
	"os"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/evilmonkeyinc/openapi-go-gen/pkg/builder/components"
	"github.com/getkin/kin-openapi/openapi3"
)

func WriteSchemaFile(output, module, componentName string, schemaRef *openapi3.SchemaRef) error {
	file := jen.NewFile("schemas")
	file.HeaderComment("// Code generated by github.com/evilmonkeyinc/openapi-go-gen. DO NOT EDIT.")

	schemaBuilder := components.NewSchemaBuilder(module, "schemas", "", schemaRef, true)
	main, extra, err := schemaBuilder.AsStruct(componentName)
	if err != nil {
		return err
	}
	file.Add(extra...)
	file.Add(main)

	fileWritter, err := os.Create(fmt.Sprintf("%s/schemas/%s.go", output, componentName))
	if err != nil {
		return err
	}
	defer fileWritter.Close()

	return file.Render(fileWritter)
}

func WriteResponseFile(output, module, componentName string, responseRef *openapi3.ResponseRef) error {
	file := jen.NewFile("responses")
	file.HeaderComment("// Code generated by github.com/evilmonkeyinc/openapi-go-gen. DO NOT EDIT.")

	responseBuilder := components.NewResponseBuilder(module, "responses", "", responseRef)
	main, extra, err := responseBuilder.AsStruct(componentName)
	if err != nil {
		return err
	}
	file.Add(extra...)
	file.Add(main)

	fileWritter, err := os.Create(fmt.Sprintf("%s/responses/%s.go", output, componentName))
	if err != nil {
		return err
	}
	defer fileWritter.Close()

	return file.Render(fileWritter)
}

func WriteParameterFile(output, module, componentName string, responseRef *openapi3.ParameterRef) error {
	file := jen.NewFile("parameters")
	file.HeaderComment("// Code generated by github.com/evilmonkeyinc/openapi-go-gen. DO NOT EDIT.")

	responseBuilder := components.NewParameterBuilder(module, "parameters", "", responseRef, true)
	main, extra, err := responseBuilder.AsStruct(componentName)
	if err != nil {
		return err
	}
	file.Add(extra...)
	file.Add(main)

	fileWritter, err := os.Create(fmt.Sprintf("%s/parameters/%s.go", output, componentName))
	if err != nil {
		return err
	}
	defer fileWritter.Close()

	return file.Render(fileWritter)
}

func BuildServiceInterfaceFile(module string, specification *openapi3.Swagger) (*jen.File, error) {
	moduleSplit := strings.Split(module, "/")
	packageName := moduleSplit[len(moduleSplit)-1]

	file := jen.NewFile(packageName)
	file.HeaderComment("// Code generated by github.com/evilmonkeyinc/openapi-go-gen. DO NOT EDIT.")

	paths := specification.Paths
	for _, tag := range specification.Tags {

		apiName := fmt.Sprintf("%sAPI", strings.Title(tag.Name))

		functions := make([]jen.Code, 0)
		for _, path := range paths {
			for _, operation := range path.Operations() {

				if operation.Tags[0] != tag.Name {
					continue
				}

				operationBuilder := NewOperationBuilder(module, packageName, operation, path.Parameters)
				function, functionStructs, err := operationBuilder.AsField("")
				if err != nil {
					return nil, err
				}

				functions = append(functions, function)
				file.Add(functionStructs...)
			}
		}

		file.Commentf("%s functions linked to %s tag", apiName, tag.Name)
		file.Commentf("%s: %s", tag.Name, tag.Description)
		if tag.ExternalDocs != nil {
			file.Commentf("%s", tag.ExternalDocs.Description)
			file.Commentf("%s", tag.ExternalDocs.URL)
		}
		file.Type().Id(apiName).Interface(functions...)
		file.Line()
	}

	return file, nil
}
