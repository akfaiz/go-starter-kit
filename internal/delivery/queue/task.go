package queue

import (
	"context"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// NewTask creates a new asynq.Task with distributed tracing headers injected from the context.
func NewTask(ctx context.Context, typename string, payload []byte, opts ...asynq.Option) *asynq.Task {
	headers := make(map[string]string)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))
	return asynq.NewTaskWithHeaders(typename, payload, headers, opts...)
}

const (
	TypeMailSend = "mail:send"
)
