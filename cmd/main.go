package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/evilmonkeyinc/openapi-go-gen/pkg/builder"
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

	os.MkdirAll(*output, os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/schemas", *output), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/responses", *output), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/parameters", *output), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/operations", *output), os.ModePerm)

	for key, schema := range swagger.Components.Schemas {
		err := builder.WriteSchemaFile(*output, *modulePath, key, schema)
		if err != nil {
			panic(err)
		}
	}

	for key, responses := range swagger.Components.Responses {
		err := builder.WriteResponseFile(*output, *modulePath, key, responses)
		if err != nil {
			panic(err)
		}
	}

	for key, parameter := range swagger.Components.Parameters {
		err := builder.WriteParameterFile(*output, *modulePath, key, parameter)
		if err != nil {
			panic(err)
		}
	}

	if err := builder.WriteOperationFiles(*output, *modulePath, swagger); err != nil {
		panic(err)
	}
}
