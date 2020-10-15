package components

import "github.com/dave/jennifer/jen"

type ComponentBuilder interface {
	AsStruct(structName string) (jen.Code, []jen.Code, error)
	AsField(fieldName string) (jen.Code, []jen.Code, error)
}
