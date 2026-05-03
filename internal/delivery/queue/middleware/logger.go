package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Logger is an asynq middleware that logs the execution time and errors of tasks.
func Logger(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		attrs := []any{
			slog.String(string(semconv.MessagingSystemKey), "asynq"),
			slog.String(string(semconv.MessagingDestinationNameKey), t.Type()),
			slog.String(
				string(semconv.MessagingOperationTypeKey),
				semconv.MessagingOperationTypeDeliver.Value.AsString(),
			),
		}

		slog.InfoContext(ctx, "task processing started", attrs...)

		start := time.Now()
		err := h.ProcessTask(ctx, t)

		latency := time.Since(start)
		attrs = append(attrs, slog.Duration("latency", latency))

		if err != nil {
			attrs = append(attrs, slog.Any("error", err))
			slog.ErrorContext(ctx, "task processing failed", attrs...)
			return err
		}

		slog.InfoContext(ctx, "task processed successfully", attrs...)
		return nil
	})
}
