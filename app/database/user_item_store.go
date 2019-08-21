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
			INSERT IGNORE INTO users_items (ID, idUser, idItem, sAncestorsComputationState)
			VALUES (?, ?, ?, 'todo')`, userItemID, userID, itemID).Error
	})
}

// PropagateAttempts copies iScore & bValidated from groups_attempts and
// marks users_items as 'todo' if corresponding groups_attempts are marked as 'todo'.
// Then it marks all the groups_attempts as 'done'.
func (s *UserItemStore) PropagateAttempts() (err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	mustNotBeError(s.db.Exec(`
		UPDATE users_items
		JOIN groups_attempts ON groups_attempts.idItem = users_items.idItem
		JOIN groups_groups ON groups_groups.idGroupParent = groups_attempts.idGroup AND
			groups_groups.sType IN ('direct', 'invitationAccepted', 'requestAccepted')
		JOIN users ON users.ID = users_items.idUser AND users.idGroupSelf = groups_groups.idGroupChild
		SET users_items.sAncestorsComputationState = 'todo'
		WHERE groups_attempts.sAncestorsComputationState = 'todo'`).Error)

	mustNotBeError(s.db.Exec(`
		UPDATE users_items
		JOIN (
			SELECT
				attempt_user.ID AS idUser,
				attempts.idItem AS idItem,
				MAX(attempts.iScore) AS iScore,
				MAX(attempts.bValidated) AS bValidated
			FROM users AS attempt_user
			JOIN groups_attempts AS attempts
			JOIN groups_groups AS attempt_group
				ON attempts.idGroup = attempt_group.idGroupParent AND attempt_user.idGroupSelf = attempt_group.idGroupChild AND
					attempt_group.sType IN ('direct', 'invitationAccepted', 'requestAccepted')
			WHERE attempts.sAncestorsComputationState = 'todo'
			GROUP BY attempt_user.ID, attempts.idItem
		) AS attempts_data
		ON attempts_data.idUser = users_items.idUser AND attempts_data.idItem = users_items.idItem
		SET
			users_items.iScore = GREATEST(users_items.iScore, IFNULL(attempts_data.iScore, 0)),
			users_items.bValidated = GREATEST(users_items.bValidated, IFNULL(attempts_data.bValidated, 0))`).Error)

	return s.GroupAttempts().Where("sAncestorsComputationState = 'todo'").
		UpdateColumn("sAncestorsComputationState", "done").Error()
}
