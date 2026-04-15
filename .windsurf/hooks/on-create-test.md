---
name: On Create — Test File
trigger: on_create
file_patterns: ["**/*_test.go"]
---

### Actions

1. **Set package name** matching the source file's package
2. **Add test imports:**

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

1. **Add layer-specific imports:**
    - Handler tests: add `"net/http"`, `"net/http/httptest"`, `"encoding/json"`, `"github.com/go-chi/chi/v5"`
    - Service tests: add `"context"`, domain import
    - Repository tests: add `"context"`, `"os"`, `"cloud.google.com/go/spanner"`, domain import
2. **Add starter test function:**

```go
func TestEntityName_MethodName(t *testing.T) {
    tests := []struct {
        name    string
        wantErr bool
    }{
        // Add test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Implement
        })
    }
}
```

1. **For integration test files** (in `repository/spanner/`):
    - Add build tag: `//go:build integration`
    - Add `TestMain` function with emulator setup skeleton