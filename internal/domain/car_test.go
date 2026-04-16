package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brandon-anubis/cars-api/internal/domain"
)

// ---------------------------------------------------------------------------
// NewCar
// ---------------------------------------------------------------------------

func TestNewCar(t *testing.T) {
	currentYear := time.Now().UTC().Year()

	tests := []struct {
		name     string
		carMake  string
		model    string
		year     int
		color    string
		wantErr  bool
		errField string
	}{
		{
			name:    "valid input",
			carMake: "Toyota",
			model:   "Camry",
			year:    currentYear,
			color:   "Blue",
		},
		{
			name:    "valid input trims whitespace",
			carMake: "  Toyota  ",
			model:   "  Camry  ",
			year:    currentYear,
			color:   "  Blue  ",
		},
		{
			name:    "minimum valid year boundary",
			carMake: "Benz",
			model:   "Patent-Motorwagen",
			year:    1886,
			color:   "Black",
		},
		{
			name:    "maximum valid year boundary",
			carMake: "Tesla",
			model:   "Model 3",
			year:    currentYear + 2,
			color:   "White",
		},
		{
			name:     "empty make",
			carMake:  "",
			model:    "Camry",
			year:     currentYear,
			color:    "Blue",
			wantErr:  true,
			errField: "make",
		},
		{
			name:     "whitespace-only make",
			carMake:  "   ",
			model:    "Camry",
			year:     currentYear,
			color:    "Blue",
			wantErr:  true,
			errField: "make",
		},
		{
			name:     "empty model",
			carMake:  "Toyota",
			model:    "",
			year:     currentYear,
			color:    "Blue",
			wantErr:  true,
			errField: "model",
		},
		{
			name:     "year below minimum",
			carMake:  "Toyota",
			model:    "Camry",
			year:     1885,
			color:    "Blue",
			wantErr:  true,
			errField: "year",
		},
		{
			name:     "year above maximum",
			carMake:  "Toyota",
			model:    "Camry",
			year:     currentYear + 3,
			color:    "Blue",
			wantErr:  true,
			errField: "year",
		},
		{
			name:     "empty color",
			carMake:  "Toyota",
			model:    "Camry",
			year:     currentYear,
			color:    "",
			wantErr:  true,
			errField: "color",
		},
		{
			name:     "whitespace-only color",
			carMake:  "Toyota",
			model:    "Camry",
			year:     currentYear,
			color:    "   ",
			wantErr:  true,
			errField: "color",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			car, err := domain.NewCar(tc.carMake, tc.model, tc.year, tc.color)

			if tc.wantErr {
				requireValidationError(t, err, tc.errField)
				assert.Nil(t, car)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, car)
			assert.NotEmpty(t, car.ID)
			assert.Equal(t, int64(0), car.Mileage)
			assert.Equal(t, int64(0), car.PriceCents)
			assert.Empty(t, car.VIN)
		})
	}
}

func TestNewCar_IDAndTimestamps(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	car, err := domain.NewCar("Ford", "Mustang", 1969, "Red")
	after := time.Now().UTC().Add(time.Second)

	require.NoError(t, err)
	require.NotNil(t, car)

	assert.NotEmpty(t, car.ID, "ID must not be empty")
	assert.False(t, car.CreatedAt.IsZero(), "CreatedAt must be set")
	assert.False(t, car.UpdatedAt.IsZero(), "UpdatedAt must be set")
	assert.True(t, car.CreatedAt.After(before) && car.CreatedAt.Before(after),
		"CreatedAt must be close to now")
	assert.True(t, car.UpdatedAt.After(before) && car.UpdatedAt.Before(after),
		"UpdatedAt must be close to now")
	assert.Equal(t, car.CreatedAt, car.UpdatedAt,
		"CreatedAt and UpdatedAt must be equal on creation")
}

// ---------------------------------------------------------------------------
// Validate
// ---------------------------------------------------------------------------

