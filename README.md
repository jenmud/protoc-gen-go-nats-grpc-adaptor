# nats-grpc-adaptor

NATS protoc-gen is a protoc plugin that simplifies the generation of NATS microservices by wrapping gRPC services.

## Why Use This Plugin?

This project leverages NATS as the primary communication layer while utilizing its built-in service discovery capabilities. It enables developers to create unified services accessible directly through gRPC or as a NATS microservice.

## Build Instructions

To build the plugin, simply run:

```bash
make build
```

## Installing the Plugin

You can install the latest version of the plugin using the following command:

```bash
go install github.com/jenmud/protoc-gen-go-nats-grpc-adaptor@latest
```

It is recommended to install goimports and run it after generating the plugin files:

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

## Example Usage of goimports

To fix all files automatically:

```bash
goimports -w .
```

To fix only the generated proto files:

```bash
goimports -w ./example/example-nats-grpc-adaptor.pb.go
```

## Using the Plugin

To use the plugin, run the protoc compiler with the following command. Ensure that protoc-gen-go-nats-grpc-adaptor is in your $PATH:

```bash
# Assuming the binary is located under ./builds after running `make build`
PATH=./builds:$PATH protoc \
--proto_path=./example \
--go_out=./example \
--go_opt=paths=source_relative \
--go-nats-grpc-adaptor_out=./example \
--go-nats-grpc-adaptor_opt=paths=source_relative \
--go-grpc_out=./example \
--go-grpc_opt=paths=source_relative \
example.proto messages.proto
```

## Debugging

To enable debug logging, set the following environment variable:

```bash
export NATS_GRPC_ADAPTOR_DEBUG=true
```

The command above builds the example directory. Modify the command to point to your protobuf files as needed.

## Querying NATS Using the CLI Client

You can query NATS services using the NATS CLI client:

```bash
./nats micro list
```

### Example Output: Listing Microservices

```bash
# All Micro Services

| Name          | Version | ID                     | Description                                       |
|---------------|---------|------------------------|---------------------------------------------------|
| GreeterServer | 0.0.1   | TWaLR1B60j04SCblXqY1xP | NATS micro service adaptor wrapping GreeterServer |
```

### Fetching Microservice Information

To fetch details about a specific microservice:

```bash
./nats micro info GreeterServer
```

Example Output: Microservice Information

```bash
Service Information

          Service: GreeterServer (TWaLR1B60j04SCblXqY1xP)
      Description: NATS micro service adaptor wrapping GreeterServer
          Version: 0.0.1

Endpoints:

               Name: Greeter
            Subject: svc.greeter.sayhello
        Queue Group: demo

               Name: Greeter
            Subject: svc.greeter.sayhelloagain
        Queue Group: demo

               Name: Greeter
            Subject: svc.greeter.saygoodbye
        Queue Group: demo

Statistics for 3 Endpoint(s):

  Greeter Endpoint Statistics:

           Requests: 1 in group "demo"
    Processing Time: 53µs (average 53µs)
            Started: 2024-11-17 16:40:49 (1m4s ago)
             Errors: 0

  Greeter Endpoint Statistics:

           Requests: 1 in group "demo"
    Processing Time: 25µs (average 25µs)
            Started: 2024-11-17 16:40:49 (1m4s ago)
             Errors: 1
         Last Error: 500:some random example error

  Greeter Endpoint Statistics:

           Requests: 1 in group "demo"
    Processing Time: 17µs (average 17µs)
            Started: 2024-11-17 16:40:49 (1m4s ago)
             Errors: 0
```
