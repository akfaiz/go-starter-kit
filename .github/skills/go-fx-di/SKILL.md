---
name: go-fx-di
description: Adding new components and modules using Uber FX. Use when adding new services, repositories, handlers, or infrastructure clients.
---

# Dependency Injection with Uber FX

This project uses **Uber FX** for dependency injection and application lifecycle management.

## Module Structure
Components are organized into modules:
- `internal/delivery/http/module.go`: Handlers and server wiring.
- `internal/service/module.go`: Business services.
- `internal/repository/module.go`: Repositories.
- `internal/infra/module.go`: Infrastructure clients (DB, Redis, SMTP).

## Adding a New Component

1. **Implement Constructor**: Create a `NewXxx` function that returns the implementation or interface.
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

3. **Dependency Resolution**: FX automatically resolves dependencies based on types. If you need a specific implementation for an interface, ensure the constructor returns the interface type.

## Using `fx.Invoke`
Use `fx.Invoke` for components that need to be initialized but aren't explicitly requested by other components (e.g., registering routes).

```go
fx.Invoke(routes.Register)
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
