# Example

This is an example directory which has a simple `application` used for starting and running the example demo.

## Build

If you need to rebuild the example source, run

```bash
# in the root directory containing the `Makefile` run
$ make generate
```

## Running the example app

The example application will start a NATS server running on the default port `4222`.

```bash
# assuming that you are in the example directory
# this example app will exit with CTRL+C or will self exit after 5s
$ go run ./cmd
```
