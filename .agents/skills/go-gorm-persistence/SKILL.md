---
name: go-gorm-persistence
description: Defining models, repositories, and migrations using GORM. Use when creating new tables, adding columns, or implementing data access logic.
---

# GORM Persistence

This project uses **GORM** for PostgreSQL.

## Defining Models
Models live in `internal/model/`.
- Use `gorm` tags for column mapping and constraints.
- Always include a factory function `New[Model]FromDomain(e *domain.Entity) *Model` to map from domain entities.
- Always include a `ToDomain() *domain.Entity` method to map to domain entities.

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

### Common Patterns
- **Insert**: `gorm.G[model.User](r.db).Create(ctx, m)`
- **Select One**: `gorm.G[model.User](r.db).Where("id = ?", id).First(ctx)`
- **Update**: `gorm.G[model.User](r.db).Where("id = ?", id).Updates(ctx, m)`
- **Delete**: `gorm.G[model.User](r.db).Where("id = ?", id).Delete(ctx)`
- **Count**: `gorm.G[model.User](r.db).Count(ctx, "*")`
- **Find All**: `gorm.G[model.User](r.db).Find(ctx)`
