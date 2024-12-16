// This is a demo application impleneting the gRPC service and then running
// an embeded NATS server.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"flag"

	proto "github.com/jenmud/protoc-gen-go-nats-grpc-adaptor/example"
	server "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"google.golang.org/protobuf/types/known/structpb"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	timeout := flag.Duration("timeout", 5*time.Second, "Timeout to auto quit the demo application")
	flag.Parse()

	opts := server.Options{}
	ns, err := server.NewServer(&opts)
	if err != nil {
		panic(err)
	}

	logger := slog.With(
		slog.Group(
			"server",
			slog.String("host", ns.ClientURL()),
		),
	)

	go ns.Start()

	if !ns.ReadyForConnections(5 * time.Second) {
		logger.Error("not ready for connections")
		return
	} else {
		logger.Info("ready for connections")
	}

	/*
		Create the demo gRPC service
	*/
	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		logger.Error("connecting to NATS server", slog.String("reason", err.Error()))
		return
	}

	cfg := micro.Config{
		Name:        "GreeterServer-Demo",
		Version:     "1.0.0",
		QueueGroup:  "example",
		Description: "NATS micro service adaptor wrapping GreeterServer",
	}

	ms, err := proto.NewNATSGreeterServer(ctx, nc, &DemoService{}, cfg)
	if err != nil {
		logger.Error("creating micro service", slog.String("reason", err.Error()))
		return
	}

	logger = slog.With(
		slog.Group(
			"nats-micro-service",
			slog.String("identity", ms.Info().ID),
			slog.String("name", ms.Info().Name),
		),
	)

	logger.Info("nats micro service accepting client requests")
	client := proto.NewNATSGreeterClient(nc, cfg.Name)

	resp, err := client.SayHello(ctx, &proto.HelloRequest{Name: "FooBar"})
	if err != nil {
		logger.Error("error saying hello", slog.String("reason", err.Error()))
		return
	}

	logger.Info("first resp: " + resp.GetMessage())

	againResp, err := client.SayHelloAgain(ctx, &proto.HelloRequest{Name: "FooBar"})
	if err != nil {
		logger.Error("error saying hello AGAIN", slog.String("reason", err.Error()))
	}

	logger.Info("again resp: " + againResp.GetMessage())

	meta := map[string]any{
		"attributes": map[string]any{
			"name": "foo",
			"age":  21,
		},
	}

	metaStruct, err := structpb.NewStruct(meta)
	if err != nil {
		panic(err)
	}

	metaResp, err := client.SaveMetadata(ctx, metaStruct)
	if err != nil {
		logger.Error("error saving meta", slog.String("reason", err.Error()))
		return
	}

	logger.Info("save meta resp", slog.Any("meta", metaResp))

	byeResp, err := client.SayGoodbye(ctx, &proto.SayGoodbyeRequest{Name: "FooBar"})
	if err != nil {
		logger.Error("error saying bye", slog.String("reason", err.Error()))
		return
	}

	logger.Info("bye resp: " + byeResp.GetMessage())

	tctx, tcancel := context.WithTimeout(ctx, *timeout)
	defer tcancel()

	select {
	case <-ctx.Done():
		logger.Info("shutdown", slog.String("reason", ctx.Err().Error()))
	case <-tctx.Done():
		logger.Info("shutdown", slog.String("reason", tctx.Err().Error()))
	}
}
