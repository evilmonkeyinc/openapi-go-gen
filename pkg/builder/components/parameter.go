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

func parser(structName string, parameter *openapi3.Parameter) (jen.Code, error) {

	code := jen.Func().Parens(jen.Id(structName).Id(strings.Title(structName))).Id("Parse").Call(jen.Id("req").Op("*").Qual("net/http", "Request")).Error()

	converstionCode, err := lib.GetConvertStringToPrimitiveTypeCode(structName, parameter.Name, parameter.Required, parameter.Schema.Value)
	if err != nil {
		return nil, err
	}

	switch parameter.In {
	case "query":
		params := jen.Id("str").Op(":=").Id("req").Dot("URL").Dot("Query").Call().Dot("Get").Call(jen.Lit(parameter.Name)).Line()
		params = params.If(jen.Id("str").Op("!=").Id("\"\"")).Block(
			converstionCode,
		)
		code = code.Block(
			params,
			jen.Return(jen.Nil()),
		)
	default:
		return nil, fmt.Errorf("support for parameter in type \"%s\" is not implemented", parameter.In)
	}

	/**
	func (queryLimit *QueryLimit) Parse(request *http.Request) error {

		str := request.URL.Query().Get("limit");
		if str != "" {
			val, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse query parameter limit %w", err)
			}
			queryLimit.Limit = &val
		}

		return nil
	}
	**/
	return code, nil
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

		param := jen.Qual(path, typeName)
		if builder.AddTags {
			param = param.Tag(map[string]string{"bson": ",inline"})
		}

		return jen.Type().Id(strings.Title(name)).Struct(
				param,
			),
			nil,
			nil
	} else if parameter := builder.ParameterRef.Value; parameter != nil {
		parameterBuilder := NewParameterBuilder(builder.Module, builder.PackageName, builder.ParentID, builder.ParameterRef, true)
		param, _, _ := parameterBuilder.AsField(parameter.Name)

		extra, _ := parser(name, parameter)

		return jen.Type().Id(strings.Title(name)).Struct(
				param,
			),
			[]jen.Code{extra},
			nil

	}
	return nil, nil, fmt.Errorf("openapi3.ParameterRef must have non-empty Ref or non-nil Value")
}

func (builder *parameterBuilder) AsField(name string) (jen.Code, []jen.Code, error) {
	if ref := builder.ParameterRef.Ref; ref != "" {
		split := strings.Split(ref, "/")
		paramType := strings.Title(split[len(split)-1])
		packageName := split[len(split)-2]

		param := jen.Id(strings.Title(name)).Qual(fmt.Sprintf("%s/%s", builder.Module, packageName), paramType)
		if builder.AddTags {
			if name == "" {
				param = param.Tag(map[string]string{
					"bson": ",inline",
				})
			} else {
				param = param.Tag(map[string]string{
					"json": fmt.Sprintf("%s,omitempty", name),
					"yaml": fmt.Sprintf("%s,omitempty", name),
				})
			}
		}
		return param,
			nil,
			nil
	} else if parameter := builder.ParameterRef.Value; parameter != nil {
		if name == "" {
			name = parameter.Name
		}

		param := jen.Comment(parameter.Description).Line().Id(strings.Title(name))

		if !parameter.Required {
			param = param.Op("*")
		}

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
			return nil, nil, fmt.Errorf("no schema specified for parameter")
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
