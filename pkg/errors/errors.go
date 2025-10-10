package pkg_errors

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrorType is the type of an error
type ErrorType uint
type ErrorCode uint

type customError struct {
	errorType     ErrorType
	errorCodes    []ErrorCode
	originalError error
	context       errorContext
}

type errorContext struct {
	Field   string
	Message string
}

func Is(err, target error) bool { return errors.Is(err, target) }

// New creates a new customError
func (errorType ErrorType) New(msg string) error {
	return customError{errorType: errorType, originalError: errors.New(msg)}
}

// Newf creates a new customError with formatted message
func (errorType ErrorType) Newf(msg string, args ...interface{}) error {
	return customError{errorType: errorType, originalError: errors.Errorf(msg, args...)}
}

// Wrap creates a new wrapped error
func (errorType ErrorType) Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return errorType.Wrapf(err, "%s", msg)
}

// Wrapf creates a new wrapped error with formatted message
func (errorType ErrorType) Wrapf(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return customError{errorType: errorType, originalError: errors.Wrapf(err, msg, args...)}
}

// Error returns the mssage of a customError
func (error customError) Error() string {
	return error.originalError.Error()
}

// Unwrap returns the wrapped error for Go 1.13+ error handling
func (error customError) Unwrap() error {
	return error.originalError
}

// Is reports whether any error in error's chain matches target
func (error customError) Is(target error) bool {
	if targetCustom, ok := target.(customError); ok {
		return error.errorType == targetCustom.errorType && errors.Is(error.originalError, targetCustom.originalError)
	}
	return errors.Is(error.originalError, target)
}

// New creates a no type error
func New(msg string) error {
	return customError{errorType: NoType, originalError: errors.New(msg)}
}

// Newf creates a no type error with formatted message
func Newf(msg string, args ...interface{}) error {
	return customError{errorType: NoType, originalError: errors.New(fmt.Sprintf(msg, args...))}
}

// Wrap an error with a string
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return Wrapf(err, "%s", msg)
}

// Cause gives the original error
func Cause(err error) error {
	if customErr, ok := err.(customError); ok {
		return errors.Cause(customErr.originalError)
	}
	return errors.Cause(err)
}

// Wrapf an error with format string
func Wrapf(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	wrappedError := errors.Wrapf(err, msg, args...)
	if customErr, ok := err.(customError); ok {
		return customError{
			errorType:     customErr.errorType,
			originalError: wrappedError,
			context:       customErr.context,
		}
	}

	return customError{errorType: NoType, originalError: wrappedError}
}

// WithMessage adds a message to an error
func WithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}

	if customErr, ok := err.(customError); ok {
		return customError{errorType: customErr.errorType, originalError: errors.WithMessage(customErr.originalError, msg)}
	}

	return customError{errorType: NoType, originalError: errors.WithMessage(err, msg)}
}

// WithMessagef adds a message to an error
func WithMessagef(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if customErr, ok := err.(customError); ok {
		return customError{errorType: customErr.errorType, originalError: errors.WithMessagef(customErr.originalError, msg, args...)}
	}

	return customError{errorType: NoType, originalError: errors.WithMessagef(err, msg, args...)}
}

// AddErrorContext adds a context to an error
func AddErrorContext(err error, field, message string) error {
	if err == nil {
		return nil
	}
	context := errorContext{Field: field, Message: message}
	if customErr, ok := err.(customError); ok {
		return customError{errorType: customErr.errorType, originalError: customErr.originalError, context: context}
	}

	return customError{errorType: NoType, originalError: err, context: context}
}

// GetErrorContext returns the error context
func GetErrorContext(err error) map[string]string {
	emptyContext := errorContext{}
	if customErr, ok := err.(customError); ok || customErr.context != emptyContext {
		return map[string]string{"field": customErr.context.Field, "message": customErr.context.Message}
	}

	return nil
}

// GetType returns the error type
func GetType(err error) ErrorType {
	if customErr, ok := err.(customError); ok {
		return customErr.errorType
	}

	return NoType
}

func IsNotFound(err error) bool {
	return GetType(err) == NotFound
}
