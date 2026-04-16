package domain

import (
	"errors"
	"fmt"
)

// ErrCarNotFound is returned when a car cannot be located by its identifier.
var ErrCarNotFound = errors.New("car not found")

// ErrCarAlreadyExists is returned when a car with the same identifier already exists.
var ErrCarAlreadyExists = errors.New("car already exists")

// ErrCarValidation is the sentinel error wrapped by ValidationError.
var ErrCarValidation = errors.New("car validation error")

// ValidationError describes a validation failure on a specific field.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// Unwrap returns ErrCarValidation, enabling errors.Is(err, ErrCarValidation) checks.
func (e *ValidationError) Unwrap() error {
	return ErrCarValidation
}
