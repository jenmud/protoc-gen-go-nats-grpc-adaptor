package helpers

import (
	"fmt"
	"log/slog"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

type Import struct {
	Path string
	Name string
}

// FormatImports generates the import statement.
func FormatImports(imports []Import) string {
	if len(imports) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("import (\n")
	for _, imp := range imports {
		if imp.Name != "" {
			result.WriteString(fmt.Sprintf("\t%s %q\n", imp.Name, imp.Path))
		} else {
			result.WriteString(fmt.Sprintf("\t%q\n", imp.Path))
		}
	}
	result.WriteString(")")
	return result.String()
}

// CollectAllImports analyzes the file and collects all required imports.
func CollectAllImports(file *protogen.File) []Import {
	imports := make(map[string]Import)

	// Add base required imports
	baseImports := map[string]Import{
		"context":                            {Path: "context"},
		"log/slog":                           {Path: "log/slog"},
		"strings":                            {Path: "strings"},
		"errors":                             {Path: "errors"},
		"google.golang.org/protobuf/proto":   {Path: "google.golang.org/protobuf/proto", Name: "googleProto"},
		"github.com/nats-io/nats.go":         {Path: "github.com/nats-io/nats.go", Name: "nats"},
		"github.com/nats-io/nats.go/micro":   {Path: "github.com/nats-io/nats.go/micro", Name: "micro"},
		"go.opentelemetry.io/otel":           {Path: "go.opentelemetry.io/otel"},
		"go.opentelemetry.io/otel/attribute": {Path: "go.opentelemetry.io/otel/attribute"},
		"go.opentelemetry.io/otel/trace":     {Path: "go.opentelemetry.io/otel/trace"},
	}

	for k, v := range baseImports {
		imports[k] = v
	}

	// Process messages recursively
	var processMessage func(message *protogen.Message)
	processMessage = func(message *protogen.Message) {
		for _, field := range message.Fields {
			// Handle message types
			if field.Message != nil {
				importPath := string(field.Message.GoIdent.GoImportPath)
				if importPath != "" && importPath != string(file.GoImportPath) {
					imports[importPath] = Import{Path: importPath}
				}
			}

			// Handle enum types
			if field.Enum != nil {
				importPath := string(field.Enum.GoIdent.GoImportPath)
				if importPath != "" && importPath != string(file.GoImportPath) {
					imports[importPath] = Import{Path: importPath}
				}
			}
		}

		// Process nested messages
		for _, nested := range message.Messages {
			processMessage(nested)
		}
	}

	// Process all top-level messages
	for _, message := range file.Messages {
		processMessage(message)
	}

	// Process services and their methods
	for _, service := range file.Services {
		for _, method := range service.Methods {
			// Process input types
			if method.Input != nil {
				importPath := string(method.Input.GoIdent.GoImportPath)
				if importPath != "" && importPath != string(file.GoImportPath) {
					imports[importPath] = Import{Path: importPath}
				}
			}
			// Process output types
			if method.Output != nil {
				importPath := string(method.Output.GoIdent.GoImportPath)
				if importPath != "" && importPath != string(file.GoImportPath) {
					imports[importPath] = Import{Path: importPath}
				}
			}
		}
	}

	// Convert map to slice
	var result []Import
	for _, imp := range imports {
		result = append(result, imp)
	}

	return result
}

// TrimPackagePath gets the last part of the import path to use as package name.
func TrimPackagePath(importPath protogen.GoImportPath) string {
	parts := strings.Split(string(importPath), "/")
	return parts[len(parts)-1]
}

// GenerateFile is the main entrypoint used for generating the .pb.go file.
func GenerateFile(gen *protogen.Plugin, file *protogen.File, tmplStr string) error {
	if len(file.Services) == 0 {
		return nil
	}

	filename := file.GeneratedFilenamePrefix + "-nats-grpc-adaptor.pb.go"
	logger := slog.With("filename", filename)
	logger.Info("generating the files",
		slog.String("current_package_path", string(file.GoImportPath)))

	// Add debug logging for each message
	for _, service := range file.Services {
		for _, method := range service.Methods {
			logger.Debug("method type info",
				slog.String("method", method.GoName),
				slog.String("input_import_path", string(method.Input.GoIdent.GoImportPath)),
				slog.String("input_go_name", method.Input.GoIdent.GoName),
				slog.String("file_go_package", string(file.GoPackageName)))
		}
	}

	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	funcMap := template.FuncMap{
		"formatImports":   FormatImports,
		"trimPackagePath": TrimPackagePath,
		"samePackage": func(msgImportPath protogen.GoImportPath, fileImportPath protogen.GoImportPath) bool {
			same := msgImportPath == fileImportPath
			slog.Debug("comparing packages",
				slog.String("msg_path", string(msgImportPath)),
				slog.String("file_path", string(fileImportPath)),
				slog.Bool("same", same))
			return same
		},
	}

	tmpl, err := template.New("nats-micro-service").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		logger.Error("failed to parse the template", slog.String("reason", err.Error()))
		return err
	}

	data := struct {
		*protogen.File
		Imports []Import
	}{
		File:    file,
		Imports: CollectAllImports(file),
	}

	if err := tmpl.Execute(g, data); err != nil {
		logger.Error("failed to execute the template", slog.String("reason", err.Error()))
		return err
	}

	return nil
}
