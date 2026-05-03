# Copilot Instructions

## Build, test, and lint commands

Use the Makefile targets as the default interface:

```bash
make tidy        # go mod tidy
make run         # go run . serve
make build       # go build -o bin/go-starter-kit .
make lint        # golangci-lint run --timeout 5m
make test        # ginkgo run -r --skip-package="internal/mocks,db/migrations,cmd,test/e2e" ./...
make test-e2e    # ginkgo run -r ./test/e2e
make coverage    # ginkgo coverage profile for non-e2e packages
make coverage-all
```

Single-test workflows:

```bash
# single Go test function (works for packages that use testing package directly)
go test ./internal/hash/jwtmanager -run TestJWTManager -v

# single Ginkgo package / focused spec
ginkgo run ./internal/service/auth --focus "RefreshToken"
```

If `ginkgo` is not installed locally:

```bash
go install github.com/onsi/ginkgo/v2/ginkgo@latest
```

## High-level architecture

- CLI entrypoint is `main.go` -> `cmd.Execute`, with two command trees: `serve` and `migrate`.
- `serve` builds an Uber Fx app (`cmd/serve/serve.go`) and wires modules in order: infra (DB/Redis/SMTP), repository, hash, security, service, telemetry, HTTP delivery. The Echo server lifecycle is managed through Fx hooks.
- HTTP routing is declared in `internal/delivery/http/routes/routes.go` using `echov5openapi`; route registration and OpenAPI metadata are centralized there.
- Request flow is layered: handler (`internal/delivery/http/handler`) -> service (`internal/service`) -> repository (`internal/repository`) through interfaces defined in `internal/domain`.
- Persistence is split by concern: PostgreSQL (Bun) for durable entities (`users`, `password_reset_tokens`) and Redis for session tokens + auth rate-limiting state. Auth session validity is enforced by matching JWTs against Redis-stored active tokens.
- Telemetry is cross-cutting: Echo middleware, Bun query hooks, and Redis instrumentation all feed OpenTelemetry (`internal/telemetry`, `internal/infra/database.go`, `internal/infra/redis_client.go`).

## Key conventions

- Domain package owns service/repository interfaces and shared business types; implementations live in layer-specific packages and are injected with Fx modules.
- Auth middleware is provided as a named Fx dependency (`name:"auth"`) and injected into protected route groups in `routes.Register`.
- Handlers are thin: bind + validate request DTOs, convert to domain via `dto.ToDomain()`, call services, and return errors upward. 
- Error Handling Standard:
  - **Services** strictly return **Domain Errors** (see `internal/domain/error.go`).
  - **Handlers** map domain errors to `validator.ValidationError` or `problem.AppError`.
  - **Unexpected errors** are wrapped in handlers using `problem.Wrap(err, problem.ErrInternalServer)`.
- Data Mapping Standard:
  - **DTO -> Domain:** `dto.ToDomain()`
  - **Domain -> Model:** `model.New[Entity]FromDomain(entity)`
  - **Model -> Domain:** `modelEntity.ToDomain()`
- API success responses use the generic envelope `dto.Response[T]` (`status`, `message`, `data`) via `dto.NewResponse` / `dto.NewMessage`.
- Partial updates use `omit` / `omitnull` types in `domain.UserUpdate`, applied in SQL via `model.ApplyUserUpdate` instead of overwriting full records.

- i18n is expected in runtime and tests: `lang.Init()` is called in startup and test suites, and request locale is injected from `Accept-Language` middleware (default `en`).
- Mocks are generated from domain interfaces using `//go:generate mockgen ...` directives into `test/mocks/`; service/handler tests consume those mocks heavily.
