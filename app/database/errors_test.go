package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database/mysqldb"
)

const (
	DuplicateEntryErrorTable = "table"
	DuplicateEntryErrorKey   = "key"
)

func TestIsDuplicateEntryError_MatchingError(t *testing.T) {
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: "Duplicate Error",
	}

	if !IsDuplicateEntryError(error(&duplicateEntryError)) {
		t.Error("should be a DuplicateEntryError")
	}
}

func TestIsDuplicateEntryError_OtherErrors(t *testing.T) {
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

func TestIsDuplicateEntryErrorForKey_MatchingError(t *testing.T) {
	table := DuplicateEntryErrorTable
	key := DuplicateEntryErrorKey
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: fmt.Sprintf("Duplicate Error for key '%s.%s'", table, key),
	}

	if !IsDuplicateEntryErrorForKey(error(&duplicateEntryError), table, key) {
		t.Errorf("should be a DuplicateEntryError with key %s.%s", table, key)
	}
}

func TestIsDuplicateEntryErrorForKey_ShouldNotMatchDuplicateWithAnotherKey(t *testing.T) {
	table := DuplicateEntryErrorTable
	otherKey := "otherkey"
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: fmt.Sprintf("Duplicate Error for key '%s.test'", table),
	}

	if IsDuplicateEntryErrorForKey(error(&duplicateEntryError), table, otherKey) {
		t.Errorf("should not match key %s.%s of DuplicateEntryError", table, otherKey)
	}
}

func TestIsDuplicateEntryErrorForKey_ShouldNotMatchDuplicateWithAnotherTable(t *testing.T) {
	otherTable := "othertable"
	key := DuplicateEntryErrorKey
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: fmt.Sprintf("Duplicate Error for key '%s.%s'", otherTable, key),
	}

	if IsDuplicateEntryErrorForKey(error(&duplicateEntryError), DuplicateEntryErrorTable, key) {
		t.Errorf("should not match key %s.%s of DuplicateEntryError", DuplicateEntryErrorTable, key)
	}
}

func TestIsDuplicateEntryErrorForKey_OtherErrors(t *testing.T) {
	table := DuplicateEntryErrorTable
	key := DuplicateEntryErrorKey

	lockDeadlockErrorWithKey := mysql.MySQLError{
		Number:  uint16(mysqldb.DeadlockError),
		Message: fmt.Sprintf("Lock Deadlock Error for key '%s.%s'", table, key),
	}

	if IsDuplicateEntryErrorForKey(error(&lockDeadlockErrorWithKey), table, key) {
		t.Error("should not match a Lock Deadlock Error")
	}

	nonMysqlErrorWithKey := fmt.Errorf("other error for key '%s.%s'", table, key)
	if IsDuplicateEntryErrorForKey(nonMysqlErrorWithKey, table, key) {
		t.Error("should not match a non-mysql error")
	}
}

func TestIsRowIsReferenced2Error_MatchingError(t *testing.T) {
	foreignKeyConstraintFailedError := mysql.MySQLError{
		Number:  uint16(mysqldb.RowIsReferenced2),
		Message: "Some message",
	}

	if !IsRowIsReferenced2Error(error(&foreignKeyConstraintFailedError)) {
		t.Error("should be a RowIsReferenced2")
	}
}

func TestIsRowIsReferenced2Error_OtherErrors(t *testing.T) {
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: "Duplicate Error",
	}

	if IsRowIsReferenced2Error(error(&duplicateEntryError)) {
		t.Error("should not match a Duplicate Entry Error")
	}

	nonMysqlError := errors.New("other error")
	if IsRowIsReferenced2Error(nonMysqlError) {
		t.Error("should not match a non-mysql error")
	}
}

func TestIsNoReferencedRow2Error_MatchingError(t *testing.T) {
	foreignKeyConstraintError := mysql.MySQLError{
		Number:  uint16(mysqldb.NoReferencedRow2),
		Message: "Some message",
	}

	if !IsNoReferencedRow2Error(error(&foreignKeyConstraintError)) {
		t.Error("should be a NoReferencedRow2")
	}
}

func TestIsNoReferencedRow2Error_OtherErrors(t *testing.T) {
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: "Duplicate Error",
	}

	if IsNoReferencedRow2Error(error(&duplicateEntryError)) {
		t.Error("should not match a Duplicate Entry Error")
	}

	nonMysqlError := errors.New("other error")
	if IsNoReferencedRow2Error(nonMysqlError) {
		t.Error("should not match a non-mysql error")
	}
}

func TestIsLockDeadlockError_MatchingError(t *testing.T) {
	lockDeadlockError := mysql.MySQLError{
		Number:  uint16(mysqldb.DeadlockError),
		Message: "Lock Deadlock Error",
	}

	if !IsDeadlockError(error(&lockDeadlockError)) {
		t.Error("should be a DeadlockError")
	}
}

func TestIsLockDeadlockError_OtherErrors(t *testing.T) {
	foreignKeyConstraintError := mysql.MySQLError{
		Number:  uint16(mysqldb.NoReferencedRow2),
		Message: "Some message",
	}

	if IsDeadlockError(error(&foreignKeyConstraintError)) {
		t.Error("should not match a Foreign Constraint Error")
	}

	nonMysqlError := errors.New("other error")
	if IsDeadlockError(nonMysqlError) {
		t.Error("should not match a non-mysql error")
	}
}
