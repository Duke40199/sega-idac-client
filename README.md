# sega-idac-client

An API client for accessing Sega's Initial D The Arcade (IDAC) APIs, which also
supports more advanced query options. It exposes its functionality over gRPC.

The project follows the same clean-architecture layout as `mcqueens_tea_cup`,
minus the Discord bot layer, plus a gRPC server for exposing endpoints.

## Architecture

```
cmd/sega-idac-client/main.go        # entrypoint: load config, start gRPC server + HTTP gateway
scripts/gen-proto.sh                # compiles proto/**/*.proto into Go, gRPC, gateway + OpenAPI
buf.yaml                            # buf lint + breaking-change rules
.github/workflows/ci.yml            # build/test, proto staleness, buf breaking gate
third_party/googleapis/google/api/  # vendored annotations.proto/http.proto (grpc-gateway deps)
docs/openapi/sega-idac-client.swagger.json  # generated OpenAPI v2 spec (load in Postman/Swagger UI)
proto/
├── health/v1/health.proto          # health service definition (+ google.api.http bindings)
├── sega_idac/v1/sega_idac.proto    # SEGA IDAC service/message definitions (+ HTTP bindings)
└── all_net/v1/all_net.proto        # All.Net store-location service definitions (+ HTTP bindings)
gen/                                # generated protobuf/gRPC/gateway code (do not edit) — importable by consumers
├── healthv1/                       #   *.pb.go, *_grpc.pb.go, *.pb.gw.go
├── segaidacv1/
└── allnetv1/
internal/
├── config/                         # env-driven configuration (Sega, AllNet, gRPC, HTTP)
├── domain/
│   ├── entity/                     # domain models (IDAC records, store locations, helpers)
│   └── port/                       # client interfaces (SegaIDACClient, AllNetClient)
└── adapter/
    ├── client/
    │   ├── sega_idac/              # SEGA IDAC HTTP client (replicated from mcqueens_tea_cup)
    │   └── all_net/                # All.Net HTML-scraping client (replicated)
    ├── grpc/
    │   ├── server.go               # gRPC server bootstrap + service registration
    │   ├── health/                 # Health service handler + test
    │   ├── sega_idac/              # SegaIDACService handler (maps entities <-> protobuf)
    │   └── all_net/                # AllNetService handler (maps entities <-> protobuf)
    └── http/
        └── gateway.go              # grpc-gateway REST/JSON reverse proxy -> gRPC
```

> **Generated stubs live in `gen/` (not `internal/`) on purpose:** Go forbids
> importing another module's `internal/` packages, so keeping the stubs in `gen/`
> lets consumers like `mcqueens_tea_cup` import the typed gRPC client directly:
>
> ```go
> import segaidacv1 "github.com/Duke40199/sega-idac-client/gen/segaidacv1"
> ```

## Prerequisites

- Go 1.25+
- To regenerate protobuf code: `protoc` (e.g. `brew install protobuf`). The Go
  codegen plugins (`protoc-gen-go`, `protoc-gen-go-grpc`) are installed
  automatically at pinned versions by `./scripts/gen-proto.sh` (and by
  `make proto`).

## Regenerating protobuf

