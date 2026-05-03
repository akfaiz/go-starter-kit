package problem_test

import (
	"errors"
	"testing"

	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/stretchr/testify/assert"
)

func TestAppErrorError(t *testing.T) {
	t.Run("returns title only when detail is empty", func(t *testing.T) {
		appErr := problem.New("Not Found", "about:blank", 404)
		assert.Equal(t, "Not Found", appErr.Error())
	})

	t.Run("returns formatted title and detail", func(t *testing.T) {
		appErr := problem.New("Not Found", "about:blank", 404, "User with ID 123 not found")
		assert.Equal(t, "Not Found: User with ID 123 not found", appErr.Error())
	})
}

func TestAppErrorUnwrap(t *testing.T) {
	t.Run("returns nil when no cause is set", func(t *testing.T) {
		appErr := problem.New("Bad Request", "about:blank", 400)
		assert.Nil(t, appErr.Unwrap())
	})

	t.Run("returns the cause error", func(t *testing.T) {
		cause := errors.New("original error")
		appErr := problem.New("Bad Request", "about:blank", 400).WithCause(cause)
		assert.Equal(t, cause, appErr.Unwrap())
	})
}

func TestAppErrorWithDetail(t *testing.T) {
	t.Run("sets detail on error", func(t *testing.T) {
		appErr := problem.New("Validation failed", "about:blank", 422)
		appErr.WithDetail("Email is required")
		assert.Equal(t, "Email is required", appErr.Detail)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		appErr := problem.New("Validation failed", "about:blank", 422).
			WithDetail("Email is required")
		assert.Equal(t, "Email is required", appErr.Detail)
	})
}

func TestAppErrorWithErrors(t *testing.T) {
	t.Run("sets errors field with slice", func(t *testing.T) {
		errs := []string{"Email is required", "Password is too short"}
		appErr := problem.New("Validation failed", "about:blank", 422).WithErrors(errs)
		assert.Equal(t, errs, appErr.Errors)
	})

	t.Run("sets errors field with map", func(t *testing.T) {
		errs := map[string]string{"email": "is required", "password": "too short"}
		appErr := problem.New("Validation failed", "about:blank", 422).WithErrors(errs)
		assert.Equal(t, errs, appErr.Errors)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		errs := []string{"Error 1"}
		appErr := problem.New("Bad Request", "about:blank", 400).
			WithErrors(errs).
			WithDetail("Some detail")
		assert.Equal(t, errs, appErr.Errors)
		assert.Equal(t, "Some detail", appErr.Detail)
	})
}

func TestAppErrorWithCause(t *testing.T) {
	t.Run("sets cause error", func(t *testing.T) {
		cause := errors.New("database connection failed")
		appErr := problem.New("Internal Server Error", "about:blank", 500).WithCause(cause)
		assert.Equal(t, cause, appErr.Unwrap())
	})

	t.Run("supports method chaining", func(t *testing.T) {
		cause := errors.New("db error")
		appErr := problem.New("Internal Server Error", "about:blank", 500).
			WithCause(cause).
			WithDetail("Something went wrong")
		assert.Equal(t, cause, appErr.Unwrap())
		assert.Equal(t, "Something went wrong", appErr.Detail)
	})
}

func TestAppErrorWithInstance(t *testing.T) {
	t.Run("sets instance field", func(t *testing.T) {
		appErr := problem.New("Not Found", "about:blank", 404).
			WithInstance("/users/123")
		assert.Equal(t, "/users/123", appErr.Instance)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		appErr := problem.New("Not Found", "about:blank", 404).
			WithInstance("/users/123").
			WithDetail("User not found")
		assert.Equal(t, "/users/123", appErr.Instance)
		assert.Equal(t, "User not found", appErr.Detail)
	})
}

