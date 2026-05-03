---
name: auth-jwt-session
description: >
  Use this skill when changing login, refresh token, logout, auth middleware,
  JWT manager, Redis-backed sessions, or auth rate limiting behavior.
---

# Auth JWT Session

## Purpose

Preserve the project's token-pair auth model and session validity guarantees.

## Apply these rules

1. Keep access/refresh token generation and validation in hash/security components.
2. Treat Redis token/session state as source of truth for active sessions.
3. Ensure refresh, revoke, and logout flows update Redis/session records consistently.
4. Use domain/service interfaces for auth behaviors; avoid handler-side auth logic.
5. Keep middleware behavior consistent for protected routes (`name:"auth"` injection).

## Files usually involved

- `internal/hash/jwtmanager/`
- `internal/service/auth/`
- `internal/repository/session/` and related token repositories
- `internal/delivery/http/middleware/`
- `internal/delivery/http/routes/routes.go`

## Verification commands

```bash
ginkgo run ./internal/service/auth
make test
```

