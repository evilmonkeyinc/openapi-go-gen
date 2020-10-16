package builder

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/evilmonkeyinc/openapi-go-gen/pkg/builder/components"
	"github.com/getkin/kin-openapi/openapi3"
)

// NewOperationBuilder creates a new Builder for the openapi3.Operation
func NewOperationBuilder(module, packageName string, operation *openapi3.Operation, pathParameters openapi3.Parameters) *OperationBuilder {
	return &OperationBuilder{
		Module:         module,
		PackageName:    packageName,
		Operation:      operation,
		PathParameters: pathParameters,
	}
}

type OperationBuilder struct {
	Module         string
	PackageName    string
	Operation      *openapi3.Operation
	PathParameters openapi3.Parameters
}

func (builder *OperationBuilder) RequestStruct() (jen.Code, []jen.Code, error) {
	operationID := strings.Title(builder.Operation.OperationID)
	requestID := fmt.Sprintf("%sRequest", operationID)

	extras := make([]jen.Code, 0)
	requestParams := make([]jen.Code, 0)
	for _, parameter := range builder.PathParameters {
		parameterBuilder := components.NewParameterBuilder(builder.Module, builder.PackageName, "", parameter, true)
		requestParam, requestExtras, err := parameterBuilder.AsField("")
		if err != nil {
			return nil, nil, err
		}

		requestParams = append(requestParams, requestParam)
		extras = append(extras, requestExtras...)
	}
	for _, parameter := range builder.Operation.Parameters {
		parameterBuilder := components.NewParameterBuilder(builder.Module, builder.PackageName, "", parameter, true)
		requestParam, requestExtras, err := parameterBuilder.AsField("")
		if err != nil {
			return nil, nil, err
		}

		requestParams = append(requestParams, requestParam)
		extras = append(extras, requestExtras...)
	}

	return jen.Commentf("%s encapsulates the expected request for %s()", requestID, operationID).Line().Type().Id(requestID).Struct(requestParams...).Line(),
		extras,
		nil
}

func (builder *OperationBuilder) ResponseStruct() (jen.Code, []jen.Code, error) {
	operationID := strings.Title(builder.Operation.OperationID)
	responseID := fmt.Sprintf("%sResponse", operationID)

	extras := make([]jen.Code, 0)

	responseParams := make([]jen.Code, 0)
	for statusCode, response := range builder.Operation.Responses {
		responseBuilder := components.NewResponseBuilder(builder.Module, builder.PackageName, responseID, response)
		main, extra, err := responseBuilder.AsField(statusCode)
		if err != nil {
			return nil, nil, err
		}
		responseParams = append(responseParams, main)
		extras = append(extras, extra...)
	}

	return jen.Commentf("%s encapsulates the expected response for %s()", responseID, operationID).Line().Type().Id(responseID).Struct(
			responseParams...,
		).Add(
			extras...,
		).Line(),
		nil,
		nil
}

func (builder *OperationBuilder) InterfaceField() (jen.Code, error) {
	operationID := strings.Title(builder.Operation.OperationID)
	requestID := fmt.Sprintf("%sRequest", operationID)
	responseID := fmt.Sprintf("%sResponse", operationID)

	return jen.Commentf("%s %s", operationID, builder.Operation.Description).Line().Id(operationID).Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("request").Op("*").Qual("", requestID),
		).Params(jen.Op("*").Qual("", responseID), jen.Error()),
		nil
}

// func(ResponseWriter, *Request)
func (builder *OperationBuilder) HTTPHandler() (jen.Code, error) {
	operationID := strings.Title(builder.Operation.OperationID)
	return jen.Func().Id(fmt.Sprintf("%sWrapper", operationID)).Params(
		jen.Qual("net/http", "ResponseWriter"),
		jen.Op("*").Qual("net/http", "Request"),
	), nil
}
