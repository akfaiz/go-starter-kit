---
name: go-api-layering
description: "Delivery -> service -> repository boundaries for this Go API. Use when adding endpoints, features, DTO/domain mappings, service methods, repositories, or business logic."
---

# Go API Layering

This project follows `delivery -> service -> repository`, with `internal/domain` as the contract layer between them.

## Layer Responsibilities

### 1. Delivery (HTTP Handlers)
- **Location**: `internal/delivery/http/handler/`
- **Role**: Handle HTTP requests, bind/validate DTOs, call service methods, and return response DTOs.
- **Rules**:
  - Put request/response structs in `internal/delivery/http/handler/dto/`.
  - Add `ToDomain()` on request DTOs that cross into services; never pass HTTP DTOs into services.
  - Use `c.Bind(&req)` for binding and automatic validation (the custom binder validates on bind).
  - Map domain errors to HTTP errors with `pkg/problem` or `pkg/validator`.
  - Wrap unexpected errors using `problem.Wrap(err, problem.ErrInternalServer)`.
  - Return responses using `dto.NewResponse(status, data, msg...)` or `dto.NewMessage(status, msg)`.
  - Avoid business logic; delegate to the Service layer.
  - Register routes in `internal/delivery/http/routes/routes.go` using the `oaswrap` router.

### 2. Service (Business Logic)
- **Location**: `internal/service/`
- **Role**: Orchestrate business rules, call multiple repositories, and handle domain logic.
- **Rules**:
  - Depend on interfaces defined in `internal/domain/`.
  - Accept and return domain entities.
  - Return domain errors from `internal/domain/error.go` for expected business outcomes.
  - Do not import `pkg/problem`, `pkg/validator`, `internal/model`, or HTTP delivery packages.
  - If a service enqueues work, keep the queue payload explicit and still expose domain-level behavior.

### 3. Repository (Data Access)
- **Location**: `internal/repository/`
- **Role**: GORM-specific logic, CRUD operations, and mapping between `model` (DB) and `domain` (Business) entities.
- **Rules**:
  - Use `model.NewXxxFromDomain(domainEntity)` to map Domain → Model.
  - Use `modelEntity.ToDomain()` to map Model → Domain.
  - Keep it focused on persistence.
  - Convert known DB outcomes to domain errors (e.g. duplicate keys → `domain.ErrEmailAlreadyExists`, missing rows → `domain.ErrResourceNotFound`).
  - Attach stack traces to unexpected DB failures with `cerrors.WithStack(err)` (`github.com/cockroachdb/errors`).

### 4. Domain & Model
- **Domain** (`internal/domain/`): each file defines both the entity struct **and** the repository/service interfaces for that entity. Add `//go:generate mockgen` at the top of any new domain file.
- **Model** (`internal/model/`): GORM-tagged structs. Each model must have `NewXxxFromDomain(e *domain.Xxx) *Model` factory and `ToDomain() *domain.Xxx` method.

## Route Registration (oaswrap)

Routes are registered in `internal/delivery/http/routes/routes.go`. `Register` receives all handlers via an `fx.In` struct. Use `echov5openapi.NewRouter` and attach metadata with `option.*` helpers:

```go
func Register(rc RouteConfig) {
    r := echov5openapi.NewRouter(rc.Echo,
        option.WithTitle("Go Starter Kit API"),
        option.WithVersion("1.0.0"),
        option.WithSecurity("bearerAuth", option.SecurityHTTPBearer("Bearer")),
    )

    v1 := r.Group("/api/v1")
    auth := v1.Group("/auth").With(option.GroupTags("Authentication"))
    auth.POST("/register", rc.AuthHandler.Register).With(
        option.Summary("User Registration"),
        option.Request(new(dto.RegisterRequest)),
        option.Response(201, new(dto.Response[dto.TokenResponse])),
        option.Response(422, new(dto.ValidationErrorResponse)),
    )
}
```

The `RouteConfig` struct uses `fx.In` to receive dependencies; named middleware (e.g. auth) uses the `name:"auth"` tag:

```go
type RouteConfig struct {
    fx.In
    Echo           *echo.Echo
    Config         config.Config
    AuthMiddleware echo.MiddlewareFunc `name:"auth"`
    AuthHandler    *handler.AuthHandler
}
```

## FX Wiring

Each layer has a `module.go` exporting `var Module = fx.Module(...)`. Handlers and services use `fx.Provide`; route/queue registration uses `fx.Invoke`. The HTTP module nests `handler.Module`:

```go
// internal/delivery/http/module.go
var Module = fx.Module("http",
    fx.Provide(server.New, middleware.New),
    fx.Invoke(routes.Register),
    handler.Module,
)

// internal/delivery/http/handler/module.go
var Module = fx.Module("handler",
    fx.Provide(NewAuthHandler, NewProfileHandler, NewUserHandler, NewHealthCheckHandler),
)

// internal/service/module.go
var Module = fx.Module("service",
    fx.Provide(auth.NewService, user.NewService),
)

// internal/repository/module.go
var Module = fx.Module("repository",
    fx.Provide(user.NewRepository, session.NewRepository, passwordresettoken.NewRepository),
)
```

## Workflow: Adding a New Feature

1. **Define Domain**: Create/update entities and interfaces in `internal/domain/`. Add `//go:generate mockgen -source=xxx.go -destination=../../test/mocks/xxx_mock.go -package=mocks` at the top of the file.
2. **Create Model**: Define the DB struct in `internal/model/` with GORM tags, a `NewXxxFromDomain` factory, and a `ToDomain` method.
3. **Create Migration**: Add a new migration in `db/migrations/`.
4. **Implement Repository**: Create the repository in `internal/repository/` and register it in `repository/module.go`.
5. **Implement Service**: Create the service in `internal/service/` and register it in `service/module.go`.
6. **Implement Handler + DTOs**: Create handler and DTO structs in `internal/delivery/http/handler/`. Add `ToDomain()` to request DTOs that are passed to services.
7. **Register Routes**: Add the route to `routes/routes.go` with full `oaswrap` metadata. Add the handler to `handler/module.go`.
8. **Add Tests**: Write Ginkgo unit tests for handler and service; add E2E tests for the flow.

## Boundary Checklist

- Handler imports may include `pkg/problem`, `pkg/validator`, and DTO packages.
- Service imports must stay domain-oriented; no transport or persistence model imports.
- Repository imports may include `internal/model`, GORM, and domain interfaces/errors.
- Queue handlers depend on service/domain abstractions, not HTTP handlers or DTOs.
- Run `make test` after behaviour changes; run `make lint` when imports or public APIs change.
