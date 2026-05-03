package handler_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var _ = Describe("HealthCheckHandler", Label("unit", "handler"), func() {
	newContext := func() (*echo.Context, *httptest.ResponseRecorder) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		return c, rec
	}

	newDB := func() (*bun.DB, sqlmock.Sqlmock, *sql.DB) {
		sqldb, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		Expect(err).NotTo(HaveOccurred())
		return bun.NewDB(sqldb, pgdialect.New()), mock, sqldb
	}

	It("returns ok when database ping succeeds", func() {
		db, mock, sqldb := newDB()
		defer func() {
			_ = db.Close()
			_ = sqldb.Close()
		}()

		mock.ExpectPing()
		h := handler.NewHealthCheckHandler(db)
		c, rec := newContext()

		err := h.HealthCheck(c)
		Expect(err).NotTo(HaveOccurred())
		Expect(rec.Code).To(Equal(http.StatusOK))

		var got map[string]any
		Expect(json.Unmarshal(rec.Body.Bytes(), &got)).To(Succeed())
		Expect(got["status"]).To(Equal("ok"))
		Expect(got["message"]).To(Equal("Application is healthy"))
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	It("returns error payload when database ping fails", func() {
		db, mock, sqldb := newDB()
		defer func() {
			_ = db.Close()
			_ = sqldb.Close()
		}()

		mock.ExpectPing().WillReturnError(errors.New("db unreachable"))
		h := handler.NewHealthCheckHandler(db)
		c, rec := newContext()

		err := h.HealthCheck(c)
		Expect(err).NotTo(HaveOccurred())
		Expect(rec.Code).To(Equal(http.StatusInternalServerError))

		var got map[string]any
		Expect(json.Unmarshal(rec.Body.Bytes(), &got)).To(Succeed())
		Expect(got["status"]).To(Equal("error"))
		Expect(got["message"]).To(Equal("Database connection error"))
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})
})
