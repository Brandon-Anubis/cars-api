// Package repository defines the port contracts for data persistence.
package repository

import (
	"context"

	"github.com/brandon-anubis/cars-api/internal/domain"
)

// CarRepository defines the persistence contract for Car entities.
type CarRepository interface {
	// Create inserts a new Car. Returns ErrCarAlreadyExists if the ID is taken.
	Create(ctx context.Context, car *domain.Car) error

	// GetByID retrieves a Car by its UUID. Returns ErrCarNotFound if absent.
	GetByID(ctx context.Context, id string) (*domain.Car, error)

	// List returns a paginated slice of Cars, the total count, and any error.
	List(ctx context.Context, limit, offset int) ([]*domain.Car, int, error)

	// Update persists mutations to an existing Car. Returns ErrCarNotFound if absent.
	Update(ctx context.Context, car *domain.Car) error

	// Delete removes a Car by its UUID. Returns ErrCarNotFound if absent.
	Delete(ctx context.Context, id string) error
}
