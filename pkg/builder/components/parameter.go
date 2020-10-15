package components

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/evilmonkeyinc/openapi-go-gen/pkg/builder/lib"
	"github.com/getkin/kin-openapi/openapi3"
)

func NewParameterBuilder(module, packageName, parentID string, parameterRef *openapi3.ParameterRef, addTags bool) ComponentBuilder {
	return &parameterBuilder{
		AddTags:      addTags,
		Module:       module,
		PackageName:  packageName,
		ParameterRef: parameterRef,
		ParentID:     parentID,
	}
}

type parameterBuilder struct {
	AddTags      bool
	Module       string
	PackageName  string
	ParameterRef *openapi3.ParameterRef
	ParentID     string
}

func (builder *parameterBuilder) AsStruct(name string) (jen.Code, []jen.Code, error) {
	if ref := builder.ParameterRef.Ref; ref != "" {
		split := strings.Split(ref, "/")
		typeName := split[len(split)-1]
		packageName := split[len(split)-2]
		path := fmt.Sprintf("%s/%s", builder.Module, packageName)
		if packageName == builder.PackageName {
			path = ""
		}

		return jen.Type().Id(strings.Title(name)).Struct(jen.Qual(path, typeName).Tag(map[string]string{"bson": ",inline"})),
			nil,
			nil
	} else if parameter := builder.ParameterRef.Value; parameter != nil {

		param := jen.Type().Id(strings.Title(name))

		if schema := parameter.Schema; schema != nil {
			if schema.Ref != "" {
				split := strings.Split(schema.Ref, "/")
				paramType := strings.Title(split[len(split)-1])
				packageName := split[len(split)-2]
				param = param.Qual(fmt.Sprintf("%s/%s", builder.Module, packageName), paramType)
			} else if schema.Value != nil {
				switch schema.Value.Type {
				case "object":
					fallthrough
				case "array":
					return nil, nil, fmt.Errorf("do not support array or object in params yet")
				default:
					param = lib.AddPrimitiveTypeFromSchema(param, schema.Value)
				}
			} else {
				return nil, nil, fmt.Errorf("schema supplied with no ref or value")
			}
		} else {
			param = param.Interface()
		}

		return param,
			nil,
			nil

	}
	return nil, nil, fmt.Errorf("openapi3.ParameterRef must have non-empty Ref or non-nil Value")
}

func (builder *parameterBuilder) AsField(name string) (jen.Code, []jen.Code, error) {
	if ref := builder.ParameterRef.Ref; ref != "" {
		split := strings.Split(ref, "/")
		paramType := strings.Title(split[len(split)-1])
		packageName := split[len(split)-2]

		if name == "" {
			return jen.Qual(fmt.Sprintf("%s/%s", builder.Module, packageName), paramType).Tag(map[string]string{
					"bson": ",inline",
				}),
				nil,
				nil
		}

		return jen.Id(name).Qual(fmt.Sprintf("%s/%s", builder.Module, packageName), paramType),
			nil,
			nil
	} else if parameter := builder.ParameterRef.Value; parameter != nil {
		if name == "" {
			name = parameter.Name
		}

		param := jen.Comment(parameter.Description).Line().Id(strings.Title(name))

		if schema := parameter.Schema; schema != nil {
			if schema.Ref != "" {
				split := strings.Split(schema.Ref, "/")
				paramType := strings.Title(split[len(split)-1])
				packageName := split[len(split)-2]
				param = param.Qual(fmt.Sprintf("%s/%s", builder.Module, packageName), paramType)
			} else if schema.Value != nil {
				switch schema.Value.Type {
				case "object":
					fallthrough
				case "array":
					return nil, nil, fmt.Errorf("do not support array or object in params yet")
				default:
					param = lib.AddPrimitiveTypeFromSchema(param, schema.Value)
				}
			} else {
				return nil, nil, fmt.Errorf("schema supplied with no ref or value")
			}
		} else {
			param = param.Interface()
		}

		if builder.AddTags {
			required := ",omitempty"
			if parameter.Required {
				required = ""
			}
			param.Tag(map[string]string{
				"json": fmt.Sprintf("%s%s", name, required),
				"yaml": fmt.Sprintf("%s%s", name, required),
			})
		}

		return param, nil, nil
	}
	return nil, nil, fmt.Errorf("openapi3.ParameterRef must have non-empty Ref or non-nil Value")
}
