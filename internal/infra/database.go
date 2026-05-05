package infra

import (
	"context"
	"log/slog"

	"github.com/akfaiz/go-starter-kit/internal/config"
	cerrors "github.com/cockroachdb/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

func NewDatabase(cfg config.Config) (*gorm.DB, error) {
	logger := logger.NewSlogLogger(slog.Default(), logger.Config{
		SlowThreshold:             200 * 1e6, // 200ms
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
	gormCfg := &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger,
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), gormCfg)
	if err != nil {
		return nil, cerrors.WithStack(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, cerrors.WithStack(err)
	}

	if err := sqlDB.PingContext(context.Background()); err != nil {
		return nil, cerrors.WithStack(err)
	}

	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, cerrors.WithStack(err)
	}

	return db, nil
}
