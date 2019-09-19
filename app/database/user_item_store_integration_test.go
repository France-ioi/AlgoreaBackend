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
		UserID                    int64
		ItemID                    int64
		AncestorsComputationState string
	}
	var insertedUserItem userItem
	assert.NoError(t,
		userItemStore.Select("user_id, item_id, ancestors_computation_state").
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
			- {id: 111, item_id: 1, user_id: 500, ancestors_computation_state: done}
			- {id: 112, item_id: 2, user_id: 500, ancestors_computation_state: done}
			- {id: 113, item_id: 1, user_id: 501, ancestors_computation_state: done}
		groups_attempts:
			- {id: 222, item_id: 1, group_id: 100, score: 10.0, validated: 1, ancestors_computation_state: todo, order: 0}
			- {id: 223, item_id: 1, group_id: 101, score: 20.0, validated: 0, ancestors_computation_state: todo, order: 0}
			- {id: 224, item_id: 1, group_id: 102, score: 30.0, validated: 0, ancestors_computation_state: done, order: 0}
			- {id: 225, item_id: 1, group_id: 103, score: 40.0, validated: 0, ancestors_computation_state: done, order: 0}
		groups_groups:
			- {id: 333, group_parent_id: 100, group_child_id: 200, type: direct}
			- {id: 334, group_parent_id: 101, group_child_id: 200, type: invitationAccepted}
			- {id: 335, group_parent_id: 102, group_child_id: 200, type: requestAccepted}
			- {id: 336, group_parent_id: 103, group_child_id: 200, type: joinedByCode}
		users:
			- {id: 500, group_self_id: 200}`)
	defer func() { _ = db.Close() }()

	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		err := store.UserItems().PropagateAttempts()
		assert.NoError(t, err)
		return nil
	}))

	type userItem struct {
		UserID                    int64
		ItemID                    int64
		Score                     float32
		Validated                 bool
		AncestorsComputationState string
	}
	var userItems []userItem
	assert.NoError(t,
		database.NewDataStore(db).UserItems().
			Select("user_id, item_id, score, validated, ancestors_computation_state").
			Order("user_id, item_id").Scan(&userItems).Error())
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
		ItemID                    int64
		GroupID                   int64
		AncestorsComputationState string
	}
	var groupAttempts []groupAttempt
	assert.NoError(t,
		database.NewDataStore(db).GroupAttempts().
			Select("item_id, group_id, ancestors_computation_state").
			Order("item_id, group_id").Scan(&groupAttempts).Error())
	assert.Equal(t, []groupAttempt{
		{ItemID: 1, GroupID: 100, AncestorsComputationState: "done"},
		{ItemID: 1, GroupID: 101, AncestorsComputationState: "done"},
		{ItemID: 1, GroupID: 102, AncestorsComputationState: "done"},
		{ItemID: 1, GroupID: 103, AncestorsComputationState: "done"},
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
