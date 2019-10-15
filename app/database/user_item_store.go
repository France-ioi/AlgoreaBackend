package database

// UserItemStore implements database operations on `users_items`
type UserItemStore struct {
	*DataStore
}

// SetActiveAttempt inserts a new userID-itemID pair with the given active_attempt_id
// into `users_items` or updates active_attempt_id for existing one.
func (s *UserItemStore) SetActiveAttempt(userID, itemID, groupAttemptID int64) error {
	return s.db.Exec(`
		INSERT INTO users_items (user_id, item_id, active_attempt_id)
		VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE active_attempt_id = VALUES(active_attempt_id)`,
		userID, itemID, groupAttemptID).Error
}
