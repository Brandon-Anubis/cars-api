---
name: On Save — Go File Quality
trigger: on_save
file_patterns: ["**/*.go"]
---

### Actions

1. **Format the saved file** — Run `gofmt` and `goimports`
2. **Vet the package** — Run `go vet` on the containing package
3. **Quick compile check** — Run `go build ./...`
4. **Architecture violation check:**
    - If file is in `internal/domain/`: Verify it does NOT import any package outside stdlib (except `github.com/google/uuid`)
    - If file is in `internal/service/`: Verify it does NOT import `internal/handler/` or `internal/repository/spanner/`
    - If file is in `internal/handler/`: Verify it does NOT import `internal/repository/spanner/`

### Error Reporting

- Prefix with `[ARCH]` for architecture violations
- Prefix with `[FMT]` for formatting issues
- Prefix with `[VET]` for vet warnings