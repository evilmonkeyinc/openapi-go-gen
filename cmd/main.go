package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/evilmonkeyinc/openapi-go-gen/pkg/builder"
	"github.com/evilmonkeyinc/openapi-go-gen/pkg/parser"
	"github.com/evilmonkeyinc/openapi-go-gen/pkg/utils"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	spec := flag.String("spec", "", "The openapi v3 specification file")
	modulePath := flag.String("go-module", "github.com/example/openapi", "The go module for the generated files")
	output := flag.String("output", "", "The location for generated files (default \"out/{gopackage}\")")
	flag.Parse()

	if spec == nil || *spec == "" {
		panic(fmt.Errorf("spec has not been specified"))
	}

	if *output == "" {
		tmp := fmt.Sprintf("out/%s", *modulePath)
		output = &tmp
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(*spec)
	if err != nil {
		panic(err)
	}
	if err := swagger.Validate(context.Background()); err != nil {
		panic(err)
	}

	swaggerParser, err := parser.New(swagger)
	if err != nil {
		panic(err)
	}

	os.MkdirAll(*output, os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/schemas", *output), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/responses", *output), os.ModePerm)
	for key, schema := range swaggerParser.Schemas {
		fileString := builder.BuildSchemaFile(*modulePath, key, schema)
		err := utils.WriteFile(fileString, fmt.Sprintf("%s/schemas/%s.go", *output, key))
		if err != nil {
			panic(err)
		}
	}

	for key, responses := range swaggerParser.Responses {
		fileString := builder.BuildResponseFile(*modulePath, key, responses)
		err := utils.WriteFile(fileString, fmt.Sprintf("%s/responses/%s.go", *output, key))
		if err != nil {
			panic(err)
		}
	}

	for key, api := range swaggerParser.APIs {
		fileString := builder.BuildAPIFile(*modulePath, api)
		err := utils.WriteFile(fileString, fmt.Sprintf("%s/%s.go", *output, key))
		if err != nil {
			panic(err)
		}
	}
}
