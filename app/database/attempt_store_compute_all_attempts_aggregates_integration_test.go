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
	ID                     int64
	LatestActivityAt       *database.Time
	TasksTried             int64
	TasksWithHelp          int64
	TasksSolved            int64
	ChildrenValidated      int64
	ResultPropagationState string
	ScoreComputed          float32
}

func TestAttemptStore_ComputeAllAttempts_Aggregates(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("attempts_propagation/_common", "attempts_propagation/aggregates")
	defer func() { _ = db.Close() }()

	attemptStore := database.NewDataStore(db).Attempts()

	currentDate := time.Now().Round(time.Second).UTC()
	oldDate := currentDate.AddDate(-1, -1, -1)

	assert.NoError(t, attemptStore.Where("id=11").Updates(map[string]interface{}{
		"latest_activity_at": oldDate,
		"tasks_tried":        1,
		"tasks_with_help":    2,
		"tasks_solved":       3,
		"children_validated": 4,
		"score_computed":     10,
		"validated_at":       "2019-05-30 11:00:00",
	}).Error())
	assert.NoError(t, attemptStore.Where("id IN (13, 15)").Updates(map[string]interface{}{
		"latest_activity_at": currentDate,
		"tasks_tried":        5,
		"tasks_with_help":    6,
		"tasks_solved":       7,
		"children_validated": 8,
		"score_computed":     20,
	}).Error())
	assert.NoError(t, attemptStore.Where("id IN (14, 16)").Updates(map[string]interface{}{
		"latest_activity_at": nil,
		"tasks_tried":        9,
		"tasks_with_help":    10,
		"tasks_solved":       11,
		"children_validated": 12,
		"score_computed":     30,
		"validated_at":       "2019-05-30 11:00:00",
	}).Error())

	err := attemptStore.InTransaction(func(s *database.DataStore) error {
		return s.Attempts().ComputeAllAttempts()
	})
	assert.NoError(t, err)

	expected := []aggregatesResultRow{
		{ID: 11, LatestActivityAt: (*database.Time)(&oldDate), TasksTried: 1, TasksWithHelp: 2, TasksSolved: 3,
			ChildrenValidated: 4, ScoreComputed: 10, ResultPropagationState: "done"},
		{ID: 12, LatestActivityAt: (*database.Time)(&currentDate), TasksTried: 1 + 5 + 9, TasksWithHelp: 2 + 6 + 10,
			TasksSolved: 3 + 7 + 11, ChildrenValidated: 2, ScoreComputed: 23.3333, /* (10*1 + 20*2 + 30*3) / (1 + 2 + 3) */
			ResultPropagationState: "done"}, // from 1, 3, 4
		{ID: 13, LatestActivityAt: (*database.Time)(&currentDate), TasksTried: 5, TasksWithHelp: 6, TasksSolved: 7,
			ChildrenValidated: 8, ScoreComputed: 20, ResultPropagationState: "done"},
		{ID: 14, LatestActivityAt: nil, TasksTried: 9, TasksWithHelp: 10, TasksSolved: 11, ChildrenValidated: 12,
			ScoreComputed: 30, ResultPropagationState: "done"},
		{ID: 15, LatestActivityAt: (*database.Time)(&currentDate), TasksTried: 5, TasksWithHelp: 6, TasksSolved: 7,
			ChildrenValidated: 8, ScoreComputed: 20, ResultPropagationState: "done"},
		{ID: 16, LatestActivityAt: nil, TasksTried: 9, TasksWithHelp: 10, TasksSolved: 11, ChildrenValidated: 12,
			ScoreComputed: 30, ResultPropagationState: "done"},
		// another user
		{ID: 22, LatestActivityAt: nil, ResultPropagationState: "done"},
	}

	assertAggregatesEqual(t, attemptStore, expected)
}

func TestAttemptStore_ComputeAllAttempts_Aggregates_OnCommonData(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("attempts_propagation/_common")
	defer func() { _ = db.Close() }()

	attemptStore := database.NewDataStore(db).Attempts()
	err := attemptStore.InTransaction(func(s *database.DataStore) error {
		return s.Attempts().ComputeAllAttempts()
	})
	assert.NoError(t, err)

	expected := []aggregatesResultRow{
		{ID: 11, ResultPropagationState: "done"},
		{ID: 12, ResultPropagationState: "done"},
		{ID: 22, ResultPropagationState: "done"},
	}
	assertAggregatesEqual(t, attemptStore, expected)
}

func TestAttemptStore_ComputeAllAttempts_Aggregates_EditScore(t *testing.T) {
	for _, test := range []struct {
		name                  string
		editRule              string
		editValue             float32
		expectedComputedScore float32
	}{
		{name: "set positive", editRule: "set", editValue: 20, expectedComputedScore: 20},
		{name: "set negative", editRule: "set", editValue: -10, expectedComputedScore: 0},
		{name: "diff positive", editRule: "diff", editValue: 20, expectedComputedScore: 30},
		{name: "diff negative", editRule: "diff", editValue: -5, expectedComputedScore: 5},
		{name: "diff big negative", editRule: "diff", editValue: -20, expectedComputedScore: 0},
		{name: "diff big positive", editRule: "diff", editValue: 95, expectedComputedScore: 100},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixture("attempts_propagation/_common")
			defer func() { _ = db.Close() }()

			attemptStore := database.NewDataStore(db).Attempts()
			assert.NoError(t, attemptStore.Where("id=11").Updates(map[string]interface{}{
				"score_computed": 10,
			}).Error())
			assert.NoError(t, attemptStore.Where("id=12").Updates(map[string]interface{}{
				"score_edit_rule":  test.editRule,
				"score_edit_value": test.editValue,
			}).Error())

			err := attemptStore.InTransaction(func(s *database.DataStore) error {
				return s.Attempts().ComputeAllAttempts()
			})
			assert.NoError(t, err)

			expected := []aggregatesResultRow{
				{ID: 11, ScoreComputed: 10, ResultPropagationState: "done"},
				{ID: 12, ScoreComputed: test.expectedComputedScore, ResultPropagationState: "done"},
				{ID: 22, ResultPropagationState: "done"},
			}
			assertAggregatesEqual(t, attemptStore, expected)
		})
	}
}

func assertAggregatesEqual(t *testing.T, attemptStore *database.AttemptStore, expected []aggregatesResultRow) {
	var result []aggregatesResultRow
	assert.NoError(t, attemptStore.
		Select(`
			id, latest_activity_at, tasks_tried, tasks_with_help, tasks_solved, children_validated, score_computed,
			result_propagation_state`).
		Scan(&result).Error())
	assert.Equal(t, expected, result)
}
