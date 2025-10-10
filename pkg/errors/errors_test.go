package pkg_errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorType_New(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		message   string
		want      string
	}{
		{
			name:      "创建BadRequest错误",
			errorType: BadRequest,
			message:   "invalid parameter",
			want:      "invalid parameter",
		},
		{
			name:      "创建NotFound错误",
			errorType: NotFound,
			message:   "user not found",
			want:      "user not found",
		},
		{
			name:      "创建DBError错误",
			errorType: DBError,
			message:   "database connection failed",
			want:      "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorType.New(tt.message)
			if err.Error() != tt.want {
				t.Errorf("ErrorType.New() error = %v, want %v", err.Error(), tt.want)
			}
			if GetType(err) != tt.errorType {
				t.Errorf("GetType() = %v, want %v", GetType(err), tt.errorType)
			}
		})
	}
}

func TestErrorType_Newf(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		format    string
		args      []interface{}
		want      string
	}{
		{
			name:      "格式化BadRequest错误",
			errorType: BadRequest,
			format:    "invalid parameter: %s",
			args:      []interface{}{"username"},
			want:      "invalid parameter: username",
		},
		{
			name:      "格式化NotFound错误",
			errorType: NotFound,
			format:    "user %d not found",
			args:      []interface{}{123},
			want:      "user 123 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorType.Newf(tt.format, tt.args...)
			if err.Error() != tt.want {
				t.Errorf("ErrorType.Newf() error = %v, want %v", err.Error(), tt.want)
			}
			if GetType(err) != tt.errorType {
				t.Errorf("GetType() = %v, want %v", GetType(err), tt.errorType)
			}
		})
	}
}

func TestErrorType_Wrap(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name      string
		errorType ErrorType
		err       error
		message   string
		wantNil   bool
	}{
		{
			name:      "包装现有错误",
			errorType: DBError,
			err:       originalErr,
			message:   "database operation failed",
			wantNil:   false,
		},
		{
			name:      "包装nil错误",
			errorType: DBError,
			err:       nil,
			message:   "should return nil",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorType.Wrap(tt.err, tt.message)
			if tt.wantNil {
				if err != nil {
					t.Errorf("ErrorType.Wrap() should return nil for nil input")
				}
				return
			}

			if err == nil {
				t.Errorf("ErrorType.Wrap() should not return nil for valid input")
				return
			}

			if GetType(err) != tt.errorType {
				t.Errorf("GetType() = %v, want %v", GetType(err), tt.errorType)
			}
		})
	}
}

func TestErrorType_Wrapf(t *testing.T) {
	originalErr := errors.New("connection timeout")

	err := DBError.Wrapf(originalErr, "failed to connect to database %s", "mysql")

	if err == nil {
		t.Fatal("Wrapf should not return nil")
	}

	if GetType(err) != DBError {
		t.Errorf("GetType() = %v, want %v", GetType(err), DBError)
	}

	// 测试nil错误
	nilErr := DBError.Wrapf(nil, "should return nil")
	if nilErr != nil {
		t.Errorf("Wrapf should return nil for nil input")
	}
}

func TestNew(t *testing.T) {
	err := New("test error")

	if err == nil {
		t.Fatal("New should not return nil")
	}

	if err.Error() != "test error" {
		t.Errorf("Error() = %v, want %v", err.Error(), "test error")
	}

	if GetType(err) != NoType {
		t.Errorf("GetType() = %v, want %v", GetType(err), NoType)
	}
}

func TestNewf(t *testing.T) {
	err := Newf("user %s not found with id %d", "john", 123)

	if err == nil {
		t.Fatal("Newf should not return nil")
	}

	expected := "user john not found with id 123"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}

	if GetType(err) != NoType {
		t.Errorf("GetType() = %v, want %v", GetType(err), NoType)
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")

	// 测试包装现有错误
	err := Wrap(originalErr, "wrapped message")
	if err == nil {
		t.Fatal("Wrap should not return nil")
	}

	if GetType(err) != NoType {
		t.Errorf("GetType() = %v, want %v", GetType(err), NoType)
	}

	// 测试包装nil错误
	nilErr := Wrap(nil, "should return nil")
	if nilErr != nil {
		t.Errorf("Wrap should return nil for nil input")
	}
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("connection failed")

	err := Wrapf(originalErr, "failed to connect to %s:%d", "localhost", 3306)

	if err == nil {
		t.Fatal("Wrapf should not return nil")
	}

	if GetType(err) != NoType {
		t.Errorf("GetType() = %v, want %v", GetType(err), NoType)
	}

	// 测试包装自定义错误类型
	customErr := DBError.New("database error")
	wrappedErr := Wrapf(customErr, "operation failed")

	if GetType(wrappedErr) != DBError {
		t.Errorf("GetType() = %v, want %v", GetType(wrappedErr), DBError)
	}
}

