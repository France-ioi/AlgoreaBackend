// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestGroupAttemptStore_CreateNew(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups_attempts:
			- {ID: 1, idGroup: 10, idItem: 20, iOrder: 1}
			- {ID: 2, idGroup: 10, idItem: 30, iOrder: 3}
			- {ID: 3, idGroup: 20, idItem: 20, iOrder: 4}`)
	defer func() { _ = db.Close() }()

	var newID int64
	var err error
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		newID, err = store.GroupAttempts().CreateNew(10, 20)
		return err
	}))
	assert.True(t, newID > 0)
	type resultType struct {
		GroupID             int64 `gorm:"column:idGroup"`
		ItemID              int64 `gorm:"column:idItem"`
		StartDateSet        bool  `gorm:"column:startDateSet"`
		LastActivityDateSet bool  `gorm:"column:lastActivityDateSet"`
		Order               int32 `gorm:"column:iOrder"`
	}
	var result resultType
	assert.NoError(t, database.NewDataStore(db).GroupAttempts().ByID(newID).
		Select(`
			idGroup, idItem, ABS(sStartDate - NOW()) < 3 AS startDateSet,
			ABS(sLastActivityDate - NOW()) < 3 AS lastActivityDateSet, iOrder`).
		Take(&result).Error())
	assert.Equal(t, resultType{
		GroupID:             10,
		ItemID:              20,
		StartDateSet:        true,
		LastActivityDateSet: true,
		Order:               2,
	}, result)
}
