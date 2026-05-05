---
name: go-i18n-and-localized-errors
description: "ctxi18n locale flow, translation catalogs, localized validation labels, and localized handler errors. Use when adding user-facing messages or locale-aware errors."
---

# I18n and Localized Errors

This project uses `github.com/akfaiz/go-starter-kit/pkg/i18n` (wrapping `ctxi18n`) for request-scoped locale handling and `pkg/validator` for translated validation errors.

## Core Pieces
- `internal/delivery/http/middleware/i18n.go`: reads `Accept-Language`, falls back to `"en"`, stores the locale via `i18n.WithLocale`.
- `internal/lang/lang.go`: embeds `*.yml` files and calls `i18n.LoadWithDefault(fs, "en")`.
- `internal/lang/en.yml` and `internal/lang/id.yml`: message catalogs (must be kept structurally identical).
- `pkg/i18n/i18n.go`: exposes `T(c *echo.Context, key)` and `TCtx(ctx context.Context, key)`.
- `internal/delivery/http/handler/auth_handler.go`: reference for all localized handler patterns.

## Locale Flow
1. HTTP middleware reads `Accept-Language` header; defaults to `"en"` when absent or unrecognised.
2. `i18n.WithLocale(ctx, locale)` stores the locale on the request context.
3. Handlers call `i18n.T(c, key)` — passing the `*echo.Context` directly.
4. `pkg/validator` uses the same context locale when formatting validation errors.

## Translation Helpers

| Helper | Signature | Use when |
|--------|-----------|----------|
| `i18n.T` | `T(c *echo.Context, key string, args ...any) string` | Inside an Echo handler (most common) |
| `i18n.TCtx` | `TCtx(ctx context.Context, key string, args ...any) string` | Outside Echo handlers (queue workers, services — avoid; localize at handler boundary) |

Import: `"github.com/akfaiz/go-starter-kit/pkg/i18n"`

## When Adding a New Message
1. Add the key to **both** `internal/lang/en.yml` and `internal/lang/id.yml` at the same nested path.
2. Call it from the handler via `i18n.T(c, "section.key")`.

Examples from existing handlers:
```go
// field validation error
return validator.NewError("email", i18n.T(c, "auth.failed"))

// problem response
return problem.ErrUnauthorized(i18n.T(c, "auth.invalid_refresh_token"))
return problem.ErrBadRequest(i18n.T(c, "passwords.token"))

// success response message
res := dto.NewResponse(200, dto.NewTokenResponse(token), i18n.T(c, "auth.login_success"))
res := dto.NewMessage(200, i18n.T(c, "passwords.sent"))
```

## Current Message Keys

```
auth.failed                       auth.password
auth.email_exists                 auth.login_success
auth.register_success             auth.invalid_refresh_token

passwords.reset                   passwords.sent
passwords.token                   passwords.user
passwords.otp_valid               passwords.reset_success

profile.update_success            profile.email_exists
profile.current_password_invalid  profile.password_changed

middleware.auth.missing_header    middleware.auth.invalid_format
middleware.auth.missing_token     middleware.auth.invalid_session
middleware.auth.token_expired     middleware.auth.token_invalid
```

## Localized Error Rules
- Prefer `validator.NewError("field", i18n.T(c, key))` for user input problems (renders as field-level error).
- Use `problem.ErrBadRequest(i18n.T(c, key))` or other `pkg/problem` helpers for request-level failures.
- Keep services language-agnostic; **localize only at the handler boundary**.
- DTO `label` tags are field labels for validation output; keep them clear and stable.

## Practical Rules
- Do not hardcode user-facing strings — use a translation key.
- Always pass `*echo.Context` (`c`) to `i18n.T`, not the raw `context.Context`.
- Keep English and Indonesian catalogs structurally identical at all times.
- Add or update handler tests when changing localized error behavior.
- Ginkgo handler test suites must call `lang.Init()` in `BeforeSuite` before any handler is exercised.
