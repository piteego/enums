package enums

import (
	"fmt"
	"reflect"
)

// typeNameOf returns the type name of the given value using
//
//   - reflect.TypeOf(e).PkgPath() if the descriptive argument is true
//
//   - fmt.Sprintf and %T verb if the descriptive argument is false.
func typeNameOf(e any, descriptive bool) string {
	if !descriptive {
		return fmt.Sprintf("%T", e)
	}
	reflectType := reflect.TypeOf(e)
	return fmt.Sprintf("%s.%s", reflectType.PkgPath(), reflectType.Name())
}

// isUnique checks if all elements in the slice are unique
func isUnique[T comparable](slice []T) bool {
	seen := make(map[T]bool) // Map to track seen elements
	for i := range slice {
		if seen[slice[i]] {
			return false // Duplicate found
		}
		seen[slice[i]] = true
	}
	return true // No duplicates found
}

func Is[E enumable](enum, target E, or ...E) bool {
	if enum == target {
		return true
	}
	for i := range or {
		if enum == or[i] {
			return true
		}
	}
	return false
}
