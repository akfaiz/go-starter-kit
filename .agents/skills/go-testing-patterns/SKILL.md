---
name: go-testing-patterns
description: "Project test patterns using Ginkgo/Gomega, Testify, gomock, Testcontainers, miniredis, and E2E helpers. Use when writing or updating handler, service, repository, queue, telemetry, or E2E tests."
---

# Testing Patterns

This project uses a mix of **Ginkgo/Gomega** for BDD-style tests and standard **Testify** for assertion-based tests.

## Unit Testing

### Handlers and Services (Ginkgo)
Handlers and services usually use Ginkgo for descriptive test suites and `gomock` for dependency mocking.

- **Location**: Next to implementation (`xxx_test.go`).
- **Mocks**: Located in `test/mocks/`. Run `go generate ./...` if domain interfaces changed.

Example (Handler Test):
```go
var _ = Describe("UserHandler", Label("unit"), func() {
    var ctrl *gomock.Controller
    var service *mocks.MockUserService

    BeforeEach(func() {
        ctrl = gomock.NewController(GinkgoT())
        service = mocks.NewMockUserService(ctrl)
    })

    It("returns 200 on success", func() {
        // Setup expectations
        service.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.User{ID: 1}, nil)
        // Execute and assert
    })
})
```

### Repositories
Repositories are tested as integration tests against a real PostgreSQL database using Testcontainers.

- **Setup**: Use `test.NewDBContainer` in `BeforeSuite` to spin up a container.
- **Isolation**: Use `dbContainer.TruncateAll(ctx)` in `BeforeEach` to ensure a clean state for each test.

Example (Repository Test):
```go
var _ = BeforeSuite(func() {
	ctx = context.Background()
	dbContainer = test.NewDBContainer(ctx, GinkgoT())
	r = user.NewRepository(dbContainer.DB)
})

var _ = Describe("UserRepository", func() {
	BeforeEach(func() {
		err := dbContainer.TruncateAll(ctx)
		Expect(err).NotTo(HaveOccurred())
	})
	// tests...
})
```

### Queue and Infrastructure
- Queue client/worker tests may use `miniredis` and Testify.
- Handler tests should cover malformed payloads with `asynq.SkipRetry`, successful calls, and retryable downstream errors.
- Telemetry tests use in-memory OTel exporters from `tracetest`.

## E2E Testing
E2E tests run against a real database and Redis (usually in Docker).

- **Location**: `test/e2e/`
- **Tooling**: Uses Ginkgo and a custom test helper (`e2eExpect`).

Example:
```go
It("registers successfully", func() {
    e2eExpect.POST("/api/v1/auth/register").
        WithJSON(map[string]any{...}).
        Expect().
        Status(http.StatusCreated)
})
```

## Running Tests
- **All tests**: `make test`
- **E2E only**: `make test-e2e`
- **Coverage**: `make coverage` or `make coverage-html`
- **Full coverage merge**: `make coverage-all`

`make test` runs the Ginkgo suite across the repo while skipping generated mocks, migrations, CLI commands, and E2E tests.

## Practical Rules

- Use table-driven tests for compact service/repository cases when it improves clarity.
- Assert domain errors in service and repository tests; assert mapped problem/validation responses in handler tests.
- Keep tests colocated with implementation unless they are full-stack E2E tests.
- Add E2E coverage when API contracts, auth flows, or persistence behavior change end to end.
