package errors

import (
	"fmt"

	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/pkg/errors"
)

// ErrorType is the type of an error
type ErrorType uint

type customError struct {
	errorType     ErrorType
	i18nKey       i18n.I18nKey
	httpStatus    int
	originalError error
	context       errorContext
}

type errorContext struct {
	Field   string
	Message string
}

func Is(err, target error) bool { return errors.Is(err, target) }

// As 查找 err 链中是否有与 target 匹配的错误
func As(err error, target any) bool { return errors.As(err, target) }

// New 创建一个新的customError
func (errorType ErrorType) New(msg string) error {
	return customError{errorType: errorType, originalError: errors.New(msg)}
}

// Newf 使用格式化消息创建新的customError
func (errorType ErrorType) Newf(msg string, args ...interface{}) error {
	return customError{errorType: errorType, originalError: errors.Errorf(msg, args...)}
}

// Wrap 创建一个新的包装错误
func (errorType ErrorType) Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return errorType.Wrapf(err, "%s", msg)
}

// Wrapf 创建新的包装错误，格式化消息
func (errorType ErrorType) Wrapf(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return customError{errorType: errorType, originalError: errors.Wrapf(err, msg, args...)}
}

// Error 返回自定义错误的消息
func (error customError) Error() string {
	return error.originalError.Error()
}

// Unwrap 返回Go 1.13错误处理的包装错误
func (error customError) Unwrap() error {
	return error.originalError
}

// Is 报告错误链中是否有匹配目标的错误
func (error customError) Is(target error) bool {
	if targetCustom, ok := target.(customError); ok {
		return error.errorType == targetCustom.errorType && errors.Is(error.originalError, targetCustom.originalError)
	}
	return errors.Is(error.originalError, target)
}

// New 创建无类型错误
func New(msg string) error {
	return customError{errorType: NoType, originalError: errors.New(msg)}
}

// Newf 创建无类型错误，格式化消息
func Newf(msg string, args ...interface{}) error {
	return customError{errorType: NoType, originalError: errors.New(fmt.Sprintf(msg, args...))}
}

// Wrap 包装字符串错误
func Wrap(err error, msg i18n.I18nKey) error {
	if err == nil {
		return nil
	}
	return Wrapf(err, "%s", msg)
}

// Wrapf 包装格式字符串错误
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

// Cause 提供原始错误
func Cause(err error) error {
	if customErr, ok := err.(customError); ok {
		return errors.Cause(customErr.originalError)
	}
	return errors.Cause(err)
}

// WithMessage 添加错误消息
func WithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}

	if customErr, ok := err.(customError); ok {
		return customError{errorType: customErr.errorType, originalError: errors.WithMessage(customErr.originalError, msg)}
	}

	return customError{errorType: NoType, originalError: errors.WithMessage(err, msg)}
}

// WithI18nKey 添加国际化键
func WithI18nKey(err error, i18nKey i18n.I18nKey) error {
	if customErr, ok := err.(customError); ok {
		return customError{
			errorType:     customErr.errorType,
			i18nKey:       i18nKey,
			httpStatus:    customErr.httpStatus,
			originalError: customErr.originalError,
			context:       customErr.context,
		}
	}

	// 如果 err 为 nil，创建一个新的错误，使用 i18nKey 作为消息
	if err == nil {
		return customError{errorType: NoType, i18nKey: i18nKey, originalError: errors.New(string(i18nKey))}
	}

	return customError{errorType: NoType, i18nKey: i18nKey, originalError: err}
}

// WithHttpStatus 添加HTTP状态码
func WithHttpStatus(err error, httpStatus int) error {
	if err == nil {
		return nil
	}

	if customErr, ok := err.(customError); ok {
		return customError{
			errorType:     customErr.errorType,
			i18nKey:       customErr.i18nKey,
			httpStatus:    httpStatus,
			originalError: customErr.originalError,
			context:       customErr.context,
		}
	}

	return customError{errorType: NoType, httpStatus: httpStatus, originalError: err}
}

// GetHttpStatus 返回HTTP状态码
func GetHttpStatus(err error) int {
	if customErr, ok := err.(customError); ok {
		return customErr.httpStatus
	}

	return defaultHttpStatus
}

// GetI18nKey 返回国际化键
func GetI18nKey(err error) i18n.I18nKey {
	if customErr, ok := err.(customError); ok {
		return customErr.i18nKey
	}

	return defaultI18nKey
}

// WithMessagef 添加错误消息
func WithMessagef(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if customErr, ok := err.(customError); ok {
		return customError{errorType: customErr.errorType, originalError: errors.WithMessagef(customErr.originalError, msg, args...)}
	}

	return customError{errorType: NoType, originalError: errors.WithMessagef(err, msg, args...)}
}

// AddErrorContext 添加错误上下文
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

// GetErrorContext 返回错误上下文
func GetErrorContext(err error) map[string]string {
	emptyContext := errorContext{}
	if customErr, ok := err.(customError); ok || customErr.context != emptyContext {
		return map[string]string{"field": customErr.context.Field, "message": customErr.context.Message}
	}

	return nil
}

// GetType 返回错误类型
func GetType(err error) ErrorType {
	if customErr, ok := err.(customError); ok {
		return customErr.errorType
	}

	return NoType
}

func IsNotFound(err error) bool {
	return GetType(err) == NotFound
}

// IsConflict 检查错误是否为冲突类型错误
func IsConflict(err error) bool {
	return GetType(err) == Conflict
}
