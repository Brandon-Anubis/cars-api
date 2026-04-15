---
name: Prepare Commit
description: Format, lint, test, and stage changes for a clean commit
inputs:
  - name: message
    type: string
    description: Conventional commit message (e.g., feat add car domain model)
---

# Prepare Commit

## Steps

1. **Format all modified Go files**
```bash
gofmt -l -w .
goimports -l -w .
```

2. **Run vet**
```bash
go vet ./...
```

3. **Run tests**
```bash
go test -race -short ./...
```

4. **Validate commit message format**
   - Must follow Conventional Commits: `type: description`
   - Valid types: `feat`, `fix`, `test`, `docs`, `build`, `chore`, `refactor`
   - Description must be lowercase, no period at end

5. **Stage and summarize**
   - List all files to be committed
   - Show a summary of changes
   - Suggest the git commands:
   - 
```bash
git add -A
git commit -m "{message}"
```

6. **Post-commit reminder**
   - Remind which phase this completes from the Phased Prompt Playbook
   - State what the next phase is
