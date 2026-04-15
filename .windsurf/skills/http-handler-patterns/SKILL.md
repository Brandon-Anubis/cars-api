---
name: http-handler-patterns
description: Production patterns for Go HTTP handlers, middleware, response formatting, and chi router setup. Use when working with HTTP handlers, middleware, routing, or response formatting.
---

### Router Setup (chi)

```go
import "github.com/go-chi/chi/v5"

func NewRouter(h *CarHandler) *chi.Mux {
    r := chi.NewRouter()
    r.Use(RecoveryMiddleware)
    r.Use(RequestIDMiddleware)
    r.Use(LoggingMiddleware)
    r.Get("/health", h.HealthCheck)
    r.Route("/api/v1", func(r chi.Router) {
        r.Route("/cars", func(r chi.Router) {
            r.Post("/", h.CreateCar)
            r.Get("/", h.ListCars)
            r.Get("/{id}", h.GetCar)
            r.Put("/{id}", h.UpdateCar)
            r.Delete("/{id}", h.DeleteCar)
        })
    })
    return r
}
```

### Standard Response Helpers

```go
package handler

type Response struct {
    Data  interface{}    `json:"data,omitempty"`
    Error *ErrorResponse `json:"error,omitempty"`
    Meta  Meta           `json:"meta"`
}

type ErrorResponse struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type Meta struct {
    RequestID string `json:"request_id"`
    Total     *int   `json:"total,omitempty"`
    Limit     *int   `json:"limit,omitempty"`
    Offset    *int   `json:"offset,omitempty"`
}

func respondJSON(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(Response{
        Data: data,
        Meta: Meta{RequestID: GetRequestID(r.Context())},
    })
}

func respondError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(Response{
        Error: &ErrorResponse{Code: code, Message: message},
        Meta:  Meta{RequestID: GetRequestID(r.Context())},
    })
}
```

### Middleware — Request ID

```go
type contextKey string
const requestIDKey contextKey = "request_id"

func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := uuid.New().String()
        ctx := context.WithValue(r.Context(), requestIDKey, id)
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return id
    }
    return ""
}
```

### Middleware — Structured Logging

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(ww, r)
        slog.Info("request",
            "method", r.Method,
            "path", r.URL.Path,
            "status", ww.status,
            "duration_ms", time.Since(start).Milliseconds(),
            "request_id", GetRequestID(r.Context()),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}
```

### Middleware — Recovery

```go
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                slog.Error("panic recovered",
                    "error", err,
                    "request_id", GetRequestID(r.Context()),
                    "stack", string(debug.Stack()),
                )
                respondError(w, r, http.StatusInternalServerError,
                    "INTERNAL_ERROR", "an unexpected error occurred")
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Handler Pattern

```go
func (h *CarHandler) CreateCar(w http.ResponseWriter, r *http.Request) {
    var req CreateCarRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, r, http.StatusBadRequest, "INVALID_JSON", "request body must be valid JSON")
        return
    }
    car, err := h.service.Create(r.Context(), req.Make, req.Model, req.Year, req.Color)
    if err != nil {
        var validationErr *domain.ValidationError
        if errors.As(err, &validationErr) {
            respondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
            return
        }
        slog.Error("create car", "error", err, "request_id", GetRequestID(r.Context()))
        respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create car")
        return
    }
    respondJSON(w, r, http.StatusCreated, car)
}
```