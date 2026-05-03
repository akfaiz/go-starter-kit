package middleware

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Otel is an asynq middleware that adds OpenTelemetry tracing to task processing.
func Otel(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		headers := t.Headers()
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(headers))

		spanName := fmt.Sprintf("%s process", t.Type())
		tracer := otel.Tracer("asynq-worker")
		ctx, span := tracer.Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				semconv.MessagingSystemKey.String("asynq"),
				semconv.MessagingDestinationNameKey.String(t.Type()),
				semconv.MessagingOperationTypeDeliver,
			),
		)
		defer span.End()

		err := h.ProcessTask(ctx, t)
		if err != nil {
			span.RecordError(err)
		}
		return err
	})
}