func TestCar_Validate(t *testing.T) {
	currentYear := time.Now().UTC().Year()

	validBase := func() domain.Car {
		return domain.Car{
			Make:       "Honda",
			Model:      "Civic",
			Year:       currentYear,
			Color:      "White",
			VIN:        "",
			Mileage:    0,
			PriceCents: 0,
		}
	}

	tests := []struct {
		name     string
		mutate   func(*domain.Car)
		wantErr  bool
		errField string
	}{
		{
			name:   "valid car minimal",
			mutate: func(_ *domain.Car) {},
		},
		{
			name: "valid car with all optional fields",
			mutate: func(c *domain.Car) {
				c.VIN = "1HGBH41JXMN109186"
				c.Mileage = 50000
				c.PriceCents = 2599000
			},
		},
		{
			name:     "empty make",
			mutate:   func(c *domain.Car) { c.Make = "" },
			wantErr:  true,
			errField: "make",
		},
		{
			name:     "whitespace-only model",
			mutate:   func(c *domain.Car) { c.Model = "   " },
			wantErr:  true,
			errField: "model",
		},
		{
			name:     "year below minimum",
			mutate:   func(c *domain.Car) { c.Year = 1885 },
			wantErr:  true,
			errField: "year",
		},
		{
			name:     "year above maximum",
			mutate:   func(c *domain.Car) { c.Year = currentYear + 3 },
			wantErr:  true,
			errField: "year",
		},
		{
			name:     "empty color",
			mutate:   func(c *domain.Car) { c.Color = "" },
			wantErr:  true,
			errField: "color",
		},
		{
			name:     "negative mileage",
			mutate:   func(c *domain.Car) { c.Mileage = -1 },
			wantErr:  true,
			errField: "mileage",
		},
		{
			name:     "negative price cents",
			mutate:   func(c *domain.Car) { c.PriceCents = -1 },
			wantErr:  true,
			errField: "price_cents",
		},
		{
			name:     "vin wrong length",
			mutate:   func(c *domain.Car) { c.VIN = "12345" },
			wantErr:  true,
			errField: "vin",
		},
		{
			name:     "vin contains I",
			mutate:   func(c *domain.Car) { c.VIN = "1HGBH41IXMN109186" },
			wantErr:  true,
			errField: "vin",
		},
		{
			name:     "vin contains O",
			mutate:   func(c *domain.Car) { c.VIN = "1HGBH41OXMN109186" },
			wantErr:  true,
			errField: "vin",
		},
		{
			name:     "vin contains Q",
			mutate:   func(c *domain.Car) { c.VIN = "1HGBH41QXMN109186" },
			wantErr:  true,
			errField: "vin",
		},
		{
			name:     "vin contains lowercase",
			mutate:   func(c *domain.Car) { c.VIN = "1hgbh41jxmn109186" },
			wantErr:  true,
			errField: "vin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			car := validBase()
			tc.mutate(&car)

			err := car.Validate()

			if tc.wantErr {
				requireValidationError(t, err, tc.errField)
				return
			}

			require.NoError(t, err)
		})
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestCar_Update(t *testing.T) {
	currentYear := time.Now().UTC().Year()

	validInput := func() domain.CarUpdateInput {
		return domain.CarUpdateInput{
			Make:       "Chevrolet",
			Model:      "Malibu",
			Year:       currentYear,
			Color:      "Black",
			VIN:        "1HGBH41JXMN109186",
			Mileage:    15000,
			PriceCents: 2500000,
		}
	}

	tests := []struct {
		name     string
		mutate   func(*domain.CarUpdateInput)
		wantErr  bool
		errField string
	}{
		{
			name:   "valid update all fields",
			mutate: func(_ *domain.CarUpdateInput) {},
		},
		{
			name:   "valid update without vin",
			mutate: func(i *domain.CarUpdateInput) { i.VIN = "" },
		},
		{
			name:   "valid update normalizes vin to uppercase",
			mutate: func(i *domain.CarUpdateInput) { i.VIN = "1hgbh41jxmn109186" },
		},
		{
			name:     "empty make",
			mutate:   func(i *domain.CarUpdateInput) { i.Make = "" },
			wantErr:  true,
			errField: "make",
		},
		{
			name:     "whitespace-only model",
			mutate:   func(i *domain.CarUpdateInput) { i.Model = "   " },
			wantErr:  true,
			errField: "model",
		},
		{
			name:     "year below minimum",
			mutate:   func(i *domain.CarUpdateInput) { i.Year = 1885 },
			wantErr:  true,
			errField: "year",
		},
		{
			name:     "year above maximum",
			mutate:   func(i *domain.CarUpdateInput) { i.Year = currentYear + 3 },
			wantErr:  true,
			errField: "year",
		},
		{
			name:     "empty color",
			mutate:   func(i *domain.CarUpdateInput) { i.Color = "" },
			wantErr:  true,
			errField: "color",
		},
		{
			name:     "negative mileage",
			mutate:   func(i *domain.CarUpdateInput) { i.Mileage = -1 },
			wantErr:  true,
			errField: "mileage",
		},
		{
			name:     "negative price cents",
			mutate:   func(i *domain.CarUpdateInput) { i.PriceCents = -500 },
			wantErr:  true,
			errField: "price_cents",
		},
		{
			name:     "vin wrong length",
			mutate:   func(i *domain.CarUpdateInput) { i.VIN = "BADVIN" },
			wantErr:  true,
			errField: "vin",
		},
		{
			name:     "vin contains disallowed letter",
			mutate:   func(i *domain.CarUpdateInput) { i.VIN = "1HGBH41IXMN109186" },
			wantErr:  true,
			errField: "vin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			car, err := domain.NewCar("Toyota", "Camry", currentYear, "Blue")
			require.NoError(t, err)

			before := *car
			input := validInput()
			tc.mutate(&input)

			err = car.Update(input)

			if tc.wantErr {
				requireValidationError(t, err, tc.errField)
				assert.Equal(t, before, *car, "car must not be mutated on failed update")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, input.Make, car.Make)
			assert.Equal(t, input.Model, car.Model)
			assert.Equal(t, input.Year, car.Year)
			assert.Equal(t, input.Color, car.Color)
			assert.Equal(t, input.Mileage, car.Mileage)
			assert.Equal(t, input.PriceCents, car.PriceCents)
			assert.Equal(t, before.ID, car.ID, "ID must not change")
			assert.Equal(t, before.CreatedAt, car.CreatedAt, "CreatedAt must not change")
			assert.True(t, car.UpdatedAt.After(before.UpdatedAt),
				"UpdatedAt must advance on successful update")
		})
	}
}

// ---------------------------------------------------------------------------
// Test helper
// ---------------------------------------------------------------------------

func requireValidationError(t *testing.T, err error, wantField string) {
	t.Helper()

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrCarValidation),
		"expected ErrCarValidation in error chain, got: %v", err)

	var ve *domain.ValidationError
	require.True(t, errors.As(err, &ve), "expected *ValidationError")
	assert.Equal(t, wantField, ve.Field)
}
