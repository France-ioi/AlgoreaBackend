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
	ID                        int64
	LatestActivityAt          *database.Time
	TasksTried                int64
	TasksWithHelp             int64
	TasksSolved               int64
	ChildrenValidated         int64
	AncestorsComputationState string
	Score                     float32
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_Aggregates(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("groups_attempts_propagation/_common", "groups_attempts_propagation/aggregates")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()

	currentDate := time.Now().Round(time.Second).UTC()
	oldDate := currentDate.AddDate(-1, -1, -1)

	assert.NoError(t, groupAttemptStore.Where("id=11").Updates(map[string]interface{}{
		"latest_activity_at": oldDate,
		"tasks_tried":        1,
		"tasks_with_help":    2,
		"tasks_solved":       3,
		"children_validated": 4,
		"score":              10,
		"validated_at":       "2019-05-30 11:00:00",
	}).Error())
	assert.NoError(t, groupAttemptStore.Where("id IN (13, 15)").Updates(map[string]interface{}{
		"latest_activity_at": currentDate,
		"tasks_tried":        5,
		"tasks_with_help":    6,
		"tasks_solved":       7,
		"children_validated": 8,
		"score":              20,
	}).Error())
	assert.NoError(t, groupAttemptStore.Where("id IN (14, 16)").Updates(map[string]interface{}{
		"latest_activity_at": nil,
		"tasks_tried":        9,
		"tasks_with_help":    10,
		"tasks_solved":       11,
		"children_validated": 12,
		"score":              30,
		"validated_at":       "2019-05-30 11:00:00",
	}).Error())

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	expected := []aggregatesResultRow{
		{ID: 11, LatestActivityAt: (*database.Time)(&oldDate), TasksTried: 1, TasksWithHelp: 2, TasksSolved: 3,
			ChildrenValidated: 4, Score: 10, AncestorsComputationState: "done"},
		{ID: 12, LatestActivityAt: (*database.Time)(&currentDate), TasksTried: 1 + 5 + 9, TasksWithHelp: 2 + 6 + 10,
			TasksSolved: 3 + 7 + 11, ChildrenValidated: 2, Score: 23.3333, /* (10*1 + 20*2 + 30*3) / (1 + 2 + 3) */
			AncestorsComputationState: "done"}, // from 1, 3, 4
		{ID: 13, LatestActivityAt: (*database.Time)(&currentDate), TasksTried: 5, TasksWithHelp: 6, TasksSolved: 7,
			ChildrenValidated: 8, Score: 20, AncestorsComputationState: "done"},
		{ID: 14, LatestActivityAt: nil, TasksTried: 9, TasksWithHelp: 10, TasksSolved: 11, ChildrenValidated: 12,
			Score: 30, AncestorsComputationState: "done"},
		{ID: 15, LatestActivityAt: (*database.Time)(&currentDate), TasksTried: 5, TasksWithHelp: 6, TasksSolved: 7,
			ChildrenValidated: 8, Score: 20, AncestorsComputationState: "done"},
		{ID: 16, LatestActivityAt: nil, TasksTried: 9, TasksWithHelp: 10, TasksSolved: 11, ChildrenValidated: 12,
			Score: 30, AncestorsComputationState: "done"},
		// another user
		{ID: 22, LatestActivityAt: nil, AncestorsComputationState: "done"},
	}

	assertAggregatesEqual(t, groupAttemptStore, expected)
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_Aggregates_OnCommonData(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("groups_attempts_propagation/_common")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()
	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	expected := []aggregatesResultRow{
		{ID: 11, AncestorsComputationState: "done"},
		{ID: 12, AncestorsComputationState: "done"},
		{ID: 22, AncestorsComputationState: "done"},
	}
	assertAggregatesEqual(t, groupAttemptStore, expected)
}

func assertAggregatesEqual(t *testing.T, groupAttemptStore *database.GroupAttemptStore, expected []aggregatesResultRow) {
	var result []aggregatesResultRow
	assert.NoError(t, groupAttemptStore.
		Select("id, latest_activity_at, tasks_tried, tasks_with_help, tasks_solved, children_validated, score, ancestors_computation_state").
		Scan(&result).Error())
	assert.Equal(t, expected, result)
}
