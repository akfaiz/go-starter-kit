package problem

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

var (
	ErrBadRequest          = register("Bad Request", "about:blank", 400)
	ErrUnauthorized        = register("Unauthorized access", "about:blank", 401)
	ErrForbidden           = register("Forbidden access", "about:blank", 403)
	ErrNotFound            = register("Resource not found", "about:blank", 404)
	ErrConflict            = register("Resource conflict", "about:blank", 409)
	ErrTooManyRequests     = register("Too Many Requests", "about:blank", 429)
	ErrUnprocessableEntity = register("Unprocessable entity", "about:blank", 422)
	ErrInternalServer      = register("Internal Server Error", "about:blank", 500)

	ErrInvalidRequestBody = register(
		"Invalid request body",
		"about:blank",
		400,
		"Your request body is malformed. Please check your JSON format.",
	)
	ErrInvalidQueryParam = register(
		"Invalid query parameter",
		"about:blank",
		400,
		"Your query parameter is invalid. Please check your request.",
	)
	ErrValidation = register(
		"Validation failed",
		"about:blank",
		422,
		"One or more fields are invalid. Please check your input and try again.",
	)

	ErrTokenExpired = register("Unauthorized", "about:blank", 401, "Your session has expired. Please log in again.")
	ErrTokenInvalid = register("Unauthorized", "about:blank", 401, "Your token is invalid. Please log in again.")
)

// AppError represents a structured error response for the application.
//
// It is based on RFC 7807 (Problem Details for HTTP APIs). (https://datatracker.ietf.org/doc/html/rfc7807)
type AppError struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Errors   any    `json:"errors,omitempty"`

	cause error
}

// AppErrorFunc is a function type that generates an AppError with optional details.
type AppErrorFunc func(details ...string) *AppError

// register creates a new AppErrorFunc with the provided title, type, status, and optional detail.
func register(title, typeName string, status int, detail ...string) AppErrorFunc {
	return func(details ...string) *AppError {
		useDetail := ""
		if len(details) > 0 {
			useDetail = details[0]
		} else if len(detail) > 0 {
			useDetail = detail[0]
		}
		return New(title, typeName, status, useDetail)
	}
}

// New creates a new AppError with the provided title, type, status, and optional detail.
func New(title, typeName string, status int, detail ...string) *AppError {
	var errDetail string
	if len(detail) > 0 {
		errDetail = detail[0]
	}

	return &AppError{
		Type:   typeName,
		Title:  title,
		Status: status,
		Detail: errDetail,
	}
}

// Error implements the error interface for AppError, returning a string representation of the error.
func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Title, e.Detail)
	}
	return e.Title
}

// Unwrap lets `errors.Is` and `errors.As` work.
func (e *AppError) Unwrap() error {
	return e.cause
}

// WithDetail sets the Detail field of the AppError and returns the modified error for chaining.
func (e *AppError) WithDetail(detail string) *AppError {
	e.Detail = detail
	return e
}

// WithErrors sets the Errors field of the AppError and returns the modified error for chaining.
func (e *AppError) WithErrors(errors any) *AppError {
	e.Errors = errors
	return e
}

// WithCause sets the cause of the AppError and returns the modified error for chaining.
func (e *AppError) WithCause(cause error) *AppError {
	e.cause = cause
	return e
}

// WithInstance sets the Instance field of the AppError and returns the modified error for chaining.
func (e *AppError) WithInstance(instance string) *AppError {
	e.Instance = instance
	return e
}

// Clone creates a copy of the AppError, allowing you to modify the copy without affecting the original error.
func (e *AppError) Clone() *AppError {
	return &AppError{
		Type:     e.Type,
		Title:    e.Title,
		Status:   e.Status,
		Detail:   e.Detail,
		Instance: e.Instance,
		Errors:   e.Errors,
		cause:    e.cause,
	}
}

// Wrap takes a standard error and an AppErrorFunc, and returns a new AppError that wraps the original error with a stack trace.
func Wrap(err error, appErr AppErrorFunc) *AppError {
	if err == nil || appErr == nil {
		return nil
	}
	return appErr().WithCause(errors.WithStack(err))
}
