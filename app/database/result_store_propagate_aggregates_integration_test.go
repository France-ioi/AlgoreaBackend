//go:build !unit

package database_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

type aggregatesResultRow struct {
	ParticipantID    int64
	AttemptID        int64
	ItemID           int64
	LatestActivityAt database.Time
	TasksTried       int64
	TasksWithHelp    int64
	State            string
	ScoreComputed    float32
}

func TestResultStore_Propagate_Aggregates(t *testing.T) {
	for _, markAllResultsAsToBePropagated := range []bool{false, true} {
		markAllResultsAsToBePropagated := markAllResultsAsToBePropagated
		t.Run(fmt.Sprintf("mark all as to_be_propagated=%v", markAllResultsAsToBePropagated), func(t *testing.T) {
			db := testhelpers.SetupDBWithFixture("results_propagation/_common", "results_propagation/aggregates")
			defer func() { _ = db.Close() }()

			resultStore := database.NewDataStore(db).Results()

			currentDate := time.Now().Round(time.Second).UTC()
			oldDate := currentDate.AddDate(-1, -1, -1)

			assert.NoError(t, resultStore.Where("attempt_id = 1 AND item_id = 1 AND participant_id = 101").
				Updates(map[string]interface{}{
					"latest_activity_at": oldDate,
					"tasks_tried":        1,
					"tasks_with_help":    2,
					"score_computed":     10,
					"validated_at":       "2019-05-30 11:00:00",
				}).Error())
			assert.NoError(t, resultStore.
				Where("(item_id = 3 AND participant_id = 101 AND attempt_id = 1) OR (item_id = 3 AND participant_id = 101 AND attempt_id = 2)").
				Updates(map[string]interface{}{
					"latest_activity_at": currentDate,
					"tasks_tried":        5,
					"tasks_with_help":    6,
					"score_computed":     20,
				}).Error())
			assert.NoError(t, resultStore.
				Where("(item_id = 4 AND participant_id = 101 AND attempt_id = 1) OR (item_id = 4 AND participant_id = 101 AND attempt_id = 2)").
				Updates(map[string]interface{}{
					"latest_activity_at": oldDate,
					"tasks_tried":        9,
					"tasks_with_help":    10,
					"score_computed":     30,
					"validated_at":       "2019-05-30 11:00:00",
				}).Error())

			assert.NoError(t, resultStore.Where("item_id = 2 AND participant_id = 102 AND attempt_id = 1").Updates(map[string]interface{}{
				"latest_activity_at": oldDate,
			}).Error())

			err := resultStore.InTransaction(func(s *database.DataStore) error {
				assert.NoError(t, resultStore.Exec(
					"INSERT IGNORE INTO results_propagate SELECT participant_id, attempt_id, item_id, 'to_be_propagated' AS state FROM results").Error())
				s.ScheduleResultsPropagation()
				return nil
			})
			assert.NoError(t, err)

			expected := []aggregatesResultRow{
				{
					ParticipantID: 101, AttemptID: 1, ItemID: 1, LatestActivityAt: database.Time(oldDate), TasksTried: 1, TasksWithHelp: 2,
					ScoreComputed: 10, State: "done",
				},
				{
					ParticipantID: 101, AttemptID: 1, ItemID: 2, LatestActivityAt: database.Time(currentDate), TasksTried: 1 + 5 + 9,
					TasksWithHelp: 2 + 6 + 10,
					ScoreComputed: 23.3333, /* (10*1 + 20*2 + 30*3) / (1 + 2 + 3) */
					State:         "done",
				}, // from 1, 3, 4
				{
					ParticipantID: 101, AttemptID: 1, ItemID: 3, LatestActivityAt: database.Time(currentDate), TasksTried: 5, TasksWithHelp: 6,
					ScoreComputed: 20, State: "done",
				},
				{
					ParticipantID: 101, AttemptID: 1, ItemID: 4, LatestActivityAt: database.Time(oldDate), TasksTried: 9, TasksWithHelp: 10,
					ScoreComputed: 30, State: "done",
				},
				{
					ParticipantID: 101, AttemptID: 2, ItemID: 3, LatestActivityAt: database.Time(currentDate), TasksTried: 5, TasksWithHelp: 6,
					ScoreComputed: 20, State: "done",
				},
				{
					ParticipantID: 101, AttemptID: 2, ItemID: 4, LatestActivityAt: database.Time(oldDate), TasksTried: 9, TasksWithHelp: 10,
					ScoreComputed: 30, State: "done",
				},
				// another user
				{ParticipantID: 102, AttemptID: 1, ItemID: 2, LatestActivityAt: database.Time(oldDate), State: "done"},
			}

			assertAggregatesEqual(t, resultStore, expected)
		})
	}
}

