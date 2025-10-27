package errors

import "errors"

// Database related errors
var (
	// ErrResourceNotFound 表示资源未找到
	ErrResourceNotFound = errors.New("resource not found")

	// ErrDuplicate 表示违反唯一性约束（重复记录）
	ErrDuplicate = errors.New("resource already exists")

	// ErrForeignKeyViolation 表示违反外键约束
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")

	// ErrInvalidTransaction 表示无效的事务操作
	ErrInvalidTransaction = errors.New("invalid transaction")

	// ErrInvalidData 表示无效的数据
	ErrInvalidData = errors.New("invalid data")

	// ErrConnectionFailed 表示数据库连接失败
	ErrConnectionFailed = errors.New("database connection failed")

	// ErrDeadlock 表示发生死锁
	ErrDeadlock = errors.New("database deadlock detected")

	// ErrLockWaitTimeout 表示锁等待超时
	ErrLockWaitTimeout = errors.New("lock wait timeout exceeded")

	// ErrCheckConstraintViolation 表示违反检查约束
	ErrCheckConstraintViolation = errors.New("check constraint violation")

	// ErrNotNullViolation 表示违反非空约束
	ErrNotNullViolation = errors.New("not null constraint violation")

	// ErrDataTooLong 表示数据过长
	ErrDataTooLong = errors.New("data too long for column")

	// ErrInvalidDataType 表示数据类型不匹配
	ErrInvalidDataType = errors.New("invalid data type")
)
