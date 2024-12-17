// This is a demo application impleneting the gRPC service and then running
// an embeded NATS server.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
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

	addr := flag.String("address", "localhost:4222", "NATS server address")
	workers := flag.Int("worker-pool", 1, "Worker pool size")
	flag.Parse()

	host, portStr, err := net.SplitHostPort(*addr)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	opts := server.Options{
		Host: host,
		Port: port,
	}

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

	ms, err := proto.NewNATSGreeterServer(ctx, nc, &DemoService{}, cfg, proto.WithConcurrentJobs(*workers))
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

	now := time.Now()

	wg := sync.WaitGroup{}
	logger.Info(" ---------- start ----------")

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		logger.Info("sending hello request")
		resp, err := client.SayHello(ctx, &proto.HelloRequest{Name: "FooBar"})
		if err != nil {
			logger.Error("error saying hello", slog.String("reason", err.Error()))
			return
		}
		logger.Info("first resp: " + resp.GetMessage())
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		logger.Info("sending hello request again")
		againResp, err := client.SayHelloAgain(ctx, &proto.HelloRequest{Name: "FooBar"})
		if err != nil {
			logger.Error("error saying hello AGAIN", slog.String("reason", err.Error()))
		}
		logger.Info("again resp: " + againResp.GetMessage())
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
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

		logger.Info("sending hello save meta request")
		metaResp, err := client.SaveMetadata(ctx, metaStruct)
		if err != nil {
			logger.Error("error saving meta", slog.String("reason", err.Error()))
			return
		}

		logger.Info("save meta resp", slog.Any("meta", metaResp))
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		logger.Info("sending say goodbye request")
		byeResp, err := client.SayGoodbye(ctx, &proto.SayGoodbyeRequest{Name: "FooBar"})
		if err != nil {
			logger.Error("error saying bye", slog.String("reason", err.Error()))
			return
		}
		logger.Info("bye resp: " + byeResp.GetMessage())
	}(&wg)

	wg.Wait()
	logger.Info("------ done ----------", slog.Duration("duration", time.Since(now)))
}
