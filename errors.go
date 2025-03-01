package enums

import (
	"errors"
	"fmt"
)

var (
	// ErrFailedToRegister is returned when an Enum type could not be registered.
	ErrFailedToRegister = errors.New("failed to register")
	// ErrInvalidValue is returned when the given value an Enum is not one of its valid values.
	ErrInvalidValue = errors.New("invalid enum value")
)

func errFailedToRegister(enumId string, causedBy string) error {
	return fmt.Errorf("[Enum] %w %q: %s", ErrFailedToRegister, enumId, causedBy)
}

func errInvalidValue(enumId string, expected any, got any) error {
	return fmt.Errorf("[Enum] %w for %s: must be one of %v, got %v", ErrInvalidValue, enumId, expected, got)
}
