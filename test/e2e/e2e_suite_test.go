package e2e_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	deliveryhttp "github.com/akfaiz/go-starter-kit/internal/delivery/http"
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue"
	"github.com/akfaiz/go-starter-kit/internal/hash"
	"github.com/akfaiz/go-starter-kit/internal/infra"
	"github.com/akfaiz/go-starter-kit/internal/lang"
	"github.com/akfaiz/go-starter-kit/internal/repository"
	"github.com/akfaiz/go-starter-kit/internal/service"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"github.com/akfaiz/go-starter-kit/test"
	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v5"
	"github.com/oaswrap/spec/adapter/echov5openapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

var (
	e2eCtx              context.Context
	e2eDBContainer      *test.DBContainer
	e2eRedisC           *rediscontainer.RedisContainer
	e2eDB               *gorm.DB
	e2eRDB              *redisclient.Client
	e2eEcho             *echo.Echo
	e2eFXApp            *fx.App
	e2eServer           *httptest.Server
	e2eExpect           *httpexpect.Expect
	e2eOpenAPIGenerator echov5openapi.Generator
)

var _ = BeforeSuite(func() {
	e2eCtx = context.Background()
	lang.Init()

	var err error
	e2eDBContainer = test.NewDBContainer(e2eCtx, GinkgoT())
	Expect(e2eDBContainer).NotTo(BeNil())

	e2eRedisC, err = rediscontainer.Run(e2eCtx, "redis:7-alpine")
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
		Database: e2eDBContainer.Config,
		Hash: config.Hash{
			Driver:            "bcrypt",
			Argon2Memory:      64 * 1024,
			Argon2Iteration:   3,
			Argon2Parallelism: 1,
			BcryptCost:        10,
		},
		Mail: config.Mail{
			Driver: "log",
			SMTP: config.MailSMTP{
				Host:     "localhost",
				Port:     2525,
				Username: "",
				Password: "",
				TLSMode:  "none",
			},
			From: config.MailFrom{Address: "noreply@example.com", Name: "E2E"},
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
		fx.Supply(cfg, cfg.Auth, cfg.Auth.JWT, cfg.Database, cfg.Redis),
		infra.Module,
		repository.Module,
		hash.Module,
		service.Module,
		telemetry.Module,
		deliveryhttp.Module,
		queue.ClientModule,
		fx.Populate(&e2eEcho, &e2eDB, &e2eRDB, &e2eOpenAPIGenerator),
	)
	Expect(e2eFXApp.Err()).NotTo(HaveOccurred())

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
	if e2eDBContainer != nil {
		_ = e2eDBContainer.Close(context.Background())
	}
	if e2eRedisC != nil {
		_ = testcontainers.TerminateContainer(e2eRedisC)
	}
	if e2eOpenAPIGenerator != nil {
		err := e2eOpenAPIGenerator.Validate()
		Expect(err).NotTo(HaveOccurred())
		err = e2eOpenAPIGenerator.WriteSchemaTo("../../docs/openapi.yml")
		Expect(err).NotTo(HaveOccurred())
		err = e2eOpenAPIGenerator.WriteSchemaTo("../../docs/openapi.json")
		Expect(err).NotTo(HaveOccurred())
	}
})
