// This is a demo application impleneting the gRPC service and then running
// an embeded NATS server.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	proto "github.com/jenmud/nats-protoc-gen/example"
	server "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

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

	ms, err := proto.NewNATSGreeterServer(ctx, nc, &DemoService{}, "0.0.1", "demo")
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
	client := proto.NewNATSGreeterClient(nc)

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

	byeResp, err := client.SayGoodbye(ctx, &proto.SayGoodbyeRequest{Name: "FooBar"})
	if err != nil {
		logger.Error("error saying bye", slog.String("reason", err.Error()))
		return
	}

	logger.Info("bye resp: " + byeResp.GetMessage())

	tctx, tcancel := context.WithTimeout(ctx, 5*time.Second)
	defer tcancel()

	select {
	case <-ctx.Done():
		logger.Info("shutdown", slog.String("reason", ctx.Err().Error()))
	case <-tctx.Done():
		logger.Info("shutdown", slog.String("reason", tctx.Err().Error()))
	}
}
