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
			INSERT IGNORE INTO users_items (id, user_id, item_id, ancestors_computation_state)
			VALUES (?, ?, ?, 'todo')`, userItemID, userID, itemID).Error
	})
}

// PropagateAttempts copies score & validated from groups_attempts and
// marks users_items as 'todo' if corresponding groups_attempts are marked as 'todo'.
// Then it marks all the groups_attempts as 'done'.
func (s *UserItemStore) PropagateAttempts() (err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	mustNotBeError(s.db.Exec(`
		UPDATE users_items
		JOIN groups_attempts ON groups_attempts.item_id = users_items.item_id
		JOIN groups_groups ON groups_groups.parent_group_id = groups_attempts.group_id AND
			groups_groups.type` + GroupRelationIsActiveCondition + ` AND NOW() < groups_groups.expires_at
		JOIN users ON users.id = users_items.user_id AND users.self_group_id = groups_groups.child_group_id
		SET users_items.ancestors_computation_state = 'todo'
		WHERE groups_attempts.ancestors_computation_state = 'todo'`).Error)

	mustNotBeError(s.db.Exec(`
		UPDATE users_items
		JOIN (
			SELECT
				attempt_user.id AS user_id,
				attempts.item_id AS item_id,
				MAX(attempts.score) AS score,
				MAX(attempts.validated) AS validated
			FROM users AS attempt_user
			JOIN groups_attempts AS attempts
			JOIN groups_groups AS attempt_group
				ON attempts.group_id = attempt_group.parent_group_id AND attempt_user.self_group_id = attempt_group.child_group_id AND
					attempt_group.type` + GroupRelationIsActiveCondition + ` AND NOW() < attempt_group.expires_at
			WHERE attempts.ancestors_computation_state = 'todo'
			GROUP BY attempt_user.id, attempts.item_id
		) AS attempts_data
		ON attempts_data.user_id = users_items.user_id AND attempts_data.item_id = users_items.item_id
		SET
			users_items.score = GREATEST(users_items.score, IFNULL(attempts_data.score, 0)),
			users_items.validated = GREATEST(users_items.validated, IFNULL(attempts_data.validated, 0))`).Error)

	return s.GroupAttempts().Where("ancestors_computation_state = 'todo'").
		UpdateColumn("ancestors_computation_state", "done").Error()
}
