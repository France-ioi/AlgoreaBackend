// +build !unit

package database_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type validatedResultRow struct {
	ParticipantID int64
	AttemptID     int64
	ItemID        int64
	Validated     bool
	State         string
}

func (r validatedResultRow) LessThan(other validatedResultRow) bool {
	return r.ParticipantID < other.ParticipantID ||
		r.ParticipantID == other.ParticipantID && r.AttemptID < other.AttemptID ||
		r.ParticipantID == other.ParticipantID && r.AttemptID == other.AttemptID && r.ItemID < other.ItemID
}

func testResultStorePropagateValidated(t *testing.T, fixtures []string,
	validationType string,
	prepareFunc func(*testing.T, *database.ResultStore), expectedResults []validatedResultRow) {
	db := testhelpers.SetupDBWithFixture(fixtures...)
	defer func() { _ = db.Close() }()

	resultStore := database.NewDataStore(db).Results()
	assert.NoError(t,
		resultStore.Items().Where("id=2").
			UpdateColumn("validation_type", validationType).Error())
	if prepareFunc != nil {
		prepareFunc(t, resultStore)
	}

	err := resultStore.InTransaction(func(s *database.DataStore) error {
		return s.Results().Propagate()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	queryResultsAndStatesForTests(t, resultStore, &result, "validated")
	assert.Equal(t, expectedResults, result)
}

func TestResultStore_Propagate_ValidatedStaysNonValidatedFor(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "None"},
		{name: "Manual"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testResultStorePropagateValidated(t,
				[]string{"results_propagation/_common"},
				tt.name,
				func(t *testing.T, resultStore *database.ResultStore) {
					assert.NoError(t, resultStore.Where("attempt_id = 1 AND item_id = 1 AND participant_id = 101").
						UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
				},
				buildExpectedValidatedResultRows(map[string]bool{
					"101_1_1": true, "101_1_2": false,
				}))
		})
	}
}

func TestResultStore_Propagate_ValidatedWithValidationTypeOneBecomesValidatedWhenThereIsAtLeastOneValidatedChild(
	t *testing.T) {
	testResultStorePropagateValidated(t,
		[]string{"results_propagation/_common", "results_propagation/validated/one"},
		"One",
		func(t *testing.T, resultStore *database.ResultStore) {
			assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id = 3").
				UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
		},
		buildExpectedValidatedResultRows(map[string]bool{
			"101_1_1": false, "101_1_2": true, "101_1_3": true,
		}))
}

func TestResultStore_Propagate_ValidatedWithValidationTypeOneStaysNonValidatedWhenThereAreNoValidatedChildren(
	t *testing.T) {
	testResultStorePropagateValidated(t,
		[]string{"results_propagation/_common", "results_propagation/validated/one"},
		"One",
		nil,
		buildExpectedValidatedResultRows(map[string]bool{
			"101_1_1": false, "101_1_2": false, "101_1_3": false,
		}))
}

