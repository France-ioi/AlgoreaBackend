package users

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// Store implements database operations on `users`
type Store struct {
	*database.DataStore
}

// NewStore returns an users.Store
func NewStore(s *database.DataStore) *Store {
	return &Store{s}
}

// ByID returns a composable query of users filtered by userID
func (s *Store) ByID(userID int64) database.DB {
	return s.All().Where("users.ID = ?", userID)
}

// All creates a composable query without filtering
func (s *Store) All() database.DB {
	return s.Table("users")
}
