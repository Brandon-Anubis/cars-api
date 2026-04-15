---
trigger: always_on
---

# API Standards — Cars API

## REST Conventions
- API prefix: `/api/v1/`
- Resource naming: lowercase plural nouns (`/cars`, not `/car`)
- Use HTTP methods correctly: GET (read), POST (create), PUT (full update), DELETE (remove)
- Health endpoint: `/health` (outside API prefix)

## Request/Response Format
- Content-Type: `application/json` for all endpoints
- Accept: `application/json`
- All timestamps in RFC 3339 / ISO 8601 format (UTC)
- UUIDs for resource IDs (v4)

## Response Envelope
All responses MUST use this structure:

```json
// Success (single resource)
{
"data": { ... },
"meta": { "request_id": "uuid" }
}

// Success (collection)
{
"data": [ ... ],
"meta": {
    "request_id": "uuid",
    "total": 42,
    "limit": 20,
    "offset": 0
    }
}

// Error

{
"error": {
    "code": "VALIDATION_ERROR",
    "message": "human-readable description",
    "details": [ ... ]  // optional field-level errors
},

"meta": { "request_id": "uuid" }

}

```

## Status Codes
- 200: Success (GET, PUT)
- 201: Created (POST)
- 204: No Content (DELETE)
- 400: Bad Request (validation failure, malformed JSON)
- 404: Not Found
- 405: Method Not Allowed
- 415: Unsupported Media Type (non-JSON request)
- 500: Internal Server Error (unexpected failures)

## Pagination
- Query params: `?limit=20&offset=0`
- Default limit: 20, max limit: 100
- Response meta includes: total, limit, offset

## Error Codes
Use uppercase snake_case error codes:
- `NOT_FOUND` — Resource does not exist
- `VALIDATION_ERROR` — Input validation failed
- `INVALID_JSON` — Request body is not valid JSON
- `INTERNAL_ERROR` — Unexpected server error
- `METHOD_NOT_ALLOWED` — HTTP method not supported for this endpoint

## Middleware Stack (order matters)
1. Recovery (panic → 500)
2. Request ID (generate UUID, add to context + response header)
3. Logging (structured log with request ID, method, path, status, duration)
4. Content-Type enforcement (reject non-JSON for POST/PUT)
