# Architecture

Vertical slices with hexagonal layering **inside each feature**, plus shared process glue.

```
internal/
  item/                  # feature slice (example)
    presentation  →  application  →  domain
                           ↑
                    infrastructure
  shared/                # cross-cutting
    lifecycle/           # Component Start/Stop + App runner
    bootstrap/           # composition root
    platform/            # logger, OTel
    presentation/http/   # middleware.Register, response, routes shell
    server/              # HTTP as lifecycle.Component
    infrastructure/      # Mongo connect, bsonutil, …
  config/
cmd/server/main.go       # flags + signal ctx + bootstrap.Run
```

Enforced by `.go-arch-lint.yml` (`make lint-arch`). Path globs apply per slice (`**/domain/**`, …). `shared/bootstrap`, `shared/server`, and `shared/lifecycle` are excluded (composition / process wiring).

## Feature slice layers

| Layer | Responsibility | May import |
|-------|----------------|------------|
| **domain** | Entities, value IDs, repository **ports** | domain only |
| **application** | Use cases / services | domain (+ application) |
| **infrastructure** | Adapters (Mongo, …) | domain, application |
| **presentation** | HTTP handlers, DTOs, slice `Register` | domain, application, shared presentation/platform |

Add a new feature (e.g. `account/`) with the same four folders; mount its `Register` from shared routes.

## Lifecycle + graceful shutdown

```go
type Component interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

`lifecycle.App` starts components in order and **stops everything in reverse** on signal or start failure. Bootstrap only `Add`s components and calls `Run` — no manual `defer` disconnect/OTel shutdown.

This template registers: Mongo → OTel → HTTP (HTTP wires the item stack after Mongo `Start`).

## Middleware registry

```go
middleware.Register(app, middleware.Deps{Logger: logger, CORS: cfg.Server.CORS})
routes.RegisterRoutes(app, cfg, logger, itemHandler)
```

Global stack (OTel → recover → request scope → CORS) lives in one place.

## Shared packages (`go-pkgs`)

| Module | Role |
|--------|------|
| `errors` | Domain error types |
| `apiresponse` | RFC 9457 Problem Details |
| `logging` | Structured logger, correlation |
| `observability` | OTel Init (format-agnostic `Config`) |

Composition (logger ↔ OTel, Fiber OTel MW) stays in **`internal/shared/platform`**.

## Persistence

- Domain port: `item/domain.Repository`
- Mongo adapter: `item/infrastructure/mongodb`
- Connect / lifecycle: `shared/infrastructure/persistence/mongodb`

Swap Postgres later by implementing the same port and constructing that adapter in the HTTP component (or a dedicated component) inside bootstrap.

## HTTP conventions

- Handlers use presentation DTOs, not Mongo models.
- Success: bare resource JSON; errors: RFC 9457 via `apiresponse`.
- Correlation: `X-Correlation-ID` header only.
- Routes under `/api/v1/...`; probes and swagger at the root.

## Swagger

- Annotations on `cmd/server/main.go` and handlers.
- Regenerate: `make swagger`.
