package database

import (
	"fmt"

	"github.com/France-ioi/AlgoreaBackend/app/database/mysqldb"
)

// IsDuplicateEntryError checks whether an error corresponds to a duplicate of primary keys on insertion.
func IsDuplicateEntryError(err error) bool {
	return mysqldb.IsMysqlError(err, mysqldb.DuplicateEntryError)
}

// IsDuplicateEntryErrorForKey checks whether an error corresponds to a duplicate of primary keys on insertion
//                             for a certain key.
func IsDuplicateEntryErrorForKey(err error, key string) bool {
	return IsDuplicateEntryError(err) && mysqldb.ErrorContains(err, fmt.Sprintf("for key '%s'", key))
}

// IsForeignConstraintError checks whether an error corresponds to a foreign key constraint fail on insert/update.
func IsForeignConstraintError(err error) bool {
	return mysqldb.IsMysqlError(err, mysqldb.ForeignConstraintError)
}

// IsLockDeadlockError checks whether an error corresponds to a deadlock when trying to get a lock.
func IsLockDeadlockError(err error) bool {
	return mysqldb.IsMysqlError(err, mysqldb.LockDeadlockError)
}
