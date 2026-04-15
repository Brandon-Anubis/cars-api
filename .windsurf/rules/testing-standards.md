---
trigger: glob
globs: "**/*_test.go"
description: Testing conventions including table-driven tests, mocking, integration test setup, and coverage targets
---

# Testing Standards — Cars API

## Coverage Target
- Minimum 85% overall coverage
- 100% coverage on domain validation logic
- Every public function must have at least one test

## Test Organization
- Test files live next to source: `car.go` → `car_test.go`
- Same package for unit tests (access unexported)
- `_test` suffix package for black-box tests when appropriate
- Integration tests use build tag `//go:build integration`

## Unit Tests
- Use table-driven tests for multiple input/output scenarios:

```go
tests := []struct {
    name    string    
    input   CreateCarRequest
    wantErr bool
}{ ... }

for _, tt := range tests {
    [t.Run](http://t.Run)([tt.name](http://tt.name), func(t *testing.T) { ... })
}
```

- Mock interfaces, not implementations
- Use `testify/assert` and `testify/require` for assertions
- `require` for fatal assertions (stops test), `assert` for non-fatal
- Test error paths, not just happy paths
- Test edge cases: empty strings, zero values, negative numbers, max lengths

## Integration Tests
- Require running Spanner emulator
- Use `TestMain` for setup/teardown:

```go
func TestMain(m *testing.M) {

    // setup emulator connection
    code := [m.Run](http://m.Run)()
    // cleanup
    os.Exit(code)
}
```

- Clean up test data between tests
- Test actual database operations: create, read, update, delete
- Test not-found scenarios, duplicate key handling

## Handler Tests
- Use `net/http/httptest` for HTTP testing
- Create test server with mocked service layer
- Verify: status codes, response body structure, headers
- Test: valid requests, invalid JSON, missing fields, not found

## Test Naming
- Format: `Test{Function}_{Scenario}` or `Test{Function}/{Scenario}`
- Examples: `TestNewCar_ValidInput`, `TestNewCar_EmptyMake`, `TestGetByID/NotFound`

## Running Tests
- All tests: `go test -v -race -coverprofile=coverage.out ./...`
- Unit only: `go test -v -race -short ./...`
- Integration: `go test -v -race -tags=integration ./...`
- Always use `-race` flag