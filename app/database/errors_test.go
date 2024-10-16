package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database/mysqldb"
)

const DuplicateEntryErrorKey = "key"

func TestIsDuplicateEntryError_matchError(t *testing.T) {
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: "Duplicate Error",
	}

	if !IsDuplicateEntryError(error(&duplicateEntryError)) {
		t.Error("should be a DuplicateEntryError")
	}
}

func TestIsDuplicateEntryError_otherErrors(t *testing.T) {
	lockDeadlockError := mysql.MySQLError{
		Number:  uint16(mysqldb.DeadlockError),
		Message: "Lock Deadlock Error",
	}

	if IsDuplicateEntryError(error(&lockDeadlockError)) {
		t.Error("should not match a Lock Deadlock Error")
	}

	nonMysqlError := errors.New("other error")
	if IsDuplicateEntryError(nonMysqlError) {
		t.Error("should not match a non-mysql error")
	}
}

func TestIsDuplicateEntryErrorForKey_matchError(t *testing.T) {
	key := DuplicateEntryErrorKey
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: fmt.Sprintf("Duplicate Error for key '%s'", key),
	}

	if !IsDuplicateEntryErrorForKey(error(&duplicateEntryError), key) {
		t.Errorf("should be a DuplicateEntryError with key %s", key)
	}
}

func TestIsDuplicateEntryErrorForKey_matchDuplicateButWithoutKey(t *testing.T) {
	otherKey := "otherkey"
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: "Duplicate Error for key 'test'",
	}

	if IsDuplicateEntryErrorForKey(error(&duplicateEntryError), otherKey) {
		t.Errorf("should not match key %s of DuplicateEntryError", otherKey)
	}
}

func TestIsDuplicateEntryErrorForKey_otherErrors(t *testing.T) {
	key := DuplicateEntryErrorKey

	lockDeadlockErrorWithKey := mysql.MySQLError{
		Number:  uint16(mysqldb.DeadlockError),
		Message: fmt.Sprintf("Lock Deadlock Error for key '%s'", key),
	}

	if IsDuplicateEntryErrorForKey(error(&lockDeadlockErrorWithKey), key) {
		t.Error("should not match a Lock Deadlock Error")
	}

	nonMysqlErrorWithKey := fmt.Errorf("other error for key '%s'", key)
	if IsDuplicateEntryErrorForKey(nonMysqlErrorWithKey, key) {
		t.Error("should not match a non-mysql error")
	}
}

func TestIsForeignConstraintError_matchError(t *testing.T) {
	foreignConstraintError := mysql.MySQLError{
		Number:  uint16(mysqldb.ForeignConstraintError),
		Message: "Foreign Constraint Error",
	}

	if !IsForeignConstraintError(error(&foreignConstraintError)) {
		t.Error("should be a ForeignConstraintError")
	}
}

func TestIsForeignConstraintError_otherError(t *testing.T) {
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: "Duplicate Error",
	}

	if IsForeignConstraintError(error(&duplicateEntryError)) {
		t.Error("should not match a Duplicate Entry Error")
	}

	nonMysqlError := errors.New("other error")
	if IsForeignConstraintError(nonMysqlError) {
		t.Error("should not match a non-mysql error")
	}
}

func TestIsLockDeadlockError_matchError(t *testing.T) {
	lockDeadlockError := mysql.MySQLError{
		Number:  uint16(mysqldb.DeadlockError),
		Message: "Lock Deadlock Error",
	}

	if !IsDeadlockError(error(&lockDeadlockError)) {
		t.Error("should be a DeadlockError")
	}
}

func TestIsLockDeadlockError_otherError(t *testing.T) {
	foreignConstraintError := mysql.MySQLError{
		Number:  uint16(mysqldb.ForeignConstraintError),
		Message: "Foreign Constraint Error",
	}

	if IsDeadlockError(error(&foreignConstraintError)) {
		t.Error("should not match a Foreign Constraint Error")
	}

	nonMysqlError := errors.New("other error")
	if IsDeadlockError(nonMysqlError) {
		t.Error("should not match a non-mysql error")
	}
}
