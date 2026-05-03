package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/hibiken/asynq"
	"go.uber.org/fx"
)

func NewServer(cfg config.Config) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
		asynq.Config{
			Concurrency:    cfg.Queue.Concurrency,
			StrictPriority: cfg.Queue.Strict,
			Logger:         &asynqLogger{l: slog.Default()},
		},
	)
}

func RegisterServer(lc fx.Lifecycle, srv *asynq.Server, mux *asynq.ServeMux) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.Run(mux); err != nil {
					slog.ErrorContext(ctx, "asynq server run failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			srv.Shutdown()
			return nil
		},
	})
}

var _ asynq.Logger = (*asynqLogger)(nil)

type asynqLogger struct {
	l *slog.Logger
}

func (l *asynqLogger) log(level func(string, ...any), args ...any) {
	if len(args) == 0 {
		level("")
		return
	}

	msg, ok := args[0].(string)
	if !ok {
		// fallback: stringify everything
		level(fmt.Sprint(args...))
		return
	}

	level(msg, args[1:]...)
}

func (l *asynqLogger) Debug(args ...any) {
	l.log(l.l.Debug, args...)
}

func (l *asynqLogger) Info(args ...any) {
	l.log(l.l.Info, args...)
}

func (l *asynqLogger) Warn(args ...any) {
	l.log(l.l.Warn, args...)
}

func (l *asynqLogger) Error(args ...any) {
	l.log(l.l.Error, args...)
}

func (l *asynqLogger) Fatal(args ...any) {
	// slog has no Fatal → map to Error
	l.log(l.l.Error, args...)
}
