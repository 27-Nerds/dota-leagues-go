// Package e defines application errors and their classification.
package e

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

const (
	ECONFLICT = "conflict"  // action cannot be performed
	EINTERNAL = "internal"  // internal error
	EINVALID  = "invalid"   // validation failed
	ENOTFOUND = "not_found" // entity does not exist
)

// Error defines a standard application error.
type Error struct {
	// Machine-readable error code.
	Code string

	// Human-readable message.
	Message string

	// Logical operation and nested error.
	Op  string
	Err error
}

// Unwrap exposes the underlying error to errors.Is and errors.As.
func (e *Error) Unwrap() error { return e.Err }

// Error returns the string representation of the error message.
func (e *Error) Error() string {
	var buf bytes.Buffer

	// Print the current operation in our stack, if any.
	if e.Op != "" {
		fmt.Fprintf(&buf, "%s: ", e.Op)
	}

	// If wrapping an error, print its Error() message.
	// Otherwise print the error code & message.
	if e.Err != nil {
		buf.WriteString(e.Err.Error())
	} else {
		if e.Code != "" {
			fmt.Fprintf(&buf, "<%s> ", e.Code)
		}
		buf.WriteString(e.Message)
	}
	return buf.String()
}

// ErrorMessage returns the human-readable message of the error, if available.
// Otherwise returns a generic error message.
func ErrorMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return err.Error()
	}

	var appError *Error
	if err == nil {
		return ""
	} else if errors.As(err, &appError) {
		if appError.Message != "" {
			return appError.Message
		}
		if appError.Err != nil {
			return ErrorMessage(appError.Err)
		}
	}
	return "An internal error has occurred. Please contact technical support."
}

// ErrorCode returns the code of the root error, if available. Otherwise returns EINTERNAL.
func ErrorCode(err error) string {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return EINVALID
	}

	var appError *Error
	if err == nil {
		return ""
	} else if errors.As(err, &appError) {
		if appError.Code != "" {
			return appError.Code
		}
		if appError.Err != nil {
			return ErrorCode(appError.Err)
		}
	}
	return EINTERNAL
}

// IsNotFound return true if error has notfound code
func IsNotFound(err error) bool {

	return ErrorCode(err) == ENOTFOUND
}
