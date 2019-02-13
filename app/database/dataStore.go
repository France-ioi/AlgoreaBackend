package database

import (
	"math/rand"

	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// DataStore gather all stores for database operations on business data
type DataStore struct {
	DB
}

// NewDataStore returns a DataStore
func NewDataStore(conn DB) *DataStore {
	return &DataStore{conn}
}

// Users returns a UserStore
func (s *DataStore) Users() *UserStore {
	return &UserStore{s}
}

// Items returns a ItemStore
func (s *DataStore) Items() *ItemStore {
	return &ItemStore{s}
}

// Groups returns a GroupStore
func (s *DataStore) Groups() *GroupStore {
	return &GroupStore{s}
}

// GroupItems returns a GroupItemStore
func (s *DataStore) GroupItems() *GroupItemStore {
	return &GroupItemStore{s}
}

// ItemStrings returns a ItemStringStore
func (s *DataStore) ItemStrings() *ItemStringStore {
	return &ItemStringStore{s}
}

// ItemItems returns a ItemItemStore
func (s *DataStore) ItemItems() *ItemItemStore {
	return &ItemItemStore{s}
}

// GroupAncestors returns a GroupAncestorStore
func (s *DataStore) GroupAncestors() *GroupAncestorStore {
	return &GroupAncestorStore{s}
}

// UserAnswers returns a UserAnswerStore
func (s *DataStore) UserAnswers() *UserAnswerStore {
	return &UserAnswerStore{s}
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
func (s *DataStore) InTransaction(txFunc func(*DataStore) error) error {
	return s.inTransaction(func(db DB) error {
		return txFunc(&DataStore{db})
	})
}
