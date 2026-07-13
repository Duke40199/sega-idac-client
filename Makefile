.PHONY: proto build run tidy test

# Directory where generated protobuf/gRPC Go code lives.
GEN_DIR := internal/adapter/grpc/gen

# Generate Go code from the .proto definitions.
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc on PATH.
proto:
	protoc \
		--go_out=. --go_opt=module=sega_idac_client \
		--go-grpc_out=. --go-grpc_opt=module=sega_idac_client \
		proto/health/v1/health.proto

tidy:
	go mod tidy

build:
	go build ./...

run:
	go run ./cmd/sega_idac_client

test:
	go test ./...
