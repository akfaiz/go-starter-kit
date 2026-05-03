---
name: go-api-layering
description: >
  Use this skill when changing API flow in handlers, services, repositories,
  domain interfaces, routes, or Fx wiring. It enforces this project's layering
  and response/error conventions.
---

# Go API Layering

## Purpose

Keep request handling consistent with this architecture:

`handler -> service -> repository`

## Apply these rules

1. Keep handlers thin: bind/validate DTO, call service, return error upward.
2. Put business logic in `internal/service`, not in handler or repository.
3. Keep data access in `internal/repository`.
4. Put cross-layer contracts in `internal/domain` interfaces/types.
5. Register routes in `internal/delivery/http/routes/routes.go`.
6. Use `dto.NewResponse` / `dto.NewMessage` for success responses.
7. Let centralized HTTP error handling shape failures (`application/problem+json`).

## Typical edit map

- Route changes: `internal/delivery/http/routes/routes.go`
- HTTP handler changes: `internal/delivery/http/handler/`
- Business behavior changes: `internal/service/`
- Data persistence changes: `internal/repository/`, `internal/model/`, `db/migrations/`
- Dependency wiring: `cmd/serve/serve.go` and module providers

## Verification commands

```bash
make build
make test
```
