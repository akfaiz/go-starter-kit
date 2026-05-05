package problem

import (
	"fmt"
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

// Error represents a structured error response for the application.
//
// It is based on RFC 7807 (Problem Details for HTTP APIs). (https://datatracker.ietf.org/doc/html/rfc7807)
type Error struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Errors   any    `json:"errors,omitempty"`

	cause error
}

// ErrorFunc is a function type that generates an Error with optional details.
type ErrorFunc func(details ...string) *Error

// register creates a new ErrorFunc with the provided title, type, status, and optional detail.
func register(title, typeName string, status int, detail ...string) ErrorFunc {
	return func(details ...string) *Error {
		useDetail := ""
		if len(details) > 0 {
			useDetail = details[0]
		} else if len(detail) > 0 {
			useDetail = detail[0]
		}
		return New(title, typeName, status, useDetail)
	}
}

// New creates a new Error with the provided title, type, status, and optional detail.
func New(title, typeName string, status int, detail ...string) *Error {
	var errDetail string
	if len(detail) > 0 {
		errDetail = detail[0]
	}

	return &Error{
		Type:   typeName,
		Title:  title,
		Status: status,
		Detail: errDetail,
	}
}

// Error implements the error interface for Error, returning a string representation of the error.
func (e *Error) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Title, e.Detail)
	}
	return e.Title
}

// Unwrap lets `errors.Is` and `errors.As` work.
func (e *Error) Unwrap() error {
	return e.cause
}

// WithDetail sets the Detail field of the Error and returns the modified error for chaining.
func (e *Error) WithDetail(detail string) *Error {
	e.Detail = detail
	return e
}

// WithErrors sets the Errors field of the Error and returns the modified error for chaining.
func (e *Error) WithErrors(errors any) *Error {
	e.Errors = errors
	return e
}

// WithCause sets the cause of the Error and returns the modified error for chaining.
func (e *Error) WithCause(cause error) *Error {
	e.cause = cause
	return e
}

// WithInstance sets the Instance field of the Error and returns the modified error for chaining.
func (e *Error) WithInstance(instance string) *Error {
	e.Instance = instance
	return e
}

// Clone creates a copy of the Error, allowing you to modify the copy without affecting the original error.
func (e *Error) Clone() *Error {
	return &Error{
		Type:     e.Type,
		Title:    e.Title,
		Status:   e.Status,
		Detail:   e.Detail,
		Instance: e.Instance,
		Errors:   e.Errors,
		cause:    e.cause,
	}
}

// Wrap takes a standard error and an ErrorFunc, and returns a new Error that wraps the original error.
func Wrap(err error, appErr ErrorFunc) *Error {
	if err == nil || appErr == nil {
		return nil
	}
	return appErr().WithCause(err)
}
