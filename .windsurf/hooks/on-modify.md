---
name: On Modify — Auto-Format & Lint
trigger: on_modify
file_patterns: ["**/*.go"]
---

### Actions

1. **If SQL migration was modified:**
    - Parse column names and types from the SQL DDL
    - Compare with struct fields in `internal/domain/car.go`
    - WARN if any column in SQL doesn't have a corresponding struct field
    - WARN if any struct field doesn't have a corresponding SQL column
    - Check Spanner type → Go type mapping:
        - `STRING(N)` → `string` or `spanner.NullString`
        - `INT64` → `int64` or `spanner.NullInt64`
        - `FLOAT64` → `float64` or `spanner.NullFloat64`
        - `TIMESTAMP` → `time.Time`
        - `BOOL` → `bool` or `spanner.NullBool`
2. **If domain model was modified:**
    - Same comparison in reverse
    - Suggest SQL migration update if new fields were added
3. **Check Spanner struct tags** — Verify every field has a `spanner:"ColumnName"` tag matching the SQL column name exactly