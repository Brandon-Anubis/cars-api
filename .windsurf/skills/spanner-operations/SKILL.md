---
name: spanner-operations
description: Deep knowledge for Cloud Spanner Go client operations, emulator setup, schema management, and CRUD patterns. Use when working with Spanner client code, database operations, or emulator setup.
---

### Go Client Library

- Package: `cloud.google.com/go/spanner`
- Admin Package: `cloud.google.com/go/spanner/admin/database/apiv1`
- Instance Admin: `cloud.google.com/go/spanner/admin/instance/apiv1`

### Emulator Setup Sequence

**1. Start emulator (Docker)**

```bash
docker run -d --name spanner-emulator \
  -p 9010:9010 -p 9020:9020 \
  gcr.io/cloud-spanner-emulator/emulator
```

**2. Create instance (programmatic Go)**

```go
import (
    instance "cloud.google.com/go/spanner/admin/instance/apiv1"
    instancepb "cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
    "google.golang.org/api/option"
)

func createInstance(ctx context.Context, projectID, instanceID string) error {
    instanceAdmin, err := instance.NewInstanceAdminClient(ctx,
        option.WithEndpoint("localhost:9010"),
        option.WithoutAuthentication(),
        option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
    )
    if err != nil {
        return fmt.Errorf("create instance admin client: %w", err)
    }
    defer instanceAdmin.Close()

    op, err := instanceAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
        Parent:     fmt.Sprintf("projects/%s", projectID),
        InstanceId: instanceID,
        Instance: &instancepb.Instance{
            Config:      fmt.Sprintf("projects/%s/instanceConfigs/emulator-config", projectID),
            DisplayName: instanceID,
            NodeCount:   1,
        },
    })
    if err != nil {
        return fmt.Errorf("create instance: %w", err)
    }
    if _, err := op.Wait(ctx); err != nil {
        return fmt.Errorf("wait for instance: %w", err)
    }
    return nil
}
```

**3. Create database with schema**

```go
import (
    database "cloud.google.com/go/spanner/admin/database/apiv1"
    databasepb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
)

func createDatabase(ctx context.Context, projectID, instanceID, databaseID string, ddlStatements []string) error {
    adminClient, err := database.NewDatabaseAdminClient(ctx,
        option.WithEndpoint("localhost:9010"),
        option.WithoutAuthentication(),
        option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
    )
    if err != nil {
        return fmt.Errorf("create admin client: %w", err)
    }
    defer adminClient.Close()

    op, err := adminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
        Parent:          fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID),
        CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", databaseID),
        ExtraStatements: ddlStatements,
    })
    if err != nil {
        return fmt.Errorf("create database: %w", err)
    }
    if _, err := op.Wait(ctx); err != nil {
        return fmt.Errorf("wait for database: %w", err)
    }
    return nil
}
```

**4. Create data client**

```go
func newSpannerClient(ctx context.Context, dbPath string) (*spanner.Client, error) {
    // When SPANNER_EMULATOR_HOST is set, the client auto-connects
    // to the emulator without TLS or authentication
    client, err := spanner.NewClient(ctx, dbPath)
    if err != nil {
        return nil, fmt.Errorf("create spanner client: %w", err)
    }
    return client, nil
}
```

### CRUD Operation Patterns

**Create (Insert)**

```go
func (r *carRepo) Create(ctx context.Context, car *domain.Car) error {
    m, err := spanner.InsertStruct("Cars", car)
    if err != nil {
        return fmt.Errorf("build mutation: %w", err)
    }
    if _, err := r.client.Apply(ctx, []*spanner.Mutation{m}); err != nil {
        if spanner.ErrCode(err) == codes.AlreadyExists {
            return domain.ErrCarAlreadyExists
        }
        return fmt.Errorf("apply insert: %w", err)
    }
    return nil
}
```

**Read (Single)**

