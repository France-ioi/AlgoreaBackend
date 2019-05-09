package database

// UserItemStore implements database operations on `users_items`
type UserItemStore struct {
	*DataStore
}

// CreateIfMissing inserts a new userID-itemID pair into `users_items` if it doesn't exist.
func (s *UserItemStore) CreateIfMissing(userID, itemID int64) {
	mustNotBeError(s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		userItemID := NewDataStore(db).NewID()
		return db.db.Exec(`
			INSERT IGNORE INTO users_items (ID, idUser, idItem, sAncestorsComputationState)
			VALUES (?, ?, ?, 'todo')`, userItemID, userID, itemID).Error
	}))
}
