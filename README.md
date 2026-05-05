# Go API Starter Kit

A production-ready, layered Go API starter with authentication, queues, migrations,
localized validation, and OpenTelemetry tracing.

## Stack

| Layer | Technology |
|---|---|
| HTTP | Echo v5 |
| Database | GORM + PostgreSQL |
| Cache & Queue | Redis (sessions, Asynq jobs) |
| Auth | JWT (access + refresh tokens), OTP forgot-password flow |
| Observability | OpenTelemetry -> Jaeger |
| DI | Uber FX |
| Email | go-mailgen |
| Migrations | Migris |

## Prerequisites

- Go 1.25.x
- Docker and Docker Compose for PostgreSQL, Redis, and Jaeger
- `ginkgo` for `make test`
- `golangci-lint` for `make lint`

## Quick Start

```bash
cp .env.example .env
docker compose up -d db redis jaeger
make tidy
make migrate-up
make run
```

| Service | URL |
|---|---|
| API | `http://localhost:8080` |
| OpenAPI docs | `http://localhost:8080/docs` |
| Jaeger UI | `http://localhost:16686` |

See the Docker section for the containerized startup flow. The app container
does not run migrations automatically.

## Project Structure

```
.
|-- main.go
|-- cmd/               # CLI commands (serve, serve-all, queue, migrate)
|-- db/migrations/
|-- pkg/               # Shared packages (env, problem, validator)
|-- test/              # E2E tests and shared test helpers
`-- internal/
    |-- config/
    |-- delivery/
    |   |-- http/      # Handlers, routes, middleware
    |   `-- queue/     # Asynq worker and handlers
    |-- domain/        # Business entities and errors
    |-- hash/          # Password hashing, JWT
    |-- infra/         # PostgreSQL, Redis, SMTP clients
    |-- lang/          # i18n (English, Indonesian)
    |-- logger/        # slog setup
    |-- model/         # DB models
    |-- repository/    # Data persistence
    |-- service/       # Business logic
    `-- telemetry/     # OTel setup
```

## Features

- **Auto-validation** - Custom binder validates request DTOs on `c.Bind`, no handler boilerplate
- **RFC 7807 errors** - Standardized `application/problem+json` responses
- **Localized validation** - English and Indonesian error messages
- **Distributed tracing** - OTel spans across HTTP, GORM, Redis, and queue layers with trace-log correlation
- **Background jobs** - Asynq task processing managed via FX lifecycle
- **Health checks** - Endpoint verifying both DB and Redis connectivity

## API

Health:

```text
GET /health
```

## Auth

Sessions store both access and refresh tokens in Redis. Refresh tokens rotate on use. Password reset revokes the active session.

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh-token
POST /api/v1/auth/forgot-password/send-otp
POST /api/v1/auth/forgot-password/verify-otp
POST /api/v1/auth/forgot-password/reset-password
```

Profile and users:

```text
GET /api/v1/profile
PUT /api/v1/profile
PUT /api/v1/profile/password
GET /api/v1/users
```

Protected endpoints require an `Authorization: Bearer <access_token>` header.
Validation messages follow the request locale from `Accept-Language`; supported
locales are English and Indonesian.

## CLI

```bash
go run . serve          # HTTP server only
go run . serve-all      # HTTP + queue worker
go run . queue          # Queue worker only
go run . migrate status
go run . migrate up
go run . migrate down
```

## Make Targets

```bash
make run           # run the HTTP API locally
make test          # run unit/package tests
make test-e2e      # run E2E tests
make coverage      # unit coverage
make coverage-all  # merged unit + E2E coverage
make coverage-html # render coverage.html from coverage.out
make lint          # run golangci-lint
make lint-fix      # run golangci-lint --fix
make fmt           # go fmt ./...
make tidy          # go mod tidy
make build         # compile bin/go-starter-kit
make migrate-up    # apply migrations
make migrate-down  # roll back migrations
make docker-build  # build the app image
make docker-run    # run the app image with .env
```

## Docker

Start dependencies for local development:

```bash
docker compose up -d db redis jaeger
```

Then run migrations and start the API:

```bash
make migrate-up
make run
```

To run the app container too, after migrations:

```bash
docker compose up --build
```

The app container does not run migrations automatically.

## Observability

Configure via `.env`:

```env
OTEL_ENABLED=true
OTEL_SERVICE_NAME=my-service
OTEL_EXPORTER=otlp                          # otlp | none
OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
OTEL_EXPORTER_OTLP_INSECURE=true
OTEL_TRACES_SAMPLER_RATIO=1.0
OTEL_EXPORT_TIMEOUT=5s
```
