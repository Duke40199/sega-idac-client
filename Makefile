MODULE     := github.com/Duke40199/sega-idac-client
CMD_PATH   := ./cmd/sega-idac-client

# All .proto files under proto/ are compiled by the `proto` target.
PROTO_DIR   := proto
PROTO_FILES := $(shell find $(PROTO_DIR) -name '*.proto')
# Vendored google/api/*.proto so grpc-gateway annotations resolve offline.
THIRD_PARTY_DIR := third_party/googleapis
OPENAPI_DIR     := docs/openapi

# Pinned codegen plugin versions (kept in sync with go.mod).
PROTOC_GEN_GO_VERSION           := v1.35.2
PROTOC_GEN_GO_GRPC_VERSION      := v1.5.1
PROTOC_GEN_GRPC_GATEWAY_VERSION := v2.20.0
PROTOC_GEN_OPENAPIV2_VERSION    := v2.20.0

# protoc plugins are installed into $(GOBIN); make sure protoc can find them.
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

.PHONY: proto proto-tools proto-check dev-cert build run tidy test fmt

## Compile all protobuf definitions into Go code (messages, gRPC, gateway) + OpenAPI
proto: proto-tools
	@command -v protoc >/dev/null 2>&1 || { echo "❌ protoc not found. Install it (e.g. 'brew install protobuf')."; exit 1; }
	mkdir -p $(OPENAPI_DIR)
	protoc \
		-I . -I $(THIRD_PARTY_DIR) \
		--go_out=. --go_opt=module=$(MODULE) \
		--go-grpc_out=. --go-grpc_opt=module=$(MODULE) \
		--grpc-gateway_out=. --grpc-gateway_opt=module=$(MODULE) \
		--openapiv2_out=$(OPENAPI_DIR) --openapiv2_opt=allow_merge=true,merge_file_name=sega-idac-client \
		$(PROTO_FILES)
	@echo "✅ Generated Go code + gateway for: $(PROTO_FILES)"

## Fail if the committed stubs under gen/ are stale vs the .proto files.
## Run in CI so a proto change can never land without regenerated code.
proto-check: proto
	@if ! git diff --quiet -- gen; then \
		echo "❌ Generated code in gen/ is out of date. Run 'make proto' and commit the result:"; \
		git --no-pager diff --stat -- gen; \
		exit 1; \
	fi
	@echo "✅ gen/ is up to date with proto/."

## Install the protoc Go plugins at the pinned versions
proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(PROTOC_GEN_GRPC_GATEWAY_VERSION)
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(PROTOC_GEN_OPENAPIV2_VERSION)

## Generate a self-signed TLS cert for local HTTPS testing (certs/dev.crt, certs/dev.key)
dev-cert:
	go run scripts/gen_dev_cert.go

## Tidy go modules
tidy:
	go mod tidy

## Build all packages
build:
	go build ./...

## Run the gRPC server + HTTP gateway
run:
	go run $(CMD_PATH)

## Run tests
test:
	go test ./...

## Format code
fmt:
	go fmt ./...
