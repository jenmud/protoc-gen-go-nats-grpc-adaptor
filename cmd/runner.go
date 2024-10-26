package cmd

import (
	"log/slog"
	"text/template"

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

const templ = `
// Code generated by protoc-gen-go-nats. DO NOT EDIT.
// source: {{.GeneratedFilenamePrefix}}.proto

package {{.GoPackageName}}

import (
    "context"
    "log/slog"
    proto "google.golang.org/protobuf/proto"
	nats "github.com/nats-io/nats.go"
	micro "github.com/nats-io/nats.go/micro"
)

{{ range .Services }}

// NATSMicroService extends the gRPC Server buy adding a method on the service which
// returns a NATS registered micro.Service.
func (s *{{ .GoName }}Server) NATSMicroService(nc *nats.Conn) (micro.Service, error) {
    cfg := micro.Config{
    	Name: "{{ .GoName }}Server",
        Version: "0.0.0",
    }

    srv, err := micro.AddService(nc, cfg)
    if err != nil {
        return nil, err
    }

    {{ range .Methods }}
    err = srv.AddEndpoint(
        "svc.{{ .GoName }}",
        micro.HandlerFunc(
        	func(req micro.Request){
         		r := &{{ .Input.GoIdent.GoName }}{}

           		/*
             		Unmarshal the request.
             	*/
         		if err := proto.Unmarshal(req.Data(), r); err != nil {
           			if err := req.Error("500", err.Error(), nil); err != nil {
              			slog.Error("error sending response error", slog.String("reason", err.Error()))
                 		return
              		}

                	return
                }

                /*
                	Forward on the original request to the original gRPC service.
                */
                resp, err := s.{{ .GoName }}(context.TODO(), r)
                if err != nil {
               		if err := req.Error("500", err.Error(), nil); err != nil {
              			slog.Error("error sending response error", slog.String("reason", err.Error()))
                 		return
                 	}
                }

                /*
                	Take the response from the gRPC service and dump it as a byte array.
                */
                respDump, err := proto.Marshal(resp)
                if err != nil {
               		if err := req.Error("500", err.Error(), nil); err != nil {
              			slog.Error("error sending response error", slog.String("reason", err.Error()))
                 		return
                 	}
                }

                /*
                	Finally response with the original response from the gRPC service.
                */
                if err := req.Respond(respDump); err != nil {
               		if err := req.Error("500", err.Error(), nil); err != nil {
              			slog.Error("error sending response error", slog.String("reason", err.Error()))
                 		return
                 	}
                }
         	},
        ),
    )

    if err != nil {
        return nil, err
    }
    {{ end }}

    return srv, nil
}

{{ end }}
`

// generateFile generates a .pb.go file.
func generateFile(gen *protogen.Plugin, file *protogen.File) error {

	tmpl, err := template.New("nats-micro-service").Parse(templ)
	if err != nil {
		return err
	}

	filename := file.GeneratedFilenamePrefix + ".nats.pb.go"

	logger := slog.With("filename", filename)
	logger.Info("generating the files")

	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	if err := tmpl.Execute(g, file); err != nil {
		logger.Error("failed to execute the template", slog.String("reason", err.Error()))
		return err
	}

	return nil
}
