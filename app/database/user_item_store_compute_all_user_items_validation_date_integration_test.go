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
	ValidationDate            *database.Time
	AncestorsComputationState string
}

func TestUserItemStore_ComputeAllUserItems_ValidationDateStaysTheSameIfItWasNotNull(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	expectedDate := time.Now().Round(time.Second).UTC()
	expectedOldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, userItemStore.Where("id=12").UpdateColumn("validation_date", expectedOldDate).Error())
	assert.NoError(t, userItemStore.Where("id=11").UpdateColumn("validation_date", expectedDate).Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, userItemStore.Select("id, validation_date, ancestors_computation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidationDate: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 12, ValidationDate: (*database.Time)(&expectedOldDate), AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidationDate: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_NonCategories_SetsValidationDateToMaxOfChildrenValidationDates(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/validation_date")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, userItemStore.Where("id=13").UpdateColumn("validation_date", oldDate).Error())
	assert.NoError(t, userItemStore.Where("id=11").UpdateColumn("validation_date", expectedDate).Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, userItemStore.Select("id, validation_date, ancestors_computation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidationDate: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 12, ValidationDate: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 13, ValidationDate: (*database.Time)(&oldDate), AncestorsComputationState: "done"},
		{ID: 14, ValidationDate: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidationDate: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_Categories_SetsValidationDateToMaxOfValidationDatesOfChildrenWithCategoryValidation_NoSuitableChildren( // nolint:lll
	t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/validation_date")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, userItemStore.Where("id=13").UpdateColumn("validation_date", oldDate).Error())
	assert.NoError(t, userItemStore.Where("id=11").UpdateColumn("validation_date", expectedDate).Error())
	assert.NoError(
		t, database.NewDataStore(db).Items().Where("id=2").UpdateColumn("validation_type", "Categories").
			Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, userItemStore.Select("id, validation_date, ancestors_computation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidationDate: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 12, ValidationDate: nil, AncestorsComputationState: "done"},
		{ID: 13, ValidationDate: (*database.Time)(&oldDate), AncestorsComputationState: "done"},
		{ID: 14, ValidationDate: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidationDate: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_Categories_SetsValidationDateToMaxOfValidationDatesOfChildrenWithCategoryValidation(
	t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/validation_date")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, userItemStore.Where("id=13").UpdateColumn("validation_date", oldDate).Error())
	assert.NoError(t, userItemStore.Where("id=11").UpdateColumn("validation_date", expectedDate).Error())
	assert.NoError(
		t, database.NewDataStore(db).Items().Where("id=2").UpdateColumn("validation_type", "Categories").
			Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("id IN (23,24)").UpdateColumn("category", "Validation").Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, userItemStore.Select("id, validation_date, ancestors_computation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidationDate: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 12, ValidationDate: (*database.Time)(&oldDate), AncestorsComputationState: "done"},
		{ID: 13, ValidationDate: (*database.Time)(&oldDate), AncestorsComputationState: "done"},
		{ID: 14, ValidationDate: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidationDate: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_Categories_SetsValidationDateToMaxOfValidationDatesOfChildrenWithCategoryValidation_IgnoresCoursesAndNoScoreItems( // nolint:lll
	t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/validation_date")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	itemStore := database.NewDataStore(db).Items()
	assert.NoError(t, userItemStore.Where("id=13").UpdateColumn("validation_date", oldDate).Error())
	assert.NoError(t, userItemStore.Where("id=11").UpdateColumn("validation_date", expectedDate).Error())
	assert.NoError(t, itemStore.Where("id=2").UpdateColumn("validation_type", "Categories").Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("id IN (21,23,24)").UpdateColumn("category", "Validation").Error())

	assert.NoError(t, itemStore.Where("id=1").Updates(map[string]interface{}{
		"type": "Course",
	}).Error())
	assert.NoError(t, itemStore.Where("id=3").Updates(map[string]interface{}{
		"no_score": true,
	}).Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, userItemStore.Select("id, validation_date, ancestors_computation_state").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidationDate: (*database.Time)(&expectedDate), AncestorsComputationState: "done"},
		{ID: 12, ValidationDate: nil, AncestorsComputationState: "done"},
		{ID: 13, ValidationDate: (*database.Time)(&oldDate), AncestorsComputationState: "done"},
		{ID: 14, ValidationDate: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidationDate: nil, AncestorsComputationState: "done"},
	}, result)
}
