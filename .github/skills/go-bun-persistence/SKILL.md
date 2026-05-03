---
name: go-bun-persistence
description: Defining models, repositories, and migrations using Bun ORM. Use when creating new tables, adding columns, or implementing data access logic.
---

# Bun Persistence

This project uses **Bun ORM** for PostgreSQL.

## Defining Models
Models live in `internal/model/`.
- Use `bun` tags for column mapping and constraints.
- Always include a factory function `New[Model]FromDomain(e *domain.Entity) *Model` to map from domain entities.
- Always include a `ToDomain() *domain.Entity` method to map to domain entities.
- Use `ApplyXxxUpdate` helper for partial updates (utilizing `github.com/aarondl/opt`).

Example:
```go
type Post struct {
    ID        int64     `bun:"id,pk,autoincrement"`
    Title     string    `bun:"title,notnull"`
    CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
}
```

## Migrations
Migrations live in `db/migrations/` and use `migris`.
- Use `schema` package for a fluent API.
- Always implement `up` and `down`.

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
- Receive `*bun.DB` in constructor.
- Implement interfaces from `internal/domain/`.
- Handle DB-specific errors (e.g., integrity violations) and return domain-friendly errors.

### Common Patterns
- **Insert**: `r.db.NewInsert().Model(m).Exec(ctx)`
- **Select**: `r.db.NewSelect().Model(m).Where("id = ?", id).Scan(ctx)`
- **Update**: `r.db.NewUpdate().Model(m).WherePK().Exec(ctx)`
- **Delete**: `r.db.NewDelete().Model(m).WherePK().Exec(ctx)`
