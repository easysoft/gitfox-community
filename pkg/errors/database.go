package errors

// Database related errors
var (
    // ErrResourceNotFound 表示资源未找到
    ErrResourceNotFound = WithI18nKey(DBError.New("resource not found"), "db.notfound")

    // ErrDuplicate 表示违反唯一性约束（重复记录）
    ErrDuplicate = WithI18nKey(DBError.New("resource already exists"), "db.duplicate")

    // ErrForeignKeyViolation 表示违反外键约束
    ErrForeignKeyViolation = WithI18nKey(DBError.New("foreign key constraint violation"), "db.foreign_key_violation")

    // ErrInvalidTransaction 表示无效的事务操作
    ErrInvalidTransaction = WithI18nKey(DBError.New("invalid transaction"), "db.invalid_transaction")

    // ErrInvalidData 表示无效的数据
    ErrInvalidData = WithI18nKey(DBError.New("invalid data"), "db.invalid_data")

    // ErrConnectionFailed 表示数据库连接失败
    ErrConnectionFailed = WithI18nKey(DBError.New("database connection failed"), "db.connection_failed")

    // ErrDeadlock 表示发生死锁
    ErrDeadlock = WithI18nKey(DBError.New("database deadlock detected"), "db.deadlock")

    // ErrLockWaitTimeout 表示锁等待超时
    ErrLockWaitTimeout = WithI18nKey(DBError.New("lock wait timeout exceeded"), "db.lock_wait_timeout")

    // ErrCheckConstraintViolation 表示违反检查约束
    ErrCheckConstraintViolation = WithI18nKey(DBError.New("check constraint violation"), "db.check_constraint_violation")

    // ErrNotNullViolation 表示违反非空约束
    ErrNotNullViolation = WithI18nKey(DBError.New("not null constraint violation"), "db.not_null_violation")

    // ErrDataTooLong 表示数据过长
    ErrDataTooLong = WithI18nKey(DBError.New("data too long for column"), "db.data_too_long")

    // ErrInvalidDataType 表示数据类型不匹配
    ErrInvalidDataType = WithI18nKey(DBError.New("invalid data type"), "db.invalid_data_type")
)
