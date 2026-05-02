package migrate

import (
	"context"
	"database/sql"

	_ "github.com/akfaiz/go-starter-kit/db/migrations"
	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/migris/extra/migriscli"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/urfave/cli/v3"
)

const migrationDir = "db/migrations"

func Command() *cli.Command {
	cfg, err := config.Load()
	if err != nil {
		return &cli.Command{
			Name:  "migrate",
			Usage: "Database migration commands",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return err
			},
		}
	}
	db, err := sql.Open("pgx", cfg.Database.DSN())
	if err != nil {
		return &cli.Command{
			Name:  "migrate",
			Usage: "Database migration commands",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return err
			},
		}
	}

	cliCfg := migriscli.Config{
		MigrationsDir: migrationDir,
		DB:            db,
		Dialect:       "pgx",
	}

	return migriscli.NewCLI(cliCfg)
}
