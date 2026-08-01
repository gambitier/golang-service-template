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

API base: `http://localhost:8080/api/v1`

Swagger UI (dev): `http://localhost:8080/swagger/index.html`  
(basic auth: `admin` / `changeme` by default)

### Example

```bash
curl -s -X POST http://localhost:8080/api/v1/items \
  -H 'Content-Type: application/json' \
  -d '{"name":"Notebook","description":"ruled"}'
```

## Layout

```
cmd/server/                  # tiny entry: flags + bootstrap.Run
internal/
  item/                      # vertical slice (domain/application/infrastructure/presentation)
  shared/
    bootstrap/               # composition root
    lifecycle/               # Component Start/Stop + graceful shutdown
    platform/                # logger + OTel
    presentation/http/       # middleware.Register, response, routes shell
    server/                  # HTTP lifecycle.Component
    infrastructure/          # Mongo connect, bsonutil, …
  config/
swagger/                     # generated OpenAPI (swag)
bruno/                       # Bruno API collection (local Item CRUD)
docs/ARCHITECTURE.md
```


## Bruno

Open [`bruno/`](bruno/) in Bruno, select the **local** environment, and run the Item requests against `make run`.

## Make targets

| Target | Description |
|--------|-------------|
| `make run` | Run with `-config config.yaml -env development` |
| `make test` | Unit tests (+ Mongo repo test if Mongo is up) |
| `make fmt` | Format Go sources (`gofmt`) |
| `make hooks` | Enable `.githooks` (runs `make fmt` on every commit) |
| `make swagger` | Regenerate `swagger/` from annotations |
| `make lint-arch` | Enforce DDD import rules (`go-arch-lint`) |
| `make check` | lint + test |

`make setup` enables hooks automatically. Otherwise run `make hooks` once after clone.

## Rename for a new service

1. Replace module path: `github.com/gambitier/golang-service-template` → yours (`go mod edit -module ...` + find/replace).
2. Rename the toy `item` slice under `internal/item/` to your aggregate (same four layers).
3. Adjust API routes (`/api/v1/...`) and Swagger `@BasePath`.
4. Wire new slices in `shared/bootstrap` / shared routes; keep lifecycle components for process deps.

## Adding Postgres later

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md). Domain and application stay unchanged; implement `item/domain.Repository` under the item slice and construct that adapter where the HTTP component is built in bootstrap.
