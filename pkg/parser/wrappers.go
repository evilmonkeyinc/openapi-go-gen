package parser

import "github.com/getkin/kin-openapi/openapi3"

type APIWrapper struct {
	Tag       *openapi3.Tag
	Functions map[string][]*OperationWrapper
}

func (wrapper *APIWrapper) AddFunction(path string, pathParameters openapi3.Parameters, method string, operation *openapi3.Operation) {
	if wrapper.Functions == nil {
		wrapper.Functions = make(map[string][]*OperationWrapper)
	}

	pathArray := wrapper.Functions[path]
	if pathArray == nil {
		pathArray = make([]*OperationWrapper, 0)
	}
	pathArray = append(pathArray, &OperationWrapper{
		Method:         method,
		Operation:      operation,
		PathParameters: pathParameters,
	})
	wrapper.Functions[path] = pathArray
}

type OperationWrapper struct {
	Method         string
	Operation      *openapi3.Operation
	PathParameters openapi3.Parameters
}
