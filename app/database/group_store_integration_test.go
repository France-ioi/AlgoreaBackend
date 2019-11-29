// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestGroupStore_CreateNew(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString()
	defer func() { _ = db.Close() }()

	var newID int64
	var err error
	dataStore := database.NewDataStore(db)
	assert.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
		newID, err = store.Groups().CreateNew(ptrString("Some group"), ptrString("Class"), ptrInt64(123))
		return err
	}))
	assert.True(t, newID > 0)
	type resultType struct {
		Name         *string
		Type         *string
		TeamItemID   *int64
		CreatedAtSet bool
	}
	var result resultType
	assert.NoError(t, dataStore.Groups().ByID(newID).
		Select("name, type, team_item_id, ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 AS created_at_set").
		Take(&result).Error())
	assert.Equal(t, resultType{
		Name:         ptrString("Some group"),
		Type:         ptrString("Class"),
		TeamItemID:   ptrInt64(123),
		CreatedAtSet: true,
	}, result)

	found, err := dataStore.GroupAncestors().
		Where("ancestor_group_id = ?", newID).
		Where("child_group_id = ?", newID).HasRows()
	assert.NoError(t, err)
	assert.True(t, found)
}

func ptrString(s string) *string { return &s }
