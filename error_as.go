package xerr

import (
	"errors"
	"slices"
)

// As works like errors.As but also validates against ErrorType if target is *XErr
func As(err error, target any, types ...ErrorType) bool {
	if err == nil {
		return false
	}

	// Try normal errors.As first
	if !errors.As(err, target) {
		return false
	}

	// If caller passed specific types, check them
	if len(types) > 0 {
		var xe *XErr
		if errors.As(err, &xe) {
			// Check if error type is in types
			return slices.Contains(types, xe.Type)
		}
		return false
	}

	return true
}
