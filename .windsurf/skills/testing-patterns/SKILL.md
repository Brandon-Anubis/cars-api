---
name: testing-patterns
description: Production Go testing patterns including table-driven tests, mock repositories, httptest handler tests, and Spanner emulator integration test setup. Use when writing or modifying tests.
---

### Table-Driven Unit Tests

```go
func TestNewCar(t *testing.T) {
    tests := []struct {
        name    string
        make    string
        model   string
        year    int
        color   string
        wantErr bool
        errMsg  string
    }{
        {
            name:  "valid car",
            make:  "Toyota",
            model: "Camry",
            year:  2024,
            color: "Blue",
        },
        {
            name:    "empty make",
            make:    "",
            model:   "Camry",
            year:    2024,
            color:   "Blue",
            wantErr: true,
            errMsg:  "make is required",
        },
        {
            name:    "year too low",
            make:    "Toyota",
            model:   "Camry",
            year:    1800,
            color:   "Blue",
            wantErr: true,
            errMsg:  "invalid year",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            car, err := domain.NewCar(tt.make, tt.model, tt.year, tt.color)
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
                assert.Nil(t, car)
            } else {
                require.NoError(t, err)
                assert.NotEmpty(t, car.ID)
                assert.Equal(t, tt.make, car.Make)
            }
        })
    }
}
```

### Mock Repository

```go
type mockCarRepo struct {
    createFn  func(ctx context.Context, car *domain.Car) error
    getByIDFn func(ctx context.Context, id string) (*domain.Car, error)
    listFn    func(ctx context.Context, limit, offset int) ([]*domain.Car, int, error)
    updateFn  func(ctx context.Context, car *domain.Car) error
    deleteFn  func(ctx context.Context, id string) error
}

func (m *mockCarRepo) Create(ctx context.Context, car *domain.Car) error {
    if m.createFn != nil {
        return m.createFn(ctx, car)
    }
    return nil
}
// ... implement all interface methods with the same pattern
```

### HTTP Handler Tests

```go
func TestCarHandler_GetCar(t *testing.T) {
    testCar := &domain.Car{ID: "test-id", Make: "Toyota", Model: "Camry", Year: 2024}
    tests := []struct {
        name       string
        carID      string
        setupMock  func(*mockCarService)
        wantStatus int
        wantCode   string
    }{
        {
            name:  "success",
            carID: "test-id",
            setupMock: func(m *mockCarService) {
                m.getByIDFn = func(ctx context.Context, id string) (*domain.Car, error) {
                    return testCar, nil
                }
            },
            wantStatus: http.StatusOK,
        },
        {
            name:  "not found",
            carID: "missing-id",
            setupMock: func(m *mockCarService) {
                m.getByIDFn = func(ctx context.Context, id string) (*domain.Car, error) {
                    return nil, domain.ErrCarNotFound
                }
            },
            wantStatus: http.StatusNotFound,
            wantCode:   "NOT_FOUND",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockCarService{}
            tt.setupMock(mock)
            handler := NewCarHandler(mock)
            req := httptest.NewRequest(http.MethodGet, "/api/v1/cars/"+tt.carID, nil)
            rctx := chi.NewRouteContext()
            rctx.URLParams.Add("id", tt.carID)
            req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
            rec := httptest.NewRecorder()
            handler.GetCar(rec, req)
            assert.Equal(t, tt.wantStatus, rec.Code)
        })
    }
}
```

### Integration Test Setup (Spanner Emulator)

```go
//go:build integration

package spanner_test

var testClient *spanner.Client

func TestMain(m *testing.M) {
    ctx := context.Background()
    if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
        fmt.Println("SPANNER_EMULATOR_HOST not set, skipping integration tests")
        os.Exit(0)
    }
    // Create test instance and database using admin clients
    client, err := spanner.NewClient(ctx, dbPath)
    if err != nil {
        log.Fatalf("create client: %v", err)
    }
    testClient = client
    code := m.Run()
    client.Close()
    os.Exit(code)
}
```