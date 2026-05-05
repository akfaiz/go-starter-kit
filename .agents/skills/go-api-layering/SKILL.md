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
  - Use `c.Bind(&req)` for binding and automatic validation.
  - Map domain errors to HTTP errors with `pkg/problem` or `pkg/validator`.
  - Wrap unexpected errors using `problem.Wrap(err, problem.ErrInternalServer)`.
  - Avoid business logic; delegate to the Service layer.
  - Register routes in `internal/delivery/http/routes/routes.go` with `oaswrap` request/response metadata.

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
- **Role**: GORM specific logic, CRUD operations, and mapping between `model` (DB) and `domain` (Business) entities.
- **Rules**:
  - Use `model.New[Entity]FromDomain(domainEntity)` to map Domain -> Model.
  - Use `modelEntity.ToDomain()` to map Model -> Domain.
  - Keep it focused on persistence.
  - Convert known DB outcomes to domain errors, for example duplicate keys to `domain.ErrEmailAlreadyExists` and missing rows to `domain.ErrResourceNotFound`.
  - Attach stack traces to unexpected DB failures with `github.com/cockroachdb/errors`.

### 4. Domain & Model
- **Domain**: `internal/domain/` contains interfaces and business entities used across layers.
- **Model**: `internal/model/` contains GORM-tagged structs representing DB tables.

## Workflow: Adding a New Feature

1. **Define Domain**: Create/update entities and interfaces in `internal/domain/`.
2. **Create Model**: Define the DB struct in `internal/model/` with GORM tags.
3. **Create Migration**: Add a new migration in `db/migrations/`.
4. **Implement Repository**: Create the repository implementation in `internal/repository/`.
5. **Implement Service**: Create the service implementation in `internal/service/`.
6. **Implement Handler**: Create the handler and DTOs in `internal/delivery/http/handler/`.
7. **Wire with FX**: Register constructors in the relevant `module.go` files and expose HTTP routes in `routes.go`.
8. **Add Tests**: Write unit tests for handler/service and E2E tests for the flow.

## Boundary Checklist

- Handler imports may include `pkg/problem`, `pkg/validator`, and DTO packages.
- Service imports should stay domain-oriented and must not import transport or persistence models.
- Repository imports may include `internal/model`, GORM, and domain interfaces/errors.
- Queue handlers should depend on service/domain abstractions, not HTTP handlers or DTOs.
- Run `make test` after behavior changes; run `make lint` when imports or public APIs change.
