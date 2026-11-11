#!/usr/bin/env bash
#
# gen-proto.sh — compile every .proto under proto/ into Go + gRPC stubs.
#
# Generated code lands next to the module source (go_package in each .proto
# controls the exact path, e.g. internal/adapter/grpc/gen/segaidacv1).
#
# Usage:
#   ./scripts/gen-proto.sh
#
# Requirements:
#   - protoc           (brew install protobuf)
#   - protoc-gen-go / protoc-gen-go-grpc  (installed automatically below)

set -euo pipefail

# Resolve repo root (parent of this script's dir) so the script works from anywhere.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

MODULE="github.com/Duke40199/sega-idac-client"
PROTO_DIR="proto"
# Vendored google/api/*.proto (annotations, http) so grpc-gateway annotations resolve offline.
THIRD_PARTY_DIR="third_party/googleapis"
OPENAPI_DIR="docs/openapi"

# Pinned codegen plugin versions (keep in sync with go.mod / Makefile).
PROTOC_GEN_GO_VERSION="v1.35.2"
PROTOC_GEN_GO_GRPC_VERSION="v1.5.1"
PROTOC_GEN_GRPC_GATEWAY_VERSION="v2.20.0"
PROTOC_GEN_OPENAPIV2_VERSION="v2.20.0"

# Make sure the Go plugin bin dir is on PATH so protoc can find the plugins.
GOBIN="$(go env GOPATH)/bin"
export PATH="${GOBIN}:${PATH}"

echo "==> Checking protoc..."
if ! command -v protoc >/dev/null 2>&1; then
  echo "❌ protoc not found. Install it first (e.g. 'brew install protobuf')." >&2
  exit 1
fi
protoc --version

echo "==> Installing protoc Go plugins (pinned)..."
go install "google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}"
go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}"
go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@${PROTOC_GEN_GRPC_GATEWAY_VERSION}"
go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@${PROTOC_GEN_OPENAPIV2_VERSION}"

echo "==> Discovering .proto files under ${PROTO_DIR}/ ..."
# shellcheck disable=SC2207
PROTO_FILES=($(find "${PROTO_DIR}" -name '*.proto' | sort))
if [ "${#PROTO_FILES[@]}" -eq 0 ]; then
  echo "❌ No .proto files found under ${PROTO_DIR}/." >&2
  exit 1
fi
printf '    %s\n' "${PROTO_FILES[@]}"

echo "==> Generating Go code (messages, gRPC, gateway) + OpenAPI..."
mkdir -p "${OPENAPI_DIR}"
protoc \
  -I . -I "${THIRD_PARTY_DIR}" \
  --go_out=. --go_opt=module="${MODULE}" \
  --go-grpc_out=. --go-grpc_opt=module="${MODULE}" \
  --grpc-gateway_out=. --grpc-gateway_opt=module="${MODULE}" \
  --openapiv2_out="${OPENAPI_DIR}" --openapiv2_opt=allow_merge=true,merge_file_name=sega-idac-client \
  "${PROTO_FILES[@]}"

echo "✅ Done. Generated stubs + gateway for ${#PROTO_FILES[@]} file(s); OpenAPI in ${OPENAPI_DIR}/."
