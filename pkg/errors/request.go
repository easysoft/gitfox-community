package errors

import "errors"

// Request related errors
var (
	// ErrParameterMissing 表示缺少必需的参数
	ErrParameterMissing = errors.New("parameter missing")

	// ErrParameterInvalid 表示参数无效
	ErrParameterInvalid = errors.New("parameter invalid")

	// ErrPathParameterNotFound 表示路径参数未找到
	ErrPathParameterNotFound = errors.New("path parameter not found")

	// ErrQueryParameterNotFound 表示查询参数未找到
	ErrQueryParameterNotFound = errors.New("query parameter not found")

	// ErrParameterDecodeFailed 表示参数解码失败
	ErrParameterDecodeFailed = errors.New("parameter decode failed")

	// ErrParameterTypeMismatch 表示参数类型不匹配
	ErrParameterTypeMismatch = errors.New("parameter type mismatch")

	// ErrInvalidIntegerParameter 表示整数参数无效
	ErrInvalidIntegerParameter = errors.New("parameter must be a valid integer")

	// ErrInvalidPositiveIntegerParameter 表示正整数参数无效
	ErrInvalidPositiveIntegerParameter = errors.New("parameter must be a positive integer")

	// ErrInvalidBooleanParameter 表示布尔参数无效
	ErrInvalidBooleanParameter = errors.New("parameter must be a boolean")

	// ErrInvalidRequestBody 表示请求体无效
	ErrInvalidRequestBody = errors.New("invalid request body")

	// ErrInvalidJSON 表示JSON格式无效
	ErrInvalidJSON = errors.New("invalid JSON format")

	// ErrInvalidContentType 表示Content-Type无效
	ErrInvalidContentType = errors.New("invalid content type")

	// ErrRequestTooLarge 表示请求体过大
	ErrRequestTooLarge = errors.New("request body too large")

	// ErrInvalidURLParameter 表示URL参数无效
	ErrInvalidURLParameter = errors.New("invalid URL parameter")
)
