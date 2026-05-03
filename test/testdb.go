package test

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/akfaiz/go-starter-kit/db/migrations" // Import migrations
	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/migris"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver for pgx
	"github.com/onsi/ginkgo/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// DBContainer holds PostgreSQL container and Bun DB connection.
type DBContainer struct {
	Container *postgres.PostgresContainer
	DB        *bun.DB
	Config    config.Database
}

// NewDBContainer creates and initializes a PostgreSQL testcontainer with migrations.
func NewDBContainer(ctx context.Context, t ginkgo.FullGinkgoTInterface) *DBContainer {
	t.Helper()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get postgres host: %v", err)
	}

	port, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("failed to get postgres port: %v", err)
	}

	portNum, err := strconv.Atoi(port.Port())
	if err != nil {
		t.Fatalf("failed to parse postgres port: %v", err)
	}

	dbConfig := config.Database{
		Host:     host,
		Port:     portNum,
		User:     "postgres",
		Password: "postgres",
		Name:     "test_db",
		SSLMode:  "disable",
	}

	// Open SQL connection for migrations
	sqlDB, err := sql.Open("pgx", dbConfig.DSN())
	if err != nil {
		t.Fatalf("failed to open sql connection: %v", err)
	}
	defer sqlDB.Close()

	// Run migrations
	migrator, err := migris.New("pgx",
		migris.WithDB(sqlDB),
		migris.WithMigrationDir("db/migrations"),
	)
	if err != nil {
		t.Fatalf("failed to create migrator: %v", err)
	}

	if err := migrator.UpContext(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create Bun DB connection
	sqlDB, err = sql.Open("pgx", dbConfig.DSN())
	if err != nil {
		t.Fatalf("failed to open bun connection: %v", err)
	}

	db := bun.NewDB(sqlDB, pgdialect.New())

	return &DBContainer{
		Container: pgContainer,
		DB:        db,
		Config:    dbConfig,
	}
}

// TruncateAll truncates all tables in the database to reset state between tests.
// It checks for table existence before truncating to avoid errors when some migrations
// haven't created all tables.
func (tc *DBContainer) TruncateAll(ctx context.Context) error {
	tables := []string{
		"password_reset_tokens",
		"users",
	}

	for _, table := range tables {
		// Check if table exists using to_regclass
		var reg sql.NullString
		q := fmt.Sprintf("SELECT to_regclass('public.%s')", table)
		if err := tc.DB.QueryRowContext(ctx, q).Scan(&reg); err != nil {
			return err
		}
		if reg.Valid {
			if _, err := tc.DB.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Close closes the database connection and terminates the container.
func (tc *DBContainer) Close(_ context.Context) error {
	if tc.DB != nil {
		if err := tc.DB.Close(); err != nil {
			return err
		}
	}
	if tc.Container != nil {
		return testcontainers.TerminateContainer(tc.Container)
	}
	return nil
}
