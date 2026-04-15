---
name: On Create — Go File Boilerplate
trigger: on_create
file_patterns: ["internal/**/*.go"]
exclude_patterns: ["**/*_test.go"]
---

### Actions

1. **Set package name** from directory:
    - `internal/domain/*.go` → `package domain`
    - `internal/service/*.go` → `package service`
    - `internal/repository/*.go` → `package repository`
    - `internal/repository/spanner/*.go` → `package spanner`
    - `internal/handler/*.go` → `package handler`
    - `internal/config/*.go` → `package config`
    - `cmd/server/*.go` → `package main`
2. **Add standard imports** based on layer:
    - domain: `"errors"`, `"time"`, `"github.com/google/uuid"`
    - service: `"context"`, `"fmt"`, domain import
    - repository/spanner: `"context"`, `"fmt"`, `"cloud.google.com/go/spanner"`, domain import
    - handler: `"encoding/json"`, `"net/http"`, `"log/slog"`, domain + service imports
    - config: `"os"`, `"strconv"`
3. **Add file header comment:**

```go
// Package {name} provides {layer description}.
```