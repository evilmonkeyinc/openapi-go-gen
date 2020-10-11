package parser

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func New(swagger *openapi3.Swagger) (*SwaggerParser, error) {
	parser := &SwaggerParser{
		swagger:    swagger,
		APIs:       map[string]*APIWrapper{},
		Parameters: map[string]*openapi3.Parameter{},
		Responses:  map[string]*openapi3.Response{},
		Schemas:    map[string]*openapi3.Schema{},
	}

	if err := parser.parseComponents(); err != nil {
		return nil, fmt.Errorf("failed to parse components %w", err)
	}

	if err := parser.parsePaths(); err != nil {
		return nil, fmt.Errorf("failed to parse paths: %w", err)
	}
	return parser, nil
}

type SwaggerParser struct {
	swagger    *openapi3.Swagger
	APIs       map[string]*APIWrapper
	Parameters map[string]*openapi3.Parameter
	Responses  map[string]*openapi3.Response
	Schemas    map[string]*openapi3.Schema
}

func (gen *SwaggerParser) parseComponents() error {
	components := gen.swagger.Components

	for componentKey, ref := range components.Parameters {
		if ref.Ref != "" {
			return fmt.Errorf("components/parameters/%s is a reference, must be value", componentKey)
		}

		gen.Parameters[componentKey] = ref.Value
	}

	for componentKey, ref := range components.Responses {
		if ref.Ref != "" {
			return fmt.Errorf("components/responses/%s is a reference, must be value", componentKey)
		}

		gen.Responses[componentKey] = ref.Value
	}

	for componentKey, ref := range components.Schemas {
		if ref.Ref != "" {
			return fmt.Errorf("components/schemas/%s is a reference, must be value", componentKey)
		}

		gen.Schemas[componentKey] = ref.Value
	}

	return nil
}

func (gen *SwaggerParser) parsePaths() error {
	for _, tag := range gen.swagger.Tags {
		api := new(APIWrapper)
		api.Tag = tag
		api.Functions = make(map[string][]*OperationWrapper)
		gen.APIs[tag.Name] = api
	}

	for pathString, path := range gen.swagger.Paths {
		operations := path.Operations()
		for method, operation := range operations {
			if operation.Deprecated {
				continue
			}
			if len(operation.Tags) == 0 {
				return fmt.Errorf("operation %s from path %s does not have any tags, at least one tag is required", method, pathString)
			}
			if operation.OperationID == "" {
				return fmt.Errorf("operation %s from path %s does not have an operation ID", method, pathString)
			}
			for _, tag := range operation.Tags {
				subAPI := gen.APIs[tag]
				if subAPI == nil {
					fmt.Printf("cannot add operation %s from path %s to sub API '%s' as it does not exist\n", method, pathString, tag)
					continue
				}
				subAPI.AddFunction(pathString, path.Parameters, method, operation)
			}
		}
	}

	return nil
}
