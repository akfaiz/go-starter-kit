---
name: go-testing-ginkgo
description: >
  Use this skill when adding or updating tests in this repository, especially
  service/handler tests with mocks and focused Ginkgo workflows.
---

# Go Testing with Ginkgo and Testify

## Purpose

Write tests that match this repository's existing patterns and tooling.

## Apply these rules

1. Keep tests near implementation (`*_test.go`).
2. Prefer table-driven tests for service/repository behavior where practical.
3. Use generated mocks from `test/mocks/` for dependency isolation.
4. Keep assertions clear and behavior-focused.
5. Initialize i18n in suites that depend on localized behavior (`lang.Init()` pattern).

## Useful commands

```bash
make test
make test-e2e
ginkgo run ./internal/service/auth --focus "RefreshToken"
go test ./internal/hash/jwtmanager -run TestJWTManager -v
```

