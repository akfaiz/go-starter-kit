package server

import (
	"context"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestCollectQueueInfo_Error(t *testing.T) {
	c := &QueueMetricsCollector{
		enabled: true,
		inspector: asynq.NewInspector(asynq.RedisClientOpt{
			Addr: "127.0.0.1:1",
		}),
	}
	defer func() { _ = c.inspector.Close() }()

	_, err := c.collectQueueInfo()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestObserveQueueMetrics_CollectErrorReturnsNil(t *testing.T) {
	c := &QueueMetricsCollector{
		enabled: true,
		inspector: asynq.NewInspector(asynq.RedisClientOpt{
			Addr: "127.0.0.1:1",
		}),
	}
	defer func() { _ = c.inspector.Close() }()

	if err := c.observeQueueMetrics(context.Background(), nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCollectQueueInfo_Success(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	_, err = client.Enqueue(asynq.NewTask("mail:send", []byte(`{}`)), asynq.Queue("mail"))
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	collector, err := NewQueueMetricsCollector(
		config.Config{Queue: config.Queue{MetricsEnabled: true}, Redis: config.Redis{Addr: mr.Addr()}},
		noop.NewMeterProvider(),
	)
	if err != nil {
		t.Fatalf("new collector: %v", err)
	}
	defer func() { _ = collector.inspector.Close() }()

	infos, err := collector.collectQueueInfo()
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(infos) == 0 {
		t.Fatal("expected at least one queue info")
	}
}
