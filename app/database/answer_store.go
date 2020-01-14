package database

// AnswerStore implements database operations on `answers`
type AnswerStore struct {
	*DataStore
}

// WithUsers creates a composable query for getting answers joined with users (via author_id)
func (s *AnswerStore) WithUsers() *AnswerStore {
	return &AnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN users ON users.group_id = answers.author_id"), s.tableName,
		),
	}
}

// WithAttempts creates a composable query for getting answers joined with attempts
func (s *AnswerStore) WithAttempts() *AnswerStore {
	return &AnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN attempts ON attempts.id = answers.attempt_id"), s.tableName,
		),
	}
}

// WithItems joins `items` through `attempts`
func (s *AnswerStore) WithItems() *AnswerStore {
	return &AnswerStore{
		NewDataStoreWithTable(
			s.WithAttempts().Joins("JOIN items ON items.id = attempts.item_id"), s.tableName,
		),
	}
}

// SubmitNewAnswer inserts a new row with type='Submission', created_at=NOW()
// into the `answers` table.
func (s *AnswerStore) SubmitNewAnswer(authorID, attemptID int64, answer string) (int64, error) {
	var answerID int64
	err := s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		store := NewDataStore(db)
		answerID = store.NewID()
		return db.db.Exec(`
				INSERT INTO answers (id, author_id, attempt_id, answer, created_at, type)
				VALUES (?, ?, ?, ?, NOW(), 'Submission')`,
			answerID, authorID, attemptID, answer).Error
	})
	return answerID, err
}

// Visible returns a composable query for getting answers with the following access rights
// restrictions:
// 1) the user should have at least 'content' access rights to the answers.item_id item,
// 2) the user is able to see answers related to his group's attempts, so
//    the user should be a member of the attempts.group_id team or
//    attempts.group_id should be equal to the user's self group
func (s *AnswerStore) Visible(user *User) *DB {
	usersGroupsQuery := s.GroupGroups().WhereUserIsMember(user).Select("parent_group_id")
	// the user should have at least 'content' access to the item
	itemsQuery := s.Items().WhereUserHasViewPermissionOnItems(user, "content")

	return s.
		// the user should have at least 'content' access to the answers.item_id
		Joins("JOIN attempts ON attempts.id = answers.attempt_id").
		Joins("JOIN ? AS items ON items.id = attempts.item_id", itemsQuery.SubQuery()).
		// attempts.group_id should be one of the authorized user's groups or the user's self group
		Where("attempts.group_id = ? OR attempts.group_id IN ?",
			user.GroupID, usersGroupsQuery.SubQuery())
}
