package main

import (
	"context"
	"errors"

	proto "github.com/jenmud/nats-protoc-gen/example"
)

type DemoService struct {
	proto.UnimplementedGreeterServer
}

func (s *DemoService) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloReply, error) {
	return &proto.HelloReply{Message: "Hi " + req.GetName()}, nil
}

func (s *DemoService) SayHelloAgain(ctx context.Context, req *proto.HelloRequest) (*proto.HelloReply, error) {
	return nil, errors.New("some random example error")
}

func (s *DemoService) SayGoodbye(ctx context.Context, req *proto.SayGoodbyeRequest) (*proto.SayGoodbyeReply, error) {
	return &proto.SayGoodbyeReply{Message: "Later " + req.GetName()}, nil
}
