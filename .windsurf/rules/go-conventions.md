---
trigger: always_on
---

# Go Conventions — Cars API

## Language & Version
- Go 1.22+ with modules enabled
- Use `gofmt` and `goimports` formatting (non-negotiable)
- Follow Effective Go, Google Go Style Guide, and Uber Go Style Guide

## Error Handling
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Never use "failed to" prefix — it piles up in error chains
- Handle errors exactly once: either wrap+return OR log+handle, never both
- Use sentinel errors with `Err` prefix: `var ErrNotFound = errors.New("not found")`
- Custom error types use `Error` suffix: `type ValidationError struct{}`
- Always use comma-ok pattern for type assertions: `v, ok := i.(Type)`
- Return errors; never panic except in truly irrecoverable situations

## Interfaces
- Define interfaces where they are CONSUMED, not where implemented
- Keep interfaces small — prefer single-method interfaces
- Accept interfaces, return concrete types
- Never use pointer to interface (`*Interface` is almost always wrong)
- Verify interface compliance at compile time:
```

var _ CarRepository = (*spannerCarRepo)(nil)

```

## Structs
- Design for zero-value usability
- Use `var s Struct` for zero values, not `s := Struct{}`
- Specify field names in composite literals (never positional)
- Always use JSON struct tags: `json:"field_name"`
- Group related fields logically
- Never embed types in public structs — use composition with explicit delegation
- Keep mutexes as named fields, never embedded: `mu sync.Mutex`

## Functions
- Keep functions focused — one thing, done well
- Return early with guard clauses — success path stays at left margin
- Maximum 3-4 levels of nesting
- Avoid unnecessary `else` after return statements
- Use functional options pattern for complex initialization:
```

type Option func(*Server)

func WithTimeout(d time.Duration) Option { ... }

```

## Context
- Always first parameter: `func DoWork(ctx context.Context, ...)`
- Never store context in structs
- Never pass nil context — use `context.Background()` or `context.TODO()`
- Always `defer cancel()` immediately after creating cancellable contexts
- Check `ctx.Done()` in long-running operations

## Concurrency
- Never fire-and-forget goroutines — always provide cancellation
- Use `sync.WaitGroup` or done channels to wait for goroutine exit
- Channel size should be 0 (unbuffered) or 1 — justify any other size
- Close channels from sender side only
- Use `select` with `ctx.Done()` for cancellation-aware operations
- No goroutines in `init()`

## Naming
- Package names: lowercase, single-word, no underscores
- No getter prefix: `Owner()` not `GetOwner()`, setters can be `SetOwner()`
- MixedCaps not underscores for multi-word names
- Exported = uppercase first letter, unexported = lowercase
- Don't shadow built-in names (`error`, `string`, `len`)

## Imports
- Group: (1) standard library, (2) third-party, (3) local packages
- Separate groups with blank lines
- Use `goimports` to manage automatically

## Performance
- Prefer `strconv` over `fmt` for number conversions
- Specify container capacity when size is known: `make([]T, 0, size)`
- Copy slices and maps at boundaries to prevent mutations
- Call `os.Exit` or `log.Fatal` only in `main()`
- Avoid `init()` functions — initialize in `main()` explicitly