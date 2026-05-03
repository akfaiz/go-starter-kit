---
name: go-errors-validation
description: Using pkg/problem for standardized RFC 7807 errors and pkg/validator for localized input validation. Use when handling errors in handlers or services, and when validating request DTOs.
---

# Errors and Validation

This project uses **RFC 7807 (Problem Details)** for error responses and `go-playground/validator` for localized input validation.

## Standardized Errors (`pkg/problem`)
Use `pkg/problem` in handlers and transport-layer code to shape HTTP responses.

### Common Error Types
- `problem.ErrBadRequest`: 400 Bad Request
- `problem.ErrUnauthorized`: 401 Unauthorized
- `problem.ErrForbidden`: 403 Forbidden
- `problem.ErrNotFound`: 404 Not Found
- `problem.ErrConflict`: 409 Conflict
- `problem.ErrUnprocessableEntity`: 422 Unprocessable Entity
- `problem.ErrInternalServer`: 500 Internal Server Error

## Localized Input Validation (`pkg/validator`)
The project uses a custom binder that automatically triggers validation after binding. This ensures all request DTOs are validated with context-aware localization before they reach the handler logic.

### Handler Implementation
Simply use `c.Bind` to handle both binding and validation. If validation fails, the global error handler will catch the error and return a standardized RFC 7807 response.

```go
func (h *MyHandler) Create(c *echo.Context) error {
    var req dto.CreateRequest
    // Bind automatically triggers validation
    if err := c.Bind(&req); err != nil {
        return err // Returns *validator.ValidationError or Bind error
    }
    // ...
}
```

### Supported Locales
- English (`en`) - Default
- Indonesian (`id`)

### DTO Definition
Use `label` tag for localized field names in error messages.
```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required" label:"Full Name"`
    Email    string `json:"email" validate:"required,email"`
}
```

## Domain Errors in Services
**Crucial Rule:** Services must *never* return HTTP-specific errors, `pkg/problem`, or `pkg/validator` errors. Services must return **Domain Errors** defined in `internal/domain/error.go` (e.g., `domain.ErrUserNotFound`).

## Mapping Errors in Handlers
Handlers are responsible for mapping Domain Errors to HTTP Problems or Validation Errors, and wrapping unexpected errors:

```go
func (h *MyHandler) Create(c *echo.Context) error {
    // ...
    err := h.service.Create(ctx, req.ToDomain())
    if err != nil {
        // 1. Map known domain errors to HTTP/Validation errors
        if errors.Is(err, domain.ErrEmailAlreadyExists) {
            return validator.NewError("email", "Email already registered")
        }
        // 2. Wrap unknown/unexpected errors
        return problem.Wrap(err, problem.ErrInternalServer)
    }
    // ...
}
```

## Current Handler Pattern
The existing handlers in `internal/delivery/http/handler/` follow this pattern:
- `c.Bind(&req)` handles both binding and validation.
- Domain errors are converted to localized validation or problem responses in the handler.
- Unexpected errors are wrapped with `problem.Wrap(err, problem.ErrInternalServer)`.

Example mappings used in the codebase:
- `domain.ErrEmailAlreadyExists` -> `validator.NewError("email", "Email already registered")`
- `domain.ErrInvalidToken` -> `problem.ErrUnauthorized(...)` or `problem.ErrBadRequest(...)` depending on the route
- `domain.ErrUserNotFound` -> localized validation error for the email field in forgot-password flows
