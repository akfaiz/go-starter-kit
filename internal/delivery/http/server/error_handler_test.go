package server_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/server"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestCustomHTTPErrorHandler(t *testing.T) {
	e := echo.New()

	t.Run("AppError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := problem.ErrBadRequest().WithDetail("bad things")
		server.CustomHTTPErrorHandler(c, err)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "bad things")
		assert.Equal(t, server.ContentTypeProblemJSON, rec.Header().Get(echo.HeaderContentType))
	})

	t.Run("ValidationError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		vErr := validator.NewErrors(validator.FieldError{Field: "name", Message: "required"})
		server.CustomHTTPErrorHandler(c, vErr)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "Validation failed")
		assert.Contains(t, rec.Body.String(), "name")
	})

	t.Run("EchoHTTPError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := echo.NewHTTPError(http.StatusNotFound, "not found buddy")
		server.CustomHTTPErrorHandler(c, err)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "not found buddy")
	})

	t.Run("InternalServerError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := errors.New("something went boom")
		server.CustomHTTPErrorHandler(c, err)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Internal Server Error")
	})

	t.Run("RequestIDHeader", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		rec.Header().Set(echo.HeaderXRequestID, "req-123")
		c := e.NewContext(req, rec)

		err := problem.ErrForbidden()
		server.CustomHTTPErrorHandler(c, err)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "req-123")
	})

	t.Run("CommittedResponse", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Response().WriteHeader(http.StatusOK) // This commits the response

		err := problem.ErrBadRequest()
		server.CustomHTTPErrorHandler(c, err)

		// Response was already committed, so error handler should return without writing more
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
