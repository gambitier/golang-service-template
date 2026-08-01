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

Composition root is `cmd/server/main.go`: load config → connect DB → build adapters → build services → start HTTP.

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
- Errors are `*domainerr.Error`; `response.Write` maps them to HTTP envelopes.
- Routes live under `/items/api/v1/...` (service prefix + version).

## Swagger

- Annotations on `cmd/server/main.go` and handlers.
- Regenerate: `make swagger`.
- UI gated by `swagger.enabled` and development environment.
