package queue

import (
	"context"
	"fmt"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	client *asynq.Client
}

func NewClient(opt config.Redis) *Client {
	return &Client{
		client: asynq.NewClient(asynq.RedisClientOpt{
			Addr:     opt.Addr,
			Password: opt.Password,
			DB:       opt.DB,
		}),
	}
}

func (c *Client) EnqueueContext(ctx context.Context, t *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	tracer := otel.Tracer("asynq-client")
	ctx, span := tracer.Start(
		ctx,
		fmt.Sprintf("%s publish", t.Type()),
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystemKey.String("asynq"),
			semconv.MessagingDestinationNameKey.String(t.Type()),
			semconv.MessagingOperationTypePublish,
		),
	)
	defer span.End()

	info, err := c.client.EnqueueContext(ctx, t, opts...)
	if err != nil {
		telemetry.RecordSpanError(span, err)
	}
	return info, err
}

func (c *Client) Enqueue(t *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.EnqueueContext(context.Background(), t, opts...)
}

func (c *Client) Close() error {
	return c.client.Close()
}
