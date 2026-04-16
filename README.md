# Cars API

<!-- [![CI](https://github.com/Brandon-Anubis/cars-api/actions/workflows/ci.yml/badge.svg)](https://github.com/Brandon-Anubis/cars-api/actions/workflows/ci.yml) -->
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Spanner](https://img.shields.io/badge/Cloud_Spanner-Emulated-4285F4?logo=googlecloud&logoColor=white)](https://cloud.google.com/spanner/docs/emulator)

Production-grade Go CRUD API for car inventory management. Clean Architecture, Cloud Spanner, comprehensive testing.

> **Status:**  Actively in development. Check commit history for progress.

---

## Table of Contents

- [Cars API](#cars-api)
  - [Table of Contents](#table-of-contents)
  - [Architecture](#architecture)
  - [Prerequisites](#prerequisites)
  - [Quick Start](#quick-start)
  - [Manual Setup](#manual-setup)
  - [API Reference](#api-reference)
    - [Endpoints](#endpoints)
    - [Pagination](#pagination)
    - [Response Examples](#response-examples)
    - [Error Codes](#error-codes)
  - [Testing](#testing)
  - [Project Structure](#project-structure)
    - [Directory Structure](#directory-structure)
  - [Configuration](#configuration)
  - [Design Decisions](#design-decisions)
  - [What I Would Add With More Time](#what-i-would-add-with-more-time)
  - [Author](#author)

---

## Architecture

Clean Architecture with strict dependency rules. Inner layers never depend on outer layers.

```
┌─────────────────────────────────────┐
│           cmd/server                │  Entry point, wiring, DI
└───────────────┬─────────────────────┘
                │
┌───────────────▼─────────────────────┐
│  ┌─────────┐  ┌─────────────────┐   │
│  │ handler │  │     config      │   │  HTTP transport, middleware
│  └────┬────┘  └─────────────────┘   │
└───────┼─────────────────────────────┘
        │
┌───────▼─────────────────────────────┐
│           service                   │  Business logic, use cases
└───────┬─────────────────────────────┘
        │
┌───────▼─────────────────────────────┐
│      repository (interface)       │  Port - abstraction
└───────┬─────────────────────────────┘
        │
┌───────▼─────────────────────────────┐
│       repository/spanner            │  Adapter - implementation
└───────┬─────────────────────────────┘
        │
┌───────▼─────────────────────────────┐
│            domain                   │  Entities, rules (zero deps)
└─────────────────────────────────────┘
```

**Dependency flow:**

- `domain` imports nothing (stdlib + uuid only)
- `service` imports `domain` + repository interface
- `handler` imports `domain` + `service` 
- `repository/spanner` imports `domain` + Spanner SDK
- `cmd/server` wires everything together

Swapping Spanner for another database means writing one new adapter. Zero changes to business logic.

---

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | 1.26+ | Language runtime |
| [Docker](https://docs.docker.com/get-docker/) | 20.10+ | Container runtime |
| [Docker Compose](https://docs.docker.com/compose/install/) | v2+ | Local development stack |
| [golangci-lint](https://golangci-lint.run/docs/welcome/install/local/) | v2 | Linting (optional for `make lint`) |

---

## Quick Start

One command runs the entire stack (Spanner emulator + API):

```bash
docker compose up --build
```

The API is available at `http://localhost:8080`. The Spanner emulator starts automatically and the schema is applied on startup.

**Verify it works:**

```bash
curl http://localhost:8080/health

# {"data":{"status":"ok"},"meta":{"request_id":"..."}}
```

**Create a car:**

```bash
curl -X POST http://localhost:8080/api/v1/cars \
  -H "Content-Type: application/json" \
  -d '{
    "make": "Toyota",
    "model": "Camry",
    "year": 2024,
    "color": "Blue"
  }'
```

**List all cars:**

```bash
curl http://localhost:8080/api/v1/cars
```

**Stop everything:**

```bash
docker compose down
```

---

## Manual Setup

Running without Docker:

```bash
# 1. Start the Spanner emulator
docker run -d --name spanner-emulator \
  -p 9010:9010 -p 9020:9020 \
  gcr.io/cloud-spanner-emulator/emulator

# 2. Set environment variables
export SPANNER_EMULATOR_HOST=localhost:9010
export SPANNER_PROJECT=cars-project
export SPANNER_INSTANCE=cars-instance
export SPANNER_DATABASE=cars-db

# 3. Build and run
make build
./bin/cars-api
```

The server creates the Spanner instance, database, and schema automatically when `SPANNER_EMULATOR_HOST` is set.

---

## API Reference

All responses use a standard envelope: `data`, `error`, and `meta` fields.

### Endpoints

| Method | Path | Description | Success |
|--------|------|-------------|---------|
| `POST` | `/api/v1/cars` | Create a car | `201` |
| `GET` | `/api/v1/cars` | List cars (paginated) | `200` |
| `GET` | `/api/v1/cars/{id}` | Get car by ID | `200` |
| `PUT` | `/api/v1/cars/{id}` | Update a car | `200` |
| `DELETE` | `/api/v1/cars/{id}` | Delete a car | `204` |
| `GET` | `/health` | Health check | `200` |

### Pagination

```bash
GET /api/v1/cars?limit=20&offset=0
```

Default limit: 20. Max limit: 100.

### Response Examples

**Success (single resource):**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "make": "Toyota",
    "model": "Camry",
    "year": 2024,
    "color": "Blue",
    "created_at": "2026-04-15T18:00:00Z",
    "updated_at": "2026-04-15T18:00:00Z"
  },
  "meta": {
    "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}
```

**Success (collection):**

```json
{
  "data": [
    { "id": "...", "make": "Toyota", "model": "Camry", "year": 2024, "color": "Blue" }
  ],
  "meta": {
    "request_id": "...",
    "total": 42,
    "limit": 20,
    "offset": 0
  }
}
```

**Error:**

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "car not found"
  },
  "meta": {
    "request_id": "..."
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|------------|-------------|
| `VALIDATION_ERROR` | 400 | Input validation failed |
| `INVALID_JSON` | 400 | Request body is not valid JSON |
| `NOT_FOUND` | 404 | Resource does not exist |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Testing

```bash
# Run all unit tests
make test

# Unit tests only (skip integration)
make test-unit

# Integration tests (requires Spanner emulator)
docker compose up spanner-emulator -d
make test-integration

# Full quality pipeline (fmt, vet, lint, test, build)
make check

# View coverage report
make coverage
```

**Coverage target:** 85%+

**Test layers:**

- **Unit tests** - Domain validation, service logic (mocked repo), handler HTTP (mocked service)
- **Integration tests** - Repository operations against real Spanner emulator
- **Middleware tests** - Request ID, recovery, logging

Integration tests use `//go:build integration` and are skipped when the emulator is not running.

---

## Project Structure

```
cars-api/
```

### Directory Structure

```
cars-api/
```

---

## Configuration

All configuration via environment variables. No config files. No hardcoded values.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `SPANNER_PROJECT` | `cars-project` | GCP project ID |
| `SPANNER_INSTANCE` | `cars-instance` | Spanner instance name |
| `SPANNER_DATABASE` | `cars-db` | Spanner database name |
| `SPANNER_EMULATOR_HOST` | *(none)* | Emulator address. When set, skips TLS/auth. |

---

## Design Decisions

| Decision | Choice | Why |
|----------|--------|-----|
| **Database** | Cloud Spanner (emulated) | GCP-native relational DB. Demonstrates fluency with Google Cloud enterprise services. Emulator gives local dev parity with production. |
| **Architecture** | Clean / Hexagonal | Strict layer separation. Swap the database adapter without touching business logic. This is how platform services should be built. |
| **Router** | chi v5 | Lightweight, idiomatic, `net/http` compatible. Most popular Go router in production. |
| **Error handling** | Sentinels + `ValidationError` with `Unwrap()` | `errors.Is()` for category matching. `errors.As()` for field-level detail. Idiomatic Go 1.13+. |
| **Logging** | `slog` (stdlib) | Go standard structured logger. JSON output with request IDs for log correlation. Zero external deps. |
| **Testing** | 3-layer (unit + integration + E2E) | Unit tests are fast (mocked interfaces). Integration tests catch real SDK issues. Build tags separate them. |
| **CI** | GitHub Actions | Mirrors Spinnaker pipeline stages: lint, test, build. |

Full decision records with alternatives considered: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

---

## What I Would Add With More Time

- OpenAPI / Swagger spec generation
- gRPC endpoints alongside REST
- OpenTelemetry distributed tracing
- Rate limiting per client
- API key authentication
- Prometheus metrics endpoint (`/metrics`)
- PATCH for partial updates
- Filtering and sorting on list endpoint
- Connection pool tuning

---

## Author

**Brandon Coburn** - Senior Software Engineer

- GitHub: [@Brandon-Anubis](https://github.com/Brandon-Anubis)
- LinkedIn: [coburnbrandon](https://linkedin.com/in/coburnbrandon)

Built with [Windsurf](https://windsurf.com/) AI-assisted development. Every line reviewed and owned by me. See `.windsurf/` for the automation setup.
