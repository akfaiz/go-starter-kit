---
name: go-errors-validation
description: Using pkg/problem for standardized RFC 7807 errors and pkg/validator for input validation. Use when handling errors in handlers or services, and when validating request DTOs.
---

# Errors and Validation

This project uses **RFC 7807 (Problem Details)** for error responses and `go-playground/validator` for input validation.

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

### Usage Examples
```go
// Direct return
return problem.ErrNotFound("User with ID 123 not found")

// Wrapping an existing error
if err != nil {
    return problem.Wrap(err, problem.ErrInternalServer)
}

// Adding field-level errors (Validation)
return problem.ErrValidation().WithErrors(validationErrors)
```

## Input Validation (`pkg/validator`)
Use `pkg/validator` in handlers to validate request DTOs.

### DTO Definition
```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2" label:"Full Name"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}
```

### Handler Implementation
```go
func (h *Handler) Create(c *echo.Context) error {
    var req dto.CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return err
    }
    // Echo is configured to use pkg/validator
    if err := c.Validate(&req); err != nil {
        return err // returns *validator.ValidationError
    }
    // ...
}
```

### Custom Errors in Services
If a service needs to return a validation error (e.g., "email already taken"):
```go
if exists {
    return validator.NewError("email", "Email already registered")
}
```
The handler or error middleware will automatically map this to a `problem.ErrValidation`.
