# Project Overview

This is a **Go API Starter Kit**, an API-only project built with Go.

## Main Technologies
- **Language:** Go 1.25.7
- **Web Framework:** Echo v5
- **API Documentation:** OpenAPI via `github.com/oaswrap/spec/adapter/echov5openapi`
- **Database / ORM:** PostgreSQL with GORM
- **Caching & Session:** Redis (used for session storage)
- **Migrations:** Migris
- **Authentication:** JWT (Access and Refresh tokens) with token pair session management in Redis.
- **Dependency Injection:** Uber FX
- **Observability:** OpenTelemetry (HTTP, GORM DB, Redis tracing) with Jaeger for local visualization.
- **Email:** go-mailgen for email content

## Architecture & Structure
The project follows a standard structured layout common in Go APIs:
- `cmd/`: Command-line entry points (`root.go`, `migrate/`, `serve/`).
- `internal/`: Core application logic, separated by concerns:
  - `config/`: Configuration mapping.
  - `delivery/http/`: Handlers, middleware, routes, and server logic.
  - `domain/`: Domain interfaces and models.
  - `hash/`: Password hashing (Argon2id) and JWT management.
  - `infra/`: Infrastructure integrations (Database, Redis, SMTP).
  - `repository/`: Data access layer for User, Session, and PasswordResetToken.
  - `service/`: Business logic layer.
- `db/migrations/`: Database migration files.
- `test/`: End-to-end tests and mocks.

## Building and Running

Commands are available via the `Makefile` and `go run`:

- **Install Dependencies:** `make tidy` or `go mod tidy`
- **Run Locally:** `make run` or `go run . serve`
- **Run Migrations:** 
  - `make migrate-up` or `go run . migrate up`
  - `make migrate-down` or `go run . migrate down`
- **Build Binary:** `make build`
- **Docker Compose:** `docker compose up --build` or `make docker-run`

## Testing

Testing is primarily driven by Ginkgo and Go test framework:

- **Run Unit Tests:** `make test`
- **Run E2E Tests:** `make test-e2e`
- **Generate Coverage:** `make coverage` or `make coverage-all`
- **View Coverage in HTML:** `make coverage-html`

## Development Conventions

- Uses `golangci-lint` for code quality checks (`make lint`, `make lint-fix`).
- Follows standard Go formatting (`make fmt`).
- Use of Dependency Injection (`go.uber.org/fx`).
- Mocks are managed and likely generated using `go.uber.org/mock`.
- End-to-end testing practices are maintained alongside unit tests.