func TestAppErrorClone(t *testing.T) {
	t.Run("creates independent copy", func(t *testing.T) {
		original := problem.New("Bad Request", "about:blank", 400, "Invalid input").
			WithDetail("Specific detail").
			WithInstance("/endpoint").
			WithErrors([]string{"error1"})

		cloned := original.Clone()

		assert.Equal(t, original.Title, cloned.Title)
		assert.Equal(t, original.Type, cloned.Type)
		assert.Equal(t, original.Status, cloned.Status)
		assert.Equal(t, original.Detail, cloned.Detail)
		assert.Equal(t, original.Instance, cloned.Instance)
		assert.Equal(t, original.Errors, cloned.Errors)
	})

	t.Run("clone has independent state", func(t *testing.T) {
		original := problem.New("Bad Request", "about:blank", 400, "Original")
		cloned := original.Clone()

		cloned.Detail = "Modified"
		assert.Equal(t, "Original", original.Detail)
		assert.Equal(t, "Modified", cloned.Detail)
	})

	t.Run("clones cause error", func(t *testing.T) {
		cause := errors.New("original cause")
		original := problem.New("Error", "about:blank", 500).WithCause(cause)
		cloned := original.Clone()

		assert.Equal(t, cause, cloned.Unwrap())
	})
}

