package queue

import (
	"context"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

var tracer = otel.Tracer("asynq-client")

func (c *Client) EnqueueContext(ctx context.Context, t *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx, span := tracer.Start(ctx, "asynq:enqueue", trace.WithAttributes(
		attribute.String("asynq.task_type", t.Type()),
	))
	defer span.End()

	return c.client.EnqueueContext(ctx, t, opts...)
}

func (c *Client) Enqueue(t *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.EnqueueContext(context.Background(), t, opts...)
}

func (c *Client) Close() error {
	return c.client.Close()
}
