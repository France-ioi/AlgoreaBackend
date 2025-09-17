// Package mysqldb contains database-related functions that are specific to mysql.
package mysqldb

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// MysqlErrorNumber represents the code number of a mysql error.
type MysqlErrorNumber uint16

const (
	// DuplicateEntryError represents a mysql Duplicate Entry Error.
	DuplicateEntryError MysqlErrorNumber = 1062
	// LockWaitTimeoutExceededError represents a mysql Lock Wait Timeout Exceeded Error.
	LockWaitTimeoutExceededError MysqlErrorNumber = 1205
	// DeadlockError represents the mysql error "Deadlock found when trying to get lock; try restarting transaction".
	DeadlockError MysqlErrorNumber = 1213

	// RowIsReferenced2 represents a mysql foreign constraint error
	// of deleting or updating a parent row when the current DB user has table-level privileges for all parent tables.
	RowIsReferenced2 MysqlErrorNumber = 1451
	// NoReferencedRow2 represents a mysql foreign constraint error
	// of adding or updating a child row when the current DB user has table-level privileges for all parent tables.
	NoReferencedRow2 MysqlErrorNumber = 1452
)

// IsMysqlError checks whether an error is a Mysql error of a certain type.
func IsMysqlError(err error, mysqlErrorNumber MysqlErrorNumber) bool {
	var e *mysql.MySQLError
	ok := errors.As(err, &e)
	return ok && e.Number == uint16(mysqlErrorNumber)
}

// ErrorContains checks whether an error contains a needle.
func ErrorContains(err error, needle string) bool {
	var e *mysql.MySQLError
	ok := errors.As(err, &e)
	return ok && strings.Contains(e.Message, needle)
}
