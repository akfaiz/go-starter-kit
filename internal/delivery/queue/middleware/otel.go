package middleware

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("asynq-worker")

// Otel is an asynq middleware that adds OpenTelemetry tracing to task processing.
func Otel(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		headers := t.Headers()
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(headers))

		spanName := fmt.Sprintf("asynq: %s", t.Type())
		ctx, span := tracer.Start(ctx, spanName, trace.WithAttributes(
			attribute.String("asynq.task_type", t.Type()),
		))
		defer span.End()

		err := h.ProcessTask(ctx, t)
		if err != nil {
			span.RecordError(err)
		}
		return err
	})
}
