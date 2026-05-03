---
name: go-errors-validation
description: Using pkg/problem for standardized RFC 7807 errors and pkg/validator for localized input validation. Use when handling errors in handlers or services, and when validating request DTOs.
---

# Errors and Validation

This project uses **RFC 7807 (Problem Details)** for error responses and `go-playground/validator` for localized input validation.

## Standardized Errors (`pkg/problem`)
Always use `pkg/problem` to return errors from handlers or map domain errors in services.

### Common Error Types
- `problem.ErrBadRequest`: 400 Bad Request
- `problem.ErrUnauthorized`: 401 Unauthorized
- `problem.ErrForbidden`: 403 Forbidden
- `problem.ErrNotFound`: 404 Not Found
- `problem.ErrConflict`: 409 Conflict
- `problem.ErrUnprocessableEntity`: 422 Unprocessable Entity
- `problem.ErrInternalServer`: 500 Internal Server Error

## Localized Input Validation (`pkg/validator`)
Use `pkg/validator` in handlers to validate request DTOs with context-aware localization.

### Handler Implementation
Inject the validator into your handler and use `ValidateWithContext` to respect the user's locale (provided by `ctxi18n` middleware).

```go
type MyHandler struct {
    validator *validator.Validate
    // ...
}

func (h *MyHandler) Create(c *echo.Context) error {
    var req dto.CreateRequest
    if err := c.Bind(&req); err != nil {
        return err
    }
    
    // Use context-aware validation for localized error messages (EN, ID supported)
    if err := h.validator.ValidateContext(c.Request().Context(), &req); err != nil {
        return err // Returns *validator.ValidationError mapped to RFC 7807
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