func TestResultStore_Propagate_Aggregates_OnCommonData(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("results_propagation/_common")
	defer func() { _ = db.Close() }()

	resultStore := database.NewDataStore(db).Results()
	err := resultStore.InTransaction(func(s *database.DataStore) error {
		s.ScheduleResultsPropagation()
		return nil
	})
	assert.NoError(t, err)

	expectedLatestActivityAt1 := database.Time(time.Date(2019, 5, 29, 11, 0, 0, 0, time.UTC))
	expectedLatestActivityAt2 := database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))

	expected := []aggregatesResultRow{
		{ParticipantID: 101, AttemptID: 1, ItemID: 1, State: "done", LatestActivityAt: expectedLatestActivityAt1},
		{ParticipantID: 101, AttemptID: 1, ItemID: 2, State: "done", LatestActivityAt: expectedLatestActivityAt1},
		{ParticipantID: 102, AttemptID: 1, ItemID: 2, State: "done", LatestActivityAt: expectedLatestActivityAt2},
	}
	assertAggregatesEqual(t, resultStore, expected)
}

func TestResultStore_Propagate_Aggregates_KeepsLastActivityAtIfItIsGreater(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("results_propagation/_common")
	defer func() { _ = db.Close() }()

	expectedLatestActivityAt1 := database.Time(time.Date(2019, 5, 29, 11, 0, 0, 0, time.UTC))
	expectedLatestActivityAt2 := database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))

	resultStore := database.NewDataStore(db).Results()
	assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id = 2").Updates(map[string]interface{}{
		"latest_activity_at": time.Time(expectedLatestActivityAt2),
	}).Error())

	err := resultStore.InTransaction(func(s *database.DataStore) error {
		s.ScheduleResultsPropagation()
		return nil
	})
	assert.NoError(t, err)

	expected := []aggregatesResultRow{
		{ParticipantID: 101, AttemptID: 1, ItemID: 1, State: "done", LatestActivityAt: expectedLatestActivityAt1},
		{ParticipantID: 101, AttemptID: 1, ItemID: 2, State: "done", LatestActivityAt: expectedLatestActivityAt2},
		{ParticipantID: 102, AttemptID: 1, ItemID: 2, State: "done", LatestActivityAt: expectedLatestActivityAt2},
	}
	assertAggregatesEqual(t, resultStore, expected)
}

func TestResultStore_Propagate_Aggregates_EditScore(t *testing.T) {
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
			db := testhelpers.SetupDBWithFixture("results_propagation/_common")
			defer func() { _ = db.Close() }()

			resultStore := database.NewDataStore(db).Results()
			assert.NoError(t, resultStore.Where("attempt_id = 1 AND item_id = 1 AND participant_id = 101").
				Updates(map[string]interface{}{
					"score_computed": 10,
				}).Error())
			assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id = 2").
				Updates(map[string]interface{}{
					"score_edit_rule":  test.editRule,
					"score_edit_value": test.editValue,
				}).Error())

			err := resultStore.InTransaction(func(s *database.DataStore) error {
				s.ScheduleResultsPropagation()
				return nil
			})
			assert.NoError(t, err)

			expectedLatestActivityAt1 := database.Time(time.Date(2019, 5, 29, 11, 0, 0, 0, time.UTC))
			expectedLatestActivityAt2 := database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))

			expected := []aggregatesResultRow{
				{
					ParticipantID: 101, AttemptID: 1, ItemID: 1, ScoreComputed: 10, State: "done",
					LatestActivityAt: expectedLatestActivityAt1,
				},
				{
					ParticipantID: 101, AttemptID: 1, ItemID: 2, ScoreComputed: test.expectedComputedScore,
					State: "done", LatestActivityAt: expectedLatestActivityAt1,
				},
				{
					ParticipantID: 102, AttemptID: 1, ItemID: 2, State: "done",
					LatestActivityAt: expectedLatestActivityAt2,
				},
			}
			assertAggregatesEqual(t, resultStore, expected)
		})
	}
}

func assertAggregatesEqual(t *testing.T, resultStore *database.ResultStore, expected []aggregatesResultRow) {
	var result []aggregatesResultRow
	queryResultsAndStatesForTests(t, resultStore, &result, "latest_activity_at, tasks_tried, tasks_with_help, score_computed")
	assert.Equal(t, expected, result)
}
