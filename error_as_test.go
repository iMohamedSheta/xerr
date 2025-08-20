package xerr_test

import (
	"errors"
	"testing"

	"github.com/iMohamedSheta/xerr"
	"github.com/stretchr/testify/assert"
)

// Define custom ErrorType
const (
	TypeNotFound xerr.ErrorType = iota + 2000
	TypeInvalid
)

func TestAs_NoError(t *testing.T) {
	var err error
	var target *xerr.XErr
	ok := xerr.As(err, &target)
	assert.False(t, ok, "As should return false for nil error")
}

func TestAs_NormalError(t *testing.T) {
	err := errors.New("plain error")
	var target *xerr.XErr
	ok := xerr.As(err, &target)
	assert.False(t, ok, "As should return false for normal errors")
}

func TestAs_XErrWithoutTypes(t *testing.T) {
	err := xerr.New("missing item", TypeNotFound, nil)
	var target *xerr.XErr
	ok := xerr.As(err, &target)
	assert.True(t, ok, "As should return true for *XErr with no type filtering")
	assert.Equal(t, TypeNotFound, target.Type)
}

func TestAs_XErrWithMatchingType(t *testing.T) {
	err := xerr.New("invalid input", TypeInvalid, nil)
	var target *xerr.XErr
	ok := xerr.As(err, &target, TypeInvalid, TypeNotFound)
	assert.True(t, ok, "As should return true when type matches")
	assert.Equal(t, TypeInvalid, target.Type)
}

func TestAs_XErrWithNonMatchingType(t *testing.T) {
	err := xerr.New("missing item", TypeNotFound, nil)
	var target *xerr.XErr
	ok := xerr.As(err, &target, TypeInvalid)

	assert.False(t, ok, "As should return false when type does not match")
	assert.NotNil(t, target, "target should still be set by errors.As")
	assert.Equal(t, TypeNotFound, target.Type, "target type should remain the actual error type")
}
