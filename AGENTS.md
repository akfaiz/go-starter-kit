# Repository Guidelines

## Project Structure & Module Organization
- `main.go` is the CLI entrypoint; commands live in `cmd/` (`serve`, `migrate`, root wiring).
- Core app code is in `internal/`: `delivery/http/` (handlers/routes/middleware/server), `service/` (business logic), `repository/` (data access), plus `domain/`, `model/`, `config/`, and helpers like `hash/`, `validator/`, `logger/`.
- Shared reusable utilities are in `pkg/` (for example `pkg/env`).
- Database migrations are in `db/migrations/`.
- Tests are colocated with code as `*_test.go`; test helpers and generated mocks are in `test/mocks/`.

## Build, Test, and Development Commands
- `make tidy`: sync module dependencies with `go.mod`/`go.sum`.
- `make run`: run the HTTP API locally (`go run . serve`).
- `make test`: run all package tests (`go test ./...`).
- `make migrate-up` / `make migrate-down`: apply or roll back schema migrations.
- `make build`: compile binary to `bin/go-starter-kit`.
- `docker compose up --build`: start local stack using containers.

## Coding Style & Naming Conventions
- Follow standard Go formatting; run `gofmt` (or `go fmt ./...`) before pushing.
- Use Go naming idioms: exported identifiers in `CamelCase`, unexported in `camelCase`, package names short/lowercase.
- Keep layer boundaries explicit (`delivery -> service -> repository`) and avoid bypassing service logic from handlers.
- Use descriptive snake_case filenames for multiword files (for example `auth_handler.go`, `user_repository.go`).

## Architectural Boundaries & Mappings
- **DTO to Domain:** Handlers must convert request DTOs to domain entities using `dto.ToDomain()` before calling services.
- **Domain to Model:** Repositories must convert domain entities to database models using factory functions like `model.New[Entity]FromDomain(entity)`.
- **Model to Domain:** Repositories must convert database models to domain entities using `modelEntity.ToDomain()`.
- **Domain Errors:** Services and internal logic must strictly return **Domain Errors** (defined in `internal/domain/error.go`).
- **Error Mapping:** Handlers are responsible for mapping domain errors to HTTP-specific responses (using `pkg/problem` or `pkg/validator`).
- **Service Isolation:** The service layer must never import `pkg/problem`, `pkg/validator`, or `internal/model`.

## Testing Guidelines
- Primary test command: `make test`.
- Tests use Go `testing` with `testify`; Ginkgo/Gomega suites are also present for some modules.
- Name tests with `TestXxx` and keep them next to implementation files.
- Prefer table-driven tests for service/repository behavior; use `test/mocks/` for dependency isolation.

## Commit & Pull Request Guidelines
- Current history is minimal (`initial project`) and uses short lowercase commit messages.
- Keep commit subjects concise and imperative, optionally scoped, e.g. `auth: validate refresh token expiry`.
- PRs should include purpose and behavioral impact, linked issue/ticket, test evidence (`make test`), and request/response examples when API contracts change.

## Security & Configuration Tips
- Copy `.env.example` to `.env` for local setup; do not commit secrets.
- Run migrations before `make run` to prevent runtime schema errors.
