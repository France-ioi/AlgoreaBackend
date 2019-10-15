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
	ID                        int64
	Validated                 bool
	AncestorsComputationState string
}

func testUserItemStoreComputeAllUserItemsValidated(t *testing.T, fixtures []string,
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
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated, ancestors_computation_state").
		Order("id").Scan(&result).Error())
	assert.Equal(t, expectedResults, result)
}

func TestUserItemStore_ComputeAllUserItems_bValidatedStaysValidated(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "None"},
		{name: "All"},
		{name: "AllButOne"},
		{name: "Categories"},
		{name: "One"},
		{name: "Manual"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testUserItemStoreComputeAllUserItemsValidated(t,
				[]string{"users_items_propagation/_common"},
				tt.name,
				func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
					assert.NoError(t, groupAttemptStore.Where("id=12").UpdateColumn("validated", true).Error())
				},
				buildExpectedValidatedResultRows(map[int64]bool{
					11: false, 12: true,
					// another user
					22: tt.name == "Categories" || tt.name == "AllButOne",
				}))
		})
	}
}

func TestUserItemStore_ComputeAllUserItems_bValidatedStaysNonValidatedFor(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "None"},
		{name: "Manual"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testUserItemStoreComputeAllUserItemsValidated(t,
				[]string{"users_items_propagation/_common"},
				tt.name,
				func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
					assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn("validated", true).Error())
				},
				buildExpectedValidatedResultRows(map[int64]bool{
					11: true, 12: false,
				}))
		})
	}
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeOneStaysNonValidatedWhenThereAreNoValidatedChildren(t *testing.T) {
	testUserItemStoreComputeAllUserItemsValidated(t,
		[]string{"users_items_propagation/_common", "users_items_propagation/validated/one"},
		"One",
		func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
			assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated", true).Error())
		},
		buildExpectedValidatedResultRows(map[int64]bool{
			11: false, 12: true, 13: true,
		}))
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeOneBecomesValidatedWhenThereIsAtLeastOneValidatedChild(
	t *testing.T) {
	testUserItemStoreComputeAllUserItemsValidated(t,
		[]string{"users_items_propagation/_common", "users_items_propagation/validated/one"},
		"One",
		nil,
		buildExpectedValidatedResultRows(map[int64]bool{
			11: false, 12: false, 13: false,
		}))
}

func TestUserItemStore_ComputeAllUserItems_bValidated(t *testing.T) {
	tests := []struct {
		name            string
		fixtures        []string
		validationType  string
		prepareFunc     func(*testing.T, *database.GroupAttemptStore)
		expectedResults []validatedResultRow
	}{
		{
			name:           "user_item with ValidationType=AllButOne stays non-validated when there are two non-validated children",
			fixtures:       []string{"users_items_propagation/_common", "users_items_propagation/validated/all_and_category"},
			validationType: "AllButOne",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated", true).Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: false, 12: false, 13: true, 14: false,
			}),
		},
		{
			name:           "user_item with ValidationType=AllButOne becomes validated when there are less than two non-validated children",
			fixtures:       []string{"users_items_propagation/_common", "users_items_propagation/validated/all_and_category"},
			validationType: "AllButOne",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id IN (11, 13)").UpdateColumn("validated", true).Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: true, 12: true, 13: true, 14: false,
			}),
		},
		{
			name:           "user_item with ValidationType=All stays non-validated when there is at least one non-validated child",
			fixtures:       []string{"users_items_propagation/_common", "users_items_propagation/validated/all_and_category"},
			validationType: "All",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id IN (11,13)").UpdateColumn("validated", true).Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: true, 12: false, 13: true, 14: false,
			}),
		},
		{
			name:           "user_item with ValidationType=All becomes validated when all its children are validated",
			fixtures:       []string{"users_items_propagation/_common", "users_items_propagation/validated/all_and_category"},
			validationType: "All",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id IN (11,13,14)").UpdateColumn("validated", true).Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: true, 12: true, 13: true, 14: true,
			}),
		},
		{
			name: "user_item with ValidationType=Categories stays non-validated when " +
				"there is at least one non-validated child item with Category=Validation",
			fixtures:       []string{"users_items_propagation/_common", "users_items_propagation/validated/all_and_category"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id IN (11,13)").UpdateColumn("validated", true).Error())
				assert.NoError(t, groupAttemptStore.ItemItems().Where("id IN (23,24)").
					UpdateColumn("category", "Validation").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: true, 12: false, 13: true, 14: false,
			}),
		},
		{
			name: "user_item with ValidationType=Categories becomes validated when all its children " +
				"having Category=Validation are validated (should ignore items with NoScore=1",
			fixtures:       []string{"users_items_propagation/_common", "users_items_propagation/validated/all_and_category"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				itemStore := groupAttemptStore.Items()
				assert.NoError(t, itemStore.Where("id=4").UpdateColumn("no_score", true).Error())
				assert.NoError(t, groupAttemptStore.Where("id IN (11,13)").UpdateColumn("validated", true).Error())
				assert.NoError(t, groupAttemptStore.ItemItems().Where("id IN (23,24)").UpdateColumn("category", "Validation").Error())
			},
			expectedResults: buildExpectedValidatedResultRows(map[int64]bool{
				11: true, 12: true, 13: true, 14: false,
			}),
		},
		{
			name: "user_item with ValidationType=Categories becomes validated when all its children" +
				"having Category=Validation are validated",
			fixtures:       []string{"users_items_propagation/_common", "users_items_propagation/validated/all_and_category"},
			validationType: "Categories",
			prepareFunc: func(t *testing.T, groupAttemptStore *database.GroupAttemptStore) {
				assert.NoError(t, groupAttemptStore.Where("id IN (13,14)").UpdateColumn("validated", true).Error())
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
			testUserItemStoreComputeAllUserItemsValidated(t, testCase.fixtures,
				testCase.validationType, testCase.prepareFunc, testCase.expectedResults)
		})
	}
}

func buildExpectedValidatedResultRows(validatedMap map[int64]bool) []validatedResultRow {
	result := make([]validatedResultRow, 0, len(validatedMap)+1)
	addResultForAnotherUser := true
	for id, validated := range validatedMap {
		result = append(result, validatedResultRow{ID: id, Validated: validated, AncestorsComputationState: "done"})
		if id == 22 {
			addResultForAnotherUser = false
		}
	}

	// another user
	if addResultForAnotherUser {
		result = append(result, validatedResultRow{ID: 22, Validated: false, AncestorsComputationState: "done"})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}
