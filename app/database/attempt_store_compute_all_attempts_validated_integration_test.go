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
	ParticipantID          int64
	AttemptID              int64
	ItemID                 int64
	Validated              bool
	ResultPropagationState string
}

func testAttemptStoreComputeAllAttemptsValidated(t *testing.T, fixtures []string,
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
		return s.Attempts().ComputeAllAttempts()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, resultStore.Select("participant_id, attempt_id, item_id, validated, result_propagation_state").
		Order("participant_id, attempt_id, item_id").Scan(&result).Error())
	assert.Equal(t, expectedResults, result)
}

func TestAttemptStore_ComputeAllAttempts_ValidatedStaysNonValidatedFor(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "None"},
		{name: "Manual"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testAttemptStoreComputeAllAttemptsValidated(t,
				[]string{"attempts_propagation/_common"},
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

func TestAttemptStore_ComputeAllAttempts_ValidatedWithValidationTypeOneBecomesValidatedWhenThereIsAtLeastOneValidatedChild(
	t *testing.T) {
	testAttemptStoreComputeAllAttemptsValidated(t,
		[]string{"attempts_propagation/_common", "attempts_propagation/validated/one"},
		"One",
		func(t *testing.T, resultStore *database.ResultStore) {
			assert.NoError(t, resultStore.Where("participant_id = 101 AND attempt_id = 1 AND item_id = 3").
				UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
		},
		buildExpectedValidatedResultRows(map[string]bool{
			"101_1_1": false, "101_1_2": true, "101_1_3": true,
		}))
}

func TestAttemptStore_ComputeAllAttempts_ValidatedWithValidationTypeOneStaysNonValidatedWhenThereAreNoValidatedChildren(
	t *testing.T) {
	testAttemptStoreComputeAllAttemptsValidated(t,
		[]string{"attempts_propagation/_common", "attempts_propagation/validated/one"},
		"One",
		nil,
		buildExpectedValidatedResultRows(map[string]bool{
			"101_1_1": false, "101_1_2": false, "101_1_3": false,
		}))
}

func TestAttemptStore_ComputeAllAttempts_Validated(t *testing.T) {
	tests := []struct {
		name            string
		fixtures        []string
		validationType  string
		prepareFunc     func(*testing.T, *database.ResultStore)
		expectedResults []validatedResultRow
	}{
		{
			name:           "for ValidationType=AllButOne stays non-validated when there are two non-validated children",
			fixtures:       []string{"attempts_propagation/_common", "attempts_propagation/validated/all_and_category"},
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
			fixtures:       []string{"attempts_propagation/_common"},
			validationType: "AllButOne",
			expectedResults: buildExpectedValidatedResultRows(map[string]bool{
				"101_1_1": false, "101_1_2": false,
			}),
		},
		{
			name:           "for ValidationType=AllButOne becomes validated when there are less than two non-validated children",
			fixtures:       []string{"attempts_propagation/_common", "attempts_propagation/validated/all_and_category"},
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
			fixtures:       []string{"attempts_propagation/_common", "attempts_propagation/validated/all_and_category"},
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
			fixtures:       []string{"attempts_propagation/_common"},
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
			fixtures:       []string{"attempts_propagation/_common", "attempts_propagation/validated/all_and_category"},
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
			fixtures:       []string{"attempts_propagation/_common", "attempts_propagation/validated/all_and_category"},
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
			fixtures:       []string{"attempts_propagation/_common", "attempts_propagation/validated/all_and_category"},
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
			fixtures:       []string{"attempts_propagation/_common"},
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
			fixtures:       []string{"attempts_propagation/_common", "attempts_propagation/validated/all_and_category"},
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
			testAttemptStoreComputeAllAttemptsValidated(t, testCase.fixtures,
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
			Validated: validated, ResultPropagationState: "done",
		})
		if participantID == 102 && attemptID == 1 && itemID == 2 {
			addResultForAnotherUser = false
		}
	}

	// another user
	if addResultForAnotherUser {
		result = append(result, validatedResultRow{
			ParticipantID: 102, AttemptID: 1, ItemID: 2, Validated: false, ResultPropagationState: "done",
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ParticipantID < result[j].ParticipantID ||
			result[i].ParticipantID == result[j].ParticipantID && result[i].AttemptID < result[j].AttemptID ||
			result[i].ParticipantID == result[j].ParticipantID && result[i].AttemptID == result[j].AttemptID &&
				result[i].ItemID < result[j].ItemID
	})

	return result
}
