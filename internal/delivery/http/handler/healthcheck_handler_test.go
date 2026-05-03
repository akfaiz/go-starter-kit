package handler_test

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var _ = Describe("HealthCheckHandler", Label("unit", "handler"), func() {
	var (
		h      *handler.HealthCheckHandler
		e      *echo.Echo
		expect *httpexpect.Expect
	)

	newDB := func() (*bun.DB, sqlmock.Sqlmock, *sql.DB) {
		sqldb, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		Expect(err).NotTo(HaveOccurred())
		return bun.NewDB(sqldb, pgdialect.New()), mock, sqldb
	}

	BeforeEach(func() {
		e = setupEcho()
		expect = newExpect(e)
	})

	It("returns ok when database ping succeeds", func() {
		db, mock, sqldb := newDB()
		defer func() {
			_ = db.Close()
			_ = sqldb.Close()
		}()

		mock.ExpectPing()
		h = handler.NewHealthCheckHandler(db)
		e.GET("/health", h.HealthCheck)

		expect.GET("/health").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("status", "ok").
			HasValue("message", "Application is healthy")

		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	It("returns error payload when database ping fails", func() {
		db, mock, sqldb := newDB()
		defer func() {
			_ = db.Close()
			_ = sqldb.Close()
		}()

		mock.ExpectPing().WillReturnError(errors.New("db unreachable"))
		h = handler.NewHealthCheckHandler(db)
		e.GET("/health", h.HealthCheck)

		expect.GET("/health").
			Expect().
			Status(http.StatusInternalServerError).
			JSON(httpexpect.ContentOpts{MediaType: "application/problem+json"}).
			Object().
			HasValue("detail", "Database connection error")

		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})
})
