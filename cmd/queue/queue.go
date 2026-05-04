package queue

import (
	"context"
	"log/slog"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue"
	"github.com/akfaiz/go-starter-kit/internal/hash"
	"github.com/akfaiz/go-starter-kit/internal/infra"
	"github.com/akfaiz/go-starter-kit/internal/lang"
	"github.com/akfaiz/go-starter-kit/internal/logger"
	"github.com/akfaiz/go-starter-kit/internal/repository"
	"github.com/akfaiz/go-starter-kit/internal/service"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var Command = &cli.Command{
	Name:  "queue",
	Usage: "Start the queue worker",
	Action: func(_ context.Context, _ *cli.Command) error {
		app, err := newApp()
		if err != nil {
			return err
		}
		app.Run()
		return nil
	},
}

func newApp() (*fx.App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	options := appOptions(cfg)
	lang.Init()
	logger.Init(cfg.App)
	if err := fx.ValidateApp(options...); err != nil {
		return nil, err
	}

	return fx.New(options...), nil
}

func appOptions(cfg config.Config) []fx.Option {
	return []fx.Option{
		fx.WithLogger(func() fxevent.Logger {
			slogLogger := &fxevent.SlogLogger{Logger: slog.Default()}
			slogLogger.UseLogLevel(slog.LevelDebug)
			return slogLogger
		}),
		fx.Supply(cfg, cfg.Auth, cfg.Auth.JWT, cfg.Database, cfg.Redis),
		infra.Module,
		repository.Module,
		hash.Module,
		service.Module,
		telemetry.Module,
		queue.ClientModule,
		queue.WorkerModule,
	}
}
