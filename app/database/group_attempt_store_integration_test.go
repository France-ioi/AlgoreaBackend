// +build !unit

package database_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestGroupAttemptStore_CreateNew(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups_attempts:
			- {ID: 1, idGroup: 10, idItem: 20, iOrder: 1}
			- {ID: 2, idGroup: 10, idItem: 30, iOrder: 3}
			- {ID: 3, idGroup: 20, idItem: 20, iOrder: 4}`)
	defer func() { _ = db.Close() }()

	var newID int64
	var err error
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		newID, err = store.GroupAttempts().CreateNew(10, 20)
		return err
	}))
	assert.True(t, newID > 0)
	type resultType struct {
		GroupID             int64 `gorm:"column:idGroup"`
		ItemID              int64 `gorm:"column:idItem"`
		StartDateSet        bool  `gorm:"column:startDateSet"`
		LastActivityDateSet bool  `gorm:"column:lastActivityDateSet"`
		Order               int32 `gorm:"column:iOrder"`
	}
	var result resultType
	assert.NoError(t, database.NewDataStore(db).GroupAttempts().ByID(newID).
		Select(`
			idGroup, idItem, ABS(sStartDate - NOW()) < 3 AS startDateSet,
			ABS(sLastActivityDate - NOW()) < 3 AS lastActivityDateSet, iOrder`).
		Take(&result).Error())
	assert.Equal(t, resultType{
		GroupID:             10,
		ItemID:              20,
		StartDateSet:        true,
		LastActivityDateSet: true,
		Order:               2,
	}, result)
}

func TestGroupAttemptStore_GetAttemptItemIDIfUserHasAccess(t *testing.T) {
	tests := []struct {
		name           string
		fixture        string
		attemptID      int64
		userID         int64
		expectedFound  bool
		expectedItemID int64
	}{
		{
			name: "okay (full access)",
			fixture: `
				users_items: [{idUser: 11, idItem: 50}]
				groups_attempts: [{ID: 100, idGroup: 111, idItem: 50}]`,
			attemptID:      100,
			userID:         11,
			expectedFound:  true,
			expectedItemID: 50,
		},
		{
			name: "okay (partial access)",
			fixture: `
				users_items: [{idUser: 10, idItem: 50}]
				groups_attempts: [{ID: 100, idGroup: 101, idItem: 50}]`,
			attemptID:      100,
			userID:         10,
			expectedFound:  true,
			expectedItemID: 50,
		},
		{
			name:      "okay (bHasAttempts=1, groups_groups.sType=requestAccepted)",
			userID:    10,
			attemptID: 200,
			fixture: `
				users_items:
					- {idUser: 10, idItem: 60}
				groups_attempts:
					- {ID: 200, idGroup: 102, idItem: 60}`,
			expectedFound:  true,
			expectedItemID: 60,
		},
		{
			name:      "okay (bHasAttempts=1, groups_groups.sType=invitationAccepted)",
			userID:    10,
			attemptID: 200,
			fixture: `
				users_items:
					- {idUser: 10, idItem: 60}
				groups_attempts:
					- {ID: 200, idGroup: 110, idItem: 60}`,
			expectedFound:  true,
			expectedItemID: 60,
		},
		{
			name:          "user not found",
			fixture:       `groups_attempts: [{ID: 100, idGroup: 121, idItem: 50}]`,
			userID:        404,
			attemptID:     100,
			expectedFound: false,
		},
		{
			name:      "user doesn't have access to the item",
			userID:    12,
			attemptID: 100,
			fixture: `
				users_items: [{idUser: 12, idItem: 50}]
				groups_attempts: [{ID: 100, idGroup: 121, idItem: 50}]`,
			expectedFound: false,
		},
		{
			name:          "no groups_attempts",
			userID:        10,
			attemptID:     100,
			fixture:       `users_items: [{idUser: 10, idItem: 50}]`,
			expectedFound: false,
		},
		{
			name:      "wrong item in groups_attempts",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{idUser: 10, idItem: 50}]
				groups_attempts: [{ID: 100, idGroup: 101, idItem: 51}]`,
			expectedFound: false,
		},
		{
			name:      "no users_items",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items:
					- {idUser: 10, idItem: 51}
					- {idUser: 11, idItem: 50}
				groups_attempts: [{ID: 100, idGroup: 101, idItem: 50}]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (invitationSent)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{idUser: 10, idItem: 60}]
				groups_attempts: [ID: 100, idGroup: 103, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (requestSent)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{idUser: 10, idItem: 60}]
				groups_attempts: [ID: 100, idGroup: 104, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (invitationRefused)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{idUser: 10, idItem: 60}]
				groups_attempts: [ID: 100, idGroup: 105, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (requestRefused)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{idUser: 10, idItem: 60}]
				groups_attempts: [ID: 100, idGroup: 106, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (removed)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{idUser: 10, idItem: 60}]
				groups_attempts: [ID: 100, idGroup: 107, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (left)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{idUser: 10, idItem: 60}]
				groups_attempts: [ID: 100, idGroup: 108, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (direct)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{idUser: 10, idItem: 60}]
				groups_attempts: [ID: 100, idGroup: 109, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:      "groups_attempts.idGroup is not user's self group",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{idUser: 10, idItem: 50}]
				groups_attempts: [ID: 100, idGroup: 102, idItem: 50]`,
			expectedFound: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				users:
					- {ID: 10, sLogin: "john", idGroupSelf: 101}
					- {ID: 11, sLogin: "jane", idGroupSelf: 111}
					- {ID: 12, sLogin: "guest", idGroupSelf: 121}
				groups_groups:
					- {idGroupParent: 102, idGroupChild: 101, sType: requestAccepted}
					- {idGroupParent: 103, idGroupChild: 101, sType: invitationSent}
					- {idGroupParent: 104, idGroupChild: 101, sType: requestSent}
					- {idGroupParent: 105, idGroupChild: 101, sType: invitationRefused}
					- {idGroupParent: 106, idGroupChild: 101, sType: requestRefused}
					- {idGroupParent: 107, idGroupChild: 101, sType: removed}
					- {idGroupParent: 108, idGroupChild: 101, sType: left}
					- {idGroupParent: 109, idGroupChild: 101, sType: direct}
					- {idGroupParent: 110, idGroupChild: 101, sType: invitationAccepted}
				groups_ancestors:
					- {idGroupAncestor: 101, idGroupChild: 101, bIsSelf: 1}
					- {idGroupAncestor: 102, idGroupChild: 101, bIsSelf: 0}
					- {idGroupAncestor: 102, idGroupChild: 102, bIsSelf: 1}
					- {idGroupAncestor: 111, idGroupChild: 111, bIsSelf: 1}
					- {idGroupAncestor: 121, idGroupChild: 121, bIsSelf: 1}
					- {idGroupAncestor: 109, idGroupChild: 101, bIsSelf: 0}
					- {idGroupAncestor: 109, idGroupChild: 109, bIsSelf: 1}
				items:
					- {ID: 10, bHasAttempts: 0}
					- {ID: 50, bHasAttempts: 0}
					- {ID: 60, bHasAttempts: 1}
				groups_items:
					- {idGroup: 101, idItem: 50, sCachedPartialAccessDate: "2017-05-29T06:38:38Z"}
					- {idGroup: 101, idItem: 60, sCachedPartialAccessDate: "2017-05-29T06:38:38Z"}
					- {idGroup: 111, idItem: 50, sCachedFullAccessDate: "2017-05-29T06:38:38Z"}
					- {idGroup: 121, idItem: 50, sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}`,
				test.fixture)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			user := database.NewUser(test.userID, store.Users(), nil)
			found, itemID, err := store.GroupAttempts().GetAttemptItemIDIfUserHasAccess(test.attemptID, user)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedFound, found)
			assert.Equal(t, test.expectedItemID, itemID)
		})
	}
}

