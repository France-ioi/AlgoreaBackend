package database

// AnswerStore implements database operations on `answers`.
type AnswerStore struct {
	*DataStore
}

// WithGradings creates a composable query for getting answers joined with gradings (via answer_id).
func (s *AnswerStore) WithGradings() *AnswerStore {
	return &AnswerStore{
		NewDataStoreWithTable(
			s.Select(`
					answers.id, answers.author_id, answers.item_id, answers.attempt_id, answers.participant_id,
					answers.type, answers.state, answers.answer, answers.created_at, gradings.score,
					gradings.graded_at
				`).
				Joins("LEFT JOIN gradings ON gradings.answer_id = answers.id"), s.tableName,
		),
	}
}

// WithUsers creates a composable query for getting answers joined with users (via author_id).
func (s *AnswerStore) WithUsers() *AnswerStore {
	return &AnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN users ON users.group_id = answers.author_id"), s.tableName,
		),
	}
}

// WithResults creates a composable query for getting answers joined with results.
func (s *AnswerStore) WithResults() *AnswerStore {
	return &AnswerStore{
		NewDataStoreWithTable(
			s.Joins(`
				JOIN results ON results.participant_id = answers.participant_id AND
					results.attempt_id = answers.attempt_id AND results.item_id = answers.item_id`), s.tableName,
		),
	}
}

// WithItems joins `items`.
func (s *AnswerStore) WithItems() *AnswerStore {
	return &AnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN items ON items.id = answers.item_id"), s.tableName,
		),
	}
}

// GetCurrentAnswer returns the current answer of a participant-item-attempt triplet.
func (s *AnswerStore) GetCurrentAnswer(participantID, itemID, attemptID int64) (map[string]interface{}, bool) {
	var result []map[string]interface{}
	err := s.
		WithGradings().
		Where("participant_id = ?", participantID).
		Where("attempt_id = ?", attemptID).
		Where("item_id = ?", itemID).
		Where("type = 'Current'").
		Order("created_at DESC").
		Limit(1).
		ScanIntoSliceOfMaps(&result).
		Error()
	mustNotBeError(err)

	if len(result) == 0 {
		return map[string]interface{}{}, false
	}

	return result[0], true
}

// SubmitNewAnswer inserts a new row with type='Submission', created_at=NOW()
// into the `answers` table.
func (s *AnswerStore) SubmitNewAnswer(authorID, participantID, attemptID, itemID int64, answer string) (int64, error) {
	var answerID int64
	err := s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		store := NewDataStore(db)
		answerID = store.NewID()
		return db.db.Exec(`
				INSERT INTO answers (id, author_id, participant_id, attempt_id, item_id, answer, created_at, type)
				VALUES (?, ?, ?, ?, ?, ?, NOW(), 'Submission')`,
			answerID, authorID, participantID, attemptID, itemID, answer).Error
	})
	return answerID, err
}
