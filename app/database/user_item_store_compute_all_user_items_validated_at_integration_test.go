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
	ID                        int64
	ValidatedAt               *database.Time
	AncestorsComputationState string
}

func TestUserItemStore_ComputeAllUserItems_ValidatedAtStaysTheSameIfItWasNotNull(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()

	expectedDate := time.Now().Round(time.Second).UTC()
	expectedOldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, groupAttemptStore.Where("id=12").UpdateColumn("validated_at", expectedOldDate).Error())
	assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn("validated_at", expectedDate).Error())

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated_at, ancestors_computation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidatedAt: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 12, ValidatedAt: (*database.Time)(&expectedOldDate), AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidatedAt: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_NonCategories_SetsValidatedAtToMaxOfChildrenValidatedAts(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/validated_at")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated_at", oldDate).Error())
	assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn("validated_at", expectedDate).Error())

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated_at, ancestors_computation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidatedAt: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 12, ValidatedAt: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 13, ValidatedAt: (*database.Time)(&oldDate), AncestorsComputationState: "done"},
		{ID: 14, ValidatedAt: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidatedAt: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_Categories_SetsValidatedAtToMaxOfValidatedAtsOfChildrenWithCategoryValidation_NoSuitableChildren( // nolint:lll
	t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/validated_at")
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
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated_at, ancestors_computation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidatedAt: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 12, ValidatedAt: nil, AncestorsComputationState: "done"},
		{ID: 13, ValidatedAt: (*database.Time)(&oldDate), AncestorsComputationState: "done"},
		{ID: 14, ValidatedAt: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidatedAt: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_Categories_SetsValidatedAtToMaxOfValidatedAtsOfChildrenWithCategoryValidation(
	t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/validated_at")
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
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated_at, ancestors_computation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidatedAt: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 12, ValidatedAt: (*database.Time)(&oldDate), AncestorsComputationState: "done"},
		{ID: 13, ValidatedAt: (*database.Time)(&oldDate), AncestorsComputationState: "done"},
		{ID: 14, ValidatedAt: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidatedAt: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_Categories_SetsValidatedAtToMaxOfValidatedAtsOfChildrenWithCategoryValidation_IgnoresCoursesAndNoScoreItems( // nolint:lll
	t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/validated_at")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	itemStore := database.NewDataStore(db).Items()
	assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn("validated_at", oldDate).Error())
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
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, validated_at, ancestors_computation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidatedAt: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 12, ValidatedAt: nil, AncestorsComputationState: "done"},
		{ID: 13, ValidatedAt: (*database.Time)(&oldDate), AncestorsComputationState: "done"},
		{ID: 14, ValidatedAt: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidatedAt: nil, AncestorsComputationState: "done"},
	}, result)
}
