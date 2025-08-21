package xerr

import (
	"fmt"
	"runtime"
)

// ErrorType is an enum for categorizing errors
type ErrorType int

const (
	ErrUnknown ErrorType = iota
)

// XErr is a custom error with stack trace and type
type XErr struct {
	Type          ErrorType
	Message       string
	PublicMessage string
	Err           error
	stack         []uintptr
	Details       map[string]any
}

// Error creates a new XErr with stack trace
func New(msg string, t ErrorType, err error) *XErr {
	stack := make([]uintptr, 32)
	n := runtime.Callers(2, stack[:])
	return &XErr{
		Type:    t,
		Message: msg,
		Err:     err,
		stack:   stack[:n],
	}
}

// Adds public message to the error
func (e *XErr) WithPublicMessage(msg string) *XErr {
	e.PublicMessage = msg
	return e
}

// Adds details to the error
func (e *XErr) WithDetails(details map[string]any) *XErr {
	e.Details = details
	return e
}

func (e *XErr) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s - %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *XErr) Unwrap() error {
	return e.Err
}

// StackTrace builds structured frames (like your ErrorHandler does)
func (e *XErr) StackTrace(showSource bool) []Frame {
	frames := runtime.CallersFrames(e.stack)
	var result []Frame
	for {
		fr, more := frames.Next()
		if fr.File != "" {
			frame := Frame{
				Function: fr.Function,
				File:     fr.File,
				Line:     fr.Line,
			}
			if showSource {
				frame.Snippet = codeSnippet(fr.File, fr.Line)
			}
			result = append(result, frame)
		}
		if !more {
			break
		}
	}
	return result
}
