package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/hibiken/asynq"
	asynqxmetrics "github.com/hibiken/asynq/x/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

func RegisterQueueMetrics(lc fx.Lifecycle, cfg config.Config) {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	collector := asynqxmetrics.NewQueueMetricsCollector(inspector)
	if err := prometheus.Register(collector); err != nil {
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredError) {
			slog.Error("failed to register asynq queue metrics collector", "error", err)
		}
	}

	if cfg.Queue.MetricsEnabled {
		addr := fmt.Sprintf(":%d", cfg.Queue.MetricsPort)
		srv := &http.Server{
			Addr:              addr,
			Handler:           promhttp.Handler(),
			ReadHeaderTimeout: 10 * time.Second,
		}

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
						slog.ErrorContext(ctx, "queue metrics server stopped unexpectedly", "error", err)
					}
				}()
				slog.InfoContext(ctx, "queue metrics server started", "port", cfg.Queue.MetricsPort)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return srv.Shutdown(ctx)
			},
		})
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return inspector.Close()
		},
	})
}
