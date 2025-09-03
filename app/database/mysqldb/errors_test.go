package mysqldb

import (
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestIsMysqlError_shouldMatchMysqlError(t *testing.T) {
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(DuplicateEntryError),
		Message: "Duplicate Error",
	}

	if !IsMysqlError(error(&duplicateEntryError), DuplicateEntryError) {
		t.Error("expected to match DuplicateEntryError")
	}
}

func TestIsMysqlError_shouldNotMatchDifferentMysqlError(t *testing.T) {
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(DuplicateEntryError),
		Message: "Duplicate Error",
	}

	if IsMysqlError(error(&duplicateEntryError), ForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError) {
		t.Error("DuplicateEntryError should not match a ForeignKeyConstraintFailedOnAddingOrUpdatingChildRowError")
	}
}

func TestIsMysqlError_shouldNotMatchNonMysqlError(t *testing.T) {
	if IsMysqlError(errors.New("error"), DuplicateEntryError) {
		t.Error("should not match a non-mysql error")
	}
}

func TestErrorContains(t *testing.T) {
	duplicateEntryError := mysql.MySQLError{
		Number:  uint16(DuplicateEntryError),
		Message: "Duplicate Error",
	}

	contains := "Duplicate"
	if !ErrorContains(error(&duplicateEntryError), contains) {
		t.Errorf("expected %s contains %s", duplicateEntryError.Error(), contains)
	}

	contains = "Error"
	if !ErrorContains(error(&duplicateEntryError), contains) {
		t.Errorf("expected %s contains %s", duplicateEntryError.Error(), contains)
	}

	contains = "e E"
	if !ErrorContains(error(&duplicateEntryError), contains) {
		t.Errorf("expected %s contains %s", duplicateEntryError.Error(), contains)
	}

	doesntContain := "not"
	if ErrorContains(error(&duplicateEntryError), doesntContain) {
		t.Errorf("expected %s doesn't contain %s", duplicateEntryError.Error(), doesntContain)
	}
}
