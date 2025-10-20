package database

import (
	"github.com/jinzhu/gorm"
)

// AttemptStore implements database operations on `attempts`.
type AttemptStore struct {
	*DataStore
}

// CreateNew creates a new attempt (with id > 0) with parent_attempt_id = parentAttemptID and a new result.
// It also sets attempts.created_at, results.started_at, results.latest_activity_at, so the result should be propagated.
func (s *AttemptStore) CreateNew(participantID, parentAttemptID, itemID, creatorID int64) (attemptID int64, err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	mustNotBeError(s.InsertMap(map[string]interface{}{
		"id": gorm.Expr("(SELECT next_id FROM ? AS max_attempt)", s.Attempts().Select("IFNULL(MAX(id)+1, 0) AS next_id").
			Where("participant_id = ?", participantID).WithExclusiveWriteLock().SubQuery()),
		"participant_id": participantID, "creator_id": creatorID,
		"parent_attempt_id": parentAttemptID, "root_item_id": itemID, "created_at": Now(),
	}))
	mustNotBeError(s.Where("participant_id = ?", participantID).PluckFirst("MAX(id)", &attemptID).Error())

	mustNotBeError(s.Results().InsertMap(map[string]interface{}{
		"participant_id": participantID, "attempt_id": attemptID, "item_id": itemID,
		"started_at": Now(), "latest_activity_at": Now(),
	}))
	mustNotBeError(s.Results().MarkAsToBePropagated(participantID, attemptID, itemID, true))
	return attemptID, nil
}
