# syntax=docker/dockerfile:1
# Multi-arch static build for sega-idac-client (HTTP gateway :8080 + gRPC :50051).
FROM --platform=$BUILDPLATFORM golang:1.25 AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/sega-idac-client ./cmd/sega-idac-client

# Distroless runtime: no shell, nonroot (matches idac-amane-bot). Config comes from
# env (ConfigMap + Secret); no .env is baked in. certs/ is dev-only, so it's omitted.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/sega-idac-client /sega-idac-client
EXPOSE 8080 50051
USER nonroot:nonroot
ENTRYPOINT ["/sega-idac-client"]
