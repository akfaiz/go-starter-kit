---
name: go-telemetry-tracing
description: OpenTelemetry, Jaeger, and tracing integration for observability. Use when adding tracing to new services, database queries, or external calls.
---

# Telemetry and Tracing

This project uses **OpenTelemetry (OTel)** for distributed tracing, with **Jaeger** for local visualization.

## Tracing Architecture
- **Location**: `internal/telemetry/`
- **Exporters**: Configured to send traces to a collector (e.g., Jaeger) over OTLP.
- **Instrumentations**:
  - `Echo`: Traces incoming HTTP requests.
  - `Bun`: Traces database queries.
  - `Redis`: Traces cache operations.

## Adding Custom Tracing
To add a custom span in a service or repository:

```go
func (s *service) ComplexOperation(ctx context.Context) error {
    ctx, span := telemetry.StartSpan(ctx, s.tracer)
    defer span.End()

    // span.SetAttributes(...)
    // ...
    return nil
}
```

## Viewing Traces
When running locally with Docker Compose, Jaeger is usually available at `http://localhost:16686`.

## Configuration
OTel configuration (endpoint, service name, sampling) is managed in `internal/config/telemetry.go` and initialized in `internal/telemetry/tracing.go`.
