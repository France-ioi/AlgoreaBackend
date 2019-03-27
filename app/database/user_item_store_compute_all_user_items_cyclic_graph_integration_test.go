// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type stateResultRow struct {
	ID                        int64  `gorm:"column:ID"`
	AncestorsComputationState string `gorm:"column:sAncestorsComputationState"`
}

func TestUserItemStore_ComputeAllUserItems_WithCyclicGraph(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/cyclic")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []stateResultRow
	assert.NoError(t, userItemStore.Select("ID, sAncestorsComputationState").Order("ID").Scan(&result).Error())
	assert.Equal(t, []stateResultRow{
		{ID: 11, AncestorsComputationState: "todo"},
		{ID: 12, AncestorsComputationState: "todo"},
		// another user
		{ID: 22, AncestorsComputationState: "done"},
	}, result)
}
