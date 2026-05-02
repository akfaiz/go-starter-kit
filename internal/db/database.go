package db

import (
	"database/sql"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/pkg/bunslog"
	"github.com/akfaiz/go-starter-kit/pkg/env"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func NewDatabase(cfg config.Database) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(cfg.DSN()),
	))
	if err := sqldb.Ping(); err != nil {
		return nil, err
	}
	db := bun.NewDB(sqldb, pgdialect.New())
	if env.GetBool("APP_DEBUG") {
		db.AddQueryHook(bunslog.NewQueryHook(
			bunslog.WithVerbose(true),
		))
	}

	return db, nil
}
