---
name: Add REST Endpoint
description: Add a new endpoint with handler, route registration, and tests
inputs:
  - name: method
    type: enum
    values: [GET, POST, PUT, DELETE]
    description: HTTP method
  - name: path
    type: string
    description: URL path (e.g., "/api/v1/cars/{id}")
  - name: handler_name
    type: string
    description: Handler method name (e.g., "GetByID")
  - name: description
    type: string
    description: What this endpoint does
---

# Add REST Endpoint

## Steps

1. **Add handler method** to `internal/handler/car_handler.go`
   - Method signature: `func (h *CarHandler) {handler_name}(w http.ResponseWriter, r *http.Request)`
   - Parse path params / query params / request body as appropriate
   - Call service layer method
   - Use standard response helpers from `response.go`
   - Handle all error cases (400, 404, 500)

2. **Register route** in `internal/handler/router.go`
   - Add route with correct HTTP method and path
   - Ensure middleware stack is applied

3. **Add handler tests** to `internal/handler/car_handler_test.go`
   - Test success case with expected status code and response body
   - Test validation error (400)
   - Test not found (404) if applicable
   - Test server error (500)

4. **Update API documentation** in `docs/API.md`
   - Add endpoint to the table
   - Add curl example
   - Document request/response shapes

5. **Verify**
   - Run `go vet ./internal/handler/...`
   - Run handler tests: `go test -v -race ./internal/handler/...`