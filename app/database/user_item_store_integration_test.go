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
	defer func() { _ = db.Close() }()

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

func TestUserItemStore_PropagateAttempts(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		users_items:
			- {ID: 111, idItem: 1, idUser: 500, sAncestorsComputationState: done}
			- {ID: 112, idItem: 2, idUser: 500, sAncestorsComputationState: done}
			- {ID: 113, idItem: 1, idUser: 501, sAncestorsComputationState: done}
		groups_attempts:
			- {ID: 222, idItem: 1, idGroup: 100, iScore: 10.0, bValidated: 1, sAncestorsComputationState: todo}
			- {ID: 223, idItem: 1, idGroup: 101, iScore: 20.0, bValidated: 0, sAncestorsComputationState: todo}
			- {ID: 224, idItem: 1, idGroup: 102, iScore: 30.0, bValidated: 0, sAncestorsComputationState: done}
		groups_groups:
			- {ID: 333, idGroupParent: 100, idGroupChild: 200, sType: direct}
			- {ID: 334, idGroupParent: 101, idGroupChild: 200, sType: invitationAccepted}
			- {ID: 335, idGroupParent: 102, idGroupChild: 200, sType: requestAccepted}
		users:
			- {ID: 500, idGroupSelf: 200}`)
	defer func() { _ = db.Close() }()

	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		err := store.UserItems().PropagateAttempts()
		assert.NoError(t, err)
		return nil
	}))

	type userItem struct {
		UserID                    int64   `gorm:"column:idUser"`
		ItemID                    int64   `gorm:"column:idItem"`
		Score                     float32 `gorm:"column:iScore"`
		Validated                 bool    `gorm:"column:bValidated"`
		AncestorsComputationState string  `gorm:"column:sAncestorsComputationState"`
	}
	var userItems []userItem
	assert.NoError(t,
		database.NewDataStore(db).UserItems().
			Select("idUser, idItem, iScore, bValidated, sAncestorsComputationState").
			Order("idUser, idItem").Scan(&userItems).Error())
	assert.Equal(t, []userItem{
		{
			UserID:                    500,
			ItemID:                    1,
			Score:                     20.0,
			Validated:                 true,
			AncestorsComputationState: "todo",
		},
		{
			UserID:                    500,
			ItemID:                    2,
			AncestorsComputationState: "done",
		},
		{
			UserID:                    501,
			ItemID:                    1,
			AncestorsComputationState: "done",
		},
	}, userItems)

	type groupAttempt struct {
		ItemID                    int64  `gorm:"column:idItem"`
		GroupID                   int64  `gorm:"column:idGroup"`
		AncestorsComputationState string `gorm:"column:sAncestorsComputationState"`
	}
	var groupAttempts []groupAttempt
	assert.NoError(t,
		database.NewDataStore(db).GroupAttempts().
			Select("idItem, idGroup, sAncestorsComputationState").
			Order("idItem, idGroup").Scan(&groupAttempts).Error())
	assert.Equal(t, []groupAttempt{
		{ItemID: 1, GroupID: 100, AncestorsComputationState: "done"},
		{ItemID: 1, GroupID: 101, AncestorsComputationState: "done"},
		{ItemID: 1, GroupID: 102, AncestorsComputationState: "done"},
	}, groupAttempts)
}

func TestUserItemStore_PropagateAttemps_MustBeInTransaction(t *testing.T) {
	db, dbMock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, database.ErrNoTransaction, func() {
		_ = database.NewDataStore(db).UserItems().PropagateAttempts()
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}
