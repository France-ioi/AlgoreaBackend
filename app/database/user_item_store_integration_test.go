// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestUserItemStore_SetActiveAttempt(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 121}]
		users: [{group_id: 121}]
		items: [{id: 34}]
		groups_attempts:
			- {id: 56, group_id: 1, item_id: 34, order: 1}
			- {id: 57, group_id: 1, item_id: 34, order: 1}`)
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()
	for _, groupAttemptID := range []int64{56, 57} {
		err := userItemStore.SetActiveAttempt(121, 34, groupAttemptID)
		assert.NoError(t, err)

		type userItem struct {
			UserGroupID     int64
			ItemID          int64
			ActiveAttemptID int64
		}
		var insertedUserItem userItem
		assert.NoError(t,
			userItemStore.Select("user_group_id, item_id, active_attempt_id").
				Scan(&insertedUserItem).Error())
		assert.Equal(t, userItem{
			UserGroupID:     121,
			ItemID:          34,
			ActiveAttemptID: groupAttemptID,
		}, insertedUserItem)
	}
}
