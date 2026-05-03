package test

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/akfaiz/go-starter-kit/db/migrations" // Import migrations
	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/migris"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver for pgx
	"github.com/onsi/ginkgo/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBContainer holds PostgreSQL container and GORM DB connection.
type DBContainer struct {
	Container *postgres.PostgresContainer
	DB        *gorm.DB
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

	// Create GORM DB connection
	db, err := gorm.Open(gormpostgres.Open(dbConfig.DSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm connection: %v", err)
	}

	return &DBContainer{
		Container: pgContainer,
		DB:        db,
		Config:    dbConfig,
	}
}

// TruncateAll truncates all tables in the database to reset state between tests.
// It dynamically fetches all tables in the public schema except the migration metadata.
func (tc *DBContainer) TruncateAll(ctx context.Context) error {
	var tables []string
	query := `
		SELECT tablename 
		FROM pg_catalog.pg_tables 
		WHERE schemaname = 'public' 
		AND tablename != 'migris_migrations'`

	if err := tc.DB.WithContext(ctx).Raw(query).Scan(&tables).Error; err != nil {
		return err
	}

	if len(tables) == 0 {
		return nil
	}

	// Truncate all tables in a single command for better performance
	truncateQuery := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join(tables, ", "))
	return tc.DB.WithContext(ctx).Exec(truncateQuery).Error
}

// Close closes the database connection and terminates the container.
func (tc *DBContainer) Close(_ context.Context) error {
	if tc.DB != nil {
		sqlDB, err := tc.DB.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			return err
		}
	}
	if tc.Container != nil {
		return testcontainers.TerminateContainer(tc.Container)
	}
	return nil
}
