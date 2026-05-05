# Go API Starter Kit - Project Context

This is a production-ready, layered Go API starter kit. It follows clean architecture principles and utilizes Uber FX for dependency injection.

## Project Overview

- **Core Technologies:** Go 1.25+, Echo v5 (HTTP), GORM (PostgreSQL), Redis (Asynq, Sessions), Uber FX (DI), OpenTelemetry (Tracing).
- **Architecture:** Layered Architecture
  - `delivery/http`: Echo handlers, routes, and middleware.
  - `delivery/queue`: Asynq task workers and handlers.
  - `service`: Business logic layer.
  - `repository`: Data persistence layer (GORM).
  - `domain`: Core business entities, interfaces, and errors.
  - `model`: GORM database models.
  - `infra`: External clients (DB, Redis, SMTP).

## Building and Running

- **Environment Setup:** `cp .env.example .env` and start dependencies via `docker compose up -d db redis jaeger`.
- **Run API:** `make run` or `go run . serve`.
- **Run Worker:** `go run . queue`.
- **Run All:** `go run . serve-all`.
- **Migrations:**
  - `make migrate-up` / `go run . migrate up`
  - `go run . migrate status`
- **Linting:** `make lint` or `make lint-fix`.
- **Formatting:** `make fmt`.

## Development Conventions

- **Dependency Injection:** Use `go.uber.org/fx`. Register new components in the `Module` variable of the respective package.
- **API Layering:**
  - **Handlers:** Use `c.Bind` for automatic request binding and validation (via `pkg/validator`). Return `error` from handlers.
  - **DTOs:** Located in `internal/delivery/http/handler/dto`. Use for request/response bodies.
  - **Errors:** Map business errors from the `service` layer to `problem` (RFC 7807) or `validator` errors in the `handler`.
- **Business Logic:** Encapsulated in the `service` layer. Services interact with `repository` interfaces defined in the `domain` layer.
- **Persistence:** Use GORM in the `repository` layer. Models are defined in `internal/model`.
- **Validation:** Use `github.com/go-playground/validator/v10`. Localized validation messages are supported.
- **I18n:** Use `github.com/invopop/ctxi18n` for localized strings. Catalogs are in `internal/lang`.
- **Tracing:** Spans are automatically propagated across layers (HTTP -> Service -> Repository -> DB/Redis/Queue). Use `telemetry.StartSpan(ctx, "name")` for custom spans.

## Testing Practices

- **Framework:** Ginkgo & Gomega.
- **Unit Tests:** `make test`. Located alongside the code (e.g., `*_test.go`).
- **E2E Tests:** `make test-e2e`. Located in `test/e2e`.
- **Mocks:** Generated via `go.uber.org/mock`. Located in `test/mocks`.

## Key Files

- `main.go`: Entry point, initializes FX app.
- `cmd/`: CLI command definitions (using `urfave/cli/v3`).
- `internal/config/config.go`: Configuration loading and validation.
- `internal/domain/error.go`: Centralized business errors.
- `pkg/problem/errors.go`: RFC 7807 problem response helpers.
- `pkg/validator/binder.go`: Custom Echo binder for auto-validation.
