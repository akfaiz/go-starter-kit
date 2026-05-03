---
name: bun-migration-data
description: >
  Use this skill when changing schema, Bun models, SQL mapping, or repository
  persistence behavior that touches PostgreSQL migrations and data updates.
---

# Bun Model and Migration Discipline

## Purpose

Keep schema, model mapping, and repository behavior aligned.

## Apply these rules

1. Add/modify schema in `db/migrations/` for durable DB changes.
2. Keep Bun model structs in `internal/model/` synced with migration changes.
3. Update repository queries and scans to match schema/model updates.
4. Preserve partial-update semantics (`omit` / `omitnull`, `model.ApplyUserUpdate`).
5. Avoid mixing persistence concerns into handlers or routing.

## Verification commands

```bash
make migrate-up
make build
make test
```
