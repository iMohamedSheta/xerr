package xerr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iMohamedSheta/xerr"
)

// TestErrorCreation ensures that creating a new XErr sets fields correctly
func TestErrorCreation(t *testing.T) {
	err := xerr.New("invalid input", xerr.ErrUnknown, nil)

	assert.NotNil(t, err)
	assert.Equal(t, xerr.ErrUnknown, err.Type)
	assert.Equal(t, "invalid input", err.Message)
	assert.Nil(t, err.Unwrap())
}

// TestErrorWrapping ensures that wrapping another error works with Unwrap()
func TestErrorWrapping(t *testing.T) {
	base := errors.New("database failed")
	err := xerr.New("could not save user", xerr.ErrUnknown, base)

	assert.NotNil(t, err)
	assert.Equal(t, base, err.Unwrap())
	assert.Contains(t, err.Error(), "could not save user")
	assert.Contains(t, err.Error(), "database failed")
}

// TestErrorAsIs ensures errors.As and errors.Is work properly
func TestErrorAsIs(t *testing.T) {
	base := errors.New("record missing")
	err := xerr.New("user not found", xerr.ErrUnknown, base)

	var target *xerr.XErr
	assert.True(t, errors.As(err, &target))
	assert.Equal(t, xerr.ErrUnknown, target.Type)

	assert.True(t, errors.Is(err, base))
}

// TestStackTraceContainsFunction ensures stack trace contains the current function name
func TestStackTraceContainsFunction(t *testing.T) {
	err := xerr.New("something broke", xerr.ErrUnknown, nil)
	frames := err.StackTrace(false)

	assert.NotEmpty(t, frames)
	found := false
	for _, f := range frames {
		if f.Function != "" && f.File != "" && f.Line > 0 {
			found = true
			break
		}
	}
	assert.True(t, found, "expected at least one valid frame in stack trace")
}

// TestStackTraceWithSnippet ensures stack trace includes snippets when enabled
func TestStackTraceWithSnippet(t *testing.T) {
	err := xerr.New("snippet test", xerr.ErrUnknown, nil)
	frames := err.StackTrace(true)

	assert.NotEmpty(t, frames)
	foundSnippet := false
	for _, f := range frames {
		if f.Snippet != "" {
			foundSnippet = true
			break
		}
	}
	assert.True(t, foundSnippet, "expected at least one frame to contain a snippet")
}

// Define custom error types outside the xerr package
const (
	ErrPaymentFailed xerr.ErrorType = iota + 1000
	ErrRateLimited
)

// TestBuiltinErrorType ensures built-in error type works
func TestBuiltinErrorType(t *testing.T) {
	err := xerr.New("something went wrong", xerr.ErrUnknown, nil)

	assert.NotNil(t, err)
	assert.Equal(t, xerr.ErrUnknown, err.Type)
	assert.Equal(t, "something went wrong", err.Error())
}

// TestCustomErrorTypes ensures custom error types can be created outside the package
func TestCustomErrorTypes(t *testing.T) {
	paymentErr := xerr.New("credit card declined", ErrPaymentFailed, nil)
	rateLimitErr := xerr.New("too many requests", ErrRateLimited, nil)

	assert.NotNil(t, paymentErr)
	assert.NotNil(t, rateLimitErr)

	assert.Equal(t, ErrPaymentFailed, paymentErr.Type)
	assert.Equal(t, "credit card declined", paymentErr.Error())

	assert.Equal(t, ErrRateLimited, rateLimitErr.Type)
	assert.Equal(t, "too many requests", rateLimitErr.Error())
}

// TestWrappedErrorWithCustomType ensures custom error types still work with wrapping
func TestWrappedErrorWithCustomType(t *testing.T) {
	base := errors.New("db timeout")
	err := xerr.New("failed to charge user", ErrPaymentFailed, base)

	assert.NotNil(t, err)
	assert.Equal(t, ErrPaymentFailed, err.Type)
	assert.Contains(t, err.Error(), "failed to charge user")
	assert.Contains(t, err.Error(), "db timeout")
	assert.Equal(t, base, err.Unwrap())
}

// TestWithPublicMessageError ensures setting the custom public messages to error
func TestWithPublicMessageError(t *testing.T) {
	paymentPubMsg := "this is public message for credit card decline"
	rateLimitPubMsg := "this is public message fro too many requests"

	paymentErr := xerr.New("credit card declined", ErrPaymentFailed, nil).WithPublicMessage(paymentPubMsg)
	rateLimitErr := xerr.New("too many requests", ErrRateLimited, nil).WithPublicMessage(rateLimitPubMsg)

	assert.NotNil(t, paymentErr)
	assert.NotNil(t, rateLimitErr)

	assert.Equal(t, ErrPaymentFailed, paymentErr.Type)
	assert.Equal(t, "credit card declined", paymentErr.Error())

	assert.Equal(t, paymentPubMsg, paymentErr.PublicMessage)
	assert.Equal(t, rateLimitPubMsg, rateLimitErr.PublicMessage)

	assert.Equal(t, ErrRateLimited, rateLimitErr.Type)
	assert.Equal(t, "too many requests", rateLimitErr.Error())
}

// TestXErrWithDetailsPublicMessage ensures XErr correctly stores Details and PublicMessage
func TestXErrWithDetailsPublicMessage(t *testing.T) {
	assert := assert.New(t)

	details := map[string]any{
		"email":    "Email is assertd",
		"password": "Password must be at least 6 characters",
	}
	publicMsg := "Some fields are invalid. Please check your input."

	baseErr := errors.New("base error")

	xe := xerr.New("Validation failed", xerr.ErrUnknown, baseErr).
		WithDetails(details).
		WithPublicMessage(publicMsg)

	assert.NotNil(xe)
	assert.Equal(xerr.ErrUnknown, xe.Type)
	assert.Equal("Validation failed - base error", xe.Error())

	assert.Equal(publicMsg, xe.PublicMessage)

	assert.Equal(details, xe.Details)

	newXE := xe.WithDetails(details)
	assert.Equal(xe, newXE)
}
