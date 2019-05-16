// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestUserItemStore_ComputeAllUserItems(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "basic", wantErr: false},
	}

	db := testhelpers.SetupDBWithFixture("users_items_propagation/main")
	defer func() { _ = db.Close() }()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := database.NewDataStore(db).InTransaction(func(s *database.DataStore) error {
				return s.UserItems().ComputeAllUserItems()
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("UserItemStore.computeAllUserItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserItemStore_ComputeAllUserItems_Concurrent(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/main")
	defer func() { _ = db.Close() }()

	testhelpers.RunConcurrently(func() {
		s := database.NewDataStore(db)
		err := s.InTransaction(func(st *database.DataStore) error {
			return st.UserItems().ComputeAllUserItems()
		})
		assert.NoError(t, err)
	}, 30)
}

func TestUserItemStore_CreateIfMissing(t *testing.T) {
	db := testhelpers.SetupDBWithFixture()

	userItemStore := database.NewDataStore(db).UserItems()
	err := userItemStore.CreateIfMissing(12, 34)
	assert.NoError(t, err)

	type userItem struct {
		UserID                    int64  `gorm:"column:idUser"`
		ItemID                    int64  `gorm:"column:idItem"`
		AncestorsComputationState string `gorm:"column:sAncestorsComputationState"`
	}
	var insertedUserItem userItem
	assert.NoError(t,
		userItemStore.Select("idUser, idItem, sAncestorsComputationState").
			Scan(&insertedUserItem).Error())
	assert.Equal(t, userItem{
		UserID:                    12,
		ItemID:                    34,
		AncestorsComputationState: "todo",
	}, insertedUserItem)
}
