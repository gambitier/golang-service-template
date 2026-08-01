# Golang Service Template

Standalone hexagonal (ports & adapters) Go service scaffold. Shared foundations come from independent [`go-pkgs`](https://github.com/gambitier/go-pkgs) modules (`errors`, `logging`, `observability`); wiring glue lives in `internal/platform`.

## Quick start

```bash
# 1. MongoDB
docker run -d -p 27017:27017 --name mongo mongo:7

# 2. Env (optional; development overlay has local defaults)
cp .env.example .env

# 3. Dependencies
make setup

# 4. Run (loads config.yaml + config.development.yaml)
make run
```

If `go-pkgs` is ever private, set:

```bash
export GOPRIVATE=github.com/gambitier/*
```

API base: `http://localhost:8080/items/api/v1`

Swagger UI (dev): `http://localhost:8080/items/swagger/index.html`  
(basic auth: `admin` / `changeme` by default)

### Example

```bash
curl -s -X POST http://localhost:8080/items/api/v1/items \
  -H 'Content-Type: application/json' \
  -d '{"name":"Notebook","description":"ruled"}'
```

## Layout

```
cmd/server/                  # composition root
internal/
  domain/                    # entities + repository ports
  application/               # use cases (depend on ports only)
  infrastructure/persistence/mongodb/  # Mongo adapter (example)
  presentation/http/         # handlers, DTOs, routes
  platform/                  # glue: logger + OTel + domainerr log fields
  server/                    # HTTP server wiring
  config/
swagger/                     # generated OpenAPI (swag)
docs/ARCHITECTURE.md
```

## Make targets

| Target | Description |
|--------|-------------|
| `make run` | Run with `-config config.yaml -env development` |
| `make test` | Unit tests (+ Mongo repo test if Mongo is up) |
| `make swagger` | Regenerate `swagger/` from annotations |
| `make lint-arch` | Enforce DDD import rules (`go-arch-lint`) |
| `make check` | lint + test |

## Rename for a new service

1. Replace module path: `github.com/gambitier/golang-service-template` → yours (`go mod edit -module ...` + find/replace).
2. Rename the toy `item` domain to your aggregate.
3. Adjust HTTP prefix (`/items`) and Swagger `@BasePath`.
4. Keep wiring changes in `cmd/server/main.go` / `internal/platform` when swapping persistence or telemetry.

## Adding Postgres later

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md). Domain and application stay unchanged; implement `item.Repository` under `infrastructure/persistence/postgres` and swap construction in `main`.
