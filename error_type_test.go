package xerr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/iMohamedSheta/xerr"
	"github.com/stretchr/testify/assert"
)

const (
	TypeNotFound xerr.ErrorType = iota + 2000
	TypeInvalid
	TypeUnauthorized
	TypeTimeout
)

func TestIsType_NoTypes(t *testing.T) {
	err := xerr.New("missing item", TypeNotFound, nil)
	assert.True(t, err.IsType(), "IsType with no args should return true for non-nil XErr")
}

func TestIsType_MatchingType(t *testing.T) {
	err := xerr.New("invalid input", TypeInvalid, nil)
	assert.True(t, err.IsType(TypeInvalid), "Should match single type")
	assert.True(t, err.IsType(TypeInvalid, TypeNotFound), "Should match one of multiple types")
}

func TestIsType_NonMatchingType(t *testing.T) {
	err := xerr.New("missing item", TypeNotFound, nil)
	assert.False(t, err.IsType(TypeInvalid), "Should return false if type does not match")
	assert.True(t, err.IsType(TypeNotFound), "Should return true for exact match")
}

func TestIsType_NilError(t *testing.T) {
	var err *xerr.XErr
	assert.False(t, err.IsType(TypeNotFound), "Nil XErr should always return false")
	assert.False(t, err.IsType(), "Nil XErr with no types should return false")
}

func TestIsType_WithWrappedError(t *testing.T) {
	inner := xerr.New("timeout", TypeTimeout, nil)

	// Wrap using fmt.Errorf
	wrapped := fmt.Errorf("extra context: %w", inner)

	var xe *xerr.XErr
	ok := errors.As(wrapped, &xe)
	assert.True(t, ok, "errors.As can extract wrapped XErr")
	assert.True(t, xe.IsType(TypeTimeout), "Wrapped XErr should match type")
}

func TestIsType_MultipleTypesEdgeCases(t *testing.T) {
	err := xerr.New("unauthorized", TypeUnauthorized, nil)

	// Large number of types including correct one
	types := []xerr.ErrorType{TypeInvalid, TypeTimeout, TypeUnauthorized, TypeNotFound}
	assert.True(t, err.IsType(types...), "Should match type even in large slice")

	// All non-matching types
	nonMatch := []xerr.ErrorType{TypeInvalid, TypeTimeout, TypeNotFound}
	assert.False(t, err.IsType(nonMatch...), "Should return false if type not in slice")
}

func TestIsType_ChainedXErrs(t *testing.T) {
	// Chain errors using fmt.Errorf
	inner := xerr.New("inner", TypeNotFound, nil)
	mid := fmt.Errorf("mid layer: %w", inner)
	outer := fmt.Errorf("outer layer: %w", mid)

	var xe *xerr.XErr
	ok := errors.As(outer, &xe)
	assert.True(t, ok, "errors.As should extract the innermost XErr")
	assert.True(t, xe.IsType(TypeNotFound), "Type should match innermost XErr")
	assert.False(t, xe.IsType(TypeInvalid), "Non-matching type should return false")
}
