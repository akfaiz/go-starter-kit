---
name: go-telemetry-tracing
description: "OpenTelemetry setup, custom spans, trace propagation, span error recording, and Jaeger/OTLP configuration. Use when adding tracing to services, repositories, queues, Redis, database queries, or external calls."
---

# Telemetry and Tracing

This project uses **OpenTelemetry (OTel)** for distributed tracing. The default exporter is OTLP, and the default endpoint points at `otel-collector:4317` in local compose setups.

## Tracing Architecture
- **Location**: `internal/telemetry/`
- **Exporters**: Configured through `internal/config/telemetry.go` and initialized in `internal/telemetry/tracing.go`.
- **Instrumentations**:
  - `Echo`: Traces incoming HTTP requests via `echo-opentelemetry`.
  - `GORM`: Traces database queries.
  - `Redis`: Traces cache operations via `redisotel`.
  - `Asynq`: Traces queue enqueueing and worker processing.
- **Errors**: `internal/telemetry/error.go` records exception type, message, and stack trace when available.

## Adding Custom Tracing
To add a custom span in a service or repository:

```go
func (s *service) ComplexOperation(ctx context.Context) error {
    ctx, span := telemetry.StartSpan(ctx, s.tracer)
    defer span.End()

    if err := s.repo.Do(ctx); err != nil {
        telemetry.RecordSpanError(span, err)
        return err
    }
    return nil
}
```

Existing services use package-level tracers such as `otel.Tracer("auth-service")` and `otel.Tracer("user-service")`.

## Error Recording

- Attach stack traces at the unexpected error origin with `github.com/cockroachdb/errors`.
- Call `telemetry.RecordSpanError(span, err)` before returning errors from custom spans when useful.
- Do not stack-wrap expected domain errors such as invalid tokens, duplicate email, or not found.

## Viewing Traces
When running locally with Docker Compose, traces are available in Grafana Explore (Tempo datasource) at `http://localhost:3000`.

## Configuration
OTel configuration (enabled flag, exporter, endpoint, insecure mode, sampling ratio, export timeout) is managed in `internal/config/telemetry.go` and initialized in `internal/telemetry/tracing.go`.

## Current Defaults
- `OTEL_ENABLED=true`
- `OTEL_EXPORTER=otlp`
- `OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317`
- `OTEL_EXPORTER_OTLP_INSECURE=true`
- `OTEL_TRACES_SAMPLER_RATIO=1.0`
- `OTEL_EXPORT_TIMEOUT=5s`

## Verification

- Use `tracetest.InMemoryExporter` for span assertions.
- For queue propagation, assert producer/consumer span kinds and propagated trace context.
- Run `make test` after telemetry or error-recording changes.
