// Command seed inserts a fixed set of sample cars into the Cars table using
// InsertOrUpdate mutations, making it safe to run multiple times.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"cloud.google.com/go/spanner"

	"github.com/brandon-anubis/cars-api/internal/config"
	"github.com/brandon-anubis/cars-api/internal/domain"
)

var (
	seedCreatedAt = time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	seedUpdatedAt = seedCreatedAt
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := spanner.NewClient(ctx, cfg.DatabasePath())
	if err != nil {
		logger.Error("create spanner client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	logger.Info("seed starting",
		"database", cfg.DatabasePath(),
		"emulator_host", cfg.SpannerEmulatorHost,
	)

	if err := seed(ctx, client, logger); err != nil {
		logger.Error("seed failed", "error", err)
		os.Exit(1)
	}

	logger.Info("seed complete")
}

func seed(ctx context.Context, client *spanner.Client, logger *slog.Logger) error {
	cars := seedCars()
	mutations := make([]*spanner.Mutation, 0, len(cars))

	for _, car := range cars {
		if err := car.Validate(); err != nil {
			return fmt.Errorf("validate car %s %s: %w", car.Make, car.Model, err)
		}

		m, err := spanner.InsertOrUpdateStruct("Cars", car)
		if err != nil {
			return fmt.Errorf("build mutation for %s %s: %w", car.Make, car.Model, err)
		}
		mutations = append(mutations, m)
	}

	if _, err := client.Apply(ctx, mutations); err != nil {
		return fmt.Errorf("apply mutations: %w", err)
	}

	for _, car := range cars {
		logger.Info("upserted car",
			"id", car.ID,
			"make", car.Make,
			"model", car.Model,
			"year", car.Year,
		)
	}
	return nil
}

// seedCars returns the canonical set of sample Car records with stable IDs so
// that repeated runs remain idempotent via InsertOrUpdate.
func seedCars() []*domain.Car {
	return []*domain.Car{
		{
			ID:         "11111111-1111-1111-1111-111111111111",
			Make:       "Toyota",
			Model:      "Camry",
			Year:       2024,
			Color:      "Blue",
			VIN:        "1HGBH41JXMN109186",
			Mileage:    15000,
			PriceCents: 2699900,
			CreatedAt:  seedCreatedAt,
			UpdatedAt:  seedUpdatedAt,
		},
		{
			ID:         "22222222-2222-2222-2222-222222222222",
			Make:       "Honda",
			Model:      "Civic",
			Year:       2023,
			Color:      "Silver",
			VIN:        "2T1BURHE0JC034244",
			Mileage:    8500,
			PriceCents: 2299900,
			CreatedAt:  seedCreatedAt,
			UpdatedAt:  seedUpdatedAt,
		},
		{
			ID:         "33333333-3333-3333-3333-333333333333",
			Make:       "Ford",
			Model:      "Mustang",
			Year:       2025,
			Color:      "Red",
			VIN:        "1FATP8UH5L5100001",
			Mileage:    500,
			PriceCents: 3599900,
			CreatedAt:  seedCreatedAt,
			UpdatedAt:  seedUpdatedAt,
		},
		{
			ID:         "44444444-4444-4444-4444-444444444444",
			Make:       "Tesla",
			Model:      "Model 3",
			Year:       2022,
			Color:      "White",
			VIN:        "5YJ3E1EA7LF000001",
			Mileage:    22000,
			PriceCents: 4299900,
			CreatedAt:  seedCreatedAt,
			UpdatedAt:  seedUpdatedAt,
		},
		{
			ID:         "55555555-5555-5555-5555-555555555555",
			Make:       "BMW",
			Model:      "X5",
			Year:       2024,
			Color:      "Black",
			VIN:        "5UXCR6C55KLL00001",
			Mileage:    3200,
			PriceCents: 6599900,
			CreatedAt:  seedCreatedAt,
			UpdatedAt:  seedUpdatedAt,
		},
	}
}
