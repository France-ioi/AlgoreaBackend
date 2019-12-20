// +build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type validationDateResultRow struct {
	ID                     int64
	ValidatedAt            *database.Time
	ResultPropagationState string
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_NonCategories_SetsValidatedAtToMaxOfChildrenValidatedAts(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("groups_attempts_propagation/_common", "groups_attempts_propagation/validated_at")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()

	baseDate := time.Now().Round(time.Second).UTC()
	skippedDate := baseDate.AddDate(-2, -1, -1)
	oldestForItem3 := baseDate.AddDate(-1, -1, -1)
	skippedInItem3 := oldestForItem3.Add(24 * time.Hour)
	oldestForItem4AndWinner := baseDate.AddDate(0, -1, -1)
	skippedInItem4 := oldestForItem4AndWinner.Add(24 * time.Hour)

	assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated_at", oldestForItem3).Error())
	assert.NoError(t, groupAttemptStore.Where("id=15").UpdateColumn("validated_at", skippedInItem3).Error())

	assert.NoError(t, groupAttemptStore.Where("id=14").UpdateColumn("validated_at", oldestForItem4AndWinner).Error())
	assert.NoError(t, groupAttemptStore.Where("id=16").UpdateColumn("validated_at", skippedInItem4).Error())

	assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn("validated_at", skippedDate).Error())

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated_at, result_propagation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidatedAt: (*database.Time)(&skippedDate), ResultPropagationState: "done"},
		{ID: 12, ValidatedAt: (*database.Time)(&oldestForItem4AndWinner), ResultPropagationState: "done"}, // the result
		{ID: 13, ValidatedAt: (*database.Time)(&oldestForItem3), ResultPropagationState: "done"},
		{ID: 14, ValidatedAt: (*database.Time)(&oldestForItem4AndWinner), ResultPropagationState: "done"},
		{ID: 15, ValidatedAt: (*database.Time)(&skippedInItem3), ResultPropagationState: "done"},
		{ID: 16, ValidatedAt: (*database.Time)(&skippedInItem4), ResultPropagationState: "done"},
		// another user
		{ID: 22, ValidatedAt: nil, ResultPropagationState: "done"},
	}, result)
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_Categories_SetsValidatedAtToMaxOfValidatedAtsOfChildrenWithCategoryValidation_NoSuitableChildren( // nolint:lll
	t *testing.T) {
	db := testhelpers.SetupDBWithFixture("groups_attempts_propagation/_common", "groups_attempts_propagation/validated_at")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated_at", oldDate).Error())
	assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn("validated_at", expectedDate).Error())
	assert.NoError(
		t, database.NewDataStore(db).Items().Where("id=2").UpdateColumn("validation_type", "Categories").
			Error())

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated_at, result_propagation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidatedAt: (*database.Time)(&expectedDate), ResultPropagationState: "done"},
		{ID: 12, ValidatedAt: nil, ResultPropagationState: "done"},
		{ID: 13, ValidatedAt: (*database.Time)(&oldDate), ResultPropagationState: "done"},
		{ID: 14, ValidatedAt: nil, ResultPropagationState: "done"},
		{ID: 15, ValidatedAt: nil, ResultPropagationState: "done"},
		{ID: 16, ValidatedAt: nil, ResultPropagationState: "done"},
		// another user
		{ID: 22, ValidatedAt: nil, ResultPropagationState: "done"},
	}, result)
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_Categories_SetsValidatedAtToNull_IfSomeCategoriesAreNotValidated(
	t *testing.T) {
	db := testhelpers.SetupDBWithFixture("groups_attempts_propagation/_common", "groups_attempts_propagation/validated_at")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated_at", oldDate).Error())
	assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn("validated_at", expectedDate).Error())
	assert.NoError(
		t, database.NewDataStore(db).Items().Where("id=2").UpdateColumn("validation_type", "Categories").
			Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("id IN (23,24)").UpdateColumn("category", "Validation").Error())

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated_at, result_propagation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidatedAt: (*database.Time)(&expectedDate), ResultPropagationState: "done"},
		{ID: 12, ValidatedAt: nil, ResultPropagationState: "done"},
		{ID: 13, ValidatedAt: (*database.Time)(&oldDate), ResultPropagationState: "done"},
		{ID: 14, ValidatedAt: nil, ResultPropagationState: "done"},
		{ID: 15, ValidatedAt: nil, ResultPropagationState: "done"},
		{ID: 16, ValidatedAt: nil, ResultPropagationState: "done"},
		// another user
		{ID: 22, ValidatedAt: nil, ResultPropagationState: "done"},
	}, result)
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_Categories_ValidatedAtShouldBeMaxOfChildrensWithCategoryValidation_IfAllAreValidated(
	t *testing.T) {
	db := testhelpers.SetupDBWithFixture("groups_attempts_propagation/_common", "groups_attempts_propagation/validated_at")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated_at", oldDate).Error())
	assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn("validated_at", oldDate).Error())
	assert.NoError(t, groupAttemptStore.Where("id=16").UpdateColumn("validated_at", expectedDate).Error())
	assert.NoError(
		t, database.NewDataStore(db).Items().Where("id=2").UpdateColumn("validation_type", "Categories").
			Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("id IN (23,24)").UpdateColumn("category", "Validation").Error())

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated_at, result_propagation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidatedAt: (*database.Time)(&oldDate), ResultPropagationState: "done"},
		{ID: 12, ValidatedAt: (*database.Time)(&expectedDate), ResultPropagationState: "done"},
		{ID: 13, ValidatedAt: (*database.Time)(&oldDate), ResultPropagationState: "done"},
		{ID: 14, ValidatedAt: nil, ResultPropagationState: "done"},
		{ID: 15, ValidatedAt: nil, ResultPropagationState: "done"},
		{ID: 16, ValidatedAt: (*database.Time)(&expectedDate), ResultPropagationState: "done"},
		// another user
		{ID: 22, ValidatedAt: nil, ResultPropagationState: "done"},
	}, result)
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_Categories_SetsValidatedAtToMaxOfValidatedAtsOfChildrenWithCategoryValidation_IgnoresNoScoreItems( // nolint:lll
	t *testing.T) {
	db := testhelpers.SetupDBWithFixture("groups_attempts_propagation/_common", "groups_attempts_propagation/validated_at")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)
	oldDatePlusOneDay := oldDate.Add(24 * time.Hour)

	itemStore := database.NewDataStore(db).Items()
	assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated_at", oldDate).Error())
	assert.NoError(t, groupAttemptStore.Where("id=14").UpdateColumn("validated_at", oldDate).Error())
	assert.NoError(t, groupAttemptStore.Where("id=15").UpdateColumn("validated_at", oldDatePlusOneDay).Error()) // should be ignored
	assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn("validated_at", expectedDate).Error())
	assert.NoError(t, itemStore.Where("id=2").UpdateColumn("validation_type", "Categories").Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("id IN (21,23,24)").UpdateColumn("category", "Validation").Error())

	assert.NoError(t, itemStore.Where("id=1").Updates(map[string]interface{}{
		"type": "Course",
	}).Error())
	assert.NoError(t, itemStore.Where("id=3").Updates(map[string]interface{}{
		"no_score": true,
	}).Error())

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated_at, result_propagation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidatedAt: (*database.Time)(&expectedDate), ResultPropagationState: "done"},
		{ID: 12, ValidatedAt: (*database.Time)(&expectedDate), ResultPropagationState: "done"},
		{ID: 13, ValidatedAt: (*database.Time)(&oldDate), ResultPropagationState: "done"},
		{ID: 14, ValidatedAt: (*database.Time)(&oldDate), ResultPropagationState: "done"},
		{ID: 15, ValidatedAt: (*database.Time)(&oldDatePlusOneDay), ResultPropagationState: "done"},
		{ID: 16, ValidatedAt: nil, ResultPropagationState: "done"},
		// another user
		{ID: 22, ValidatedAt: nil, ResultPropagationState: "done"},
	}, result)
}
