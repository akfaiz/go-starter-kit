# Go API Starter Kit

API-only starter kit built with Go.

## Stack

- Echo v5
- OpenAPI router: `github.com/oaswrap/spec/adapter/echov5openapi`
- GORM + PostgreSQL
- Redis (auth rate limit + session store)
- Migris migrations
- JWT auth (access + refresh)
- Token pair session management in Redis (access + refresh)
- Forgot password with OTP via email
- OpenTelemetry tracing (HTTP, GORM DB, Redis)
- Jaeger for local trace visualization
- go-mailgen for email content
- Uber FX for dependency injection

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

## Key Features

- **Automatic Request Validation**: Leverage a custom binder that automatically validates request DTOs during `c.Bind`, reducing handler boilerplate.
- **RFC 7807 Error Responses**: Standardized, machine-readable error responses using the `application/problem+json` content type.
- **Localized Validation**: Built-in support for multiple locales (English and Indonesian) in validation error messages.
- **Dependency Injection**: Robust and modular component management using **Uber FX**.
- **Observability**: Distributed tracing integrated at every level (HTTP, Database, Redis) using **OpenTelemetry**.
- **Developer Experience**: Includes a `Makefile` for common tasks, `golangci-lint` configuration, and comprehensive unit/E2E testing setups.

## Architecture & Project Structure

This project follows a layered architecture pattern, organized into the following directory structure:

- `cmd/`: Command-line entry points (`serve`, `migrate`).
- `internal/`: Core application logic, separated by concerns:
    - `config/`: Application configuration and environment mapping.
    - `delivery/http/`: HTTP transport layer (handlers, routes, middleware).
    - `domain/`: Business entities and repository/service interfaces.
    - `service/`: Implementation of business logic.
    - `repository/`: Data persistence logic (GORM).
    - `infra/`: Infrastructure clients (PostgreSQL, Redis, SMTP).
    - `hash/`: Password hashing and JWT management.
    - `telemetry/`: Observability and tracing setup.
- `db/migrations/`: SQL/Go database migration files.
- `pkg/`: Shared utility packages (validation, error handling, environment helpers).
- `test/`: Integration and E2E tests, along with mocks.

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
