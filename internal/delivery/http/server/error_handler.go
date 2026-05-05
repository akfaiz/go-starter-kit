package server

import (
	"log/slog"
	"net/http"

	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/trace"
)

const ContentTypeProblemJSON = "application/problem+json"

func CustomHTTPErrorHandler(c *echo.Context, err error) {
	if r, _ := echo.UnwrapResponse(c.Response()); r != nil && r.Committed {
		return
	}
	telemetry.RecordSpanError(trace.SpanFromContext(c.Request().Context()), err)

	res := c.Response()
	res.Header().Set(echo.HeaderContentType, ContentTypeProblemJSON)

	instance := c.Path()
	requestID, ok := res.Header()[echo.HeaderXRequestID]
	if ok && len(requestID) > 0 {
		instance = requestID[0]
	}

	var problemError *problem.Error
	if errors.As(err, &problemError) {
		if jsonErr := c.JSON(problemError.Status, problemError.WithInstance(instance)); jsonErr != nil {
			slog.ErrorContext(c.Request().Context(), "write problem error response failed", "error", jsonErr)
		}
		return
	}

	var validationErr *validator.ValidationError
	if errors.As(err, &validationErr) {
		problemError := problem.ErrValidation().
			WithErrors(validationErr).
			WithCause(err).
			WithInstance(instance)
		if jsonErr := c.JSON(problemError.Status, problemError); jsonErr != nil {
			slog.ErrorContext(c.Request().Context(), "write validation error response failed", "error", jsonErr)
		}
		return
	}

	code := http.StatusInternalServerError
	var statusCoder echo.HTTPStatusCoder
	if errors.As(err, &statusCoder) {
		code = statusCoder.StatusCode()
		message := http.StatusText(code)

		// Extract more detailed message from echo.HTTPError if available
		var he *echo.HTTPError
		if errors.As(err, &he) {
			message = he.Message
		}

		problemError = problem.New(message, "about:blank", code)
	} else {
		problemError = problem.ErrInternalServer()
	}

	if code >= 500 {
		slog.ErrorContext(c.Request().Context(), "http handler error", "error", err, "code", code)
	}

	if jsonErr := c.JSON(code, problemError.WithInstance(instance)); jsonErr != nil {
		slog.ErrorContext(c.Request().Context(), "write generic error response failed", "error", jsonErr)
	}
}
