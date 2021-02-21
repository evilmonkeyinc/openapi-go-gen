package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/evilmonkeyinc/openapi-go-gen/pkg/builder/components"
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

	for key, schema := range swagger.Components.Schemas {
		err := components.GenerateSchema(*output, key, schema)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}
	}

}
