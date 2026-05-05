---
name: go-fx-di
description: "Uber FX wiring for services, repositories, HTTP handlers, queue workers, middleware, infrastructure clients, and lifecycle hooks. Use when adding constructors or modules."
---

# Dependency Injection with Uber FX

This project uses **Uber FX** for dependency injection and application lifecycle management.

## Module Structure
Components are organized into modules:
- `internal/delivery/http/module.go`: Handlers and server wiring.
- `internal/delivery/http/handler/module.go`: HTTP handler constructors.
- `internal/delivery/http/routes/routes.go`: Route registration through `fx.Invoke(routes.Register)`.
- `internal/delivery/queue/module.go`: Queue client and worker modules.
- `internal/service/module.go`: Business services.
- `internal/repository/module.go`: Repositories.
- `internal/infra/module.go`: Infrastructure clients such as DB, Redis, and SMTP.
- `internal/hash/module.go`: Hashing and JWT implementations.
- `internal/telemetry/module.go`: Tracer provider and lifecycle shutdown.

## Adding a New Component

1. **Implement Constructor**: Create a `NewXxx` function that returns the implementation or the domain interface.
   ```go
   func NewService(repo domain.Repository) domain.Service {
       return &service{repo: repo}
   }
   ```

2. **Register in Module**: Add the constructor to `fx.Provide` in the relevant `module.go`.
   ```go
   var Module = fx.Module("service",
       fx.Provide(
           NewService,
           // ...
       ),
   )
   ```

3. **Dependency Resolution**: FX resolves dependencies by type. If callers depend on a domain interface, make the constructor return that interface type.

4. **Named Values**: When a dependency is annotated with an FX name, keep the provider and consumer names aligned. The auth middleware is consumed as an `echo.MiddlewareFunc` named `auth`.

## Using `fx.Invoke`
Use `fx.Invoke` for components that need to be initialized but aren't explicitly requested by other components (e.g., registering routes).

```go
fx.Invoke(routes.Register)
```

Queue workers also use invoke-based registration:

```go
fx.Invoke(Register)
fx.Invoke(server.RegisterServer)
```

## Lifecycle Hooks
If a component needs to perform actions on start or stop, use `fx.Lifecycle`.

```go
func NewServer(lc fx.Lifecycle) *Server {
    s := &Server{}
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error { return s.Start() },
        OnStop:  func(ctx context.Context) error { return s.Stop() },
    })
    return s
}
```

## Practical Rules

- Register each constructor once in the smallest relevant module.
- Return interfaces at boundaries (`domain.UserService`, `domain.UserRepository`, `domain.Mailer`) when the rest of the app should not depend on concrete implementations.
- Keep HTTP route metadata in `routes.go`; do not hide route registration inside handlers.
- After wiring changes, run `make test` or at least the package tests for the affected module.
