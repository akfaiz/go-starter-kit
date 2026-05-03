package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
)

// Logger is an asynq middleware that logs the execution time and errors of tasks.
func Logger(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()

		err := h.ProcessTask(ctx, t)

		latency := time.Since(start)
		attrs := []any{
			slog.String("task_type", t.Type()),
			slog.Float64("latency_ms", float64(latency.Microseconds())/1000),
		}

		if err != nil {
			attrs = append(attrs, slog.Any("error", err))
			slog.ErrorContext(ctx, "task processing failed", attrs...)
			return err
		}

		slog.InfoContext(ctx, "task processed successfully", attrs...)
		return nil
	})
}
