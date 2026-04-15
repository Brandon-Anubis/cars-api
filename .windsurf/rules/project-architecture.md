---
trigger: always_on
---

# Project Architecture — Cars API

## Architecture: Clean Architecture (Hexagonal)

This project follows Clean Architecture with strict dependency rules.
Inner layers NEVER depend on outer layers.

### Layer Structure (inside → outside)
1. **Domain** (`internal/domain/`) — Entities, value objects, business rules, validation
2. **Service** (`internal/service/`) — Use cases, business logic orchestration
3. **Port** (`internal/repository/`) — Interface definitions (contracts)
4. **Adapter** (`internal/repository/spanner/`) — Database implementations
5. **Handler** (`internal/handler/`) — HTTP transport, request/response mapping
6. **Config** (`internal/config/`) — Environment configuration
7. **Entry** (`cmd/server/`) — Application wiring, dependency injection, startup

### Dependency Rules
- `domain` depends on: NOTHING (only stdlib)
- `service` depends on: `domain`, repository interfaces
- `repository/spanner` depends on: `domain`, Spanner SDK
- `handler` depends on: `domain`, `service`
- `cmd/server` depends on: ALL (wires everything together)

### File Organization

cars-api/
├── cmd/server/main.go
├── internal/
│   ├── domain/
│   │   ├── car.go
│   │   ├── car_test.go
│   │   └── errors.go
│   ├── service/
│   │   ├── car_service.go
│   │   └── car_service_test.go
│   ├── repository/
│   │   ├── repository.go          # Interface only
│   │   └── spanner/
│   │       ├── car_repo.go
│   │       ├── car_repo_test.go   # Integration tests
│   │       └── helpers.go
│   ├── handler/
│   │   ├── car_handler.go
│   │   ├── car_handler_test.go
│   │   ├── middleware.go
│   │   ├── response.go            # Standard response helpers
│   │   └── router.go
│   └── config/
│       └── config.go
├── migrations/
│   └── 001_create_cars.sql
├── scripts/
│   ├── [setup-emulator.sh](http://setup-emulator.sh)
│   └── [seed.sh](http://seed.sh)
├── .windsurf/
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── .golangci.yml
├── [README.md](http://README.md)
└── docs/
├── [API.md](http://API.md)
└── [ARCHITECTURE.md](http://ARCHITECTURE.md)

### Key Patterns
- **Repository Pattern**: All data access behind interfaces defined in `repository/repository.go`
- **Dependency Injection**: Constructor injection in `cmd/server/main.go`, no globals
- **Functional Options**: For complex struct initialization
- **Standard Response Envelope**: All API responses use `handler/response.go` helpers
