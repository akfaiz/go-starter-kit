---
name: go-testing-patterns
description: "Project test patterns using Ginkgo/Gomega, Testify, gomock, Testcontainers, miniredis, and E2E helpers. Use when writing or updating handler, service, repository, queue, telemetry, or E2E tests."
---

# Testing Patterns

This project uses a mix of **Ginkgo/Gomega** for BDD-style tests and standard **Testify** for assertion-based tests.

## Package Naming

Always use the external `_test` package (e.g. `package handler_test`, `package auth_test`). This enforces black-box testing and avoids accidental access to unexported internals. Only drop the suffix when you genuinely need to test unexported helpers and cannot cover them indirectly through the public API.

## Unit Testing

### Handlers (Ginkgo + httpexpect)

Handler tests live next to the implementation (`xxx_test.go`). Each handler package needs two files:
- `handler_suite_test.go` — Ginkgo bootstrap + `lang.Init()` in `BeforeSuite`
- `test_helper_test.go` — shared `setupEcho()` and `newExpect()` helpers

**Suite bootstrap** (`handler_suite_test.go`):
```go
package handler_test

func TestHandler(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "HTTP Handler Suite")
}

var _ = BeforeSuite(func() { lang.Init() })
```

**Helpers** (`test_helper_test.go`):
```go
package handler_test

func newExpect(e *echo.Echo) *httpexpect.Expect {
    return httpexpect.WithConfig(httpexpect.Config{
        Client:   &http.Client{Transport: httpexpect.NewBinder(e)},
        Reporter: httpexpect.NewRequireReporter(GinkgoT()),
    })
}

func setupEcho() *echo.Echo {
    e := echo.New()
    e.Use(middleware.I18n())
    v := appvalidator.New()
    e.Validator = v
    e.Binder = appvalidator.NewBinder(e.Binder, v)
    e.HTTPErrorHandler = server.CustomHTTPErrorHandler
    return e
}
```

**Test structure**:
```go
package handler_test

var _ = Describe("AuthHandler", Label("unit", "handler"), func() {
    var (
        ctrl    *gomock.Controller
        service *mocks.MockAuthService
        h       *handler.AuthHandler
        e       *echo.Echo
        expect  *httpexpect.Expect
    )

    BeforeEach(func() {
        ctrl = gomock.NewController(GinkgoT())
        service = mocks.NewMockAuthService(ctrl)
        h = handler.NewAuthHandler(service)
        e = setupEcho()
        expect = newExpect(e)
    })

    AfterEach(func() { ctrl.Finish() })

    Describe("Login", func() {
        BeforeEach(func() {
            e.POST("/auth/login", h.Login)
        })

        It("returns 200 with token response on success", func() {
            service.EXPECT().Login(gomock.Any(), "john@example.com", "secret123").
                Return(&domain.PairToken{AccessToken: "acc", RefreshToken: "ref"}, nil)

            expect.POST("/auth/login").
                WithJSON(map[string]any{"email": "john@example.com", "password": "secret123"}).
                Expect().
                Status(http.StatusOK).
                JSON().Object().HasValue("status", 200)
        })
    })
})
```

### Services (Ginkgo + miniredis)

Services use Ginkgo with `gomock` mocks and `miniredis` for queue clients. Mock setup and teardown use `DeferCleanup` instead of `AfterEach`.

```go
package auth_test

func TestAuth(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Auth Suite")
}

var _ = Describe("Auth", Label("unit", "usecase"), func() {
    var (
        userRepoMock *mocks.MockUserRepository
        svc          domain.AuthService
        mr           *miniredis.Miniredis
    )

    BeforeEach(func() {
        mr, _ = miniredis.Run()
        workerClient := queue.NewClient(config.Redis{Addr: mr.Addr()})

        ctrl := gomock.NewController(GinkgoT())
        userRepoMock = mocks.NewMockUserRepository(ctrl)
        svc = auth.NewService(cfg, userRepoMock, ..., workerClient)

        DeferCleanup(func() {
            ctrl.Finish()
            workerClient.Close()
            mr.Close()
        })
    })
})
```

