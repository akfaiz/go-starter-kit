---
name: go-i18n-and-localized-errors
description: "ctxi18n locale flow, translation catalogs, localized validation labels, and localized handler errors. Use when adding user-facing messages or locale-aware errors."
---

# I18n and Localized Errors

This project uses `ctxi18n` for request-scoped locale handling and `pkg/validator` for translated validation errors.

## Core Pieces
- `internal/delivery/http/middleware/i18n.go`: reads `Accept-Language` and stores the locale in the request context.
- `internal/lang/lang.go`: embeds and loads translation files.
- `internal/lang/en.yml` and `internal/lang/id.yml`: message catalogs.
- `pkg/validator/validator.go`: chooses the translator from the request context.
- `internal/delivery/http/handler/auth_handler.go`: current reference for translated messages in handlers.

## Locale Flow
1. HTTP middleware reads `Accept-Language`.
2. `ctxi18n.WithLocale` stores the locale on the request context.
3. Handlers call `i18n.T(ctx, key)` for localized user-facing strings.
4. `pkg/validator` uses the same context locale when formatting validation errors.

## When Adding a New Message
1. Add the key to both `internal/lang/en.yml` and `internal/lang/id.yml`.
2. Keep the same nested path in both files.
3. Use the key from handlers or validators via the request context.

Example:
```go
return validator.NewError("email", i18n.T(c.Request().Context(), "passwords.user"))
```

## Current Message Keys
The existing catalogs include:
- `auth.failed`
- `auth.password`
- `passwords.reset`
- `passwords.sent`
- `passwords.token`
- `passwords.user`

## Localized Error Rules
- Prefer translated validation messages for user input problems.
- Use `problem.ErrBadRequest` or other `pkg/problem` errors for route-level failures that should still be localized.
- Keep services language-agnostic; localize at the handler boundary.
- DTO `label` tags are field labels for validation output; keep them clear and stable.

## Practical Rules
- Do not hardcode user-facing strings if a translation key already exists.
- Use the request context for translation lookups, not global state.
- When a new locale is added, update the middleware fallback and validator support together.
- Keep English and Indonesian catalogs structurally identical.
- Add or update handler tests when changing localized error behavior.
