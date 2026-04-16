package spanner

import (
	"fmt"
	"time"

	"cloud.google.com/go/spanner"

	"github.com/brandon-anubis/cars-api/internal/domain"
)

// spannerCar is an unexported intermediary used exclusively for Spanner row
// decoding. It differs from domain.Car in two ways that the Spanner client
// requires: Year is int64 (Spanner INT64) and VIN is spanner.NullString
// (the DB column is nullable).
type spannerCar struct {
	ID         string             `spanner:"ID"`
	Make       string             `spanner:"MAKE"`
	Model      string             `spanner:"MODEL"`
	Year       int64              `spanner:"YEAR"`
	Color      string             `spanner:"COLOR"`
	VIN        spanner.NullString `spanner:"VIN"`
	Mileage    int64              `spanner:"MILEAGE"`
	PriceCents int64              `spanner:"PRICE_CENTS"`
	CreatedAt  time.Time          `spanner:"CREATED_AT"`
	UpdatedAt  time.Time          `spanner:"UPDATED_AT"`
}

// rowToCar decodes a Spanner row into a domain.Car, handling the nullable VIN
// column and the int64→int conversion for Year.
func rowToCar(row *spanner.Row) (*domain.Car, error) {
	var sc spannerCar
	if err := row.ToStruct(&sc); err != nil {
		return nil, fmt.Errorf("scan row: %w", err)
	}

	car := &domain.Car{
		ID:         sc.ID,
		Make:       sc.Make,
		Model:      sc.Model,
		Year:       int(sc.Year),
		Color:      sc.Color,
		Mileage:    sc.Mileage,
		PriceCents: sc.PriceCents,
		CreatedAt:  sc.CreatedAt,
		UpdatedAt:  sc.UpdatedAt,
	}

	if sc.VIN.Valid {
		car.VIN = sc.VIN.StringVal
	}

	return car, nil
}

// carToMap converts a domain.Car to the column→value map consumed by
// spanner.InsertMap and spanner.UpdateMap. An empty VIN is mapped to nil so
// that the nullable VIN column is stored as NULL rather than an empty string.
func carToMap(car *domain.Car) map[string]any {
	m := map[string]any{
		"ID":          car.ID,
		"MAKE":        car.Make,
		"MODEL":       car.Model,
		"YEAR":        int64(car.Year),
		"COLOR":       car.Color,
		"MILEAGE":     car.Mileage,
		"PRICE_CENTS": car.PriceCents,
		"CREATED_AT":  car.CreatedAt,
		"UPDATED_AT":  car.UpdatedAt,
	}

	if car.VIN != "" {
		m["VIN"] = car.VIN
	} else {
		m["VIN"] = nil
	}

	return m
}
