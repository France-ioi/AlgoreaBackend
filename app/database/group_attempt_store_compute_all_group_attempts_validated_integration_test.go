// +build !unit

package database_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type validatedResultRow struct {
	ID                     int64
	Validated              bool
	ResultPropagationState string
}

func testGroupAttemptStoreComputeAllGroupAttemptsValidated(t *testing.T, fixtures []string,
	validationType string,
	prepareFunc func(*testing.T, *database.GroupAttemptStore), expectedResults []validatedResultRow) {
	db := testhelpers.SetupDBWithFixture(fixtures...)
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()
	assert.NoError(t,
		groupAttemptStore.Items().Where("id=2").
			UpdateColumn("validation_type", validationType).Error())
	if prepareFunc != nil {
		prepareFunc(t, groupAttemptStore)
	}

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated, result_propagation_state").
		Order("id").Scan(&result).Error())
	assert.Equal(t, expectedResults, result)
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_ValidatedStaysNonValidatedFor(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "None"},
		{name: "Manual"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testGroupAttemptStoreComputeAllGroupAttemptsValidated(t,
				[]string{"groups_attempts_propagation/_common"},
				tt.name,
				func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
					assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
				},
				buildExpectedValidatedResultRows(map[int64]bool{
					11: true, 12: false,
				}))
		})
	}
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_ValidatedWithValidationTypeOneBecomesValidatedWhenThereIsAtLeastOneValidatedChild(
	t *testing.T) {
	testGroupAttemptStoreComputeAllGroupAttemptsValidated(t,
		[]string{"groups_attempts_propagation/_common", "groups_attempts_propagation/validated/one"},
		"One",
		func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
			assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
		},
		buildExpectedValidatedResultRows(map[int64]bool{
			11: false, 12: true, 13: true,
		}))
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_ValidatedWithValidationTypeOneStaysNonValidatedWhenThereAreNoValidatedChildren(
	t *testing.T) {
	testGroupAttemptStoreComputeAllGroupAttemptsValidated(t,
		[]string{"groups_attempts_propagation/_common", "groups_attempts_propagation/validated/one"},
		"One",
		nil,
		buildExpectedValidatedResultRows(map[int64]bool{
			11: false, 12: false, 13: false,
		}))
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_Validated(t *testing.T) {
	tests := []struct {
		name            string
		fixtures        []string
		validationType  string
		prepareFunc     func(*testing.T, *database.GroupAttemptStore)
		expectedResults []validatedResultRow
	}{
		{
			name:           "for ValidationType=AllButOne stays non-validated when there are two non-validated children",
			fixtures:       []string{"groups_attempts_propagation/_common", "groups_attempts_propagation/validated/all_and_category"},
			validationType: "AllButOne",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: false, 12: false, 13: true, 14: false,
			}),
		},
		{
			name:           "for ValidationType=AllButOne stays non-validated when there is just one child and it is not validated",
			fixtures:       []string{"groups_attempts_propagation/_common"},
			validationType: "AllButOne",
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: false, 12: false,
			}),
		},
		{
			name:           "for ValidationType=AllButOne becomes validated when there are less than two non-validated children",
			fixtures:       []string{"groups_attempts_propagation/_common", "groups_attempts_propagation/validated/all_and_category"},
			validationType: "AllButOne",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id IN (11, 13)").UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: true, 12: true, 13: true, 14: false,
			}),
		},
		{
			name:           "for ValidationType=All stays non-validated when there is at least one non-validated child",
			fixtures:       []string{"groups_attempts_propagation/_common", "groups_attempts_propagation/validated/all_and_category"},
			validationType: "All",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id IN (11,13)").UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: true, 12: false, 13: true, 14: false,
			}),
		},
		{
			name:           "for ValidationType=All stays non-validated when there are no children",
			fixtures:       []string{"groups_attempts_propagation/_common"},
			validationType: "All",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.ItemItems().Delete("parent_item_id=2").Error())
				assert.NoError(t, groupAttemptStore.Delete("item_id=1").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				12: false,
			}),
		},
		{
			name:           "for ValidationType=All becomes validated when all its children are validated",
			fixtures:       []string{"groups_attempts_propagation/_common", "groups_attempts_propagation/validated/all_and_category"},
			validationType: "All",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id IN (11,13,14)").UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: true, 12: true, 13: true, 14: true,
			}),
		},
		{
			name: "for ValidationType=Categories stays non-validated when " +
				"there is at least one non-validated child item with Category=Validation",
			fixtures:       []string{"groups_attempts_propagation/_common", "groups_attempts_propagation/validated/all_and_category"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id IN (11,13)").UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
				assert.NoError(t, groupAttemptStore.ItemItems().Where("id IN (23,24)").
					UpdateColumn("category", "Validation").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: true, 12: false, 13: true, 14: false,
			}),
		},
		{
			name: "for ValidationType=Categories becomes validated when all its children " +
				"having Category=Validation are validated (should ignore items with NoScore=1",
			fixtures:       []string{"groups_attempts_propagation/_common", "groups_attempts_propagation/validated/all_and_category"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				itemStore := groupAttemptStore.Items()
				assert.NoError(t, itemStore.Where("id=4").UpdateColumn("no_score", true).Error())
				assert.NoError(t, groupAttemptStore.Where("id IN (11,13)").UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
				assert.NoError(t, groupAttemptStore.ItemItems().Where("id IN (23,24)").UpdateColumn("category", "Validation").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: true, 12: true, 13: true, 14: false,
			}),
		},
		{
			name: "for ValidationType=Categories stays non-validated when there are no children " +
				"having Category=Validation (should ignore items with NoScore=1",
			fixtures:       []string{"groups_attempts_propagation/_common"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				itemStore := groupAttemptStore.Items()
				assert.NoError(t, itemStore.Where("id=1").UpdateColumn("no_score", true).Error())
				assert.NoError(t, groupAttemptStore.ItemItems().Where("id=21").UpdateColumn("category", "Validation").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: false, 12: false,
			}),
		},
		{
			name: "for ValidationType=Categories becomes validated when all its children" +
				"having Category=Validation are validated",
			fixtures:       []string{"groups_attempts_propagation/_common", "groups_attempts_propagation/validated/all_and_category"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id IN (13,14)").UpdateColumn("validated_at", "2019-05-30 11:00:00").Error())
				assert.NoError(t, groupAttemptStore.ItemItems().Where("id IN (23,24)").UpdateColumn("category", "Validation").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: false, 12: true, 13: true, 14: true,
			}),
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testGroupAttemptStoreComputeAllGroupAttemptsValidated(t, testCase.fixtures,
				testCase.validationType, testCase.prepareFunc, testCase.expectedResults)
		})
	}
}

func buildExpectedValidatedResultRows(validatedMap map[int64]bool) []validatedResultRow {
	result := make([]validatedResultRow, 0, len(validatedMap)+1)
	addResultForAnotherUser := true
	for id, validated := range validatedMap {
		result = append(result, validatedResultRow{ID: id, Validated: validated, ResultPropagationState: "done"})
		if id == 22 {
			addResultForAnotherUser = false
		}
	}

	// another user
	if addResultForAnotherUser {
		result = append(result, validatedResultRow{ID: 22, Validated: false, ResultPropagationState: "done"})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}
