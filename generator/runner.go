package generator

import (
	_ "embed"

	"github.com/jenmud/protoc-gen-go-nats-grpc-adaptor/internal/helpers"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed proto-gen.tmpl
var templ string

// Run is the main entrypoint.
func Run() error {
	protogen.Options{}.Run(
		func(gen *protogen.Plugin) error {
			for _, file := range gen.Files {
				if !file.Generate {
					continue
				}

				if err := helpers.GenerateFile(gen, file, templ); err != nil {
					return err
				}
			}
			return nil
		},
	)
	return nil
}
