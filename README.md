# Go API Starter Kit

A production-ready, layered Go API starter with authentication, queues, migrations,
localized validation, and full-stack observability.

## Stack

| Layer | Technology |
|---|---|
| HTTP | Echo v5 |
| Database | GORM + PostgreSQL (pgx driver) |
| Cache & Queue | Redis (sessions, Asynq jobs) |
| Auth | JWT (access + refresh tokens), OTP forgot-password flow |
| Monitoring | OpenTelemetry -> Mimir / Prometheus / Grafana |
| Tracing | OpenTelemetry -> Tempo / Jaeger |
| DI | Uber FX |
| Email | go-mailgen |
| Migrations | Migris |

## Prerequisites

- Go 1.25.x
- Docker and Docker Compose for the observability stack
- `ginkgo` for `make test`
- `golangci-lint` for `make lint`

## Quick Start

```bash
cp .env.example .env
docker compose up -d
make tidy
make migrate-up
make run
```

| Service | URL |
|---|---|
| API | `http://localhost:8080` |
| OpenAPI docs | `http://localhost:8080/docs` |
| Grafana | `http://localhost:3000` (admin/admin) |
| Trace UI (Grafana Explore / Tempo) | `http://localhost:3000` (admin/admin) |

## Observability

The kit is fully instrumented with **OpenTelemetry** for metrics, tracing, and logs. Data is exported via OTLP to the observability stack defined in `docker-compose.yml` (Mimir, Tempo, Loki).

Metrics include:
- **HTTP**: Request rate, latency (P95/P99), and success rate via `echo-opentelemetry`.
- **Database**: GORM/sql.DB connection pool utilization and wait times.
- **Queue**: Asynq task throughput, failure rates, and queue sizes.
- **Runtime**: Go runtime metrics including heap allocation, goroutines, and GC pause duration.

OpenTelemetry metrics are exported via OTLP to a collector or compatible backend (e.g., Mimir).

Queue worker metrics are consolidated and exported via OpenTelemetry using `asynq.Inspector`, including queue-level metrics such as:
- `asynq_tasks_enqueued_total{queue,state}`
- `asynq_tasks_processed_total{queue}`
- `asynq_tasks_failed_total{queue}`
- `asynq_queue_size{queue}`

## Performance Tuning

### Database Pooling
Control the `sql.DB` connection pool via `.env`:
```env
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=50
```

### Password Hashing
Choose hashing drivers based on your environment's resource constraints:
- **argon2id**: Maximum security (recommended for production).
- **bcrypt**: High-concurrency friendly (recommended for low-memory environments or heavy registration load).

```env
HASH_DRIVER=argon2id # argon2id | bcrypt
HASH_ARGON2_MEMORY=65536
HASH_BCRYPT_COST=10
```

## Load Testing

The project includes a **k6** script for stress testing the API and validating the observability stack.

```bash
# Run the load test using Docker
docker run --rm -i grafana/k6 run - <scripts/load-test.js
```

## Features

- **Auto-validation** - Custom binder validates request DTOs on `c.Bind`, no handler boilerplate
- **RFC 7807 errors** - Standardized `application/problem+json` responses
- **Localized validation** - English and Indonesian error messages
- **Distributed tracing** - OTel spans across HTTP, GORM, Redis, and queue layers
- **Log Correlation** - Automatic injection of `trace_id` and `span_id` into application logs
- **Background jobs** - Asynq task processing managed via FX lifecycle

## Project Structure

```
.
|-- main.go
|-- cmd/               # CLI commands (serve, serve-all, queue, migrate)
|-- db/migrations/
|-- docker/            # Loki, Mimir, Tempo, and OTEL Collector configurations
|-- scripts/           # Utility scripts (including k6 load test)
|-- pkg/               # Shared packages (env, problem, validator)
|-- internal/
    |-- config/        # Environment-based configuration
    |-- delivery/
    |   |-- http/      # Echo handlers, routes, middleware
    |   `-- queue/     # Asynq worker and handlers
    |-- domain/        # Business entities and repository interfaces
    |-- hash/          # Multidriver password hashing (Argon2id/Bcrypt)
    |-- infra/         # Database (pgx), Redis, and SMTP clients
    |-- lang/          # Translation catalogs
    |-- logger/        # slog setup with OTel correlation
    |-- repository/    # GORM implementations
    |-- service/       # Business logic
    `-- telemetry/     # OpenTelemetry initialization
```

## Make Targets

```bash
make run           # run the HTTP API locally
make test          # run unit/package tests
make test-e2e      # run E2E tests
make coverage      # unit coverage
make lint          # run golangci-lint
make migrate-up    # apply migrations
make docker-build  # build the app image
```
