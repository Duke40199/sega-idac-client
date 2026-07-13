MODULE     := sega_idac_client
CMD_PATH   := ./cmd/sega_idac_client

# All .proto files under proto/ are compiled by the `proto` target.
PROTO_DIR   := proto
PROTO_FILES := $(shell find $(PROTO_DIR) -name '*.proto')

# Pinned codegen plugin versions (kept in sync with go.mod).
PROTOC_GEN_GO_VERSION      := v1.35.2
PROTOC_GEN_GO_GRPC_VERSION := v1.5.1

# protoc plugins are installed into $(GOBIN); make sure protoc can find them.
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

.PHONY: proto proto-tools build run tidy test fmt

## Compile all protobuf definitions into Go code
proto: proto-tools
	@command -v protoc >/dev/null 2>&1 || { echo "❌ protoc not found. Install it (e.g. 'brew install protobuf')."; exit 1; }
	protoc \
		--go_out=. --go_opt=module=$(MODULE) \
		--go-grpc_out=. --go-grpc_opt=module=$(MODULE) \
		$(PROTO_FILES)
	@echo "✅ Generated Go code for: $(PROTO_FILES)"

## Install the protoc Go plugins at the pinned versions
proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

## Tidy go modules
tidy:
	go mod tidy

## Build all packages
build:
	go build ./...

## Run the gRPC server
run:
	go run $(CMD_PATH)

## Run tests
test:
	go test ./...

## Format code
fmt:
	go fmt ./...
