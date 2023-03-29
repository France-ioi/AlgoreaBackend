//go:build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type resultType struct {
	ParticipantID    int64
	AttemptID        int64
	ItemID           int64
	StartedAt        *database.Time
	LatestActivityAt database.Time
}

type attemptType struct {
	ParticipantID   int64
	ID              int64
	ParentAttemptID *int64
	RootItemID      *int64
	CreatorID       int64
	CreatedAt       *database.Time
}

func TestAttemptStore_CreateNew_CreatesNewAttempt(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups:
			- {id: 10}
			- {id: 100}
		users:
			- {group_id: 100}
		items: [{id: 20, default_language_tag: fr}, {id: 30, default_language_tag: fr}]
		attempts:
			- {id: 0, participant_id: 10}
			- {id: 0, participant_id: 20}
		results:
			- {attempt_id: 0, participant_id: 10, item_id: 20}
			- {attempt_id: 0, participant_id: 10, item_id: 30}
			- {attempt_id: 0, participant_id: 20, item_id: 20}`)
	defer func() { _ = db.Close() }()

	testhelpers.MockDBTime("2019-05-30 11:00:00")
	defer testhelpers.RestoreDBTime()

	var newAttemptID int64
	var err error
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		newAttemptID, err = store.Attempts().CreateNew(10, 200, 20, 100)
		return err
	}))
	assert.Equal(t, int64(1), newAttemptID)
	var result resultType
	expectedTime := database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))
	assert.NoError(t, database.NewDataStore(db).Results().
		Where("attempt_id = ?", newAttemptID).
		Where("participant_id = ?", 10).
		Select("participant_id, attempt_id, item_id, started_at, latest_activity_at").Take(&result).Error())
	assert.Equal(t, resultType{
		ParticipantID:    10,
		AttemptID:        1,
		ItemID:           20,
		StartedAt:        &expectedTime,
		LatestActivityAt: expectedTime,
	}, result)
	var attempt attemptType
	assert.NoError(t, database.NewDataStore(db).Attempts().ByID(newAttemptID).
		Where("participant_id = ?", 10).
		Select("participant_id, id, creator_id, parent_attempt_id, root_item_id, created_at").Take(&attempt).Error())
	assert.Equal(t, attemptType{
		ParticipantID:   10,
		ID:              1,
		ParentAttemptID: ptrInt64(200),
		RootItemID:      ptrInt64(20),
		CreatorID:       100,
		CreatedAt:       &expectedTime,
	}, attempt)
}
