---
name: go-auth-security
description: JWT, sessions, Redis, and auth guards for security and rate limiting. Use when working with authentication, protecting routes, or implementing rate limits.
---

# Auth and Security

This project implements JWT-based authentication with session management in Redis and rate limiting via Auth Guards.

## JWT and Sessions
- **Access Tokens**: Short-lived, used for authentication.
- **Refresh Tokens**: Long-lived, used to issue new access tokens.
- **Session Repository**: Stores active token pairs in Redis to allow revocation and "login from one device" if desired.

## Protecting Routes
Use the `auth` middleware to protect Echo routes.

```go
// Routes definition
api := e.Group("/api/v1")
authMiddleware := auth.New(jwtManager) // Basic JWT check
// OR
authMiddleware := auth.NewWithSession(jwtManager, sessionRepo) // Strict session check

api.GET("/profile", profileHandler.Get, authMiddleware)
```

## Auth Guard (Rate Limiting)
The `AuthGuard` handles security-related rate limiting (e.g., login attempts, refresh attempts).

### Key Methods
- `CheckLogin(ctx, ip, email)`: Checks if a login attempt should be blocked.
- `OnLoginFailure(ctx, ip, email)`: Records a failed attempt and potentially locks the account.
- `OnLoginSuccess(ctx, email)`: Resets the failure counter on success.

### Usage in Handler
```go
func (h *AuthHandler) Login(c *echo.Context) error {
    // 1. Check if limited
    check, _ := h.authGuard.CheckLogin(ctx, ip, email)
    if check.Limited {
        return tooManyRequests(c, check.RetryAfter)
    }

    // 2. Attempt login
    pair, err := h.authService.Login(ctx, email, password)
    if err != nil {
        // 3. Handle failure
        h.authGuard.OnLoginFailure(ctx, ip, email)
        return err
    }

    // 4. Record success
    h.authGuard.OnLoginSuccess(ctx, email)
    return c.JSON(http.StatusOK, pair)
}
```

## Security Configuration
Settings for JWT expiry, lockout thresholds, and window durations are in `internal/config/auth.go` and `internal/config/rate_limit.go`.
