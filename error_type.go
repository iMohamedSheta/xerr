package xerr

import (
	"slices"
)

// IsType checks if the XErr is one of the specified types.
// If no types are provided, it returns true if err is not nil.
func (err *XErr) IsType(types ...ErrorType) bool {
	if err == nil {
		return false
	}

	if types == nil || len(types) == 0 {
		return true
	}

	return slices.Contains(types, err.Type)
}
