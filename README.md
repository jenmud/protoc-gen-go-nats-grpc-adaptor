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

## Using NATS cli client

You can query NATS using the NATS cli client

```bash
$ ./nats micro list
# All Micro Services

| Name          | Version | ID                     | Description                                       |
|---------------|---------|------------------------|---------------------------------------------------|
| GreeterServer | 0.0.1   | TWaLR1B60j04SCblXqY1xP | NATS micro service adaptor wrapping GreeterServer |
```

Fetch information about the microservice
```bash
$ ./nats micro info GreeterServer
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
