// Package spanner provides a Cloud Spanner implementation of CarRepository.
package spanner

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc/codes"

	"github.com/brandon-anubis/cars-api/internal/domain"
	"github.com/brandon-anubis/cars-api/internal/repository"
)

// Compile-time verification that carRepo satisfies CarRepository.
var _ repository.CarRepository = (*carRepo)(nil)

// carRepo is a Cloud Spanner-backed implementation of repository.CarRepository.
type carRepo struct {
	client *spanner.Client
}

// NewCarRepository constructs a carRepo using the provided Spanner client.
func NewCarRepository(client *spanner.Client) repository.CarRepository {
	return &carRepo{client: client}
}

// carColumns is the complete ordered column list for the Cars table, matching
// the spanner struct tags on domain.Car.
var carColumns = []string{
	"ID", "MAKE", "MODEL", "YEAR", "COLOR", "VIN",
	"MILEAGE", "PRICE_CENTS", "CREATED_AT", "UPDATED_AT",
}

// Create inserts car into the Cars table.
// Returns domain.ErrCarAlreadyExists if a row with the same ID already exists.
func (r *carRepo) Create(ctx context.Context, car *domain.Car) error {
	m := spanner.InsertMap("Cars", carToMap(car))

	if _, err := r.client.Apply(ctx, []*spanner.Mutation{m}); err != nil {
		if spanner.ErrCode(err) == codes.AlreadyExists {
			return domain.ErrCarAlreadyExists
		}

		return fmt.Errorf("spannerRepo.Create apply: %w", err)
	}

	return nil
}

// GetByID reads a single Car by primary key.
// Returns domain.ErrCarNotFound when no row matches id.
func (r *carRepo) GetByID(ctx context.Context, id string) (*domain.Car, error) {
	txn := r.client.Single()
	defer txn.Close()

	row, err := txn.ReadRow(ctx, "Cars", spanner.Key{id}, carColumns)
	if err != nil {
		if spanner.ErrCode(err) == codes.NotFound {
			return nil, domain.ErrCarNotFound
		}

		return nil, fmt.Errorf("spannerRepo.GetByID read row: %w", err)
	}

	car, err := rowToCar(row)
	if err != nil {
		return nil, fmt.Errorf("spannerRepo.GetByID: %w", err)
	}

	return car, nil
}

// List returns a page of Cars ordered by CREATED_AT descending, plus the total row count.
// Both the count and the page queries run under a single ReadOnlyTransaction, providing
// a consistent timestamp snapshot — the total and the page always reflect the same
// database state regardless of concurrent writes.
func (r *carRepo) List(ctx context.Context, limit, offset int) ([]*domain.Car, int, error) {
	txn := r.client.ReadOnlyTransaction()
	defer txn.Close()

	total, err := r.countCars(ctx, txn)
	if err != nil {
		return nil, 0, err
	}

	cars, err := r.listCars(ctx, txn, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return cars, total, nil
}

func (r *carRepo) countCars(ctx context.Context, txn *spanner.ReadOnlyTransaction) (int, error) {
	var total int64

	stmt := spanner.Statement{SQL: "SELECT COUNT(*) FROM Cars"}
	if err := txn.Query(ctx, stmt).Do(func(row *spanner.Row) error {
		return row.Columns(&total)
	}); err != nil {
		return 0, fmt.Errorf("spannerRepo.List count: %w", err)
	}

	return int(total), nil
}

func (r *carRepo) listCars(
	ctx context.Context,
	txn *spanner.ReadOnlyTransaction,
	limit, offset int,
) ([]*domain.Car, error) {
	stmt := spanner.Statement{
		SQL: "SELECT ID, MAKE, MODEL, YEAR, COLOR, VIN, MILEAGE, PRICE_CENTS, CREATED_AT, UPDATED_AT FROM Cars ORDER BY CREATED_AT DESC LIMIT @limit OFFSET @offset",
		Params: map[string]any{
			"limit":  int64(limit),
			"offset": int64(offset),
		},
	}

	cars := make([]*domain.Car, 0)
	if err := txn.Query(ctx, stmt).Do(func(row *spanner.Row) error {
		car, err := rowToCar(row)
		if err != nil {
			return fmt.Errorf("spannerRepo.List scan: %w", err)
		}

		cars = append(cars, car)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("spannerRepo.List query: %w", err)
	}

	return cars, nil
}

func ensureCarExists(ctx context.Context, txn *spanner.ReadWriteTransaction, id string) error {
	if _, err := txn.ReadRow(ctx, "Cars", spanner.Key{id}, []string{"ID"}); err != nil {
		if spanner.ErrCode(err) == codes.NotFound {
			return domain.ErrCarNotFound
		}

		return fmt.Errorf("spannerRepo.ensureCarExists: %w", err)
	}
	return nil
}

// Update applies all mutable fields of car to the matching Spanner row inside a
// read-write transaction. Returns domain.ErrCarNotFound when no row matches car.ID.
func (r *carRepo) Update(ctx context.Context, car *domain.Car) error {
	_, err := r.client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		if err := ensureCarExists(ctx, txn, car.ID); err != nil {
			return err
		}

		return txn.BufferWrite([]*spanner.Mutation{
			spanner.UpdateMap("Cars", carToMap(car)),
		})
	})
	if err != nil {
		return fmt.Errorf("spannerRepo.Update: %w", err)
	}

	return nil
}

// Delete removes the car identified by id inside a read-write transaction.
// Returns domain.ErrCarNotFound when no row matches id.
func (r *carRepo) Delete(ctx context.Context, id string) error {
	_, err := r.client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		if err := ensureCarExists(ctx, txn, id); err != nil {
			return err
		}

		return txn.BufferWrite([]*spanner.Mutation{
			spanner.Delete("Cars", spanner.Key{id}),
		})
	})
	if err != nil {
		return fmt.Errorf("spannerRepo.Delete: %w", err)
	}

	return nil
}
