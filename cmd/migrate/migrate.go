package migrate

import (
	"database/sql"
	"log"

	_ "github.com/akfaiz/go-starter-kit/db/migrations"
	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/migris/extra/migriscli"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/urfave/cli/v3"
)

const migrationDir = "db/migrations"

func Command() *cli.Command {
	cfg := config.Load()
	db, err := sql.Open("pgx", cfg.Database.DSN())
	if err != nil {
		log.Fatal(err)
	}

	cliCfg := migriscli.Config{
		MigrationsDir: migrationDir,
		DB:            db,
		Dialect:       "pgx",
	}

	return migriscli.NewCLI(cliCfg)
}
