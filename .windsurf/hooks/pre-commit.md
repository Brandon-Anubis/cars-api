---
name: Pre Commit — Go Quality Checks
trigger: pre_commit
---

### Checks (in order)

1. **No TODO/FIXME in committed code** — Scan staged `.go` files for `TODO`, `FIXME`, `HACK`, `XXX` comments. BLOCK if found.
2. **Format verification** — Run `gofmt -l` on staged Go files. BLOCK if any need formatting.
3. **Vet check** — Run `go vet ./...`. BLOCK on any errors.
4. **Unit tests pass** — Run `go test -race -short ./...`. BLOCK if any fail.
5. **Build succeeds** — Run `CGO_ENABLED=0 go build ./cmd/server`. BLOCK if build fails.
6. **Commit message format** — Verify Conventional Commits: `type: description`. WARN (don't block) if format doesn't match.

### On Success

- Log: "✅ All pre-commit checks passed"

### On Failure

- Log which check(s) failed with specific errors
- Suggest fix commands
- BLOCK the commit