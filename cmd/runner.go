package cmd

import (
	"log/slog"

	"google.golang.org/protobuf/compiler/protogen"
)

// Run starts running the plugin
func Run() error {
	protogen.Options{}.Run(
		func(gen *protogen.Plugin) error {
			for _, file := range gen.Files {
				if !file.Generate {
					continue
				}
				if err := generateFile(gen, file); err != nil {
					return err
				}
			}
			return nil
		},
	)
	return nil
}

// generateFile generates a .pb.go file.
func generateFile(gen *protogen.Plugin, file *protogen.File) error {
	filename := file.GeneratedFilenamePrefix + ".nats.pb.go"
	slog.Info("generating file", slog.String("filename", filename))
	return nil
}
