package handler_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

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

type pingDriver struct {
	pingErr error
}

type pingConn struct {
	pingErr error
}

func (d *pingDriver) Open(_ string) (driver.Conn, error) {
	return &pingConn{pingErr: d.pingErr}, nil
}

func (c *pingConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (c *pingConn) Close() error { return nil }

func (c *pingConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

func (c *pingConn) Ping(_ context.Context) error {
	return c.pingErr
}

var pingDriverCounter atomic.Uint64

var _ = Describe("HealthCheckHandler", Label("unit", "handler"), func() {
	var (
		h      *handler.HealthCheckHandler
		e      *echo.Echo
		expect *httpexpect.Expect
		mr     *miniredis.Miniredis
		rdb    *redis.Client
	)

	newDB := func(pingErr error) (*gorm.DB, *sql.DB) {
		driverName := fmt.Sprintf("healthcheck-ping-%d", pingDriverCounter.Add(1))
		sql.Register(driverName, &pingDriver{pingErr: pingErr})

		sqldb, err := sql.Open(driverName, "")
		Expect(err).NotTo(HaveOccurred())

		db, err := gorm.Open(postgres.New(postgres.Config{
			Conn: sqldb,
		}), &gorm.Config{
			DisableAutomaticPing: true,
		})
		Expect(err).NotTo(HaveOccurred())
		return db, sqldb
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
		db, sqldb := newDB(nil)
		defer func() {
			_ = sqldb.Close()
		}()

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
	})

	It("returns error payload when database ping fails", func() {
		db, sqldb := newDB(errors.New("db unreachable"))
		defer func() {
			_ = sqldb.Close()
		}()

		h = handler.NewHealthCheckHandler(db, rdb)
		e.GET("/health", h.HealthCheck)

		expect.GET("/health").
			Expect().
			Status(http.StatusInternalServerError).
			JSON(httpexpect.ContentOpts{MediaType: "application/problem+json"}).
			Object().
			HasValue("detail", "Database connection error")
	})

	It("returns ok when redis ping fails", func() {
		db, sqldb := newDB(nil)
		defer func() {
			_ = sqldb.Close()
		}()

		mr.Close() // Simulate redis down

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
			HasValue("redis", "degraded")
	})
})
