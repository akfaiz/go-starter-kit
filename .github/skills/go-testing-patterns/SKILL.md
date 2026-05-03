---
name: go-testing-patterns
description: Ginkgo, Testify, and E2E testing patterns for this project. Use when writing unit tests for handlers, services, repositories, or adding new E2E test cases.
---

# Testing Patterns

This project uses a mix of **Ginkgo/Gomega** for BDD-style tests and standard **Testify** for assertion-based tests.

## Unit Testing

### Handlers and Services (Ginkgo)
Handlers and services usually use Ginkgo for descriptive test suites and `gomock` for dependency mocking.

- **Location**: Next to implementation (`xxx_test.go`).
- **Mocks**: Located in `test/mocks/`. Run `go generate ./...` to update them if you changed interfaces.

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
