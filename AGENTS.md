# Repository Guidelines

## Project Structure & Module Organization
- `main.go` is the CLI entrypoint; `cmd/` contains the root wiring plus `serve`, `serve-all`, `queue`, and `migrate`.
- Core app code is in `internal/`: `delivery/http/` (handlers/routes/middleware/server), `delivery/queue/` (workers, client, middleware, handlers), `service/` (business logic), `repository/` (data access), plus `domain/`, `model/`, `config/`, `infra/`, `telemetry/`, `lang/`, `hash/`, and `logger/`.
- Shared reusable utilities are in `pkg/` (for example `pkg/env`, `pkg/problem`, and `pkg/validator`).
- Database migrations are in `db/migrations/`.
- Tests are colocated with code as `*_test.go`; integration/E2E tests live in `test/e2e/`, and shared test helpers live in `test/`.
- Project-specific agent skills live in `.agents/skills/`; consult the relevant skill before changing API layering, persistence, validation/errors, FX wiring, queues, email, tracing, or tests.

## Build, Test, and Development Commands
- `make tidy`: sync module dependencies with `go.mod`/`go.sum`.
- `make fmt`: run `go fmt ./...`.
- `make lint` / `make lint-fix`: run or auto-fix `golangci-lint`.
- `make run`: run the HTTP API locally (`go run . serve`).
- `make test`: run the unit and package test suite via Ginkgo, excluding `cmd`, migrations, mocks, and E2E tests.
- `make test-e2e`: run the end-to-end suite in `test/e2e/`.
- `make coverage` / `make coverage-all`: generate coverage for unit tests or merged unit + E2E coverage.
- `make coverage-html`: render `coverage.html` from the current `coverage.out`.
- `make migrate-up` / `make migrate-down`: apply or roll back schema migrations.
- `make build`: compile binary to `bin/go-starter-kit`.
- `make docker-build` / `make docker-run`: build and run the container image.
- `docker compose up --build`: start the local stack using containers.
- For local non-container development, run `docker compose up -d db redis jaeger`, then `make migrate-up`, then `make run`.

## Coding Style & Naming Conventions
- Follow standard Go formatting; run `gofmt` (or `go fmt ./...`) before pushing.
- Use Go naming idioms: exported identifiers in `CamelCase`, unexported in `camelCase`, package names short/lowercase.
- Keep layer boundaries explicit (`delivery -> service -> repository`) and avoid bypassing service logic from handlers or queue consumers.
- Use descriptive snake_case filenames for multiword files (for example `auth_handler.go`, `user_repository.go`).

## Architectural Boundaries & Mappings
- **DTO to Domain:** Handlers must convert request DTOs to domain entities using `dto.ToDomain()` before calling services.
- **Domain to Model:** Repositories must convert domain entities to database models using factory functions like `model.New[Entity]FromDomain(entity)`.
- **Model to Domain:** Repositories must convert database models to domain entities using `modelEntity.ToDomain()`.
- **Domain Errors:** Services and internal logic must strictly return **Domain Errors** (defined in `internal/domain/error.go`).
- **Error Mapping:** Handlers are responsible for mapping domain errors to HTTP-specific responses (using `pkg/problem` or `pkg/validator`).
- **Service Isolation:** The service layer must never import `pkg/problem`, `pkg/validator`, or `internal/model`.
- **Queue Isolation:** Queue handlers should depend on service/domain abstractions, not on HTTP delivery types.
- **Repository Isolation:** Repositories own GORM calls and database error translation; services should depend on repository interfaces, not concrete storage details.
- **FX Wiring:** New handlers, services, repositories, queue handlers, middleware, and infrastructure clients must be registered in the appropriate FX module.
- **Localization:** User-facing validation and handler messages should use the existing i18n flow when locale matters; currently supported locales are English and Indonesian.
- **Tracing:** Preserve trace propagation through HTTP, queue, Redis, and repository work. Record meaningful errors on spans when adding traced operations.

## Testing Guidelines
- Primary test command: `make test`.
- Tests use Go `testing` with `testify`; Ginkgo/Gomega suites are used in several modules and E2E tests.
- Name plain Go tests with `TestXxx` and keep package tests next to implementation files.
- Prefer table-driven tests for service/repository behavior; use `test/e2e/` for full-stack flows.
- Handler tests should assert HTTP status, response shape, and `application/problem+json` error bodies where applicable.
- Repository and E2E tests may need PostgreSQL/Redis test containers; keep helpers in `test/` instead of duplicating container setup.
- Run `make fmt` after code edits and `make test` for behavior changes. Add `make test-e2e` when an API flow, persistence contract, or infrastructure integration changes.

## Commit & Pull Request Guidelines
- Current history is minimal (`initial project`) and uses short lowercase commit messages.
- Keep commit subjects concise and imperative, optionally scoped, e.g. `auth: validate refresh token expiry`.
- PRs should include purpose and behavioral impact, linked issue/ticket, test evidence (`make test`), and request/response examples when API contracts change.

## Security & Configuration Tips
- Copy `.env.example` to `.env` for local setup; do not commit secrets.
- Keep JWT secrets, database credentials, SMTP credentials, and Redis passwords out of source control.
- Run migrations before `make run` to prevent runtime schema errors.
- Do not log raw tokens, OTP values, passwords, or reset links.
