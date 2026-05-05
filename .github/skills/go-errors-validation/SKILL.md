---
name: go-errors-validation
description: "Project error and validation rules: domain errors in services, RFC 7807 problem responses in delivery, localized validator errors for DTOs. Use when handling errors, mapping domain errors, or validating request DTOs."
---

# Errors and Validation

This project uses **RFC 7807 (Problem Details)** for error responses and `go-playground/validator` for localized input validation.

## Standardized Errors (`pkg/problem`)
Use `pkg/problem` only in handlers, middleware, and transport-layer code to shape HTTP responses.

### Common Error Types
- `problem.ErrBadRequest`: 400 Bad Request
- `problem.ErrUnauthorized`: 401 Unauthorized
- `problem.ErrForbidden`: 403 Forbidden
- `problem.ErrNotFound`: 404 Not Found
- `problem.ErrConflict`: 409 Conflict
- `problem.ErrTooManyRequests`: 429 Too Many Requests
- `problem.ErrUnprocessableEntity`: 422 Unprocessable Entity
- `problem.ErrInternalServer`: 500 Internal Server Error

## Localized Input Validation (`pkg/validator`)
The project uses a custom binder that automatically triggers validation after binding. This ensures all request DTOs are validated with context-aware localization before they reach the handler logic.

### Handler Implementation
Use `c.Bind` to handle both binding and validation. If validation fails, the global error handler converts the validator error into a standardized RFC 7807 response.

```go
func (h *MyHandler) Create(c *echo.Context) error {
    var req dto.CreateRequest
    if err := c.Bind(&req); err != nil {
        return err
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
Services must not return HTTP-specific errors, `pkg/problem`, or `pkg/validator` errors. Services return domain errors from `internal/domain/error.go`, for example `domain.ErrUserNotFound`, `domain.ErrInvalidToken`, or `domain.ErrEmailAlreadyExists`.

Repositories also return domain errors for expected persistence outcomes, and attach stack traces to unexpected DB failures.

## Stack Trace Capture
Capture stack traces at the **first unexpected error origin** (repository/infra/hash/service internals), not at the HTTP mapping layer.

- Use `github.com/cockroachdb/errors` (`errors.WithStack` / `errors.Wrap`) when returning unexpected low-level errors.
- Keep expected business outcomes as domain errors without stack wrapping.
- `problem.Wrap(err, problem.ErrInternalServer)` should map errors to RFC 7807 response shape, not become the primary stack-capture point.

## Mapping Errors in Handlers
Handlers are responsible for mapping Domain Errors to HTTP Problems or Validation Errors, and wrapping unexpected errors for transport response:

```go
func (h *MyHandler) Create(c *echo.Context) error {
    // ...
    err := h.service.Create(ctx, req.ToDomain())
    if err != nil {
        if errors.Is(err, domain.ErrEmailAlreadyExists) {
            return validator.NewError("email", "Email already registered")
        }
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
- Stack traces for unexpected failures should already be attached from lower layers.

Example mappings used in the codebase:
- `domain.ErrEmailAlreadyExists` -> `validator.NewError("email", "Email already registered")`
- `domain.ErrInvalidToken` -> `problem.ErrUnauthorized(...)` or `problem.ErrBadRequest(...)` depending on the route
- `domain.ErrUserNotFound` -> localized validation error for the email field in forgot-password flows

## Review Checklist

- DTOs use `validate` tags and `label` tags where field names appear in validation messages.
- Handlers use `errors.Is` for domain error mapping.
- User-facing strings are localized with `i18n.T(ctx, key)` when a catalog key exists.
- Service tests assert domain errors; handler tests assert mapped HTTP/validation behavior.
