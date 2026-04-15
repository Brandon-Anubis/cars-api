---
trigger: glob
globs: internal/repository/**/*.go, migrations/*.sql, cmd/server/main.go, scripts/*
description: Cloud Spanner client patterns, emulator setup, read/write operations, and error mapping
---

# Cloud Spanner Patterns — Cars API

## Emulator Configuration
- Emulator host: `SPANNER_EMULATOR_HOST` environment variable
- Project: configurable via `SPANNER_PROJECT` env var
- Instance: configurable via `SPANNER_INSTANCE` env var
- Database: configurable via `SPANNER_DATABASE` env var
- Full database path: `projects/{project}/instances/{instance}/databases/{database}`

## Client Initialization
- Create Spanner client in `cmd/server/main.go`, inject into repository
- Use `option.WithoutAuthentication()` when connecting to emulator
- Set `SPANNER_EMULATOR_HOST` to skip TLS/auth automatically
- Always `defer client.Close()`

## Read Operations
- Use `client.Single()` for single reads (no transaction needed)
- Use `client.Single().Read()` for point lookups by primary key
- Use `client.Single().Query()` for SQL queries
- Always use parameterized queries with `spanner.Statement{SQL: ..., Params: ...}`
- Iterate with `iter.Do()` or manual `iter.Next()` + `defer iter.Stop()`

## Write Operations
- Use `client.Apply()` with mutations for simple writes
- Use `client.ReadWriteTransaction()` for transactional operations
- Mutations: `spanner.Insert()`, `spanner.Update()`, `spanner.InsertOrUpdate()`, `spanner.Delete()`
- For conditional updates (check existence first), use read-write transactions

## Schema Conventions
- Primary keys: `STRING(36)` for UUIDs
- Timestamps: `TIMESTAMP` type with `allow_commit_timestamp=true` where appropriate
- Use secondary indexes for common query patterns
- Use `WHERE ... IS NOT NULL` for filtered indexes

## Error Handling
- Check `spanner.ErrCode(err)` for gRPC status codes
- `codes.NotFound` → map to domain `ErrNotFound`
- `codes.AlreadyExists` → map to domain `ErrAlreadyExists`
- Wrap all Spanner errors with repository context before returning

## Testing with Emulator
- Integration tests use `TestMain` to set up emulator connection
- Create a fresh database per test run (or use a shared one with cleanup)
- Set `SPANNER_EMULATOR_HOST=localhost:9010` in test environment
- Use `gcloud` CLI or admin client to create instance/database programmatically