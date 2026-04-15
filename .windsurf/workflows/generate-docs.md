---
name: Generate Documentation
description: Generate or update project documentation from source code
inputs:
  - name: doc_type
    type: enum
    values: [readme, api, architecture, all]
    description: Which documentation to generate
---

# Generate Documentation

## Steps

### If doc_type = readme (or all)
Generate `README.md` with:
- Project title and description
- Architecture overview (brief)
- Prerequisites (Go 1.22+, Docker)
- Quick Start (docker compose up)
- Manual setup steps
- API reference table (method, path, description, status codes)
- curl examples for each endpoint
- Running tests (`make test`)
- Project structure tree
- Design decisions summary
- "What I'd Add" section

### If doc_type = api (or all)
Generate `docs/API.md` with:
- Base URL
- Authentication (none for this challenge)
- Common headers
- Each endpoint:
  - Method + Path
  - Description
  - Request body schema (if applicable)
  - Response schema
  - Status codes
  - curl example
  - Example response JSON

### If doc_type = architecture (or all)
Generate `docs/ARCHITECTURE.md` with:
- Architecture diagram (text-based)
- Layer descriptions and responsibilities
- Dependency flow
- Design decisions with rationale:
  - Why Cloud Spanner
  - Why Clean Architecture
  - Why chi router (or stdlib)
  - Error handling strategy
  - Testing strategy
- Trade-offs acknowledged
- Future improvements

## Verify
- All links in documentation are valid
- Code examples compile
- curl examples use correct paths and payloads