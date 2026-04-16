package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	minCarYear = 1886
	vinLength  = 17
)

// Car represents a vehicle entity in the domain.
type Car struct {
	ID         string    `json:"id" spanner:"ID"`
	Make       string    `json:"make" spanner:"MAKE"`
	Model      string    `json:"model" spanner:"MODEL"`
	Year       int       `json:"year" spanner:"YEAR"`
	Color      string    `json:"color" spanner:"COLOR"`
	VIN        string    `json:"vin" spanner:"VIN"`
	Mileage    int64     `json:"mileage" spanner:"MILEAGE"`
	PriceCents int64     `json:"price_cents" spanner:"PRICE_CENTS"`
	CreatedAt  time.Time `json:"created_at" spanner:"CREATED_AT"`
	UpdatedAt  time.Time `json:"updated_at" spanner:"UPDATED_AT"`
}

// CarUpdateInput contains the mutable fields for updating a car.
type CarUpdateInput struct {
	Make       string
	Model      string
	Year       int
	Color      string
	VIN        string
	Mileage    int64
	PriceCents int64
}

// NewCar constructs and validates a new Car, generating a UUID and setting timestamps.
func NewCar(carMake, model string, year int, color string) (*Car, error) {
	now := time.Now().UTC()

	car := &Car{
		ID:        uuid.NewString(),
		Make:      normalizeText(carMake),
		Model:     normalizeText(model),
		Year:      year,
		Color:     normalizeText(color),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := car.Validate(); err != nil {
		return nil, err
	}

	return car, nil
}

// Validate checks all required fields and business rules on the Car.
func (c *Car) Validate() error {
	if err := validateRequiredText("make", c.Make); err != nil {
		return err
	}

	if err := validateRequiredText("model", c.Model); err != nil {
		return err
	}

	maxYear := time.Now().UTC().Year() + 2
	if c.Year < minCarYear || c.Year > maxYear {
		return &ValidationError{
			Field:   "year",
			Message: fmt.Sprintf("year must be between %d and %d", minCarYear, maxYear),
		}
	}

	if err := validateRequiredText("color", c.Color); err != nil {
		return err
	}

	if c.Mileage < 0 {
		return &ValidationError{Field: "mileage", Message: "mileage must be greater than or equal to 0"}
	}

	if c.PriceCents < 0 {
		return &ValidationError{Field: "price_cents", Message: "price_cents must be greater than or equal to 0"}
	}

	return validateVIN(c.VIN)
}

// Update validates the provided values then applies them to the Car, refreshing UpdatedAt.
func (c *Car) Update(input CarUpdateInput) error {
	candidate := &Car{
		Make:       normalizeText(input.Make),
		Model:      normalizeText(input.Model),
		Year:       input.Year,
		Color:      normalizeText(input.Color),
		VIN:        normalizeVIN(input.VIN),
		Mileage:    input.Mileage,
		PriceCents: input.PriceCents,
	}

	if err := candidate.Validate(); err != nil {
		return err
	}

	c.Make = candidate.Make
	c.Model = candidate.Model
	c.Year = candidate.Year
	c.Color = candidate.Color
	c.VIN = candidate.VIN
	c.Mileage = candidate.Mileage
	c.PriceCents = candidate.PriceCents
	c.UpdatedAt = time.Now().UTC()

	return nil
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}

func normalizeVIN(vin string) string {
	return strings.ToUpper(strings.TrimSpace(vin))
}

func validateRequiredText(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s is required", field),
		}
	}

	return nil
}

func validateVIN(vin string) error {
	if vin == "" {
		return nil
	}

	if len(vin) != vinLength {
		return &ValidationError{Field: "vin", Message: "vin must be exactly 17 characters"}
	}

	return validateVINCharacters(vin)
}

func validateVINCharacters(vin string) error {
	for _, r := range vin {
		if err := validateVINRune(r); err != nil {
			return err
		}
	}

	return nil
}

func validateVINRune(r rune) error {
	if isDigit(r) {
		return nil
	}

	if !isUppercaseLetter(r) {
		return &ValidationError{Field: "vin", Message: "vin must contain only uppercase letters and digits"}
	}

	if isDisallowedVINLetter(r) {
		return &ValidationError{Field: "vin", Message: "vin must not contain I, O, or Q"}
	}

	return nil
}

func isDigit(r rune) bool               { return r >= '0' && r <= '9' }
func isUppercaseLetter(r rune) bool     { return r >= 'A' && r <= 'Z' }
func isDisallowedVINLetter(r rune) bool { return r == 'I' || r == 'O' || r == 'Q' }
