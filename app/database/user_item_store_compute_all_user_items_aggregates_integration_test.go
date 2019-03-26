// +build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type aggregatesResultRow struct {
	ID                        int64      `gorm:"column:ID"`
	LastActivityDate          *time.Time `gorm:"column:sLastActivityDate"`
	TasksTried                int64      `gorm:"column:nbTasksTried"`
	TasksWithHelp             int64      `gorm:"column:nbTasksWithHelp"`
	TasksSolved               int64      `gorm:"column:nbTasksSolved"`
	ChildrenValidated         int64      `gorm:"column:nbChildrenValidated"`
	AncestorsComputationState string     `gorm:"column:sAncestorsComputationState"`
}

func TestUserItemStore_ComputeAllUserItems_Aggregates(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/aggregates")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	currentDate := time.Now().Round(time.Second).UTC()
	oldDate := currentDate.AddDate(-1, -1, -1)

	assert.NoError(t, userItemStore.Where("ID=11").Updates(map[string]interface{}{
		"sLastActivityDate":   oldDate,
		"nbTasksTried":        1,
		"nbTasksWithHelp":     2,
		"nbTasksSolved":       3,
		"nbChildrenValidated": 4,
		"bValidated":          1,
	}).Error())
	assert.NoError(t, userItemStore.Where("ID=13").Updates(map[string]interface{}{
		"sLastActivityDate":   currentDate,
		"nbTasksTried":        5,
		"nbTasksWithHelp":     6,
		"nbTasksSolved":       7,
		"nbChildrenValidated": 8,
	}).Error())
	assert.NoError(t, userItemStore.Where("ID=14").Updates(map[string]interface{}{
		"sLastActivityDate":   nil,
		"nbTasksTried":        9,
		"nbTasksWithHelp":     10,
		"nbTasksSolved":       11,
		"nbChildrenValidated": 12,
		"bValidated":          1,
	}).Error())

	err := userItemStore.ComputeAllUserItems()
	assert.NoError(t, err)

	expected := []aggregatesResultRow{
		{ID: 11, LastActivityDate: &oldDate, TasksTried: 1, TasksWithHelp: 2, TasksSolved: 3, ChildrenValidated: 4, AncestorsComputationState: "done"},
		{ID: 12, LastActivityDate: &currentDate, TasksTried: 1 + 5 + 9, TasksWithHelp: 2 + 6 + 10, TasksSolved: 3 + 7 + 11, ChildrenValidated: 2, AncestorsComputationState: "done"},
		{ID: 13, LastActivityDate: &currentDate, TasksTried: 5, TasksWithHelp: 6, TasksSolved: 7, ChildrenValidated: 8, AncestorsComputationState: "done"},
		{ID: 14, LastActivityDate: nil, TasksTried: 9, TasksWithHelp: 10, TasksSolved: 11, ChildrenValidated: 12, AncestorsComputationState: "done"},
		// another user
		{ID: 22, LastActivityDate: nil, AncestorsComputationState: "done"},
	}

	assertAggregatesEqual(t, userItemStore, expected)
}

func TestUserItemStore_ComputeAllUserItems_Aggregates_OnCommonData(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()
	err := userItemStore.ComputeAllUserItems()
	assert.NoError(t, err)

	var result []aggregatesResultRow
	assert.NoError(t, userItemStore.
		Select("ID, sLastActivityDate, nbTasksTried, nbTasksWithHelp, nbTasksSolved, nbChildrenValidated, sAncestorsComputationState").
		Scan(&result).Error())

	expected := []aggregatesResultRow{
		{ID: 11, AncestorsComputationState: "done"},
		{ID: 12, AncestorsComputationState: "done"},
		{ID: 22, AncestorsComputationState: "done"},
	}
	assertAggregatesEqual(t, userItemStore, expected)
}

func assertAggregatesEqual(t *testing.T, userItemStore *database.UserItemStore, expected []aggregatesResultRow) {
	var result []aggregatesResultRow
	assert.NoError(t, userItemStore.
		Select("ID, sLastActivityDate, nbTasksTried, nbTasksWithHelp, nbTasksSolved, nbChildrenValidated, sAncestorsComputationState").
		Scan(&result).Error())
	assert.Equal(t, expected, result)
}