func TestResultStore_Propagate_Validated(t *testing.T) {
	tests := []struct {
		name            string
		fixtures        []string
		validationType  string
		prepareFunc     func(*testing.T, *database.ResultStore)
		expectedResults []validatedResultRow
	}{
		{
			name:           "for ValidationType=AllButOne stays non-validated when there are two non-validated children",
			fixtures:       []string{"results_propagation/_common", "results_propagation/validated/all_and_category"},
			validationType: "AllButOne",
			prepareFunc: func(t *testing.T, resultStore *database.ResultStore) {
				assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id = 3").
					UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_1": false, "101_1_2": false, "101_1_3": true, "101_1_4": false,
			}),
		},
		{
			name:           "for ValidationType=AllButOne stays non-validated when there is just one child and it is not validated",
			fixtures:       []string{"results_propagation/_common"},
			validationType: "AllButOne",
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_1": false, "101_1_2": false,
			}),
		},
		{
			name:           "for ValidationType=AllButOne becomes validated when there are less than two non-validated children",
			fixtures:       []string{"results_propagation/_common", "results_propagation/validated/all_and_category"},
			validationType: "AllButOne",
			prepareFunc: func(t *testing.T, resultStore *database.ResultStore) {
				assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id IN (1, 3)").
					UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_1": true, "101_1_2": true, "101_1_3": true, "101_1_4": false,
			}),
		},
		{
			name:           "for ValidationType=All stays non-validated when there is at least one non-validated child",
			fixtures:       []string{"results_propagation/_common", "results_propagation/validated/all_and_category"},
			validationType: "All",
			prepareFunc: func(t *testing.T, resultStore *database.ResultStore) {
				assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id IN (1, 3)").
					UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_1": true, "101_1_2": false, "101_1_3": true, "101_1_4": false,
			}),
		},
		{
			name:           "for ValidationType=All stays non-validated when there are no children",
			fixtures:       []string{"results_propagation/_common"},
			validationType: "All",
			prepareFunc: func(t *testing.T, resultStore *database.ResultStore) {
				assert.NoError(t, resultStore.ItemItems().Delete("parent_item_id=2").Error())
				assert.NoError(t, resultStore.Delete("participant_id = 101 AND attempt_id = 1 AND item_id = 1").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_2": false,
			}),
		},
		{
			name:           "for ValidationType=All becomes validated when all its children are validated",
			fixtures:       []string{"results_propagation/_common", "results_propagation/validated/all_and_category"},
			validationType: "All",
			prepareFunc: func(t *testing.T, resultStore *database.ResultStore) {
				assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id IN (1, 3, 4)").
					UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_1": true, "101_1_2": true, "101_1_3": true, "101_1_4": true,
			}),
		},
		{
			name: "for ValidationType=Categories stays non-validated when " +
				"there is at least one non-validated child item with Category=Validation",
			fixtures:       []string{"results_propagation/_common", "results_propagation/validated/all_and_category"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, resultStore *database.ResultStore) {
				assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id IN (1, 3)").
					UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
				assert.NoError(t, resultStore.ItemItems().Where("parent_item_id = 2 AND child_item_id IN (3, 4)").
					UpdateColumn("category", "Validation").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_1": true, "101_1_2": false, "101_1_3": true, "101_1_4": false,
			}),
		},
		{
			name: "for ValidationType=Categories becomes validated when all its children " +
				"having Category=Validation are validated (should ignore items with NoScore=1",
			fixtures:       []string{"results_propagation/_common", "results_propagation/validated/all_and_category"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, resultStore *database.ResultStore) {
				itemStore := resultStore.Items()
				assert.NoError(t, itemStore.Where("id=4").UpdateColumn("no_score", true).Error())
				assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id IN (1, 3)").
					UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
				assert.NoError(t, resultStore.ItemItems().Where("parent_item_id = 2 AND child_item_id IN (3, 4)").
					UpdateColumn("category", "Validation").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_1": true, "101_1_2": true, "101_1_3": true, "101_1_4": false,
			}),
		},
		{
			name: "for ValidationType=Categories stays non-validated when there are no children " +
				"having Category=Validation (should ignore items with NoScore=1",
			fixtures:       []string{"results_propagation/_common"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, resultStore *database.ResultStore) {
				itemStore := resultStore.Items()
				assert.NoError(t, itemStore.Where("id = 1").UpdateColumn("no_score", true).Error())
				assert.NoError(t, resultStore.ItemItems().Where("parent_item_id = 2 AND child_item_id = 1").
					UpdateColumn("category", "Validation").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_1": false, "101_1_2": false,
			}),
		},
		{
			name: "for ValidationType=Categories becomes validated when all its children" +
				"having Category=Validation are validated",
			fixtures:       []string{"results_propagation/_common", "results_propagation/validated/all_and_category"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, resultStore *database.ResultStore) {
				assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id IN (3, 4)").
					UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
				assert.NoError(t, resultStore.ItemItems().Where("parent_item_id = 2 AND child_item_id IN (3, 4)").
					UpdateColumn("category", "Validation").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_1": false, "101_1_2": true, "101_1_3": true, "101_1_4": true,
			}),
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testResultStorePropagateValidated(t, testCase.fixtures,
				testCase.validationType, testCase.prepareFunc, testCase.expectedResults)
		})
	}
}

func buildExpectedValidatedResultRows(validatedMap map[string]bool) []validatedResultRow {
	result := make([]validatedResultRow, 0, len(validatedMap)+1)
	addResultForAnotherUser := true
	for id, validated := range validatedMap {
		var participantID, attemptID, itemID int64
		_, _ = fmt.Sscanf(id, "%d_%d_%d", &participantID, &attemptID, &itemID)
		result = append(result, validatedResultRow{
			ParticipantID: participantID, AttemptID: attemptID, ItemID: itemID,
			Validated: validated, State: "done",
		})
		if participantID == 102 && attemptID == 1 && itemID == 2 {
			addResultForAnotherUser = false
		}
	}

	// another user
	if addResultForAnotherUser {
		result = append(result, validatedResultRow{
			ParticipantID: 102, AttemptID: 1, ItemID: 2, Validated: false, State: "done",
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LessThan(result[j])
	})

	return result
}
