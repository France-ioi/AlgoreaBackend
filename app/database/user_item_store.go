package database

// UserItemStore implements database operations on `users_items`
type UserItemStore struct {
	*DataStore
}

// CreateIfMissing inserts a new userID-itemID pair into `users_items` if it doesn't exist.
func (s *UserItemStore) CreateIfMissing(userID, itemID int64) error {
	return s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		userItemID := NewDataStore(db).NewID()
		return db.db.Exec(`
			INSERT IGNORE INTO users_items (id, user_id, item_id)
			VALUES (?, ?, ?)`, userItemID, userID, itemID).Error
	})
}
