package enums

import (
	"errors"
	"fmt"
)

var (
	// ErrNotRegisteredYet is returned when the given Numeric type is not registered in the internal registry.
	// Register the Numeric values to fix this error.
	ErrNotRegisteredYet = errors.New("enum not registered yet")
	// ErrInvalidValue is returned when the given value is not one of the registered values of the given Numeric type.
	ErrInvalidValue = errors.New("invalid enum value")
)

func errNotRegisteredYet(enumName string) error {
	return fmt.Errorf("[Enum] %q %w", enumName, ErrNotRegisteredYet)
}

func errInvalidValue(enumId string, expected any, got any) error {
	return fmt.Errorf("[Enum] %w for %s: must be one of %v, got %v", ErrInvalidValue, enumId, expected, got)
}
