// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package errors

import "github.com/easysoft/gitfox-community/pkg/i18n"

// Database related errors
var (
	// ErrResourceNotFound 表示资源未找到
	ErrResourceNotFound = WithI18nKey(DBError.New("resource not found"), i18n.DbNotfound)

	// ErrDuplicate 表示违反唯一性约束（重复记录）
	ErrDuplicate = WithI18nKey(DBError.New("resource already exists"), i18n.DbDuplicate)

	// ErrForeignKeyViolation 表示违反外键约束
	ErrForeignKeyViolation = WithI18nKey(DBError.New("foreign key constraint violation"), i18n.DbForeignKeyViolation)

	// ErrInvalidTransaction 表示无效的事务操作
	ErrInvalidTransaction = WithI18nKey(DBError.New("invalid transaction"), i18n.DbInvalidTransaction)

	// ErrInvalidData 表示无效的数据
	ErrInvalidData = WithI18nKey(DBError.New("invalid data"), i18n.DbInvalidData)

	// ErrConnectionFailed 表示数据库连接失败
	ErrConnectionFailed = WithI18nKey(DBError.New("database connection failed"), i18n.DbConnectionFailed)

	// ErrDeadlock 表示发生死锁
	ErrDeadlock = WithI18nKey(DBError.New("database deadlock detected"), i18n.DbDeadlock)

	// ErrLockWaitTimeout 表示锁等待超时
	ErrLockWaitTimeout = WithI18nKey(DBError.New("lock wait timeout exceeded"), i18n.DbLockWaitTimeout)

	// ErrCheckConstraintViolation 表示违反检查约束
	ErrCheckConstraintViolation = WithI18nKey(DBError.New("check constraint violation"), i18n.DbCheckConstraintViolation)

	// ErrNotNullViolation 表示违反非空约束
	ErrNotNullViolation = WithI18nKey(DBError.New("not null constraint violation"), i18n.DbNotNullViolation)

	// ErrDataTooLong 表示数据过长
	ErrDataTooLong = WithI18nKey(DBError.New("data too long for column"), i18n.DbDataTooLong)

	// ErrInvalidDataType 表示数据类型不匹配
	ErrInvalidDataType = WithI18nKey(DBError.New("invalid data type"), i18n.DbInvalidDataType)
)