func TestWithMessage(t *testing.T) {
	originalErr := errors.New("original error")

	err := WithMessage(originalErr, "additional context")

	if err == nil {
		t.Fatal("WithMessage should not return nil")
	}

	// 测试nil错误
	nilErr := WithMessage(nil, "should return nil")
	if nilErr != nil {
		t.Errorf("WithMessage should return nil for nil input")
	}

	// 测试自定义错误
	customErr := BadRequest.New("bad request")
	msgErr := WithMessage(customErr, "validation failed")

	if GetType(msgErr) != BadRequest {
		t.Errorf("GetType() = %v, want %v", GetType(msgErr), BadRequest)
	}
}

func TestWithMessagef(t *testing.T) {
	originalErr := errors.New("original error")

	err := WithMessagef(originalErr, "user %s operation failed", "john")

	if err == nil {
		t.Fatal("WithMessagef should not return nil")
	}

	// 测试nil错误
	nilErr := WithMessagef(nil, "should return nil")
	if nilErr != nil {
		t.Errorf("WithMessagef should return nil for nil input")
	}
}

func TestAddErrorContext(t *testing.T) {
	originalErr := errors.New("validation error")

	err := AddErrorContext(originalErr, "username", "username is required")

	if err == nil {
		t.Fatal("AddErrorContext should not return nil")
	}

	context := GetErrorContext(err)
	if context == nil {
		t.Fatal("GetErrorContext should not return nil")
	}

	if context["field"] != "username" {
		t.Errorf("context field = %v, want %v", context["field"], "username")
	}

	if context["message"] != "username is required" {
		t.Errorf("context message = %v, want %v", context["message"], "username is required")
	}

	// 测试nil错误
	nilErr := AddErrorContext(nil, "field", "message")
	if nilErr != nil {
		t.Errorf("AddErrorContext should return nil for nil input")
	}
}

func TestGetErrorContext(t *testing.T) {
	// 测试没有上下文的错误
	simpleErr := errors.New("simple error")
	context := GetErrorContext(simpleErr)
	if context != nil {
		t.Errorf("GetErrorContext should return nil for error without context")
	}

	// 测试有上下文的错误
	errWithContext := AddErrorContext(simpleErr, "email", "invalid email format")
	context = GetErrorContext(errWithContext)

	if context == nil {
		t.Fatal("GetErrorContext should not return nil")
	}

	if context["field"] != "email" {
		t.Errorf("context field = %v, want %v", context["field"], "email")
	}

	if context["message"] != "invalid email format" {
		t.Errorf("context message = %v, want %v", context["message"], "invalid email format")
	}
}

func TestGetType(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorType
	}{
		{
			name: "标准错误",
			err:  errors.New("standard error"),
			want: NoType,
		},
		{
			name: "BadRequest错误",
			err:  BadRequest.New("bad request"),
			want: BadRequest,
		},
		{
			name: "NotFound错误",
			err:  NotFound.New("not found"),
			want: NotFound,
		},
		{
			name: "DBError错误",
			err:  DBError.New("database error"),
			want: DBError,
		},
		{
			name: "SystemErr错误",
			err:  SystemErr.New("system error"),
			want: SystemErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetType(tt.err); got != tt.want {
				t.Errorf("GetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "NotFound错误",
			err:  NotFound.New("user not found"),
			want: true,
		},
		{
			name: "BadRequest错误",
			err:  BadRequest.New("bad request"),
			want: false,
		},
		{
			name: "标准错误",
			err:  errors.New("standard error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIs(t *testing.T) {
	targetErr := errors.New("target error")
	wrappedErr := Wrap(targetErr, "wrapped")

	if !Is(wrappedErr, targetErr) {
		t.Errorf("Is() should return true for wrapped error")
	}

	otherErr := errors.New("other error")
	if Is(wrappedErr, otherErr) {
		t.Errorf("Is() should return false for different error")
	}
}

func TestCause(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, "wrapped")

	cause := Cause(wrappedErr)
	if cause.Error() != originalErr.Error() {
		t.Errorf("Cause() = %v, want %v", cause.Error(), originalErr.Error())
	}
}

// 基准测试
func BenchmarkErrorType_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BadRequest.New("benchmark error")
	}
}

func BenchmarkErrorType_Newf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BadRequest.Newf("benchmark error %d", i)
	}
}

func BenchmarkWrap(b *testing.B) {
	originalErr := errors.New("original error")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Wrap(originalErr, "wrapped error")
	}
}

func BenchmarkAddErrorContext(b *testing.B) {
	originalErr := errors.New("original error")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = AddErrorContext(originalErr, "field", "message")
	}
}

// 示例测试
func ExampleErrorType_New() {
	err := BadRequest.New("invalid username")
	fmt.Println(err.Error())
	fmt.Println(GetType(err))
	// Output:
	// invalid username
	// 1
}

func ExampleErrorType_Wrap() {
	originalErr := errors.New("connection timeout")
	err := DBError.Wrap(originalErr, "failed to query database")

	fmt.Println(GetType(err))
	fmt.Println(IsNotFound(err))
	// Output:
	// 3
	// false
}

func ExampleAddErrorContext() {
	err := BadRequest.New("validation failed")
	errWithContext := AddErrorContext(err, "email", "email format is invalid")

	context := GetErrorContext(errWithContext)
	fmt.Println(context["field"])
	fmt.Println(context["message"])
	// Output:
	// email
	// email format is invalid
}
