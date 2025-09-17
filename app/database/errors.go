package database

import (
	"fmt"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database/mysqldb"
)

// IsDuplicateEntryError checks whether an error corresponds to a duplicate of primary keys on insertion.
func IsDuplicateEntryError(err error) bool {
	return mysqldb.IsMysqlError(err, mysqldb.DuplicateEntryError)
}

// IsDuplicateEntryErrorForKey checks whether an error corresponds to a duplicate of primary keys on insertion for a certain key.
func IsDuplicateEntryErrorForKey(err error, table, key string) bool {
	return IsDuplicateEntryError(err) && mysqldb.ErrorContains(err, fmt.Sprintf("for key '%s.%s'", table, key))
}

// IsRowIsReferenced2Error checks whether an error corresponds to a foreign key constraint failure
// on deleting/updating a parent row when the current DB user has table-level privileges for all parent tables.
func IsRowIsReferenced2Error(err error) bool {
	return mysqldb.IsMysqlError(err, mysqldb.RowIsReferenced2)
}

// IsNoReferencedRow2Error checks whether an error corresponds to a foreign key constraint failure
// on inserting/updating a child row when the current DB user has table-level privileges for all parent tables.
func IsNoReferencedRow2Error(err error) bool {
	return mysqldb.IsMysqlError(err, mysqldb.NoReferencedRow2)
}

// IsDeadlockError checks whether an error corresponds to a deadlock when trying to get a lock.
func IsDeadlockError(err error) bool {
	return mysqldb.IsMysqlError(err, mysqldb.DeadlockError)
}

// IsLockWaitTimeoutExceededError checks whether an error corresponds to a lock wait timeout exceeded.
func IsLockWaitTimeoutExceededError(err error) bool {
	return mysqldb.IsMysqlError(err, mysqldb.LockWaitTimeoutExceededError)
}
