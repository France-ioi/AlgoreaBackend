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

func TestIsDuplicateEntryErrorForKey_matchDuplicateButWithOtherKey(t *testing.T) {
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

func TestIsDuplicateEntryErrorForKey_matchDuplicateButWithOtherTable(t *testing.T) {
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

func TestIsDuplicateEntryErrorForKey_otherErrors(t *testing.T) {
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

func TestIsForeignKeyConstraintFailedOnDeletingOrUpdatingParentRowError_matchError(t *testing.T) {
	foreignKeyConstraintFailedError := mysql.MySQLError{
		Number:  uint16(mysqldb.ForeignKeyConstraintFailedOnDeletingOrUpdatingParentRowError),
		Message: "Some message",
	}

	if !IsForeignKeyConstraintFailedOnDeletingOrUpdatingParentRowError(error(&foreignKeyConstraintFailedError)) {
		t.Error("should be a ForeignKeyConstraintFailedOnDeletingOrUpdatingParentRowError")
	}
}

func TestIsForeignKeyConstraintFailedOnDeletingOrUpdatingParentRowError_otherError(t *testing.T) {
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: "Duplicate Error",
	}

	if IsForeignKeyConstraintFailedOnDeletingOrUpdatingParentRowError(error(&duplicateEntryError)) {
		t.Error("should not match a Duplicate Entry Error")
	}

	nonMysqlError := errors.New("other error")
	if IsForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError(nonMysqlError) {
		t.Error("should not match a non-mysql error")
	}
}

func TestIsForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError_matchError(t *testing.T) {
	foreignKeyConstraintError := mysql.MySQLError{
		Number:  uint16(mysqldb.ForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError),
		Message: "Some message",
	}

	if !IsForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError(error(&foreignKeyConstraintError)) {
		t.Error("should be a ForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError")
	}
}

func TestIsForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError_otherError(t *testing.T) {
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(mysqldb.DuplicateEntryError),
		Message: "Duplicate Error",
	}

	if IsForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError(error(&duplicateEntryError)) {
		t.Error("should not match a Duplicate Entry Error")
	}

	nonMysqlError := errors.New("other error")
	if IsForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError(nonMysqlError) {
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
	foreignKeyConstraintError := mysql.MySQLError{
		Number:  uint16(mysqldb.ForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError),
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
