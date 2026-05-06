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
| Monitoring | Prometheus + Mimir + Loki + Grafana |
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
| Jaeger UI | `http://localhost:16686` |

## Observability Suite

The kit includes a pre-configured monitoring stack in Grafana:

- **[Overview](http://localhost:3000/d/go-api-starter-kit)**: High-level health, QPS, and P95 latency.
- **[Error Analytics](http://localhost:3000/d/error-log-analytics)**: Real-time 4xx/5xx tracking and log patterns.
- **[Go Runtime](http://localhost:3000/d/go-runtime-deep-dive)**: GC duration, heap allocation, and goroutine tracking.
- **[Database connection](http://localhost:3000/d/db-connection-analytics)**: GORM/sql.DB pool utilization and wait times.
- **[Queue Worker](http://localhost:3000/d/queue-worker-analytics)**: Asynq task throughput, latency, and queue sizes.

Observability configs and Grafana provisioning are stored under `docker/`. Prometheus scrapes local metrics and remote-writes them to Mimir for long-term query/storage in Grafana.

Queue worker metrics are consolidated and exported via the main HTTP port (default `8080`) using `github.com/hibiken/asynq/x/metrics`, including queue-level metrics such as:
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
|-- docker/            # Loki, Mimir, Prometheus, Promtail, Tempo, Grafana provisioning
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
