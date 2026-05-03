package handler_test

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/alicebob/miniredis/v2"
	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var _ = Describe("HealthCheckHandler", Label("unit", "handler"), func() {
	var (
		h      *handler.HealthCheckHandler
		e      *echo.Echo
		expect *httpexpect.Expect
		mr     *miniredis.Miniredis
		rdb    *redis.Client
	)

	newDB := func() (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
		sqldb, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		Expect(err).NotTo(HaveOccurred())

		// GORM tries to ping the DB on opening
		mock.ExpectPing()

		db, err := gorm.Open(postgres.New(postgres.Config{
			Conn: sqldb,
		}), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		return db, mock, sqldb
	}

	BeforeEach(func() {
		e = setupEcho()
		expect = newExpect(e)

		var err error
		mr, err = miniredis.Run()
		Expect(err).NotTo(HaveOccurred())
		rdb = redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})
	})

	AfterEach(func() {
		mr.Close()
	})

	It("returns ok when database and redis ping succeeds", func() {
		db, mock, sqldb := newDB()
		defer func() {
			_ = sqldb.Close()
		}()

		mock.ExpectPing()
		h = handler.NewHealthCheckHandler(db, rdb)
		e.GET("/health", h.HealthCheck)

		expect.GET("/health").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("status", "ok").
			HasValue("message", "Application is healthy").
			Value("checks").Object().
			HasValue("database", "ok").
			HasValue("redis", "ok")

		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	It("returns error payload when database ping fails", func() {
		db, mock, sqldb := newDB()
		defer func() {
			_ = sqldb.Close()
		}()

		mock.ExpectPing().WillReturnError(errors.New("db unreachable"))
		h = handler.NewHealthCheckHandler(db, rdb)
		e.GET("/health", h.HealthCheck)

		expect.GET("/health").
			Expect().
			Status(http.StatusInternalServerError).
			JSON(httpexpect.ContentOpts{MediaType: "application/problem+json"}).
			Object().
			HasValue("detail", "Database connection error")

		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	It("returns error payload when redis ping fails", func() {
		db, mock, sqldb := newDB()
		defer func() {
			_ = sqldb.Close()
		}()

		mock.ExpectPing()
		mr.Close() // Simulate redis down

		h = handler.NewHealthCheckHandler(db, rdb)
		e.GET("/health", h.HealthCheck)

		expect.GET("/health").
			Expect().
			Status(http.StatusInternalServerError).
			JSON(httpexpect.ContentOpts{MediaType: "application/problem+json"}).
			Object().
			HasValue("detail", "Redis connection error")

		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})
})