```go
func (r *carRepo) GetByID(ctx context.Context, id string) (*domain.Car, error) {
    row, err := r.client.Single().ReadRow(ctx, "Cars",
        spanner.Key{id},
        []string{"ID", "Make", "Model", "Year", "Color", "VIN", "Mileage", "Price", "CreatedAt", "UpdatedAt"},
    )
    if err != nil {
        if spanner.ErrCode(err) == codes.NotFound {
            return nil, domain.ErrCarNotFound
        }
        return nil, fmt.Errorf("read row: %w", err)
    }
    var car domain.Car
    if err := row.ToStruct(&car); err != nil {
        return nil, fmt.Errorf("scan row: %w", err)
    }
    return &car, nil
}
```

**Read (List with SQL)**

```go
func (r *carRepo) List(ctx context.Context, limit, offset int) ([]*domain.Car, int, error) {
    var total int64
    countStmt := spanner.Statement{SQL: "SELECT COUNT(*) FROM Cars"}
    if err := r.client.Single().Query(ctx, countStmt).Do(func(row *spanner.Row) error {
        return row.Columns(&total)
    }); err != nil {
        return nil, 0, fmt.Errorf("count query: %w", err)
    }

    stmt := spanner.Statement{
        SQL:    "SELECT * FROM Cars ORDER BY CreatedAt DESC LIMIT @limit OFFSET @offset",
        Params: map[string]interface{}{"limit": int64(limit), "offset": int64(offset)},
    }
    var cars []*domain.Car
    if err := r.client.Single().Query(ctx, stmt).Do(func(row *spanner.Row) error {
        var car domain.Car
        if err := row.ToStruct(&car); err != nil {
            return err
        }
        cars = append(cars, &car)
        return nil
    }); err != nil {
        return nil, 0, fmt.Errorf("list query: %w", err)
    }
    return cars, int(total), nil
}
```

**Update (Read-Write Transaction)**

```go
func (r *carRepo) Update(ctx context.Context, car *domain.Car) error {
    _, err := r.client.ReadWriteTransaction(ctx,
        func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
            row, err := txn.ReadRow(ctx, "Cars", spanner.Key{car.ID}, []string{"ID"})
            if err != nil {
                if spanner.ErrCode(err) == codes.NotFound {
                    return domain.ErrCarNotFound
                }
                return fmt.Errorf("check existence: %w", err)
            }
            _ = row
            m, err := spanner.UpdateStruct("Cars", car)
            if err != nil {
                return fmt.Errorf("build mutation: %w", err)
            }
            return txn.BufferWrite([]*spanner.Mutation{m})
        },
    )
    if err != nil {
        return fmt.Errorf("update transaction: %w", err)
    }
    return nil
}
```

**Delete**

```go
func (r *carRepo) Delete(ctx context.Context, id string) error {
    _, err := r.client.ReadWriteTransaction(ctx,
        func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
            _, err := txn.ReadRow(ctx, "Cars", spanner.Key{id}, []string{"ID"})
            if err != nil {
                if spanner.ErrCode(err) == codes.NotFound {
                    return domain.ErrCarNotFound
                }
                return fmt.Errorf("check existence: %w", err)
            }
            m := spanner.Delete("Cars", spanner.Key{id})
            return txn.BufferWrite([]*spanner.Mutation{m})
        },
    )
    if err != nil {
        return fmt.Errorf("delete transaction: %w", err)
    }
    return nil
}
```

### Spanner Struct Tags

Use `spanner:"ColumnName"` tags on domain structs for Spanner serialization:

```go
type Car struct {
    ID        string              `json:"id" spanner:"ID"`
    Make      string              `json:"make" spanner:"Make"`
    Model     string              `json:"model" spanner:"Model"`
    Year      int64               `json:"year" spanner:"Year"`
    Color     spanner.NullString  `json:"color" spanner:"Color"`
    VIN       spanner.NullString  `json:"vin" spanner:"VIN"`
    Mileage   spanner.NullInt64   `json:"mileage" spanner:"Mileage"`
    Price     spanner.NullFloat64 `json:"price" spanner:"Price"`
    CreatedAt time.Time           `json:"created_at" spanner:"CreatedAt"`
    UpdatedAt time.Time           `json:"updated_at" spanner:"UpdatedAt"`
}
```

**Important:** Spanner nullable columns MUST use `spanner.NullString`, `spanner.NullInt64`, etc. Non-nullable columns can use Go primitives.