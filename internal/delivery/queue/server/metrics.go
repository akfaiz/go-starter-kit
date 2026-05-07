package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
)

const asynqMetricsMeterName = "asynq-metrics"

type QueueMetricsCollector struct {
	enabled bool

	inspector *asynq.Inspector
	meter     metric.Meter

	tasksEnqueuedTotal metric.Int64ObservableGauge
	queueSize          metric.Int64ObservableGauge
	queueLatency       metric.Float64ObservableGauge
	queueMemUsage      metric.Int64ObservableGauge
	tasksProcessed     metric.Int64ObservableCounter
	tasksFailed        metric.Int64ObservableCounter
	queuePaused        metric.Int64ObservableGauge

	registration metric.Registration
}

func NewQueueMetricsCollector(cfg config.Config, mp metric.MeterProvider) (*QueueMetricsCollector, error) {
	if !cfg.Queue.MetricsEnabled {
		return &QueueMetricsCollector{enabled: false}, nil
	}

	meter := mp.Meter(asynqMetricsMeterName)

	tasksEnqueuedTotal, err := meter.Int64ObservableGauge(
		"asynq_tasks_enqueued_total",
		metric.WithDescription("Number of tasks enqueued; broken down by queue and state."),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric asynq_tasks_enqueued_total: %w", err)
	}

	queueSize, err := meter.Int64ObservableGauge(
		"asynq_queue_size",
		metric.WithDescription("Number of tasks in a queue."),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric asynq_queue_size: %w", err)
	}

	queueLatency, err := meter.Float64ObservableGauge(
		"asynq_queue_latency_seconds",
		metric.WithDescription("Number of seconds the oldest pending task is waiting."),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric asynq_queue_latency_seconds: %w", err)
	}

	queueMemUsage, err := meter.Int64ObservableGauge(
		"asynq_queue_memory_usage_approx_bytes",
		metric.WithDescription("Approximate memory used by a queue."),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric asynq_queue_memory_usage_approx_bytes: %w", err)
	}

	tasksProcessed, err := meter.Int64ObservableCounter(
		"asynq_tasks_processed_total",
		metric.WithDescription("Number of processed tasks by queue."),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric asynq_tasks_processed_total: %w", err)
	}

	tasksFailed, err := meter.Int64ObservableCounter(
		"asynq_tasks_failed_total",
		metric.WithDescription("Number of failed tasks by queue."),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric asynq_tasks_failed_total: %w", err)
	}

	queuePaused, err := meter.Int64ObservableGauge(
		"asynq_queue_paused_total",
		metric.WithDescription("Queue pause state (1 paused, 0 not paused)."),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric asynq_queue_paused_total: %w", err)
	}

	return &QueueMetricsCollector{
		enabled: true,
		inspector: asynq.NewInspector(asynq.RedisClientOpt{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}),
		meter:              meter,
		tasksEnqueuedTotal: tasksEnqueuedTotal,
		queueSize:          queueSize,
		queueLatency:       queueLatency,
		queueMemUsage:      queueMemUsage,
		tasksProcessed:     tasksProcessed,
		tasksFailed:        tasksFailed,
		queuePaused:        queuePaused,
	}, nil
}

func RegisterQueueMetricsCollector(lc fx.Lifecycle, c *QueueMetricsCollector) {
	if c == nil || !c.enabled {
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			registration, err := c.meter.RegisterCallback(
				c.observeQueueMetrics,
				c.tasksEnqueuedTotal,
				c.queueSize,
				c.queueLatency,
				c.queueMemUsage,
				c.tasksProcessed,
				c.tasksFailed,
				c.queuePaused,
			)
			if err != nil {
				return fmt.Errorf("register queue metrics callback: %w", err)
			}
			c.registration = registration
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if c.registration != nil {
				if err := c.registration.Unregister(); err != nil {
					slog.WarnContext(ctx, "unregister queue metrics callback failed", "error", err)
				}
			}
			if err := c.inspector.Close(); err != nil {
				return fmt.Errorf("close asynq inspector: %w", err)
			}
			return nil
		},
	})
}

func (c *QueueMetricsCollector) observeQueueMetrics(ctx context.Context, obs metric.Observer) error {
	queueInfos, err := c.collectQueueInfo()
	if err != nil {
		slog.WarnContext(ctx, "collect queue metrics failed", "error", err)
		return nil
	}

	for _, info := range queueInfos {
		queueAttr := metric.WithAttributes(attribute.String("queue", info.Queue))
		activeAttr := metric.WithAttributes(
			attribute.String("queue", info.Queue),
			attribute.String("state", "active"),
		)
		pendingAttr := metric.WithAttributes(
			attribute.String("queue", info.Queue),
			attribute.String("state", "pending"),
		)
		scheduledAttr := metric.WithAttributes(
			attribute.String("queue", info.Queue),
			attribute.String("state", "scheduled"),
		)
		retryAttr := metric.WithAttributes(
			attribute.String("queue", info.Queue),
			attribute.String("state", "retry"),
		)
		archivedAttr := metric.WithAttributes(
			attribute.String("queue", info.Queue),
			attribute.String("state", "archived"),
		)
		completedAttr := metric.WithAttributes(
			attribute.String("queue", info.Queue),
			attribute.String("state", "completed"),
		)

		obs.ObserveInt64(c.tasksEnqueuedTotal, int64(info.Active), activeAttr)
		obs.ObserveInt64(c.tasksEnqueuedTotal, int64(info.Pending), pendingAttr)
		obs.ObserveInt64(c.tasksEnqueuedTotal, int64(info.Scheduled), scheduledAttr)
		obs.ObserveInt64(c.tasksEnqueuedTotal, int64(info.Retry), retryAttr)
		obs.ObserveInt64(c.tasksEnqueuedTotal, int64(info.Archived), archivedAttr)
		obs.ObserveInt64(c.tasksEnqueuedTotal, int64(info.Completed), completedAttr)

		obs.ObserveInt64(c.queueSize, int64(info.Size), queueAttr)
		obs.ObserveFloat64(c.queueLatency, info.Latency.Seconds(), queueAttr)
		obs.ObserveInt64(c.queueMemUsage, info.MemoryUsage, queueAttr)
		obs.ObserveInt64(c.tasksProcessed, int64(info.ProcessedTotal), queueAttr)
		obs.ObserveInt64(c.tasksFailed, int64(info.FailedTotal), queueAttr)

		paused := int64(0)
		if info.Paused {
			paused = 1
		}
		obs.ObserveInt64(c.queuePaused, paused, queueAttr)
	}

	return nil
}

func (c *QueueMetricsCollector) collectQueueInfo() ([]*asynq.QueueInfo, error) {
	queueNames, err := c.inspector.Queues()
	if err != nil {
		return nil, fmt.Errorf("get queue names: %w", err)
	}

	infos := make([]*asynq.QueueInfo, len(queueNames))
	for i, queueName := range queueNames {
		info, err := c.inspector.GetQueueInfo(queueName)
		if err != nil {
			return nil, fmt.Errorf("get queue info %s: %w", queueName, err)
		}
		infos[i] = info
	}

	return infos, nil
}
