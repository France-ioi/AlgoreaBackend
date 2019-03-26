// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type validatedResultRow struct {
	ID                        int64  `gorm:"column:ID"`
	Validated                 bool   `gorm:"column:bValidated"`
	AncestorsComputationState string `gorm:"column:sAncestorsComputationState"`
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
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixture("users_items_propagation/_common")
			defer func() { _ = db.Close() }()

			userItemStore := database.NewDataStore(db).UserItems()

			assert.NoError(t,
				database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", tt.name).
					Error())
			assert.NoError(t, userItemStore.Where("ID=12").UpdateColumn("bValidated", true).Error())

			err := userItemStore.InTransaction(func(s *database.DataStore) error {
				return s.UserItems().ComputeAllUserItems()
			})
			assert.NoError(t, err)

			var result []validatedResultRow
			assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
			assert.Equal(t, []validatedResultRow{
				{ID: 11, Validated: false, AncestorsComputationState: "done"},
				{ID: 12, Validated: true, AncestorsComputationState: "done"},
				// another user
				{ID: 22, Validated: tt.name == "Categories" || tt.name == "AllButOne", AncestorsComputationState: "done"},
			}, result)
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
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixture("users_items_propagation/_common")
			defer func() { _ = db.Close() }()

			userItemStore := database.NewDataStore(db).UserItems()

			assert.NoError(t,
				database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", tt.name).
					Error())
			assert.NoError(t, userItemStore.Where("ID=11").UpdateColumn("bValidated", true).Error())

			err := userItemStore.InTransaction(func(s *database.DataStore) error {
				return s.UserItems().ComputeAllUserItems()
			})
			assert.NoError(t, err)

			var result []validatedResultRow
			assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
			assert.Equal(t, []validatedResultRow{
				{ID: 11, Validated: true, AncestorsComputationState: "done"},
				{ID: 12, Validated: false, AncestorsComputationState: "done"},
				// another user
				{ID: 22, Validated: false, AncestorsComputationState: "done"},
			}, result)
		})
	}
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeOneStaysNonValidatedWhenThereAreNoValidatedChildren(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/bValidated/one")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	assert.NoError(t,
		database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", "One").
			Error())
	assert.NoError(t, userItemStore.Where("ID=13").UpdateColumn("bValidated", true).Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validatedResultRow{
		{ID: 11, Validated: false, AncestorsComputationState: "done"},
		{ID: 12, Validated: true, AncestorsComputationState: "done"},
		{ID: 13, Validated: true, AncestorsComputationState: "done"},
		// another user
		{ID: 22, Validated: false, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeOneBecomesValidatedWhenThereIsAtLeastOneValidatedChild(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/bValidated/one")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	assert.NoError(t,
		database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", "One").
			Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validatedResultRow{
		{ID: 11, Validated: false, AncestorsComputationState: "done"},
		{ID: 12, Validated: false, AncestorsComputationState: "done"},
		{ID: 13, Validated: false, AncestorsComputationState: "done"},
		// another user
		{ID: 22, Validated: false, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeAllButOneStaysNonValidatedWhenThereAreTwoNonValidatedChildren(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/bValidated/all_and_category")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	assert.NoError(t,
		database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", "AllButOne").
			Error())
	assert.NoError(t, userItemStore.Where("ID=13").UpdateColumn("bValidated", true).Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validatedResultRow{
		{ID: 11, Validated: false, AncestorsComputationState: "done"},
		{ID: 12, Validated: false, AncestorsComputationState: "done"},
		{ID: 13, Validated: true, AncestorsComputationState: "done"},
		{ID: 14, Validated: false, AncestorsComputationState: "done"},
		// another user
		{ID: 22, Validated: false, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeAllButOneBecomesValidatedWhenThereAreLessThanTwoNonValidatedChildren(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/bValidated/all_and_category")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	assert.NoError(t,
		database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", "AllButOne").
			Error())
	assert.NoError(t, userItemStore.Where("ID IN (11, 13)").UpdateColumn("bValidated", true).Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validatedResultRow{
		{ID: 11, Validated: true, AncestorsComputationState: "done"},
		{ID: 12, Validated: true, AncestorsComputationState: "done"},
		{ID: 13, Validated: true, AncestorsComputationState: "done"},
		{ID: 14, Validated: false, AncestorsComputationState: "done"},
		// another user
		{ID: 22, Validated: false, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeAllStaysNonValidatedWhenThereIsAtLeastOneNonValidatedChild(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/bValidated/all_and_category")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	assert.NoError(t,
		database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", "All").
			Error())
	assert.NoError(t, userItemStore.Where("ID IN (11,13)").UpdateColumn("bValidated", true).Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validatedResultRow{
		{ID: 11, Validated: true, AncestorsComputationState: "done"},
		{ID: 12, Validated: false, AncestorsComputationState: "done"},
		{ID: 13, Validated: true, AncestorsComputationState: "done"},
		{ID: 14, Validated: false, AncestorsComputationState: "done"},
		// another user
		{ID: 22, Validated: false, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeAllBecomesValidatedWhenThereAllChildrenAreValidated(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/bValidated/all_and_category")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	assert.NoError(t,
		database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", "All").
			Error())
	assert.NoError(t, userItemStore.Where("ID IN (11,13,14)").UpdateColumn("bValidated", true).Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validatedResultRow{
		{ID: 11, Validated: true, AncestorsComputationState: "done"},
		{ID: 12, Validated: true, AncestorsComputationState: "done"},
		{ID: 13, Validated: true, AncestorsComputationState: "done"},
		{ID: 14, Validated: true, AncestorsComputationState: "done"},
		// another user
		{ID: 22, Validated: false, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeCategoriesStaysNonValidatedWhenThereIsAtLeastOneNonValidatedChildWithCategoryValidation(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/bValidated/all_and_category")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	assert.NoError(t,
		database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", "Categories").
			Error())
	assert.NoError(t, userItemStore.Where("ID IN (11,13)").UpdateColumn("bValidated", true).Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("ID IN (23,24)").UpdateColumn("sCategory", "Validation").Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validatedResultRow{
		{ID: 11, Validated: true, AncestorsComputationState: "done"},
		{ID: 12, Validated: false, AncestorsComputationState: "done"},
		{ID: 13, Validated: true, AncestorsComputationState: "done"},
		{ID: 14, Validated: false, AncestorsComputationState: "done"},
		// another user
		{ID: 22, Validated: false, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeCategoriesBecomesValidatedWhenAllChildrenWithCategoryValidationAreValidated_IgnoresItemsWithTypeCourse(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/bValidated/all_and_category")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()
	itemStore := database.NewDataStore(db).Items()

	assert.NoError(t, itemStore.Where("ID=2").UpdateColumn("sValidationType", "Categories").Error())
	assert.NoError(t, itemStore.Where("ID=4").UpdateColumn("sType", "Course").Error())
	assert.NoError(t, userItemStore.Where("ID IN (11,13)").UpdateColumn("bValidated", true).Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("ID IN (23,24)").UpdateColumn("sCategory", "Validation").Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validatedResultRow{
		{ID: 11, Validated: true, AncestorsComputationState: "done"},
		{ID: 12, Validated: true, AncestorsComputationState: "done"},
		{ID: 13, Validated: true, AncestorsComputationState: "done"},
		{ID: 14, Validated: false, AncestorsComputationState: "done"},
		// another user
		{ID: 22, Validated: false, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeCategoriesBecomesValidatedWhenAllChildrenWithCategoryValidationAreValidated_IgnoresItemsWithNoScore(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/bValidated/all_and_category")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()
	itemStore := database.NewDataStore(db).Items()

	assert.NoError(t, itemStore.Where("ID=2").UpdateColumn("sValidationType", "Categories").Error())
	assert.NoError(t, itemStore.Where("ID=4").UpdateColumn("bNoScore", true).Error())
	assert.NoError(t, userItemStore.Where("ID IN (11,13)").UpdateColumn("bValidated", true).Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("ID IN (23,24)").UpdateColumn("sCategory", "Validation").Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validatedResultRow{
		{ID: 11, Validated: true, AncestorsComputationState: "done"},
		{ID: 12, Validated: true, AncestorsComputationState: "done"},
		{ID: 13, Validated: true, AncestorsComputationState: "done"},
		{ID: 14, Validated: false, AncestorsComputationState: "done"},
		// another user
		{ID: 22, Validated: false, AncestorsComputationState: "done"},
	}, result)
}

func TestUserItemStore_ComputeAllUserItems_bValidatedWithValidationTypeCategoriesBecomesValidatedWhenAllChildrenWithCategoryValidationAreValidated(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/bValidated/all_and_category")
	defer func() { _ = db.Close() }()

	userItemStore := database.NewDataStore(db).UserItems()

	assert.NoError(t,
		database.NewDataStore(db).Items().Where("ID=2").UpdateColumn("sValidationType", "Categories").
			Error())
	assert.NoError(t, userItemStore.Where("ID IN (13,14)").UpdateColumn("bValidated", true).Error())
	assert.NoError(t, database.NewDataStore(db).ItemItems().Where("ID IN (23,24)").UpdateColumn("sCategory", "Validation").Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []validatedResultRow
	assert.NoError(t, userItemStore.Select("ID, bValidated, sAncestorsComputationState").Scan(&result).Error())
	assert.Equal(t, []validatedResultRow{
		{ID: 11, Validated: false, AncestorsComputationState: "done"},
		{ID: 12, Validated: true, AncestorsComputationState: "done"},
		{ID: 13, Validated: true, AncestorsComputationState: "done"},
		{ID: 14, Validated: true, AncestorsComputationState: "done"},
		// another user
		{ID: 22, Validated: false, AncestorsComputationState: "done"},
	}, result)
}
