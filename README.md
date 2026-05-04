# Go API Starter Kit

A production-ready, layered Go API with batteries included.

## Stack

| Layer | Technology |
|---|---|
| HTTP | Echo v5 |
| Database | GORM + PostgreSQL |
| Cache & Queue | Redis (sessions, Asynq jobs) |
| Auth | JWT (access + refresh tokens), OTP forgot-password flow |
| Observability | OpenTelemetry → Jaeger |
| DI | Uber FX |
| Email | go-mailgen |
| Migrations | Migris |

## Quick Start

```bash
cp .env.example .env
go mod tidy
go run . migrate up
go run . serve
```

| Service | URL |
|---|---|
| API | `http://localhost:8080` |
| OpenAPI docs | `http://localhost:8080/docs` |
| Jaeger UI | `http://localhost:16686` |

## Project Structure

```
.
├── main.go
├── cmd/               # CLI commands (serve, serve-all, queue, migrate)
├── db/migrations/
└── internal/
    ├── config/
    ├── delivery/
    │   ├── http/      # Handlers, routes, middleware
    │   └── queue/     # Asynq worker and handlers
    ├── domain/        # Business entities and errors
    ├── hash/          # Password hashing, JWT
    ├── infra/         # PostgreSQL, Redis, SMTP clients
    ├── lang/          # i18n (English, Indonesian)
    ├── model/         # DB models
    ├── repository/    # Data persistence
    ├── service/       # Business logic
    └── telemetry/     # OTel setup
```

## Features

- **Auto-validation** — Custom binder validates request DTOs on `c.Bind`, no handler boilerplate
- **RFC 7807 errors** — Standardized `application/problem+json` responses
- **Localized validation** — English and Indonesian error messages
- **Distributed tracing** — OTel spans across HTTP, GORM, Redis, and queue layers with trace-log correlation
- **Background jobs** — Asynq task processing managed via FX lifecycle
- **Health checks** — Endpoint verifying both DB and Redis connectivity

## Auth

Sessions store both access and refresh tokens in Redis. Refresh tokens rotate on use. Password reset revokes the active session.

**Endpoints:**
```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh-token
POST /api/v1/auth/forgot-password/send-otp
POST /api/v1/auth/forgot-password/verify-otp
POST /api/v1/auth/forgot-password/reset-password
```

## CLI

```bash
go run . serve          # HTTP server only
go run . serve-all      # HTTP + queue worker
go run . queue          # Queue worker only
go run . migrate status
go run . migrate up
go run . migrate down
```

## Makefile

```bash
make run
make test
make test-e2e
make coverage-all
make lint
make fmt
make tidy
make build
make migrate-up
make migrate-down
make docker-build
make docker-run
```

## Docker

```bash
docker compose up --build
```

Starts the app, PostgreSQL, Redis, and Jaeger.

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