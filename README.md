# sega_idac_client

An API client for accessing Sega's Initial D The Arcade (IDAC) APIs, which also
supports more advanced query options. It exposes its functionality over gRPC.

The project follows the same clean-architecture layout as `mcqueens_tea_cup`,
minus the Discord bot layer, plus a gRPC server for exposing endpoints.

## Architecture

```
cmd/sega_idac_client/main.go        # entrypoint: load config, init clients, start gRPC server
proto/health/v1/health.proto        # gRPC service/message definitions
internal/
├── config/                         # env-driven configuration (Sega, AllNet, gRPC)
├── domain/
│   ├── entity/                     # domain models (IDAC records, store locations, helpers)
│   └── port/                       # client interfaces (SegaIDACClient, AllNetClient)
└── adapter/
    ├── client/
    │   ├── sega_idac/              # SEGA IDAC HTTP client (replicated from mcqueens_tea_cup)
    │   └── all_net/                # All.Net HTML-scraping client (replicated)
    └── grpc/
        ├── server.go               # gRPC server bootstrap + service registration
        ├── health/                 # Health service handler + test
        └── gen/healthv1/           # generated protobuf/gRPC code (do not edit)
```

## Prerequisites

- Go 1.25+
- To regenerate protobuf code: `protoc`, plus the Go plugins:
  ```sh
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```

## Setup

```sh
cp .env.example .env   # fill in the SEGA_IDAC_* and ALLNET_* values
make tidy
```

## Common tasks

```sh
make proto   # regenerate Go code from proto/health/v1/health.proto
make build   # go build ./...
make run     # start the gRPC server (defaults to 0.0.0.0:50051)
make test    # go test ./...
```

## gRPC endpoints

| Service                   | RPC     | Description                          |
| ------------------------- | ------- | ------------------------------------ |
| `health.v1.HealthService` | `Check` | Returns the server's serving status. |

Server reflection is enabled, so tools such as `grpcurl` can introspect the
service:

```sh
grpcurl -plaintext localhost:50051 health.v1.HealthService/Check
```
