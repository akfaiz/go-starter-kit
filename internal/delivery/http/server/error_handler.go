package server

import (
	"fmt"
	"net/http"

	"github.com/akfaiz/go-starter-kit/internal/errdefs"
	"github.com/akfaiz/go-starter-kit/internal/validator"
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

	var appError *errdefs.AppError
	if errors.As(err, &appError) {
		if jsonErr := c.JSON(appError.Status, appError.WithInstance(instance)); jsonErr != nil {
			c.Logger().Error("write app error response failed", "error", jsonErr)
		}
		return
	}

	var validationErr *validator.ValidationError
	if errors.As(err, &validationErr) {
		appError := errdefs.ErrValidation().
			WithErrors(validationErr).
			WithCause(err).
			WithInstance(instance)
		if jsonErr := c.JSON(appError.Status, appError); jsonErr != nil {
			c.Logger().Error("write validation error response failed", "error", jsonErr)
		}
		return
	}

	code := http.StatusInternalServerError
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		code = httpErr.Code
		appError = errdefs.New(fmt.Sprintf("%v", httpErr.Message), "about:blank", code)
	} else {
		appError = errdefs.ErrInternalServer()
	}
	if jsonErr := c.JSON(code, appError.WithInstance(instance)); jsonErr != nil {
		c.Logger().Error("write generic error response failed", "error", jsonErr)
	}
}
