package lib

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
)

func GetConvertStringToPrimitiveTypeCode(structName, parameterName string, isRequired bool, schema *openapi3.Schema) (jen.Code, error) {
	valOp := ""
	if !isRequired {
		valOp = "&"
	}

	switch schema.Type {
	case "boolean":
		return jen.Id("val").Op(",").Id("err").Op(":=").Qual("strconv", "ParseBool").Call(
				jen.Id("str"),
			).Line().If(
				jen.Id("err").Op("!=").Id("nil").Block(
					jen.Return(jen.Id("err")),
				),
			).Line().Id(structName).Dot(strings.Title(parameterName)).Op("=").Op(valOp).Id("val"),
			nil
	case "integer":
		switch schema.Format {
		case "int32":
			return jen.Id("val").Op(",").Id("err").Op(":=").Qual("strconv", "ParseInt").Call(
				jen.Id("str"),
				jen.Id("10"),
				jen.Id("32"),
			).Line().If(
				jen.Id("err").Op("!=").Id("nil").Block(
					jen.Return(jen.Id("err")),
				),
			).Line().Id("tmp").Op(":=").Qual("", "int32").Parens(
				jen.Id("val"),
			).Line().Id(structName).Dot(strings.Title(parameterName)).Op("=").Op(valOp).Id("tmp"), nil
		case "int64":
			fallthrough
		default:
			return jen.Id("val").Op(",").Id("err").Op(":=").Qual("strconv", "ParseInt").Call(
				jen.Id("str"),
				jen.Id("10"),
				jen.Id("64"),
			).Line().If(
				jen.Id("err").Op("!=").Id("nil").Block(
					jen.Return(jen.Id("err")),
				),
			).Line().Id(structName).Dot(strings.Title(parameterName)).Op("=").Op(valOp).Id("val"), nil
		}
	case "number":
		switch schema.Format {
		case "float":
			return jen.Id("val").Op(",").Id("err").Op(":=").Qual("strconv", "ParseFloat").Call(
				jen.Id("str"),
				jen.Id("32"),
			).Line().If(
				jen.Id("err").Op("!=").Id("nil").Block(
					jen.Return(jen.Id("err")),
				),
			).Line().Id("tmp").Op(":=").Qual("", "float32").Parens(
				jen.Id("val"),
			).Line().Id(structName).Dot(strings.Title(parameterName)).Op("=").Op(valOp).Id("tmp"), nil
		case "double":
			fallthrough
		default:
			return jen.Id("val").Op(",").Id("err").Op(":=").Qual("strconv", "ParseFloat").Call(
				jen.Id("str"),
				jen.Id("64"),
			).Line().If(
				jen.Id("err").Op("!=").Id("nil").Block(
					jen.Return(jen.Id("err")),
				),
			).Line().Id(structName).Dot(strings.Title(parameterName)).Op("=").Op(valOp).Id("val"), nil
		}
	case "string":
		switch schema.Format {
		case "byte":
			return jen.Id("tmp").Op(":=").Op("[]").Qual("", "byte").Parens(
					jen.Id("str"),
				).Line().Id(structName).Dot(strings.Title(parameterName)).Op("=").Op(valOp).Id("tmp"),
				nil
		case "date":
			fallthrough
		case "date-time":
			return jen.Id("val").Op(",").Id("err").Op(":=").Qual("time", "Parse").Call(
				jen.Qual("time", "RFC3339"),
				jen.Id("str"),
			).Line().If(
				jen.Id("err").Op("!=").Id("nil").Block(
					jen.Return(jen.Id("err")),
				),
			).Line().Id(structName).Dot(strings.Title(parameterName)).Op("=").Op(valOp).Id("val"), nil
		case "binary":
			fallthrough
		case "password":
			fallthrough
		default:
			return jen.Id(structName).Dot(strings.Title(parameterName)).Op("=").Op(valOp).Id("str"),
				nil
		}
	case "object":
		return nil, fmt.Errorf("object is not primitive type")
	case "array":
		return nil, fmt.Errorf("object is not primitive type")
	default:
		return nil, fmt.Errorf("unsupported type %s", schema.Type)
	}
}

func AddPrimitiveTypeFromSchema(code *jen.Statement, schema *openapi3.Schema) *jen.Statement {
	switch schema.Type {
	case "boolean":
		return code.Bool()
	case "integer":
		switch schema.Format {
		case "int32":
			return code.Int32()
		case "int64":
			fallthrough
		default:
			return code.Int64()
		}
	case "number":
		switch schema.Format {
		case "float":
			return code.Float32()
		case "double":
			fallthrough
		default:
			return code.Float64()
		}
	case "string":
		switch schema.Format {
		case "byte":
			return code.Op("[]").Byte()
		case "date":
			fallthrough
		case "date-time":
			return code.Qual("time", "Time")
		case "binary":
			fallthrough
		case "password":
			fallthrough
		default:
			return code.String()
		}
	case "object":
		panic("object is not primitive type")
	case "array":
		panic("object is not primitive type")
	default:
		panic(fmt.Sprintf("unsupported type %s", schema.Type))
	}
}
