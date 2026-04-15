---
name: Scaffold Architecture Layer
description: Generate a new file in the correct architecture layer with boilerplate, imports, and tests
inputs:
  - name: layer
    type: enum
    values: [domain, service, repository, handler]
    description: Which architecture layer to scaffold
  - name: entity
    type: string
    description: Entity name (e.g., "car")
---

# Scaffold Architecture Layer

Generate production-ready boilerplate for a new entity in the specified layer.

## Steps

1. **Validate inputs**
   - Entity name must be lowercase, single word
   - Layer must be one of: domain, service, repository, handler

2. **Generate files based on layer**

   ### If layer = domain
   - Create `internal/domain/{entity}.go`:
     - Struct with JSON tags and common fields (ID, CreatedAt, UpdatedAt)
     - `New{Entity}()` constructor with validation
     - `Validate()` method with comprehensive checks
     - Sentinel errors: `Err{Entity}NotFound`, `Err{Entity}Validation`
   - Create `internal/domain/{entity}_test.go`:
     - Table-driven tests for `New{Entity}()` and `Validate()`
     - Happy path + all error paths

   ### If layer = service
   - Create `internal/service/{entity}_service.go`:
     - Service struct accepting repository interface via constructor
     - CRUD methods: Create, GetByID, List, Update, Delete
     - Each method accepts `context.Context` as first param
     - Business logic validation before repository calls
   - Create `internal/service/{entity}_service_test.go`:
     - Mock repository using interface
     - Table-driven tests for each CRUD method
     - Test both success and error paths

   ### If layer = repository
   - Create `internal/repository/repository.go` (if not exists):
     - `{Entity}Repository` interface with CRUD methods
     - All methods accept `context.Context` as first param
   - Create `internal/repository/spanner/{entity}_repo.go`:
     - Struct implementing repository interface
     - Constructor accepting `*spanner.Client`
     - Compile-time interface verification
     - Proper error mapping (Spanner codes → domain errors)

   ### If layer = handler
   - Create `internal/handler/{entity}_handler.go`:
     - Handler struct accepting service interface via constructor
     - HTTP handler methods for each CRUD endpoint
     - Request parsing with validation
     - Standard response envelope usage
   - Create `internal/handler/{entity}_handler_test.go`:
     - httptest-based tests with mocked service
     - Test each endpoint: success, validation error, not found, server error

3. **Verify**
   - Run `goimports` on generated files
   - Run `go vet ./...` to verify compilation
   - Run generated tests: `go test -v ./{generated_path}/...`