func TestNew(t *testing.T) {
	t.Run("creates error with all fields", func(t *testing.T) {
		appErr := problem.New("Not Found", "https://example.com/errors/not-found", 404, "Resource not found")
		assert.Equal(t, "Not Found", appErr.Title)
		assert.Equal(t, "https://example.com/errors/not-found", appErr.Type)
		assert.Equal(t, 404, appErr.Status)
		assert.Equal(t, "Resource not found", appErr.Detail)
	})

	t.Run("creates error without detail", func(t *testing.T) {
		appErr := problem.New("Internal Server Error", "about:blank", 500)
		assert.Equal(t, "Internal Server Error", appErr.Title)
		assert.Equal(t, "", appErr.Detail)
	})

	t.Run("ignores extra details", func(t *testing.T) {
		appErr := problem.New("Error", "about:blank", 400, "First detail", "Second detail")
		assert.Equal(t, "First detail", appErr.Detail)
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("ErrBadRequest", func(t *testing.T) {
		appErr := problem.ErrBadRequest()
		assert.Equal(t, 400, appErr.Status)
		assert.Equal(t, "Bad Request", appErr.Title)
	})

	t.Run("ErrUnauthorized", func(t *testing.T) {
		appErr := problem.ErrUnauthorized()
		assert.Equal(t, 401, appErr.Status)
		assert.Equal(t, "Unauthorized access", appErr.Title)
	})

	t.Run("ErrForbidden", func(t *testing.T) {
		appErr := problem.ErrForbidden()
		assert.Equal(t, 403, appErr.Status)
		assert.Equal(t, "Forbidden access", appErr.Title)
	})

	t.Run("ErrNotFound", func(t *testing.T) {
		appErr := problem.ErrNotFound()
		assert.Equal(t, 404, appErr.Status)
		assert.Equal(t, "Resource not found", appErr.Title)
	})

	t.Run("ErrConflict", func(t *testing.T) {
		appErr := problem.ErrConflict()
		assert.Equal(t, 409, appErr.Status)
		assert.Equal(t, "Resource conflict", appErr.Title)
	})

	t.Run("ErrTooManyRequests", func(t *testing.T) {
		appErr := problem.ErrTooManyRequests()
		assert.Equal(t, 429, appErr.Status)
		assert.Equal(t, "Too Many Requests", appErr.Title)
	})

	t.Run("ErrUnprocessableEntity", func(t *testing.T) {
		appErr := problem.ErrUnprocessableEntity()
		assert.Equal(t, 422, appErr.Status)
		assert.Equal(t, "Unprocessable entity", appErr.Title)
	})

	t.Run("ErrInternalServer", func(t *testing.T) {
		appErr := problem.ErrInternalServer()
		assert.Equal(t, 500, appErr.Status)
		assert.Equal(t, "Internal Server Error", appErr.Title)
	})
}

func TestPredefinedErrorsWithCustomDetail(t *testing.T) {
	t.Run("ErrInvalidRequestBody with custom detail", func(t *testing.T) {
		appErr := problem.ErrInvalidRequestBody("Missing required field: name")
		assert.Equal(t, 400, appErr.Status)
		assert.Equal(t, "Missing required field: name", appErr.Detail)
	})

	t.Run("ErrInvalidQueryParam with custom detail", func(t *testing.T) {
		appErr := problem.ErrInvalidQueryParam("Invalid page number")
		assert.Equal(t, 400, appErr.Status)
		assert.Equal(t, "Invalid page number", appErr.Detail)
	})

	t.Run("ErrValidation with default detail", func(t *testing.T) {
		appErr := problem.ErrValidation()
		assert.Equal(t, 422, appErr.Status)
		assert.Contains(t, appErr.Detail, "invalid")
	})

	t.Run("ErrTokenExpired with default detail", func(t *testing.T) {
		appErr := problem.ErrTokenExpired()
		assert.Equal(t, 401, appErr.Status)
		assert.Contains(t, appErr.Detail, "expired")
	})

	t.Run("ErrTokenInvalid with default detail", func(t *testing.T) {
		appErr := problem.ErrTokenInvalid()
		assert.Equal(t, 401, appErr.Status)
		assert.Contains(t, appErr.Detail, "invalid")
	})
}

func TestWrap(t *testing.T) {
	t.Run("wraps error with app error func", func(t *testing.T) {
		cause := errors.New("database error")
		result := problem.Wrap(cause, problem.ErrInternalServer)
		assert.NotNil(t, result)
		assert.Equal(t, 500, result.Status)
		assert.NotNil(t, result.Unwrap())
	})

	t.Run("returns nil when cause is nil", func(t *testing.T) {
		result := problem.Wrap(nil, problem.ErrInternalServer)
		assert.Nil(t, result)
	})

	t.Run("returns nil when error func is nil", func(t *testing.T) {
		cause := errors.New("error")
		result := problem.Wrap(cause, nil)
		assert.Nil(t, result)
	})

	t.Run("wraps with different error funcs", func(t *testing.T) {
		cause := errors.New("first error")
		wrapped := problem.Wrap(cause, problem.ErrBadRequest)

		assert.NotNil(t, wrapped)
		assert.Equal(t, 400, wrapped.Status)
	})

	t.Run("preserves error stack", func(t *testing.T) {
		cause := errors.New("root cause")
		result := problem.Wrap(cause, problem.ErrInternalServer)
		assert.NotNil(t, result.Unwrap())
	})
}

func TestAppErrorChaining(t *testing.T) {
	t.Run("chains multiple methods", func(t *testing.T) {
		appErr := problem.New("Validation failed", "about:blank", 422).
			WithDetail("Email validation failed").
			WithErrors(map[string]string{"email": "invalid format"}).
			WithInstance("/users/signup").
			WithCause(errors.New("regex mismatch"))

		assert.Equal(t, "Validation failed", appErr.Title)
		assert.Equal(t, 422, appErr.Status)
		assert.Equal(t, "Email validation failed", appErr.Detail)
		assert.NotNil(t, appErr.Errors)
		assert.Equal(t, "/users/signup", appErr.Instance)
		assert.NotNil(t, appErr.Unwrap())
	})
}

func TestAppErrorRFC7807Compliance(t *testing.T) {
	t.Run("follows RFC 7807 structure", func(t *testing.T) {
		appErr := problem.New("Not Found", "https://example.com/errors/not-found", 404, "User not found")
		appErr.WithInstance("/users/123")

		assert.NotEmpty(t, appErr.Type)
		assert.NotEmpty(t, appErr.Title)
		assert.NotZero(t, appErr.Status)
		assert.Equal(t, "User not found", appErr.Detail)
		assert.Equal(t, "/users/123", appErr.Instance)
	})

	t.Run("all predefined errors have RFC 7807 fields", func(t *testing.T) {
		errorTests := []struct {
			name      string
			errorFunc problem.AppErrorFunc
			status    int
		}{
			{"ErrBadRequest", problem.ErrBadRequest, 400},
			{"ErrUnauthorized", problem.ErrUnauthorized, 401},
			{"ErrForbidden", problem.ErrForbidden, 403},
			{"ErrNotFound", problem.ErrNotFound, 404},
			{"ErrConflict", problem.ErrConflict, 409},
			{"ErrTooManyRequests", problem.ErrTooManyRequests, 429},
			{"ErrUnprocessableEntity", problem.ErrUnprocessableEntity, 422},
			{"ErrInternalServer", problem.ErrInternalServer, 500},
		}

		for _, e := range errorTests {
			t.Run(e.name, func(t *testing.T) {
				appErr := e.errorFunc()
				assert.Equal(t, e.status, appErr.Status)
				assert.NotEmpty(t, appErr.Title)
				assert.Equal(t, "about:blank", appErr.Type)
			})
		}
	})
}
