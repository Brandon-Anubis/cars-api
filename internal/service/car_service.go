// Package service implements use-case orchestration between the HTTP transport
// and the repository port.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/brandon-anubis/cars-api/internal/domain"
	"github.com/brandon-anubis/cars-api/internal/repository"
)

const (
	minLimit = 1
	maxLimit = 100
)

// CarServiceInterface defines the use-case contract consumed by HTTP handlers.
type CarServiceInterface interface {
	Create(ctx context.Context, carMake, model string, year int, color string) (*domain.Car, error)
	GetByID(ctx context.Context, id string) (*domain.Car, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Car, int, error)
	Update(ctx context.Context, id string, input domain.CarUpdateInput) (*domain.Car, error)
	Delete(ctx context.Context, id string) error
}

// CarService implements CarServiceInterface using an injected CarRepository.
type CarService struct {
	repo   repository.CarRepository
	logger *slog.Logger
}

// compile-time interface verification.
var _ CarServiceInterface = (*CarService)(nil)

// NewCarService constructs a CarService. If logger is nil, slog.Default() is used.
func NewCarService(repo repository.CarRepository, logger *slog.Logger) *CarService {
	if logger == nil {
		logger = slog.Default()
	}
	return &CarService{repo: repo, logger: logger}
}

// Create constructs a new Car from the supplied fields and persists it.
func (s *CarService) Create(ctx context.Context, carMake, model string, year int, color string) (*domain.Car, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("service.Create: %w", err)
	}

	car, err := domain.NewCar(carMake, model, year, color)
	if err != nil {
		return nil, fmt.Errorf("service.Create: %w", err)
	}

	if err := s.repo.Create(ctx, car); err != nil {
		return nil, fmt.Errorf("service.Create: %w", err)
	}

	s.logger.InfoContext(ctx, "car created", slog.String("id", car.ID))
	return car, nil
}

// GetByID retrieves a Car by its UUID. Returns a wrapped domain.ErrCarNotFound if absent.
func (s *CarService) GetByID(ctx context.Context, id string) (*domain.Car, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("service.GetByID: %w", err)
	}

	if err := requireID("GetByID", id); err != nil {
		return nil, err
	}

	car, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service.GetByID: %w", err)
	}

	return car, nil
}

// List returns a paginated slice of Cars and the total count.
// limit is clamped to [1, 100]; offset is floored at 0.
func (s *CarService) List(ctx context.Context, limit, offset int) ([]*domain.Car, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("service.List: %w", err)
	}

	limit = clampLimit(limit)
	if offset < 0 {
		offset = 0
	}

	cars, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("service.List: %w", err)
	}

	return cars, total, nil
}

// Update fetches the Car by id, applies the mutation via car.Update, and persists it.
func (s *CarService) Update(ctx context.Context, id string, input domain.CarUpdateInput) (*domain.Car, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}

	if err := requireID("Update", id); err != nil {
		return nil, err
	}

	car, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}

	if err := car.Update(input); err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}

	if err := s.repo.Update(ctx, car); err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}

	s.logger.InfoContext(ctx, "car updated", slog.String("id", car.ID))
	return car, nil
}

// Delete removes a Car by its UUID.
func (s *CarService) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("service.Delete: %w", err)
	}

	if err := requireID("Delete", id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("service.Delete: %w", err)
	}

	s.logger.InfoContext(ctx, "car deleted", slog.String("id", id))
	return nil
}

// clampLimit constrains limit to [minLimit, maxLimit].
func clampLimit(limit int) int {
	if limit < minLimit {
		return minLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

// requireID returns a ValidationError if id is empty or whitespace-only.
func requireID(method, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("service.%s: %w", method, &domain.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}
	return nil
}
