PROTOC_VERSION=28.3
PROTOC_URL=https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-linux-x86_64.zip
PROTOC_PATH=~/.local/concencus/protoc
GOROOT := $(shell go env GOROOT)
GOPATH := $(shell go env GOPATH)
PATH=$(GOROOT)/bin:$(GOPATH)/bin:$(PROTOC_PATH)/bin:/usr/bin:/usr/local/bin:$$PATH
TEMP := $(shell /usr/bin/mktemp -d)

all: generate

install: download-and-install-protoc install-tools

download-protoc:
	PATH=$(PATH) /usr/bin/mkdir -p $(PROTOC_PATH)
	PATH=$(PATH) /usr/bin/curl -L $(PROTOC_URL) -o $(TEMP)/protoc.zip

download-and-install-protoc: download-protoc
	PATH=$(PATH) /usr/bin/unzip -u $(TEMP)/protoc.zip -d $(PROTOC_PATH)
	PATH=$(PATH) /usr/bin/rm -rf $(TEMP)

install-tools:
	PATH=$(PATH) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	PATH=$(PATH) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	PATH=$(PATH) go install github.com/air-verse/air@latest
	PATH=$(PATH) go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	PATH=$(PATH) go install github.com/spf13/cobra-cli@latest

update-deps:
	PATH=$(PATH) go mod tidy

vendor: update-deps
	PATH=$(PATH) go mod vendor

build:
	go build -o builds/protoc-gen-go-nats .

generate: build
	PATH=$(PATH):./builds protoc \
	--go_out=. \
	--go_opt=paths=source_relative \
	--go-nats_out=. \
	--go-nats_opt=paths=source_relative \
	--go-grpc_out=. \
	--go-grpc_opt=paths=source_relative \
	example/example.proto
