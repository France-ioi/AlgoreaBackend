package database

import (
	"math/rand"

	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// DataStore gather all stores for database operations on business data
type DataStore struct {
	*DB
	tableName string
}

// NewDataStore returns a DataStore
func NewDataStore(conn *DB) *DataStore {
	return &DataStore{DB: conn}
}

// NewDataStoreWithTable returns a specialized DataStore
func NewDataStoreWithTable(conn *DB, tableName string) *DataStore {
	return &DataStore{conn.Table(tableName), tableName}
}

// Users returns a UserStore
func (s *DataStore) Users() *UserStore {
	return &UserStore{NewDataStoreWithTable(s.DB, "users")}
}

// Items returns a ItemStore
func (s *DataStore) Items() *ItemStore {
	return &ItemStore{NewDataStoreWithTable(s.DB, "items")}
}

// GroupAttempts returns a GroupAttemptStore
func (s *DataStore) GroupAttempts() *GroupAttemptStore {
	return &GroupAttemptStore{NewDataStoreWithTable(s.DB, "groups_attempts")}
}

// Groups returns a GroupStore
func (s *DataStore) Groups() *GroupStore {
	return &GroupStore{NewDataStoreWithTable(s.DB, "groups")}
}

// GroupAncestors returns a GroupAncestorStore
func (s *DataStore) GroupAncestors() *GroupAncestorStore {
	return &GroupAncestorStore{NewDataStoreWithTable(s.DB, "groups_ancestors")}
}

// GroupGroups returns a GroupGroupStore
func (s *DataStore) GroupGroups() *GroupGroupStore {
	return &GroupGroupStore{NewDataStoreWithTable(s.DB, "groups_groups")}
}

// GroupItems returns a GroupItemStore
func (s *DataStore) GroupItems() *GroupItemStore {
	return &GroupItemStore{NewDataStoreWithTable(s.DB, "groups_items")}
}

// ItemAncestors returns a ItemAncestorStore
func (s *DataStore) ItemAncestors() *ItemAncestorStore {
	return &ItemAncestorStore{NewDataStoreWithTable(s.DB, "items_ancestors")}
}

// ItemStrings returns a ItemStringStore
func (s *DataStore) ItemStrings() *ItemStringStore {
	return &ItemStringStore{NewDataStoreWithTable(s.DB, "items_strings")}
}

// ItemItems returns a ItemItemStore
func (s *DataStore) ItemItems() *ItemItemStore {
	return &ItemItemStore{NewDataStoreWithTable(s.DB, "items_items")}
}

// UserAnswers returns a UserAnswerStore
func (s *DataStore) UserAnswers() *UserAnswerStore {
	return &UserAnswerStore{NewDataStoreWithTable(s.DB, "users_answers")}
}

// NewID generates a positive random int64 to be used as ID
// !!! To be safe, the insertion should be be retried if the ID conflicts with an existing entry
func (s *DataStore) NewID() int64 {
	// gen a 63-bits number as we want unsigned number stored in a 64-bits signed DB attribute
	return rand.Int63()
}

// EnsureSetID does check the given ID is set. If not, generate a (random) ID for it
func (s *DataStore) EnsureSetID(id *types.Int64) {
	if !id.Set {
		*id = *types.NewInt64(s.NewID())
	}
}

// InTransaction executes the given function in a transaction and commits
func (s *DataStore) InTransaction(txFunc func(*DataStore) (interface{}, error)) (interface{}, error) {
	return s.inTransaction(func(db *DB) (interface{}, error) {
		return txFunc(NewDataStoreWithTable(db, s.tableName))
	})
}

// ByID returns a composable query for filtering by _table_.ID
func (s *DataStore) ByID(id int64) *DB {
	if len(s.tableName) == 0 {
		panic("method ByID() called for abstract DataStore")
	}
	return s.Where(s.tableName+".ID = ?", id)
}
