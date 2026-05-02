package db

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/pkg/env"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bunotel"
	"github.com/uptrace/bun/extra/bunslog"
)

func NewDatabase(cfg config.Config) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(cfg.Database.DSN()),
	))
	if err := sqldb.Ping(); err != nil {
		return nil, err
	}
	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bunotel.NewQueryHook(
		bunotel.WithDBName(cfg.Database.Name),
		bunotel.WithFormattedQueries(env.GetBool("APP_DEBUG")),
	))

	if env.GetBool("APP_DEBUG") {
		hook := bunslog.NewQueryHook(
			bunslog.WithQueryLogLevel(slog.LevelDebug),
			bunslog.WithSlowQueryLogLevel(slog.LevelWarn),
			bunslog.WithErrorQueryLogLevel(slog.LevelError),
			bunslog.WithSlowQueryThreshold(3*time.Second),
		)
		db.AddQueryHook(hook)
	}

	return db, nil
}
