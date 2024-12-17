package cmd

import (
	"fmt"
	"log/slog"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

// Run is the main entrypoint.
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

type Import struct {
	Path string
	Name string
}

// formatImports generates the import statement.
func formatImports(imports []Import) string {
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

// collectAllImports analyzes the file and collects all required imports.
func collectAllImports(file *protogen.File) []Import {
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

// trimPackagePath gets the last part of the import path to use as package name.
func trimPackagePath(importPath protogen.GoImportPath) string {
	parts := strings.Split(string(importPath), "/")
	return parts[len(parts)-1]
}

// generateFile is the main entrypoint used for generating the .pb.go file.
func generateFile(gen *protogen.Plugin, file *protogen.File) error {
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
		"formatImports":   formatImports,
		"trimPackagePath": trimPackagePath,
		"samePackage": func(msgImportPath protogen.GoImportPath, fileImportPath protogen.GoImportPath) bool {
			same := msgImportPath == fileImportPath
			slog.Debug("comparing packages",
				slog.String("msg_path", string(msgImportPath)),
				slog.String("file_path", string(fileImportPath)),
				slog.Bool("same", same))
			return same
		},
	}

	tmpl, err := template.New("nats-micro-service").Funcs(funcMap).Parse(templ)
	if err != nil {
		return err
	}

	data := struct {
		*protogen.File
		Imports []Import
	}{
		File:    file,
		Imports: collectAllImports(file),
	}

	if err := tmpl.Execute(g, data); err != nil {
		logger.Error("failed to execute the template", slog.String("reason", err.Error()))
		return err
	}

	return nil
}

const templ = `
// Code generated by protoc-gen-go-nats-grpc-adaptor. DO NOT EDIT.
// source: {{.GeneratedFilenamePrefix}}.proto

package {{.GoPackageName}}

{{ formatImports .Imports }}

var tracer = otel.Tracer("{{ .Proto.Name }}")

// handleError is a helper which response with the error.
func handleError(req micro.Request, err error) {
    if sendErr := req.Error("500", err.Error(), nil); sendErr != nil {
        slog.Error(
            "error sending response error",
            slog.String("reason", sendErr.Error()),
            slog.String("subject", req.Subject()),
        )
    }
}

// JobHandler is job handler for the concurrent service.
type JobHandler struct {
	ctx context.Context
	execute func(context.Context, micro.Request)
	msg micro.Request
}

// ConcurrentService is a wrapper around the micro.Service interface, extending with additional functionality.
type ConcurrentService struct {
	micro micro.Service
	jobs chan(JobHandler)
}

// AddEndpoint registers endpoint with given name on a specific subject.
func (m *ConcurrentService) AddEndpoint(name string, handler micro.Handler, opts ...micro.EndpointOpt) error {
	return m.micro.AddEndpoint(name, handler, opts...)
}

// AddGroup returns a Group interface, allowing for more complex endpoint topologies.
// A group can be used to register endpoints with given prefix.
 func (m *ConcurrentService) AddGroup(group string, opts ...micro.GroupOpt) micro.Group {
  return m.micro.AddGroup(group)
 }

// Info returns the service info.
 func (m *ConcurrentService) Info() micro.Info {
  return m.micro.Info()
 }

// Stats returns statistics for the service endpoint and all monitoring endpoints.
 func (m *ConcurrentService) Stats() micro.Stats {
  return m.micro.Stats()
 }

// Reset resets all statistics (for all endpoints) on a service instance.
 func (m *ConcurrentService) Reset() {
  m.micro.Reset()
 }

// Stop drains the endpoint subscriptions and marks the service as stopped.
 func (m *ConcurrentService) Stop() error {
  close(m.jobs)
  return m.micro.Stop()
 }

// Stopped informs whether [Stop] was executed on the service.
 func (m *ConcurrentService) Stopped() bool {
  return m.micro.Stopped()
 }

// ConcurrentServiceOption is a function used to configure a ConcurrentService.
type ConcurrentServiceOption func(*ConcurrentService)

// WithConcurrentJobs sets the number of concurrent jobs to be executed.
func WithConcurrentJobs(jobs int) ConcurrentServiceOption {
	return func(s *ConcurrentService) {
		s.jobs = make(chan JobHandler, jobs)
		for i := 0; i < jobs; i++ {
			go func() {
				for job := range s.jobs {
					job.execute(job.ctx, job.msg)
				}
			}()
		}
	}
}

{{ range .Services }}
// NewNATS{{ .GoName }}Server returns the gRPC server as a NATS micro service.
//
// Example:
//   nc, err := nats.Connect(ns.ClientURL())
//   if err != nil {
//     panic(err)
//   }
//
//   cfg := micro.Config{
//       Name: "{{ .GoName }}Server-Demo",
//       Version: "1.0.0",
//       QueueGroup: "example",
//       Description: "NATS micro service adaptor wrapping {{ .GoName }}Server",
//   }
//
//   mc, err := NewNATS{{ .GoName }}Server(context.Background(), nc, {{ .GoName }}Service{}, cfg)
//   if err != nil {
//     panic(err)
//   }
//
//   fmt.Printf("%s -> %s\n", mc.Info().Name, mc.Info().ID)
//
func NewNATS{{ .GoName }}Server(ctx context.Context, nc *nats.Conn, server {{ .GoName }}Server, cfg micro.Config, opts ...ConcurrentServiceOption) (micro.Service, error) {
    srv, err := micro.AddService(nc, cfg)
    if err != nil {
        return nil, err
    }

    concurrentSrv := &ConcurrentService{
        jobs: make(chan JobHandler, 1),
    }

    for _, opt := range opts {
        opt(concurrentSrv)
    }

    concurrentSrv.micro = srv

    logger := slog.With(
        slog.Group(
            "service",
            slog.String("name", cfg.Name),
            slog.String("version", cfg.Version),
            slog.String("queue-group", cfg.QueueGroup),
            slog.Int("workers", cap(concurrentSrv.jobs)),
        ),
    )

    {{ range .Methods }}
    logger.Info(
        "registring endpoint",
        slog.Group(
            "endpoint",
            slog.String("subject", cfg.Name + "." + strings.ToLower("svc.{{ .Parent.GoName }}.{{ .GoName }}")),
        ),
    )

    err = srv.AddEndpoint(
        "{{ .Parent.GoName }}",
        micro.ContextHandler(
            ctx,
            func(ctx context.Context, req micro.Request) {
            	handler := func(ctx context.Context, req micro.Request) {
	                endpointSubject := cfg.Name + "." + strings.ToLower("svc.{{ .Parent.GoName }}.{{ .GoName }}")

	                ctx, span := tracer.Start(ctx, "{{ .GoName }}", trace.WithAttributes(attribute.String("subject", endpointSubject)))
	                defer span.End()

	                hlogger := logger.With(
	                    slog.Group(
	                        "endpoint",
	                        slog.String("subject", endpointSubject),
	                    ),
	                )

	                r := new({{ if not (samePackage .Input.GoIdent.GoImportPath $.GoImportPath) }}{{ trimPackagePath .Input.GoIdent.GoImportPath }}.{{ end }}{{ .Input.GoIdent.GoName }})

	                if err := googleProto.Unmarshal(req.Data(), r); err != nil {
	                    hlogger.Error("unmarshaling request", slog.String("reason", err.Error()))
	                    handleError(req, err)
	                    return
	                }

	                resp, err := server.{{ .GoName }}(ctx, r)
	                if err != nil {
	                    hlogger.Error("service error", slog.String("reason", err.Error()))
	                    handleError(req, err)
	                    return
	                }

	                respDump, err := googleProto.Marshal(resp)
	                if err != nil {
	                    hlogger.Error("marshaling response", slog.String("reason", err.Error()))
	                    handleError(req, err)
	                    return
	                }

	                if err := req.Respond(respDump); err != nil {
	                    hlogger.Error("sending response", slog.String("reason", err.Error()))
	                    handleError(req, err)
	                    return
	                }
				}

				concurrentSrv.jobs <- JobHandler{ctx, handler, req}
            },
        ),
        micro.WithEndpointSubject(cfg.Name + "." + strings.ToLower("svc.{{ .Parent.GoName }}.{{ .GoName }}")),
        micro.WithEndpointMetadata(map[string]string{"Description": "TODO: still to be implemented - see .proto file for doco"}),
    )
    {{ end }}

    return srv, nil
}

// NewNATSGRPCClientTo{{ .GoName }}Server returns the gRPC server wrapping a gRPC client as a NATS micro service.
//
// Example:
//   nc, err := nats.Connect(ns.ClientURL())
//   if err != nil {
//     panic(err)
//   }
//
//   var opts := []grpc.DailOption
//
//   conn, err := grpc.NewClient("localhost:1234", opts...)
//   if err != nil {
//       panic(err)
//   }
//
//   defer conn.Close()
//
//   client := New{{ .GoName }}Client(conn)
//
//   cfg := micro.Config{
//       Name: "{{ .GoName }}Wrapper-Demo",
//       Version: "1.0.0",
//       QueueGroup: "example",
//       Description: "NATS micro service adaptor wrapping {{ .GoName }}Client",
//   }
//
//   mc, err := NewNATSGRPCClientTo{{ .GoName }}Server(context.Background(), nc, client, cfg)
//   if err != nil {
//     panic(err)
//   }
//
//   fmt.Printf("%s -> %s\n", mc.Info().Name, mc.Info().ID)
//
func NewNATSGRPCClientTo{{ .GoName }}Server(ctx context.Context, nc *nats.Conn, client {{ .GoName }}Client, cfg micro.Config) (micro.Service, error) {
    srv, err := micro.AddService(nc, cfg)
    if err != nil {
        return nil, err
    }

    logger := slog.With(
        slog.Group(
            "service",
            slog.String("name", cfg.Name),
            slog.String("version", cfg.Version),
            slog.String("queue-group", cfg.QueueGroup),
        ),
    )

    {{ range .Methods }}
    logger.Info(
        "registring endpoint",
        slog.Group(
            "endpoint",
            slog.String("subject", cfg.Name + "." + strings.ToLower("svc.{{ .Parent.GoName }}.{{ .GoName }}")),
        ),
    )

    err = srv.AddEndpoint(
        "{{ .Parent.GoName }}",
        micro.ContextHandler(
            ctx,
            func(ctx context.Context, req micro.Request) {
                endpointSubject := cfg.Name + "." + strings.ToLower("svc.{{ .Parent.GoName }}.{{ .GoName }}")

                ctx, span := tracer.Start(ctx, "{{ .GoName }}", trace.WithAttributes(attribute.String("subject", endpointSubject)))
                defer span.End()

                hlogger := logger.With(
                    slog.Group(
                        "endpoint",
                        slog.String("subject", endpointSubject),
                    ),
                )

                r := new({{ if not (samePackage .Input.GoIdent.GoImportPath $.GoImportPath) }}{{ trimPackagePath .Input.GoIdent.GoImportPath }}.{{ end }}{{ .Input.GoIdent.GoName }})

                if err := googleProto.Unmarshal(req.Data(), r); err != nil {
                    hlogger.Error("unmarshaling request", slog.String("reason", err.Error()))
                    handleError(req, err)
                    return
                }

                resp, err := client.{{ .GoName }}(ctx, r)
                if err != nil {
                    hlogger.Error("service error", slog.String("reason", err.Error()))
                    handleError(req, err)
                    return
                }

                respDump, err := googleProto.Marshal(resp)
                if err != nil {
                    hlogger.Error("marshaling response", slog.String("reason", err.Error()))
                    handleError(req, err)
                    return
                }

                if err := req.Respond(respDump); err != nil {
                    hlogger.Error("sending response", slog.String("reason", err.Error()))
                    handleError(req, err)
                    return
                }
            },
        ),
        micro.WithEndpointSubject(cfg.Name + "." + strings.ToLower("svc.{{ .Parent.GoName }}.{{ .GoName }}")),
        micro.WithEndpointMetadata(map[string]string{"Description": "TODO: still to be implemented - see .proto file for doco"}),
    )
    {{ end }}

    return srv, nil
}

// NATS{{ .GoName }}Client is a client connecting to a NATS {{ .GoName }}Server.
type NATS{{ .GoName }}Client struct {
    nc *nats.Conn
    name string
}

// NewNATS{{ .GoName }}Client returns a new {{ .GoName }}Server client.
// Example:
//   nc, err := nats.Connect(ns.ClientURL())
//   if err != nil {
//     panic(err)
//   }
//
//   client := NewNATS{{ .GoName }}Client(nc, "example-service-name")
//
func NewNATS{{ .GoName }}Client(nc *nats.Conn, name string) *NATS{{ .GoName }}Client {
    return &NATS{{ .GoName }}Client{
        nc: nc,
        name: name,
    }
}

{{ range .Methods }}
{{ .Comments.Leading }}func (c *NATS{{ .Parent.GoName }}Client) {{ .GoName }}(ctx context.Context, req *{{ if not (samePackage .Input.GoIdent.GoImportPath $.GoImportPath) }}{{ trimPackagePath .Input.GoIdent.GoImportPath }}.{{ end }}{{ .Input.GoIdent.GoName }}) (*{{ if not (samePackage .Output.GoIdent.GoImportPath $.GoImportPath) }}{{ trimPackagePath .Output.GoIdent.GoImportPath }}.{{ end }}{{ .Output.GoIdent.GoName }}, error) {
    subject := c.name + "." + strings.ToLower("svc.{{ .Parent.GoName }}.{{ .GoName }}")

    ctx, span := tracer.Start(ctx, "{{ .GoName }}", trace.WithAttributes(attribute.String("subject", subject)))
    defer span.End()

    payload, err := googleProto.Marshal(req)
    if err != nil {
        return nil, err
    }

    respPayload, err := c.nc.RequestWithContext(ctx, subject, payload)
    if err != nil {
        return nil, err
    }

    rpcError := respPayload.Header.Get(micro.ErrorHeader)
    if rpcError != "" {
        return nil, errors.New(rpcError)
    }

    resp := new({{ if not (samePackage .Output.GoIdent.GoImportPath $.GoImportPath) }}{{ trimPackagePath .Output.GoIdent.GoImportPath }}.{{ end }}{{ .Output.GoIdent.GoName }})
    if err := googleProto.Unmarshal(respPayload.Data, resp); err != nil {
        return nil, err
    }

    return resp, nil
}
{{ end }}

{{ end }}
`
