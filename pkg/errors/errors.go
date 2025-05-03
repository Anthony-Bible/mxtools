// Package errors provides error handling functionality for the MXToolbox clone.
package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType represents the type of error.
type ErrorType string

const (
	// ErrorTypeNetwork represents a network error.
	ErrorTypeNetwork ErrorType = "network"
	// ErrorTypeDNS represents a DNS error.
	ErrorTypeDNS ErrorType = "dns"
	// ErrorTypeBlacklist represents a blacklist error.
	ErrorTypeBlacklist ErrorType = "blacklist"
	// ErrorTypeSMTP represents an SMTP error.
	ErrorTypeSMTP ErrorType = "smtp"
	// ErrorTypeAuth represents an authentication error.
	ErrorTypeAuth ErrorType = "auth"
	// ErrorTypeTimeout represents a timeout error.
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeInternal represents an internal error.
	ErrorTypeInternal ErrorType = "internal"
)

// AppError represents an application error.
type AppError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Cause   error     `json:"-"`
	Stack   string    `json:"-"`
}

// Error returns the error message.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Cause.Error())
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError.
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Stack:   getStack(),
	}
}

// Wrap wraps an error with additional context.
func Wrap(err error, errType ErrorType, message string) *AppError {
	if err == nil {
		return nil
	}

	// If the error is already an AppError, just update the message
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Type:    appErr.Type,
			Message: message,
			Cause:   appErr,
			Stack:   appErr.Stack,
		}
	}

	return &AppError{
		Type:    errType,
		Message: message,
		Cause:   err,
		Stack:   getStack(),
	}
}

// WrapWithType wraps an error with a specific error type.
func WrapWithType(err error, errType ErrorType) *AppError {
	if err == nil {
		return nil
	}

	// If the error is already an AppError, just update the type
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Type:    errType,
			Message: appErr.Message,
			Cause:   appErr.Cause,
			Stack:   appErr.Stack,
		}
	}

	return &AppError{
		Type:    errType,
		Message: err.Error(),
		Cause:   err,
		Stack:   getStack(),
	}
}

// IsType checks if an error is of a specific type.
func IsType(err error, errType ErrorType) bool {
	if err == nil {
		return false
	}

	// Check if the error is an AppError
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errType
	}

	return false
}

// getStack returns the stack trace.
func getStack() string {
	var sb strings.Builder
	for i := 2; i < 15; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}
		name := fn.Name()
		if strings.Contains(name, "runtime.") {
			break
		}
		sb.WriteString(fmt.Sprintf("%s:%d %s\n", file, line, name))
	}
	return sb.String()
}

// FormatError formats an error for display.
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	// Check if the error is an AppError
	if appErr, ok := err.(*AppError); ok {
		if appErr.Cause != nil {
			return fmt.Sprintf("[%s] %s: %s", appErr.Type, appErr.Message, appErr.Cause.Error())
		}
		return fmt.Sprintf("[%s] %s", appErr.Type, appErr.Message)
	}

	return err.Error()
}