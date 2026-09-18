package database

import (
	"testing"

	"github.com/HuaweiCloudDeveloper/gaussdb-go/gaussdbconn"
	"github.com/easysoft/gitfox-community/pkg/errors"
)

func gaussdbError(code string) error {
	return &gaussdbconn.GaussdbError{Severity: "ERROR", Code: code, Message: "test"}
}

func TestIsSQLUniqueConstraintError_GaussDB(t *testing.T) {
	if !isSQLUniqueConstraintError(gaussdbError("23505")) {
		t.Error("expected 23505 to be a unique constraint error")
	}
	if isSQLUniqueConstraintError(gaussdbError("23503")) {
		t.Error("expected 23503 not to be a unique constraint error")
	}
}

func TestIsSQLForeignKeyViolationError_GaussDB(t *testing.T) {
	if !isSQLForeignKeyViolationError(gaussdbError("23503")) {
		t.Error("expected 23503 to be a foreign key violation error")
	}
	if isSQLForeignKeyViolationError(gaussdbError("23505")) {
		t.Error("expected 23505 not to be a foreign key violation error")
	}
}

func TestIsSQLNotNullViolation_GaussDB(t *testing.T) {
	if !isSQLNotNullViolation(gaussdbError("23502")) {
		t.Error("expected 23502 to be a not null violation error")
	}
}

func TestIsSQLCheckConstraintViolation_GaussDB(t *testing.T) {
	if !isSQLCheckConstraintViolation(gaussdbError("23514")) {
		t.Error("expected 23514 to be a check constraint violation error")
	}
}

func TestIsSQLDataTooLong_GaussDB(t *testing.T) {
	if !isSQLDataTooLong(gaussdbError("22001")) {
		t.Error("expected 22001 to be a data too long error")
	}
}

func TestIsSQLInvalidDataType_GaussDB(t *testing.T) {
	if !isSQLInvalidDataType(gaussdbError("22P02")) {
		t.Error("expected 22P02 to be an invalid data type error")
	}
}

func TestIsSQLDeadlock_GaussDB(t *testing.T) {
	if !isSQLDeadlock(gaussdbError("40P01")) {
		t.Error("expected 40P01 to be a deadlock error")
	}
}

func TestIsSQLLockWaitTimeout_GaussDB(t *testing.T) {
	if !isSQLLockWaitTimeout(gaussdbError("55P03")) {
		t.Error("expected 55P03 to be a lock wait timeout error")
	}
}

func TestIsSQLDatabaseNotFound_GaussDB(t *testing.T) {
	if !isSQLDatabaseNotFound(gaussdbError("3D000")) {
		t.Error("expected 3D000 to be a database not found error")
	}
}

func TestIsSQLTableNotFound_GaussDB(t *testing.T) {
	if !isSQLTableNotFound(gaussdbError("42P01")) {
		t.Error("expected 42P01 to be a table not found error")
	}
}

func TestProcessSQLErrorf_GaussDBUniqueViolation(t *testing.T) {
	err := ProcessSQLErrorf(t.Context(), gaussdbError("23505"))
	if err != errors.ErrDuplicate {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

func TestProcessSQLErrorf_GaussDBForeignKeyViolation(t *testing.T) {
	err := ProcessSQLErrorf(t.Context(), gaussdbError("23503"))
	if err != errors.ErrForeignKeyViolation {
		t.Errorf("expected ErrForeignKeyViolation, got %v", err)
	}
}

func TestProcessSQLErrorf_GaussDBDeadlock(t *testing.T) {
	err := ProcessSQLErrorf(t.Context(), gaussdbError("40P01"))
	if err != errors.ErrDeadlock {
		t.Errorf("expected ErrDeadlock, got %v", err)
	}
}
