---
name: go-gorm-persistence
description: "GORM models, repositories, migris migrations, and domain/model mapping. Use when creating tables, changing schema, adding persistence methods, or handling database errors."
---

# GORM Persistence

This project uses **GORM** for PostgreSQL.

## Defining Models
Models live in `internal/model/`.
- Use `gorm` tags for column mapping and constraints.
- Always include a factory function `New[Model]FromDomain(e *domain.Entity) *Model` to map from domain entities.
- Always include a `ToDomain() *domain.Entity` method to map to domain entities.
- Keep API/validation DTO tags out of models; models are database concerns only.

Example:
```go
type Post struct {
    ID        int64     `gorm:"primaryKey;autoIncrement"`
    Title     string    `gorm:"not null"`
    CreatedAt time.Time `gorm:"not null"`
}
```

## Migrations
Migrations live in `db/migrations/` and use `migris`.
- Migrations are written in Go.
- Use `schema` package for a fluent API (provided by migris).
- Always implement `up` and `down`.
- Register migrations in `init()` with `migris.AddMigrationContext`.
- Name files with a timestamp prefix, matching existing migrations.

Example:
```go
func upCreatePostsTable(c schema.Context) error {
    return schema.Create(c, "posts", func(table *schema.Blueprint) {
        table.ID()
        table.String("title")
        table.Timestamps()
    })
}
```

## Repository Pattern
Repositories live in `internal/repository/`.
- Receive `*gorm.DB` in constructor.
- Implement interfaces from `internal/domain/`.
- Handle DB-specific errors (e.g., integrity violations) and return domain-friendly errors.
- Use the generic helper `gorm.G[model.Entity](db)` to get a typed chain.
- Convert domain input to model before writes and model results back to domain before returning.
- Use `errors.Is(err, gorm.ErrRecordNotFound)` to return `domain.ErrResourceNotFound`.
- Use `pgconn.PgError` or equivalent checks for PostgreSQL constraint errors when mapping conflicts.
- Wrap unexpected DB errors with `github.com/cockroachdb/errors.WithStack`.

### Common Patterns
- **Insert**: `gorm.G[model.User](r.db).Create(ctx, m)`
- **Select One**: `gorm.G[model.User](r.db).Where("id = ?", id).First(ctx)`
- **Update**: `gorm.G[model.User](r.db).Where("id = ?", id).Updates(ctx, m)`
- **Delete**: `gorm.G[model.User](r.db).Where("id = ?", id).Delete(ctx)`
- **Count**: `gorm.G[model.User](r.db).Count(ctx, "*")`
- **Find All**: `gorm.G[model.User](r.db).Find(ctx)`

## Pagination Pattern

- Accept `domain.FindAllParams` from the service layer.
- Build a scoped GORM query for filters.
- Run `Count(ctx, "*")` before applying `Limit`/`Offset`.
- Return `domain.NewPagination(page, limit, total)` with the domain entities.

## Verification

- Repository tests use Testcontainers via `test.NewDBContainer`.
- Truncate tables in `BeforeEach` with `dbContainer.TruncateAll(ctx)`.
- Run `make test` after repository or migration changes.
