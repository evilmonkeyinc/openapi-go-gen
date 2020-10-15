package lib

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
)

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
			return code.Byte()
		case "date":
			fallthrough
		case "date-time":
			return code.Qual("time", "Duration")
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
