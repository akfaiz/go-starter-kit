package server

import (
	"net/http"

	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v5"
)

const ContentTypeProblemJSON = "application/problem+json"

func customHTTPErrorHandler(c *echo.Context, err error) {
	if r, _ := echo.UnwrapResponse(c.Response()); r != nil && r.Committed {
		return
	}

	res := c.Response()
	res.Header().Set(echo.HeaderContentType, ContentTypeProblemJSON)

	instance := c.Path()
	requestID, ok := res.Header()[echo.HeaderXRequestID]
	if ok && len(requestID) > 0 {
		instance = requestID[0]
	}

	var appError *problem.AppError
	if errors.As(err, &appError) {
		if jsonErr := c.JSON(appError.Status, appError.WithInstance(instance)); jsonErr != nil {
			c.Logger().Error("write app error response failed", "error", jsonErr)
		}
		return
	}

	var validationErr *validator.ValidationError
	if errors.As(err, &validationErr) {
		appError := problem.ErrValidation().
			WithErrors(validationErr).
			WithCause(err).
			WithInstance(instance)
		if jsonErr := c.JSON(appError.Status, appError); jsonErr != nil {
			c.Logger().Error("write validation error response failed", "error", jsonErr)
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

		appError = problem.New(message, "about:blank", code)
	} else {
		appError = problem.ErrInternalServer()
	}
	if jsonErr := c.JSON(code, appError.WithInstance(instance)); jsonErr != nil {
		c.Logger().Error("write generic error response failed", "error", jsonErr)
	}
}
