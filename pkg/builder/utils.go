package builder

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
)

func addPrimitiveTypeFromSchema(code *jen.Statement, schema *openapi3.Schema) *jen.Statement {
	switch schema.Type {
	case "string":
		return code.String()
	case "integer":
		return code.Int64()
	case "object":
		panic("object is not primitive type")
	case "array":
		panic("object is not primitive type")
	default:
		panic(fmt.Sprintf("unsupported type %s", schema.Type))
	}
}
