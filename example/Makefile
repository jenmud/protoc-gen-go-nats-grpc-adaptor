PROTOC_PATH=~/.local/nats-protoc-gen/protoc
GOROOT := $(shell go env GOROOT)
GOPATH := $(shell go env GOPATH)
PATH=$(GOROOT)/bin:$(GOPATH)/bin:$(PROTOC_PATH)/bin:/usr/bin:/usr/local/bin:$$PATH

all: generate

generate: build
	PATH=$(PATH):./builds protoc \
	--proto_path=./ \
	--go_out=./ \
	--go_opt=paths=source_relative \
	--go-nats_out=./ \
	--go-nats_opt=paths=source_relative \
	--go-grpc_out=./ \
	--go-grpc_opt=paths=source_relative \
	messages.proto example.proto
