package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type validationDateResultRow struct {
	ID                        int64      `gorm:"column:ID"`
	ValidationDate            *time.Time `gorm:"column:sValidationDate"`
	AncestorsComputationState string     `gorm:"column:sAncestorsComputationState"`
}

func TestUserItemStore_ComputeAllUserItems_ValidationDateStaysTheSameIfItWasNotNull(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	expectedDate := time.Now().Round(time.Second).UTC()
	expectedOldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, userItemStore.Where("ID=12").UpdateColumn("sValidationDate", expectedOldDate).Error())
	assert.NoError(t, userItemStore.Where("ID=11").UpdateColumn("sValidationDate", expectedDate).Error())

	err := userItemStore.ComputeAllUserItems()
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, userItemStore.Select("ID, sValidationDate, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidationDate: &expectedDate, AncestorsComputationState: "done"},
		{ID: 12, ValidationDate: &expectedOldDate, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidationDate: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_NonCategories_SetsValidationDateToMaxOfChildrenValidationDates(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/sValidationDate")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, userItemStore.Where("ID=13").UpdateColumn("sValidationDate", oldDate).Error())
	assert.NoError(t, userItemStore.Where("ID=11").UpdateColumn("sValidationDate", expectedDate).Error())

	err := userItemStore.ComputeAllUserItems()
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, userItemStore.Select("ID, sValidationDate, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidationDate: &expectedDate, AncestorsComputationState: "done"},
		{ID: 12, ValidationDate: &expectedDate, AncestorsComputationState: "done"},
		{ID: 13, ValidationDate: &oldDate, AncestorsComputationState: "done"},
		{ID: 14, ValidationDate: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidationDate: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_Categories_SetsValidationDateToMaxOfValidationDatesOfChildrenWithCategoryValidation_NoSuitableChildren(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/sValidationDate")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, userItemStore.Where("ID=13").UpdateColumn("sValidationDate", oldDate).Error())
	assert.NoError(t, userItemStore.Where("ID=11").UpdateColumn("sValidationDate", expectedDate).Error())
	assert.NoError(
		t, database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", "Categories").
			Error())

	err := userItemStore.ComputeAllUserItems()
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, userItemStore.Select("ID, sValidationDate, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidationDate: &expectedDate, AncestorsComputationState: "done"},
		{ID: 12, ValidationDate: nil, AncestorsComputationState: "done"},
		{ID: 13, ValidationDate: &oldDate, AncestorsComputationState: "done"},
		{ID: 14, ValidationDate: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidationDate: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_Categories_SetsValidationDateToMaxOfValidationDatesOfChildrenWithCategoryValidation(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/sValidationDate")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	assert.NoError(t, userItemStore.Where("ID=13").UpdateColumn("sValidationDate", oldDate).Error())
	assert.NoError(t, userItemStore.Where("ID=11").UpdateColumn("sValidationDate", expectedDate).Error())
	assert.NoError(
		t, database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", "Categories").
			Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("ID IN (23,24)").UpdateColumn("sCategory", "Validation").Error())

	err := userItemStore.ComputeAllUserItems()
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, userItemStore.Select("ID, sValidationDate, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidationDate: &expectedDate, AncestorsComputationState: "done"},
		{ID: 12, ValidationDate: &oldDate, AncestorsComputationState: "done"},
		{ID: 13, ValidationDate: &oldDate, AncestorsComputationState: "done"},
		{ID: 14, ValidationDate: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidationDate: nil, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_Categories_SetsValidationDateToMaxOfValidationDatesOfChildrenWithCategoryValidation_IgnoresCoursesAndNoScoreItems(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/sValidationDate")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	expectedDate := time.Now().Round(time.Second).UTC()
	oldDate := expectedDate.AddDate(-1, -1, -1)

	itemStore := database.NewDataStore(db).Items()
	assert.NoError(t, userItemStore.Where("ID=13").UpdateColumn("sValidationDate", oldDate).Error())
	assert.NoError(t, userItemStore.Where("ID=11").UpdateColumn("sValidationDate", expectedDate).Error())
	assert.NoError(t, itemStore.Where("ID=2").UpdateColumn("sValidationType", "Categories").Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("ID IN (21,23,24)").UpdateColumn("sCategory", "Validation").Error())

	assert.NoError(t, itemStore.Where("ID=1").Updates(map[string]interface{}{
		"sType": "Course",
	}).Error())
	assert.NoError(t, itemStore.Where("ID=3").Updates(map[string]interface{}{
		"bNoScore": true,
	}).Error())

	err := userItemStore.ComputeAllUserItems()
	assert.NoError(t, err)

	var result []validationDateResultRow
	assert.NoError(t, userItemStore.Select("ID, sValidationDate, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validationDateResultRow{
		{ID: 11, ValidationDate: &expectedDate, AncestorsComputationState: "done"},
		{ID: 12, ValidationDate: nil, AncestorsComputationState: "done"},
		{ID: 13, ValidationDate: &oldDate, AncestorsComputationState: "done"},
		{ID: 14, ValidationDate: nil, AncestorsComputationState: "done"},
		// another user
		{ID: 22, ValidationDate: nil, AncestorsComputationState: "done"},
	}, result)
}
