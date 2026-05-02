package serve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/db"
	deliveryhttp "github.com/akfaiz/go-starter-kit/internal/delivery/http"
	"github.com/akfaiz/go-starter-kit/internal/hash"
	"github.com/akfaiz/go-starter-kit/internal/lang"
	"github.com/akfaiz/go-starter-kit/internal/logger"
	"github.com/akfaiz/go-starter-kit/internal/provider"
	"github.com/akfaiz/go-starter-kit/internal/repository"
	"github.com/akfaiz/go-starter-kit/internal/security"
	"github.com/akfaiz/go-starter-kit/internal/service"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"github.com/labstack/echo/v5"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var Command = &cli.Command{
	Name:  "serve",
	Usage: "Start the API server",
	Action: func(ctx context.Context, c *cli.Command) error {
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
		fx.Supply(cfg, cfg.Auth, cfg.Auth.JWT, cfg.Database),
		fx.Provide(db.NewDatabase, db.NewRedisClient),
		repository.Module,
		hash.Module,
		provider.Module,
		security.Module,
		service.Module,
		telemetry.Module,
		deliveryhttp.Module,
		fx.Invoke(httpServerLifecycle),
	}
}

func httpServerLifecycle(lc fx.Lifecycle, e *echo.Echo, cfg config.Config) {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: e}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					slog.Error("http server stopped unexpectedly", "error", err)
				}
			}()
			slog.Info("server started", "url", fmt.Sprintf("http://localhost:%d", cfg.Server.Port))
			slog.Info("openapi docs", "url", fmt.Sprintf("http://localhost:%d/docs", cfg.Server.Port))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
