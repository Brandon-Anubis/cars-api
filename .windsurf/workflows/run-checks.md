---
name: Run Full Check Suite
description: Execute the complete quality pipeline — format, lint, vet, test, coverage
---

# Run Full Check Suite

Execute all quality checks in the correct order. Stop on first failure.

## Steps

1. **Format**
```bash
gofmt -l -w .
goimports -l -w .
```

   - If any files were modified, report them

2. **Vet**
```bash
go vet ./...
```
   - Must pass with zero warnings

3. **Lint**
```bash
golangci-lint run ./...
```

   - Must pass (uses `.golangci.yml` config)
   - If golangci-lint not installed, skip with warning

4. **Unit Tests**
```bash
go test -v -race -short -coverprofile=coverage.out ./...
```

   - Must pass with zero failures
   - Report coverage percentage

5. **Coverage Check**
```bash
go tool cover -func=coverage.out
```

   - Report total coverage
   - Warn if below 85%

6. **Build**
```bash
CGO_ENABLED=0 go build -o /dev/null ./cmd/server
```

   - Must compile successfully

7. **Summary**
   - Report pass/fail for each step
   - Report total coverage percentage
   - List any warnings or suggestions