func TestGroupAttemptStore_ByUserAndItemID(t *testing.T) {
	tests := []struct {
		name        string
		fixture     string
		attemptID   int64
		userID      int64
		itemID      int64
		expectedIDs []int64
		expectedErr error
	}{
		{
			name:        "okay (full access)",
			fixture:     `groups_attempts: [{ID: 100, idGroup: 111, idItem: 50},{ID: 101, idGroup: 111, idItem: 50}]`,
			attemptID:   100,
			userID:      11,
			expectedIDs: []int64{100, 101},
			itemID:      50,
		},
		{
			name:        "okay (partial access)",
			fixture:     `groups_attempts: [{ID: 100, idGroup: 101, idItem: 50}]`,
			attemptID:   100,
			userID:      10,
			expectedIDs: []int64{100},
			itemID:      50,
		},
		{
			name:        "okay (bHasAttempts=1, groups_groups.sType=requestAccepted)",
			userID:      10,
			attemptID:   200,
			fixture:     `groups_attempts: [{ID: 200, idGroup: 102, idItem: 60},{ID: 201, idGroup: 102, idItem: 60}]`,
			expectedIDs: []int64{200, 201},
			itemID:      60,
		},
		{
			name:        "okay (bHasAttempts=1, groups_groups.sType=invitationAccepted)",
			userID:      10,
			attemptID:   200,
			fixture:     `groups_attempts: [{ID: 200, idGroup: 110, idItem: 60}]`,
			expectedIDs: []int64{200},
			itemID:      60,
		},
		{
			name:        "user not found",
			fixture:     `groups_attempts: [{ID: 100, idGroup: 121, idItem: 50}]`,
			userID:      404,
			attemptID:   100,
			expectedIDs: []int64(nil),
			expectedErr: gorm.ErrRecordNotFound,
		},
		{
			name:        "user doesn't have access to the item",
			userID:      12,
			attemptID:   100,
			fixture:     `groups_attempts: [{ID: 100, idGroup: 121, idItem: 50}]`,
			expectedIDs: []int64(nil),
		},
		{
			name:        "no groups_attempts",
			userID:      10,
			attemptID:   100,
			fixture:     "",
			expectedIDs: nil,
		},
		{
			name:        "wrong item in groups_attempts",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [{ID: 100, idGroup: 101, idItem: 51}]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (invitationSent)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [ID: 100, idGroup: 103, idItem: 60]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (requestSent)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [ID: 100, idGroup: 104, idItem: 60]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (invitationRefused)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [ID: 100, idGroup: 105, idItem: 60]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (requestRefused)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [ID: 100, idGroup: 106, idItem: 60]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (removed)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [ID: 100, idGroup: 107, idItem: 60]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (left)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [ID: 100, idGroup: 108, idItem: 60]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (direct)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [ID: 100, idGroup: 109, idItem: 60]`,
			expectedIDs: nil,
		},
		{
			name:        "groups_attempts.idGroup is not user's self group",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [ID: 100, idGroup: 102, idItem: 50]`,
			expectedIDs: nil,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				users:
					- {ID: 10, sLogin: "john", idGroupSelf: 101}
					- {ID: 11, sLogin: "jane", idGroupSelf: 111}
					- {ID: 12, sLogin: "guest", idGroupSelf: 121}
				groups_groups:
					- {idGroupParent: 102, idGroupChild: 101, sType: requestAccepted}
					- {idGroupParent: 103, idGroupChild: 101, sType: invitationSent}
					- {idGroupParent: 104, idGroupChild: 101, sType: requestSent}
					- {idGroupParent: 105, idGroupChild: 101, sType: invitationRefused}
					- {idGroupParent: 106, idGroupChild: 101, sType: requestRefused}
					- {idGroupParent: 107, idGroupChild: 101, sType: removed}
					- {idGroupParent: 108, idGroupChild: 101, sType: left}
					- {idGroupParent: 109, idGroupChild: 101, sType: direct}
					- {idGroupParent: 110, idGroupChild: 101, sType: invitationAccepted}
				groups_ancestors:
					- {idGroupAncestor: 101, idGroupChild: 101, bIsSelf: 1}
					- {idGroupAncestor: 102, idGroupChild: 101, bIsSelf: 0}
					- {idGroupAncestor: 102, idGroupChild: 102, bIsSelf: 1}
					- {idGroupAncestor: 111, idGroupChild: 111, bIsSelf: 1}
					- {idGroupAncestor: 121, idGroupChild: 121, bIsSelf: 1}
					- {idGroupAncestor: 109, idGroupChild: 101, bIsSelf: 0}
					- {idGroupAncestor: 109, idGroupChild: 109, bIsSelf: 1}
				items:
					- {ID: 10, bHasAttempts: 0}
					- {ID: 50, bHasAttempts: 0}
					- {ID: 60, bHasAttempts: 1}
				groups_items:
					- {idGroup: 101, idItem: 50, sCachedPartialAccessDate: "2017-05-29T06:38:38Z"}
					- {idGroup: 101, idItem: 60, sCachedPartialAccessDate: "2017-05-29T06:38:38Z"}
					- {idGroup: 111, idItem: 50, sCachedFullAccessDate: "2017-05-29T06:38:38Z"}
					- {idGroup: 121, idItem: 50, sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}`,
				test.fixture)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			user := database.NewUser(test.userID, store.Users(), nil)
			var ids []int64
			err := store.GroupAttempts().ByUserAndItemID(user, test.itemID).Pluck("groups_attempts.ID", &ids).Error()
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedIDs, ids)
		})
	}
}
