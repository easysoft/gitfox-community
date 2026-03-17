package database

import (
	"context"
	"database/sql"
	stderr "errors"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// default query range limit.
const defaultLimit = 100

// limit returns the page size to a sql limit.
func Limit(size int) uint64 {
	if size == 0 {
		size = defaultLimit
	}
	return uint64(size)
}

func GormLimit(size int) int {
	if size == 0 {
		size = defaultLimit
	}
	return size
}

// offset converts the page to a sql offset.
func Offset(page, size int) uint64 {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = defaultLimit
	}
	page--
	return uint64(page * size)
}

func GormOffset(page, size int) int {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = defaultLimit
	}
	page--
	return page * size
}

// ProcessSQLErrorf 处理 SQL 错误（包括 GORM 错误）并转换为应用层 DBError 类型错误
func ProcessSQLErrorf(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// 识别特定的 SQL 错误类型，并包装底层错误
	// 按照错误出现频率排序，最常见的错误放在前面以提高性能
	var translatedError error
	switch {
	// 1. 最常见：记录未找到错误（查询操作中最常见）
	case stderr.Is(err, gorm.ErrRecordNotFound):
		translatedError = errors.ErrResourceNotFound
	case stderr.Is(err, sql.ErrNoRows):
		translatedError = errors.ErrResourceNotFound

	// 2. 常见：唯一约束冲突（创建/更新操作中常见）
	case isSQLUniqueConstraintError(err):
		translatedError = errors.ErrDuplicate

	// 3. 常见：外键约束冲突（删除/更新操作中常见）
	case isSQLForeignKeyViolationError(err):
		translatedError = errors.ErrForeignKeyViolation

	// 4. 较常见：非空约束错误
	case isSQLNotNullViolation(err):
		translatedError = errors.ErrNotNullViolation

	// 5. 较常见：数据过长错误
	case isSQLDataTooLong(err):
		translatedError = errors.ErrDataTooLong

	// 6. 一般：GORM 特定错误
	case stderr.Is(err, gorm.ErrInvalidData):
		translatedError = errors.ErrInvalidData
	case stderr.Is(err, gorm.ErrInvalidTransaction):
		translatedError = errors.ErrInvalidTransaction

	// 7. 较少见：数据类型不匹配
	case isSQLInvalidDataType(err):
		translatedError = errors.ErrInvalidDataType

	// 8. 较少见：检查约束错误
	case isSQLCheckConstraintViolation(err):
		translatedError = errors.ErrCheckConstraintViolation

	// 9. 少见但重要：并发和锁相关错误
	case isSQLDeadlock(err):
		translatedError = errors.ErrDeadlock
	case isSQLLockWaitTimeout(err):
		translatedError = errors.ErrLockWaitTimeout

	// 10. 少见：连接错误
	case stderr.Is(err, sql.ErrConnDone):
		translatedError = errors.ErrConnectionFailed

	// 12. 默认：其他数据库错误
	default:
		translatedError = err
	}

	return translatedError
}

// isSQLUniqueConstraintError 检查是否是唯一约束错误
func isSQLUniqueConstraintError(original error) bool {
	if original == nil {
		return false
	}

	// MySQL: 错误码 1062 表示唯一约束冲突
	var mysqlErr *mysql.MySQLError
	if stderr.As(original, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}

// isSQLForeignKeyViolationError 检查是否是外键约束错误
func isSQLForeignKeyViolationError(original error) bool {
	if original == nil {
		return false
	}

	// MySQL: 错误码 1451 和 1452 表示外键约束冲突
	var mysqlErr *mysql.MySQLError
	if stderr.As(original, &mysqlErr) {
		// 1451: Cannot delete or update a parent row
		// 1452: Cannot add or update a child row
		return mysqlErr.Number == 1451 || mysqlErr.Number == 1452
	}
	return false
}

// isSQLNotNullViolation 检查是否是非空约束错误
func isSQLNotNullViolation(original error) bool {
	if original == nil {
		return false
	}

	// MySQL: 错误码 1048 表示列不能为 NULL
	var mysqlErr *mysql.MySQLError
	if stderr.As(original, &mysqlErr) {
		return mysqlErr.Number == 1048
	}
	return false
}

// isSQLCheckConstraintViolation 检查是否是检查约束错误
func isSQLCheckConstraintViolation(original error) bool {
	if original == nil {
		return false
	}

	// MySQL: 错误码 3819 表示检查约束失败
	var mysqlErr *mysql.MySQLError
	if stderr.As(original, &mysqlErr) {
		return mysqlErr.Number == 3819
	}
	return false
}

// isSQLDataTooLong 检查是否是数据过长错误
func isSQLDataTooLong(original error) bool {
	if original == nil {
		return false
	}

	// MySQL: 错误码 1406 表示数据过长
	var mysqlErr *mysql.MySQLError
	if stderr.As(original, &mysqlErr) {
		return mysqlErr.Number == 1406
	}

	return false
}

// isSQLInvalidDataType 检查是否是数据类型不匹配错误
func isSQLInvalidDataType(original error) bool {
	if original == nil {
		return false
	}

	// MySQL: 错误码 1366 表示数据类型不正确
	var mysqlErr *mysql.MySQLError
	if stderr.As(original, &mysqlErr) {
		return mysqlErr.Number == 1366
	}

	return false
}

// isSQLDeadlock 检查是否是死锁错误
func isSQLDeadlock(original error) bool {
	if original == nil {
		return false
	}

	// MySQL: 错误码 1213 表示死锁
	var mysqlErr *mysql.MySQLError
	if stderr.As(original, &mysqlErr) {
		return mysqlErr.Number == 1213
	}

	return false
}

// isSQLLockWaitTimeout 检查是否是锁等待超时错误
func isSQLLockWaitTimeout(original error) bool {
	if original == nil {
		return false
	}

	// MySQL: 错误码 1205 表示锁等待超时
	var mysqlErr *mysql.MySQLError
	if stderr.As(original, &mysqlErr) {
		return mysqlErr.Number == 1205
	}

	return false
}

// isSQLDatabaseNotFound 检查是否是数据库不存在错误
func isSQLDatabaseNotFound(original error) bool {
	if original == nil {
		return false
	}

	// MySQL: 错误码 1049 表示数据库不存在
	var mysqlErr *mysql.MySQLError
	if stderr.As(original, &mysqlErr) {
		return mysqlErr.Number == 1049
	}

	return false
}

// isSQLTableNotFound 检查是否是表不存在错误
func isSQLTableNotFound(original error) bool {
	if original == nil {
		return false
	}

	// MySQL: 错误码 1146 表示表不存在
	var mysqlErr *mysql.MySQLError
	if stderr.As(original, &mysqlErr) {
		return mysqlErr.Number == 1146
	}

	return false
}
