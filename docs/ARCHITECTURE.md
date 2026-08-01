# Architecture

Hexagonal / DDD layering with dependency direction **inward**.

```
presentation  →  application  →  domain
                      ↑
               infrastructure (adapters)
```

Enforced by `.go-arch-lint.yml` (`make lint-arch`).

## Layers

| Layer | Responsibility | May import |
|-------|----------------|------------|
| **domain** | Entities, value IDs, repository **ports** | domain only |
| **application** | Use cases / services | domain (+ application) |
| **infrastructure** | Adapters (Mongo, future Postgres, email, …) | domain, application |
| **presentation** | HTTP handlers, DTOs, routes | domain, application (not infrastructure) |
| **config** | YAML / env loading | anywhere |

Composition root is `cmd/server/main.go`: load config → `internal/platform` (logger + OTel) → connect DB → build adapters → build services → start HTTP.

## Shared packages (`go-pkgs`)

Reusable modules live in [`github.com/gambitier/go-pkgs`](https://github.com/gambitier/go-pkgs) as **independent** Go modules:

| Module | Role |
|--------|------|
| `errors` | Domain error types (codes, fields, wrap helpers) |
| `apiresponse` | RFC 9457 Problem Details + code↔HTTP status mapping |
| `logging` | Structured logger, sinks, `X-Correlation-ID` |
| `observability` | OTel Init + Fiber middleware |

`apiresponse` depends on `errors`. Composition (logger ↔ OTel bridge, domain errors → log fields) stays in **`internal/platform`**.

## Persistence port vs application ports

- **Domain repository** (`domain/item.Repository`): persistence for aggregates. Mongo implements it today.
- **`application/ports`**: reserve for *external* systems (payment gateway, email, third-party APIs)—not for Mongo/Postgres.

## Mongo adapter pattern

Per aggregate under `infrastructure/persistence/mongodb/<aggregate>/`:

| File | Role |
|------|------|
| `model.go` | BSON struct + field name constants + collection name |
| `converter.go` | domain ↔ Mongo document |
| `repository.go` | implements domain repository |
| `errors.go` | constant error messages |

`factory.go` returns `domain/repository.Persistence` so growing to multiple aggregates stays consistent.

## Switching Mongo → Postgres

1. Add `internal/infrastructure/persistence/postgres/item` implementing `domain/item.Repository`.
2. In `cmd/server/main.go`, construct that adapter instead of (or based on config instead of) Mongo:

```go
// persistence.driver: mongo | postgres  (wire here only)
var itemRepo domainitem.Repository
switch cfg.Persistence.Driver { // add when you need it
case "postgres":
    itemRepo = postgresitem.New(...)
default:
    // existing Mongo InitializePersistence path
}
itemSvc := appitem.NewService(itemRepo)
```

3. Do **not** change `application/item` or `domain/item`.

## HTTP conventions

- Handlers use presentation DTOs (`handlers/item/dto.go`), not Mongo models.
- Success: bare resource JSON (`application/json`); no proprietary envelope.
- Errors: RFC 9457 problem details via `go-pkgs/apiresponse` (`application/problem+json`).
- Correlation: `X-Correlation-ID` header only (not in JSON body).
- Routes live under `/api/v1/...` (versioned API; probes and swagger stay at the root).

## Swagger

- Annotations on `cmd/server/main.go` and handlers.
- Regenerate: `make swagger`.
- UI gated by `swagger.enabled` and development environment.