### Repositories (Ginkgo + Testcontainers)

Repositories are tested against a real PostgreSQL database using Testcontainers. Always add `AfterSuite` to close the container.

```go
package user_test

func TestUserRepository(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "User Repository Suite")
}

var (
    dbContainer *test.DBContainer
    r           domain.UserRepository
    ctx         context.Context
)

var _ = BeforeSuite(func() {
    ctx = context.Background()
    dbContainer = test.NewDBContainer(ctx, GinkgoT())
    r = user.NewRepository(dbContainer.DB)
})

var _ = AfterSuite(func() {
    if dbContainer != nil {
        Expect(dbContainer.Close(ctx)).NotTo(HaveOccurred())
    }
})

var _ = Describe("User Repository", Label("unit", "repository", "integration"), func() {
    BeforeEach(func() {
        Expect(dbContainer.TruncateAll(ctx)).NotTo(HaveOccurred())
    })
})
```

### Queue and Infrastructure (Testify)

Queue client/worker and other infrastructure tests use plain Testify (not Ginkgo). Use `miniredis` for Redis and `tracetest.NewInMemoryExporter` for OTel span assertions.

```go
package queue_test

func TestClient_EnqueueContext(t *testing.T) {
    exporter := tracetest.NewInMemoryExporter()
    tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
    otel.SetTracerProvider(tp)
    t.Cleanup(func() { tp.Shutdown(context.Background()) })

    mr, _ := miniredis.Run()
    t.Cleanup(mr.Close)

    client := queue.NewClient(config.Redis{Addr: mr.Addr()})
    t.Cleanup(func() { client.Close() })
    // ...
}
```

## E2E Testing

E2E tests run against real PostgreSQL and Redis containers. They use `httpexpect` via `e2eExpect` and reset both DB and Redis in `BeforeEach`.

- **Location**: `test/e2e/`
- Flush both DB and Redis cache in `BeforeEach`: `dbContainer.TruncateAll` + `rdb.FlushDB`

```go
package e2e_test

var _ = Describe("Auth Flow E2E", Label("e2e"), func() {
    BeforeEach(func() {
        Expect(e2eDBContainer.TruncateAll(e2eCtx)).NotTo(HaveOccurred())
        Expect(e2eRDB.FlushDB(e2eCtx).Err()).NotTo(HaveOccurred())
    })

    It("registers, logs in, and fetches profile successfully", func() {
        e2eExpect.POST("/api/v1/auth/register").
            WithJSON(map[string]any{...}).
            Expect().
            Status(http.StatusCreated)
    })
})
```

## Running Tests
- **All unit tests**: `make test`
- **E2E only**: `make test-e2e`
- **Coverage**: `make coverage` or `make coverage-html`
- **Full coverage merge**: `make coverage-all`

`make test` runs Ginkgo across the repo, skipping `test/mocks`, `db/migrations`, `cmd`, and `test/e2e`.

## Practical Rules

- **Always use `package xxx_test`** (external test package). Only use the internal package when testing unexported symbols that cannot be covered indirectly.
- Use `Label("unit", "handler"|"usecase"|"repository"|"integration")` for Ginkgo test categorisation.
- Assert domain errors in service and repository tests; assert HTTP status codes, response shape, and `application/problem+json` bodies in handler tests.
- Use `DeferCleanup` in `BeforeEach` for mock/resource teardown in service tests; use `AfterEach(ctrl.Finish)` in handler tests.
- Mocks live in `test/mocks/`. Run `go generate ./...` when domain interfaces change.
- Keep tests colocated with implementation unless they are full-stack E2E tests.
- Add E2E coverage when API contracts, auth flows, or persistence behaviour change end to end.
