# nats-protoc-gen
NATS protoc gen is a protoc plugin that simplifies generating NATS microservices by wrapping gRPC services.

## Why
This project aims to leverage NATS as the primary communication layer while taking advantage of its built-in service discovery capabilities. It allows developers to create unified services that can be accessed either directly through gRPC or as a NATS microservice.

## Build
To build the plugin simply run

```bash
$ make build
```

## Using the plugin
To use the plugin, run the protoc compiler with the following command. Make sure that `protoc-gen-go-nats` is in your $PATH.

```bash
# assuming that the binary is found under ./builds after `make build`
PATH=$(PATH):./builds protoc \
--proto_path=./example \
--go_out=./example \
--go_opt=paths=source_relative \
--go-nats_out=./example \
--go-nats_opt=paths=source_relative \
--go-grpc_out=./example \
--go-grpc_opt=paths=source_relative \
example.proto messages.proto
```

The command above will build the example directory, so you will need to alter the command to point to your own protobuf files.
