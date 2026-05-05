package server_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/server"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	cerrors "github.com/cockroachdb/errors"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func TestCustomHTTPErrorHandler(t *testing.T) {
	e := echo.New()

	t.Run("ProblemError", func(t *testing.T) {
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
		prevTracerProvider := otel.GetTracerProvider()
		exporter := tracetest.NewInMemoryExporter()
		tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
		otel.SetTracerProvider(tp)
		t.Cleanup(func() {
			_ = tp.Shutdown(context.Background())
			otel.SetTracerProvider(prevTracerProvider)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		ctx, span := otel.Tracer("test").Start(req.Context(), "request")
		c.SetRequest(req.WithContext(ctx))

		err := cerrors.WithStack(errors.New("something went boom"))
		server.CustomHTTPErrorHandler(c, err)
		span.End()

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Internal Server Error")

		spans := exporter.GetSpans()
		require.Len(t, spans, 1)
		assert.Contains(t, spans[0].Attributes, semconv.ExceptionMessageKey.String("something went boom"))
		assert.NotEmpty(t, findAttr(t, spans[0].Attributes, semconv.ExceptionStacktraceKey).Value.AsString())
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

func findAttr(t *testing.T, attrs []attribute.KeyValue, key attribute.Key) attribute.KeyValue {
	t.Helper()
	for _, attr := range attrs {
		if attr.Key == key {
			return attr
		}
	}
	t.Fatalf("attribute %s not found", key)
	return attribute.KeyValue{}
}
