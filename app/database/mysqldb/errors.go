package mysqldb

import (
	"strings"

	"github.com/go-sql-driver/mysql"
)

type MysqlErrorNumber uint16

const (
	DuplicateEntryError    MysqlErrorNumber = 1062
	LockDeadlockError      MysqlErrorNumber = 1213
	ForeignConstraintError MysqlErrorNumber = 1452
)

// IsMysqlError checks whether an error is a Mysql error of a certain type.
func IsMysqlError(err error, mysqlErrorNumber MysqlErrorNumber) bool {
	e, ok := err.(*mysql.MySQLError)
	return ok && e.Number == uint16(mysqlErrorNumber)
}

// ErrorContains checks whether an error contains a needle.
func ErrorContains(err error, needle string) bool {
	e, ok := err.(*mysql.MySQLError)
	return ok && strings.Contains(e.Message, needle)
}
