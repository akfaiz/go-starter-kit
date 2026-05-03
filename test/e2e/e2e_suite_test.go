package e2e_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	_ "github.com/akfaiz/go-starter-kit/db/migrations"
	"github.com/akfaiz/go-starter-kit/internal/config"
	deliveryhttp "github.com/akfaiz/go-starter-kit/internal/delivery/http"
	"github.com/akfaiz/go-starter-kit/internal/hash"
	"github.com/akfaiz/go-starter-kit/internal/infra"
	"github.com/akfaiz/go-starter-kit/internal/lang"
	"github.com/akfaiz/go-starter-kit/internal/repository"
	"github.com/akfaiz/go-starter-kit/internal/security"
	"github.com/akfaiz/go-starter-kit/internal/service"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"github.com/akfaiz/migris"
	"github.com/gavv/httpexpect/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

var (
	e2eCtx    context.Context
	e2ePG     *postgres.PostgresContainer
	e2eRedisC *rediscontainer.RedisContainer
	e2eDB     *bun.DB
	e2eRDB    *redis.Client
	e2eEcho   *echo.Echo
	e2eFXApp  *fx.App
	e2eServer *httptest.Server
	e2eExpect *httpexpect.Expect
)

var _ = BeforeSuite(func() {
	e2eCtx = context.Background()
	lang.Init()

	var err error
	e2ePG, err = postgres.Run(e2eCtx,
		"postgres:16-alpine",
		postgres.WithDatabase("e2e_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	Expect(err).NotTo(HaveOccurred())

	e2eRedisC, err = rediscontainer.Run(e2eCtx, "redis:7-alpine")
	Expect(err).NotTo(HaveOccurred())

	pgHost, err := e2ePG.Host(e2eCtx)
	Expect(err).NotTo(HaveOccurred())
	pgPort, err := e2ePG.MappedPort(e2eCtx, "5432/tcp")
	Expect(err).NotTo(HaveOccurred())
	pgPortNum, err := strconv.Atoi(pgPort.Port())
	Expect(err).NotTo(HaveOccurred())

	redisHost, err := e2eRedisC.Host(e2eCtx)
	Expect(err).NotTo(HaveOccurred())
	redisPort, err := e2eRedisC.MappedPort(e2eCtx, "6379/tcp")
	Expect(err).NotTo(HaveOccurred())

	cfg := config.Config{
		App: config.App{
			Name:            "go-starter-kit-e2e",
			FrontendBaseURL: "http://localhost:3000",
			LogLevel:        "error",
			LogFormat:       "json",
		},
		Auth: config.Auth{
			ResetPasswordExpiration: 60 * time.Minute,
			VerificationExpiration:  60 * time.Minute,
			JWT: config.JWT{
				AccessSecret:   "e2e-access-secret",
				RefreshSecret:  "e2e-refresh-secret",
				AccessExpires:  15 * time.Minute,
				RefreshExpires: 24 * time.Hour,
			},
		},
		Database: config.Database{
			Host:     pgHost,
			Port:     pgPortNum,
			User:     "postgres",
			Password: "postgres",
			Name:     "e2e_db",
			SSLMode:  "disable",
		},
		Mail: config.Mail{
			SMTP: config.MailSMTP{
				Host:     "localhost",
				Port:     2525,
				Username: "",
				Password: "",
				TLSMode:  "none",
			},
			From: config.MailFrom{Address: "noreply@example.com", Name: "E2E"},
		},
		RateLimit: config.RateLimit{
			LoginAttempts:         5,
			LoginWindow:           10 * time.Minute,
			LoginLockoutThreshold: 5,
			LoginLockoutDuration:  15 * time.Minute,
			RefreshAttemptsPerIP:  20,
			RefreshWindow:         10 * time.Minute,
		},
		Redis: config.Redis{
			Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort.Port()),
			Password: "",
			DB:       0,
			Prefix:   "e2e",
		},
		Server: config.Server{Port: 8080, CORSOrigins: []string{"http://localhost:3000"}},
		Telemetry: config.Telemetry{
			Enabled:       false,
			ServiceName:   "go-starter-kit-e2e",
			Exporter:      "none",
			Endpoint:      "localhost:4317",
			Insecure:      true,
			SampleRatio:   1,
			ExportTimeout: 5 * time.Second,
		},
	}

	e2eFXApp = fx.New(
		fx.Supply(cfg, cfg.Auth, cfg.Auth.JWT, cfg.Database),
		infra.Module,
		repository.Module,
		hash.Module,
		security.Module,
		service.Module,
		telemetry.Module,
		deliveryhttp.Module,
		fx.Populate(&e2eEcho, &e2eDB, &e2eRDB),
	)
	Expect(e2eFXApp.Err()).NotTo(HaveOccurred())

	sqlDB, err := sql.Open("pgx", cfg.Database.DSN())
	Expect(err).NotTo(HaveOccurred())
	defer sqlDB.Close()
	migrator, err := migris.New("pgx",
		migris.WithDB(sqlDB),
		migris.WithMigrationDir("../../db/migrations"),
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(migrator.UpContext(e2eCtx)).NotTo(HaveOccurred())

	startCtx, cancel := context.WithTimeout(e2eCtx, 20*time.Second)
	defer cancel()
	Expect(e2eFXApp.Start(startCtx)).NotTo(HaveOccurred())

	e2eServer = httptest.NewServer(e2eEcho)
	e2eExpect = httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  e2eServer.URL,
		Client:   e2eServer.Client(),
		Reporter: httpexpect.NewRequireReporter(GinkgoT()),
	})
})

var _ = AfterSuite(func() {
	if e2eServer != nil {
		e2eServer.Close()
	}
	if e2eFXApp != nil {
		stopCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_ = e2eFXApp.Stop(stopCtx)
	}
	if e2eRedisC != nil {
		_ = testcontainers.TerminateContainer(e2eRedisC)
	}
	if e2ePG != nil {
		_ = testcontainers.TerminateContainer(e2ePG)
	}
})
