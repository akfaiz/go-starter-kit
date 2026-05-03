# Go API Starter Kit

API-only starter kit built with Go.

## Overview

This repository provides a layered Go API starter with:

- HTTP API built with Echo v5
- CLI entrypoints for serving the API, running migrations, and starting workers
- PostgreSQL persistence with GORM
- Redis-backed session storage, auth rate limiting, and queue support
- JWT authentication with access and refresh tokens
- Forgot-password flow with OTP email delivery
- OpenTelemetry tracing across HTTP, DB, Redis, and queue flows
- Uber FX dependency injection
- Asynq background job processing

## Stack

- Echo v5
- GORM + PostgreSQL
- Redis (auth rate limit, session store, queue)
- Migris migrations
- JWT auth (access + refresh)
- Token pair session management in Redis
- Forgot password with OTP via email
- OpenTelemetry tracing (HTTP, GORM DB, Redis, queue)
- Jaeger for local trace visualization
- go-mailgen for email content
- Uber FX for dependency injection
- Asynq (Redis-based distributed queue)

## Quick Start

```bash
cp .env.example .env
go mod tidy
go run . migrate up
go run . serve
```

Server: `http://localhost:8080`
OpenAPI docs: `http://localhost:8080/docs`
Jaeger UI: `http://localhost:16686`

### CLI Commands

```bash
go run . serve
go run . serve-all
go run . queue
go run . migrate status
go run . migrate up
go run . migrate down
```

## Key Features

- **Automatic Request Validation**: Leverage a custom binder that automatically validates request DTOs during `c.Bind`, reducing handler boilerplate.
- **RFC 7807 Error Responses**: Standardized, machine-readable error responses using the `application/problem+json` content type.
- **Localized Validation**: Built-in support for multiple locales (English and Indonesian) in validation error messages.
- **Dependency Injection**: Robust and modular component management using **Uber FX**.
- **Observability**: Distributed tracing integrated at every level (HTTP, Database, Redis) using **OpenTelemetry**, with trace-to-log correlation in structured logs.
- **Background Jobs**: Integrated background task processing using **Asynq**, managed via Uber FX lifecycle.
- **Health Checks**: Comprehensive health check endpoint monitoring both Database and Redis connectivity.
- **Security**: Pre-configured secure middleware (HSTS, CSP, etc.) and robust RFC 7807 error responses.
- **Developer Experience**: Includes a `Makefile` for common tasks, `golangci-lint` configuration, and comprehensive unit/E2E testing setups.

## Architecture & Project Structure

This project follows a layered architecture pattern, organized into the following directory structure:

```text
.
├── main.go            # CLI entrypoint
├── cmd/               # Command-line entry points (serve, serve-all, queue, migrate)
├── db/
│   └── migrations/    # Database migration files
├── internal/
│   ├── config/        # Application configuration and environment mapping
│   ├── delivery/
│   │   ├── http/      # HTTP handlers, routes, middleware, and server
│   │   └── queue/     # Queue client, worker, middleware, and handlers
│   ├── domain/        # Business entities and domain errors
│   ├── hash/          # Password hashing and JWT management
│   ├── infra/         # Infrastructure clients (PostgreSQL, Redis, SMTP)
│   ├── lang/          # Localization files
│   ├── logger/        # Structured logging setup
│   ├── model/         # Database models
│   ├── repository/    # Data persistence logic
│   ├── security/      # Auth guards and rate limiting
│   ├── service/       # Business logic
│   └── telemetry/     # OpenTelemetry setup
├── pkg/               # Shared utility packages
└── test/
    └── e2e/           # End-to-end tests
```

## Auth Security

- Access and refresh tokens are stored in Redis per user session.
- Refresh token is rotated and validated against Redis.
- Auth middleware validates access token against Redis active session.
- Password reset revokes active session (access + refresh tokens).
- Redis-backed rate limiting:
  - Login: IP limit + email lockout on repeated failures.
  - Refresh token: IP limit.
- Limited requests return `429` and `Retry-After` header.

## Auth Endpoints

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh-token`
- `POST /api/v1/auth/forgot-password/send-otp`
- `POST /api/v1/auth/forgot-password/verify-otp`
- `POST /api/v1/auth/forgot-password/reset-password`

## Common Commands

```bash
make tidy
make fmt
make lint
make test
make test-e2e
make coverage
make coverage-all
make run
make migrate-up
make migrate-down
make build
make docker-build
make docker-run
```

## Migrations

```bash
go run . migrate status
go run . migrate up
go run . migrate down
```

## Docker

```bash
docker compose up --build
```

Services started by compose:

- App: `http://localhost:8080`
- OpenAPI docs: `http://localhost:8080/docs`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- Jaeger UI: `http://localhost:16686`

## Observability (OTel)

Telemetry env vars are available in `.env.example`:

- `OTEL_ENABLED`
- `OTEL_SERVICE_NAME`
- `OTEL_EXPORTER` (`otlp` or `none`)
- `OTEL_EXPORTER_OTLP_ENDPOINT` (default: `jaeger:4317`)
- `OTEL_EXPORTER_OTLP_INSECURE`
- `OTEL_TRACES_SAMPLER_RATIO`
- `OTEL_EXPORT_TIMEOUT`
