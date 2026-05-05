---
name: go-gorm-persistence
description: "GORM models, repositories, migris migrations, and domain/model mapping. Use when creating tables, changing schema, adding persistence methods, or handling database errors."
---

# GORM Persistence

This project uses **GORM** with the [generics API](https://gorm.io/docs/the_generics_way.html) (`gorm.G[T]`) for PostgreSQL (requires GORM >= v1.30.0).

## Defining Models

Models live in `internal/model/`.
- Use `gorm` tags for column mapping and constraints.
- Always provide a `NewXxxFromDomain(e *domain.Xxx) *Model` factory to map Domain → Model.
- Always provide a `ToDomain() *domain.Xxx` method to map Model → Domain.
- Keep validation/API tags out of models; models are database concerns only.

```go
type User struct {
    ID              int64      `gorm:"primaryKey;autoIncrement"`
    Name            string     `gorm:"not null"`
    Email           string     `gorm:"uniqueIndex;not null"`
    Password        string     `gorm:"not null"`
    EmailVerifiedAt *time.Time `gorm:"index"`
    CreatedAt       time.Time  `gorm:"not null"`
    UpdatedAt       time.Time  `gorm:"not null"`
}

func NewUserFromDomain(u *domain.User) *User { ... }
func (u *User) ToDomain() *domain.User       { ... }
```

## Migrations

Migrations live in `db/migrations/` and use `migris` with the `schema` fluent API.
- **Create a new migration file**: `go run . migrate create --name=create_posts_table` — this generates a timestamped file in `db/migrations/`.
- Always implement both `up` and `down` functions.
- Register in `init()` via `migris.AddMigrationContext(up, down)`.

```go
func init() {
    migris.AddMigrationContext(upCreatePostsTable, downCreatePostsTable)
}

func upCreatePostsTable(c schema.Context) error {
    return schema.Create(c, "posts", func(t *schema.Blueprint) {
        t.ID()
        t.String("title")
        t.BigInteger("user_id").Index()
        t.Timestamp("published_at").Nullable()
        t.Timestamp("created_at").UseCurrent()
        t.Timestamps()

        t.Foreign("user_id").References("id").On("users").CascadeOnDelete()
        t.Unique("user_id")
    })
}

func downCreatePostsTable(c schema.Context) error {
    return schema.DropIfExists(c, "posts")
}
```

**Common blueprint helpers:**

| Method | Notes |
|--------|-------|
| `t.ID()` | `BIGSERIAL PRIMARY KEY` |
| `t.String("col")` | `VARCHAR` |
| `t.BigInteger("col")` | `BIGINT` |
| `t.Timestamp("col")` | Add `.Nullable()` or `.UseCurrent()` as needed |
| `t.Timestamps()` | Adds `created_at` + `updated_at` |
| `t.Foreign("col").References("id").On("table").CascadeOnDelete()` | FK with cascade |
| `t.Unique("col")` | Unique constraint |

## Repository Pattern

Repositories live in `internal/repository/`. Each file has the same structure:

```go
var tracer = otel.Tracer("user-repository") // one per file

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) domain.UserRepository {
    return &repository{db: db}
}
```

Constructors return the **domain interface**, not the concrete struct.

## Generic Query API (`gorm.G`)

`gorm.G[Model](db, ...options)` returns a type-safe query builder. Every call creates a fresh instance — no SQL pollution between calls. The `Save` and `FirstOrCreate` methods are intentionally absent from the generics API.

```go
// Create — after Create, reflect ID/timestamps back to the domain entity
m := model.NewUserFromDomain(user)
if err := gorm.G[model.User](r.db).Create(ctx, m); err != nil { ... }
user.ID = m.ID
user.CreatedAt = m.CreatedAt
user.UpdatedAt = m.UpdatedAt

// Upsert — pass clause.OnConflict as a second argument to gorm.G
err := gorm.G[model.PasswordResetToken](r.db, clause.OnConflict{
    Columns:   []clause.Column{{Name: "user_id"}},
    DoUpdates: clause.AssignmentColumns([]string{"token", "expires_at"}),
}).Create(ctx, m)

// Select one
user, err := gorm.G[model.User](r.db).Where("email = ?", email).First(ctx)

// Select many
users, err := gorm.G[model.User](r.db).Where("age <= ?", 18).Find(ctx)

// Update
gorm.G[model.User](r.db).Where("id = ?", id).Updates(ctx, model.User{Name: "Alice"})

// Delete — returns (rowsAffected int64, err error)
_, err = gorm.G[model.User](r.db).Where("id = ?", id).Delete(ctx)

// Count — run BEFORE applying Limit/Offset
total, err := query.Count(ctx, "*")
```

## Error Handling

```go
// Not found
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, domain.ErrResourceNotFound
}

// Unique constraint violation — match on the constraint name string
if strings.Contains(err.Error(), "users_email_unique") {
    return domain.ErrEmailAlreadyExists
}

// Unexpected error — attach stack trace
return cerrors.WithStack(err)   // import cerrors "github.com/cockroachdb/errors"
```

Do **not** use `pgconn.PgError` type assertions; use `strings.Contains` on the constraint name.

## Pagination Pattern

```go
query := gorm.G[model.User](r.db).Where("1=1")

// 1. Apply filters to query
if params.Search != "" {
    query = query.Where("name ILIKE ?", "%"+params.Search+"%")
}

// 2. Count before paginating
total, err := query.Count(ctx, "*")

// 3. Apply sort, limit, offset
users, err := query.
    Limit(params.Limit).
    Offset((params.Page - 1) * params.Limit).
    Find(ctx)

// 4. Return paginated domain result
return domain.NewPagination(params.Page, params.Limit, total, domainUsers), nil
```

## Tracing

Every repository method must start a span:

```go
func (r *repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    ctx, span := telemetry.StartSpan(ctx, tracer)
    defer span.End()
    // ...
}
```

## Verification

- Repository tests use Testcontainers via `test.NewDBContainer` (see `go-testing-patterns` skill).
- Truncate tables in `BeforeEach` with `dbContainer.TruncateAll(ctx)`.
- Run `make test` after repository or migration changes.
