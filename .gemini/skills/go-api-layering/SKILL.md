---
name: go-api-layering
description: Guidance on the delivery -> service -> repository flow and domain-driven design in this Go API. Use when adding new features, endpoints, or business logic.
---

# Go API Layering

This project follows a structured layering pattern: `Delivery (HTTP) -> Service (Business Logic) -> Repository (Data Access)`.

## Layer Responsibilities

### 1. Delivery (HTTP Handlers)
- **Location**: `internal/delivery/http/handler/`
- **Role**: Handle HTTP requests, bind/validate DTOs, call service methods, and return response DTOs.
- **Rules**:
  - Use `dto` package for request/response structures.
  - Return `error` from handlers; Echo middleware handles mapping to RFC 7807.
  - Avoid business logic; delegate to the Service layer.

### 2. Service (Business Logic)
- **Location**: `internal/service/`
- **Role**: Orchestrate business rules, call multiple repositories, handle domain logic, and map errors to validation/domain errors.
- **Rules**:
  - Depend on interfaces defined in `internal/domain/`.
  - Return domain entities or result pairs.

### 3. Repository (Data Access)
- **Location**: `internal/repository/`
- **Role**: Bun ORM specific logic, CRUD operations, and mapping between `model` (DB) and `domain` (Business) entities.
- **Rules**:
  - Keep it focused on persistence.
  - Handle DB-specific errors (e.g., unique constraints) and wrap them in domain errors.

### 4. Domain & Model
- **Domain**: `internal/domain/` contains interfaces and business entities used across layers.
- **Model**: `internal/model/` contains Bun-tagged structs representing DB tables.

## Workflow: Adding a New Feature

1. **Define Domain**: Create/update entities and interfaces in `internal/domain/`.
2. **Create Model**: Define the DB struct in `internal/model/` with Bun tags.
3. **Create Migration**: Add a new migration in `db/migrations/`.
4. **Implement Repository**: Create the repository implementation in `internal/repository/`.
5. **Implement Service**: Create the service implementation in `internal/service/`.
6. **Implement Handler**: Create the handler and DTOs in `internal/delivery/http/handler/`.
7. **Wire with FX**: Register the new components in the respective `module.go` files and `routes.go`.
8. **Add Tests**: Write unit tests for handler/service and E2E tests for the flow.
