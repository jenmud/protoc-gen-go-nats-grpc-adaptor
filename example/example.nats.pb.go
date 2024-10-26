// Code generated by protoc-gen-go-nats. DO NOT EDIT.
// source: example/example.proto

package example

import (
	"context"
	"log/slog"
	proto "google.golang.org/protobuf/proto"
	nats "github.com/nats-io/nats.go"
	micro "github.com/nats-io/nats.go/micro"
)

// NATSMicroService extends the gRPC Server buy adding a method on the service which
// returns a NATS registered micro.Service.
func (s *GreeterServer) NATSMicroService(nc *nats.Conn) (micro.Service, error) {
	cfg := micro.Config{
		Name:    "GreeterServer",
		Version: "0.0.0",
	}

	srv, err := micro.AddService(nc, cfg)
	if err != nil {
		return nil, err
	}

	err = srv.AddEndpoint(
		"svc.SayHello",
		micro.HandlerFunc(
			func(req micro.Request) {
				r := &HelloRequest{}

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
				resp, err := s.SayHello(context.TODO(), r)
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

	err = srv.AddEndpoint(
		"svc.SayHelloAgain",
		micro.HandlerFunc(
			func(req micro.Request) {
				r := &HelloRequest{}

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
				resp, err := s.SayHelloAgain(context.TODO(), r)
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

	return srv, nil
}