Both of these compile every `proto/**/*.proto` into Go + gRPC stubs (output
paths are controlled by each file's `go_package` option):

```sh
./scripts/gen-proto.sh   # standalone script
make proto               # equivalent Makefile target
make proto-check         # CI gate: fails if committed gen/ is stale vs proto/
```

## Versioning & CI (for consumers)

Consumers depend on this module by version (Go modules / MVS). The contract is
guarded by `.github/workflows/ci.yml`:

- **build / vet / test** — the service itself compiles and passes.
- **proto-check** — regenerates and fails if `gen/` drifts from `proto/`, so a
  proto edit can never merge without regenerated code.
- **buf lint + breaking** — on PRs, a backward-incompatible proto change fails
  against `main` (see `buf.yaml`).

To cut a release for consumers: merge to `main`, then tag (`git tag v0.1.0 &&
git push --tags`). Downstream repos pick it up via `go get ...@v0.1.0` — and a
Renovate/Dependabot bot can open that bump automatically.

## Setup

```sh
cp .env.example .env   # fill in the SEGA_IDAC_* and ALLNET_* values
make tidy
```

## Common tasks

```sh
make proto   # regenerate Go/gRPC/gateway code + OpenAPI from all proto files
make build   # go build ./...
make run     # start the gRPC server (:50051) and HTTP gateway (:8080)
make test    # go test ./...
```

## Endpoints

`make run` starts two listeners: gRPC on `GRPC_PORT` (default 50051) and the
grpc-gateway HTTP/JSON proxy on `HTTP_PORT` (default 8080). Every RPC is reachable
both ways; the HTTP bindings are `GET` with request fields mapped to query params.

| Service / RPC                                  | HTTP (gateway)                   | Description                                 |
| ---------------------------------------------- | -------------------------------- | ------------------------------------------- |
| `health.v1.HealthService/Check`                | `GET /v1/health`                 | Server serving status.                      |
| `sega_idac.v1.SegaIDACService/GetListTimeTrail`| `GET /v1/time-trail`             | Time-attack leaderboard (course/area/car).  |
| `…/GetTeamRanking`                             | `GET /v1/team-ranking`           | Team leaderboard for a round.               |
| `…/GetListOBRanking`                           | `GET /v1/ob-ranking`             | Online-battle ranking for a round/area.     |
| `…/GetOBRankingByIGN`                          | `GET /v1/ob-ranking/by-ign`      | Online-battle record for one in-game name.  |
| `…/GetCurrentRound`                            | `GET /v1/current-round`          | Active tournament round number.             |
| `…/FetchConst`                                 | `GET /v1/const`                  | IDAC car/style constant metadata.           |
| `…/GetListPlayerGrade`                         | `GET /v1/player-grade`           | Player-grade ranking for an area.           |
| `…/GetPlayerGradeByIGN`                        | `GET /v1/player-grade/by-ign`    | Player-grade record for one in-game name.   |
| `all_net.v1.AllNetService/GetListStore`        | `GET /v1/stores`                 | IDAC arcade store locations.                |

### Testing over HTTP (grpc-gateway)

```sh
curl -s localhost:8080/v1/health
curl -s localhost:8080/v1/current-round
curl -s 'localhost:8080/v1/player-grade?area_code=area-all'
curl -s 'localhost:8080/v1/ob-ranking/by-ign?ign=SHIRO&round_num=78&area_code=area-all'
curl -s 'localhost:8080/v1/stores?game_code=...&language_code=...&area_code=...'
```

An OpenAPI v2 spec is generated to `docs/openapi/sega-idac-client.swagger.json` —
import it into Postman or a Swagger UI to explore/test interactively.

#### HTTPS / TLS

The gateway serves plain HTTP by default. To serve HTTPS, set `HTTP_TLS_CERT` and
`HTTP_TLS_KEY`. For local testing, generate a self-signed cert (valid for
`localhost`/`127.0.0.1`/`::1`):

```sh
make dev-cert   # writes certs/dev.crt + certs/dev.key (gitignored, dev only)
HTTP_TLS_CERT=certs/dev.crt HTTP_TLS_KEY=certs/dev.key make run
```

```sh
curl -k https://localhost:8080/v1/health                 # -k: skip verification
curl --cacert certs/dev.crt https://localhost:8080/v1/current-round  # trust the dev cert
```

In a browser, a self-signed cert shows a warning — accept the exception, or import
`certs/dev.crt` as a trusted root. In production, point `HTTP_TLS_CERT`/`_KEY` at a
real certificate (or terminate TLS at an ingress/load balancer and leave the
gateway on plain HTTP).

> Note: `SSL_ERROR_RX_RECORD_TOO_LONG` means a client spoke HTTPS to a
> plain-HTTP gateway — either enable TLS as above, or use `http://` (and prefer
> `localhost` over `0.0.0.0`, which browsers may auto-upgrade to https).

### Testing over gRPC

Server reflection is enabled, so `grpcurl` can introspect the service:

```sh
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50051 health.v1.HealthService/Check
grpcurl -plaintext -d '{"area_code":"area-all"}' \
  localhost:50051 sega_idac.v1.SegaIDACService/GetListPlayerGrade
```
