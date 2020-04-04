// +build !unit

package database_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestGroupStore_CreateNew(t *testing.T) {
	for _, test := range []struct {
		groupType            string
		shouldCreateAttempts bool
	}{
		{groupType: "Class", shouldCreateAttempts: false},
		{groupType: "Team", shouldCreateAttempts: true},
	} {
		test := test
		t.Run(test.groupType, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString()
			defer func() { _ = db.Close() }()

			var newID int64
			var err error
			dataStore := database.NewDataStore(db)
			assert.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
				newID, err = store.Groups().CreateNew("Some group", test.groupType)
				return err
			}))
			assert.True(t, newID > 0)
			type resultType struct {
				Name         string
				Type         string
				CreatedAtSet bool
			}
			var result resultType
			assert.NoError(t, dataStore.Groups().ByID(newID).
				Select("name, type, ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 AS created_at_set").
				Take(&result).Error())
			assert.Equal(t, resultType{
				Name:         "Some group",
				Type:         test.groupType,
				CreatedAtSet: true,
			}, result)

			found, err := dataStore.GroupAncestors().
				Where("ancestor_group_id = ?", newID).
				Where("child_group_id = ?", newID).HasRows()
			assert.NoError(t, err)
			assert.True(t, found)

			var attempts []map[string]interface{}
			assert.NoError(t, dataStore.Attempts().
				Select(`
					participant_id, id, creator_id, parent_attempt_id, root_item_id,
					ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 AS created_at_set`).
				ScanIntoSliceOfMaps(&attempts).Error())
			var expectedAttempts []map[string]interface{}
			if test.shouldCreateAttempts {
				expectedAttempts = []map[string]interface{}{
					{"participant_id": strconv.FormatInt(newID, 10), "id": "0", "creator_id": nil, "parent_attempt_id": nil,
						"root_item_id": nil, "created_at_set": "1"},
				}
			}
			assert.Equal(t, expectedAttempts, attempts)
		})
	}
}

func ptrString(s string) *string { return &s }
func ptrInt64(i int64) *int64    { return &i }
