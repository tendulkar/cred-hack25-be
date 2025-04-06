package structured

import (
	"fmt"
)

// Error represents a structured error type
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new Error with the given code and message
func NewError(code, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}
