// +build !unit

package database_test

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
	"github.com/stretchr/testify/assert"
	"testing"
)

type hintsRequestedResultRow struct {
	ID                        int64   `gorm:"column:ID"`
	HintsRequested            *string `gorm:"column:sHintsRequested"`
	AncestorsComputationState string  `gorm:"column:sAncestorsComputationState"`
}

func TestUserItemStore_ComputeAllUserItems_CopiesHintsRequestedFromGroupAttempts(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/hints_requested")
	defer func() { _ = db.Close() }()

	err := database.NewDataStore(db).InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	expected := []hintsRequestedResultRow{
		{ID: 11, HintsRequested: ptrString("Hints requested for 11"), AncestorsComputationState: "done"},
		{ID: 12, HintsRequested: ptrString("Hints requested for 12"), AncestorsComputationState: "done"},
		{ID: 22, HintsRequested: ptrString("old value"), AncestorsComputationState: "done"},
	}
	var result []hintsRequestedResultRow
	assert.NoError(t, database.NewDataStore(db).UserItems().
		Select("ID, sHintsRequested, sAncestorsComputationState").
		Scan(&result).Error())
	assert.Equal(t, expected, result)
}

func ptrString(str string) *string { return &str }
