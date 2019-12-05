// +build !unit

package database_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func setupDB() *database.DB {
	return testhelpers.SetupDBWithFixture("visibility")
}

func TestItemStore_VisibleMethods(t *testing.T) {
	tests := []struct {
		methodToCall string
		args         []interface{}
		column       string
		expected     []int64
	}{
		{methodToCall: "Visible", column: "id", expected: []int64{190, 191, 192, 1900, 1901, 1902, 19000, 19001, 19002}},
		{methodToCall: "VisibleByID", args: []interface{}{int64(191)}, column: "id", expected: []int64{191}},
		{methodToCall: "VisibleChildrenOfID", args: []interface{}{int64(190)}, column: "items.id", expected: []int64{1900, 1901, 1902}},
		{methodToCall: "VisibleGrandChildrenOfID", args: []interface{}{int64(190)}, column: "items.id", expected: []int64{19000, 19001, 19002}},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.methodToCall, func(t *testing.T) {
			db := setupDB()
			defer func() { _ = db.Close() }()

			user := &database.User{GroupID: 11, DefaultLanguageID: 2}
			dataStore := database.NewDataStore(db)
			itemStore := dataStore.Items()

			var result []int64
			parameters := make([]reflect.Value, 0, len(testCase.args)+1)
			parameters = append(parameters, reflect.ValueOf(user))
			for _, arg := range testCase.args {
				parameters = append(parameters, reflect.ValueOf(arg))
			}
			db = reflect.ValueOf(itemStore).MethodByName(testCase.methodToCall).
				Call(parameters)[0].Interface().(*database.DB).Pluck(testCase.column, &result)
			assert.NoError(t, db.Error())

			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestItemStore_CheckSubmissionRights(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("item_store/check_submission_rights")
	defer func() { _ = db.Close() }()
	user := &database.User{GroupID: 10}

	tests := []struct {
		name          string
		itemID        int64
		wantHasAccess bool
		wantReason    error
		wantError     error
	}{
		{name: "normal", itemID: 13, wantHasAccess: true, wantReason: nil, wantError: nil},
		{name: "read-only", itemID: 12, wantHasAccess: false, wantReason: errors.New("item is read-only"), wantError: nil},
		{name: "no access", itemID: 10, wantHasAccess: false, wantReason: errors.New("no access to the task item"), wantError: nil},
		{name: "info access", itemID: 10, wantHasAccess: false, wantReason: errors.New("no access to the task item"), wantError: nil},
		{name: "finished time-limited", itemID: 14, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished"), wantError: nil},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			err := database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				hasAccess, reason, err := store.Items().CheckSubmissionRights(test.itemID, user)
				assert.Equal(t, test.wantHasAccess, hasAccess)
				assert.Equal(t, test.wantReason, reason)
				assert.Equal(t, test.wantError, err)
				return nil
			})
			assert.NoError(t, err)
		})
	}
}

func TestItemStore_CheckSubmissionRightsForTimeLimitedContest(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("item_store/check_submission_rights_for_time_limited_contest")
	defer func() { _ = db.Close() }()

	tests := []struct {
		name          string
		itemID        int64
		userID        int64
		wantHasAccess bool
		wantReason    error
		initFunc      func(*database.DB) error
	}{
		{name: "no items", itemID: 404, userID: 11, wantHasAccess: true, wantReason: nil},
		{name: "user has no active contest", itemID: 14, userID: 11, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active team contest has expired", itemID: 14, userID: 12, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active team contest has expired (again)", itemID: 14, userID: 12, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active contest has expired", itemID: 15, userID: 13, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active contest has expired (again)", itemID: 15, userID: 13, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active contest is OK and it is from another competition, but the user has full access to the time-limited chapter",
			initFunc: func(db *database.DB) error {
				if err := database.NewDataStore(db).GroupGroups().InsertMap(
					map[string]interface{}{
						"parent_group_id": 200, // contest participants group
						"child_group_id":  14,
						"child_order":     1,
					}); err != nil {
					return err
				}
				return database.NewDataStore(db).GroupAttempts().InsertMap(
					map[string]interface{}{
						"item_id":    500, // chapter
						"group_id":   14,
						"entered_at": database.Now(),
						"order":      1,
					})
			},
			itemID: 15, userID: 14, wantHasAccess: true, wantReason: nil},
		{name: "user's active contest is OK and it is the task's time-limited chapter",
			initFunc: func(db *database.DB) error {
				if err := database.NewDataStore(db).GroupGroups().InsertMap(
					map[string]interface{}{
						"parent_group_id": 100, // contest participants group
						"child_group_id":  15,
						"child_order":     1,
					}); err != nil {
					return err
				}
				return database.NewDataStore(db).GroupAttempts().
					InsertMap(map[string]interface{}{
						"item_id":    115,
						"group_id":   15,
						"entered_at": database.Now(),
						"order":      1,
					})
			},
			itemID: 15, userID: 15, wantHasAccess: true, wantReason: nil},
		{name: "user's active contest is OK, but it is not an ancestor of the task and the user doesn't have full access to the task's chapter",
			initFunc: func(db *database.DB) error {
				if err := database.NewDataStore(db).GroupGroups().InsertMap(
					map[string]interface{}{
						"parent_group_id": 300, // contest participants group
						"child_group_id":  17,
						"child_order":     1,
					}); err != nil {
					return err
				}
				return database.NewDataStore(db).GroupAttempts().
					InsertMap(map[string]interface{}{
						"item_id":    114,
						"group_id":   17,
						"entered_at": database.Now(),
						"order":      1,
					})
			},
			itemID: 15, userID: 17, wantHasAccess: false,
			wantReason: errors.New("the exercise for which you wish to submit an answer is a part " +
				"of a different competition than the one in progress")},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var err error
			if test.initFunc != nil {
				err = test.initFunc(db)
				if err != nil {
					t.Error(err)
					return
				}
			}
			err = database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				user := &database.User{}
				assert.NoError(t, user.LoadByID(store, test.userID))

				hasAccess, reason := store.Items().CheckSubmissionRightsForTimeLimitedContest(test.itemID, user)
				assert.Equal(t, test.wantHasAccess, hasAccess)
				assert.Equal(t, test.wantReason, reason)
				return nil
			})
			assert.NoError(t, err)
		})
	}
}

func TestItemStore_GetActiveContestInfoForUser(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 101}, {id: 102}, {id: 103}, {id: 104}, {id: 105}, {id: 106},
		         {id: 200}, {id: 300}, {id: 400}, {id: 500}]
		users:
			- {login: 1, group_id: 101}
			- {login: 2, group_id: 102}
			- {login: 3, group_id: 103}
			- {login: 4, group_id: 104}
			- {login: 5, group_id: 105}
			- {login: 6, group_id: 106}
		items:
			- {id: 12, contest_participants_group_id: 200}
			- {id: 13, contest_participants_group_id: 300}
			- {id: 14, duration: 10:00:00, contest_participants_group_id: 400}
			- {id: 15, contest_participants_group_id: 500}
		groups_ancestors:
			- {ancestor_group_id: 101, child_group_id: 101}
			- {ancestor_group_id: 102, child_group_id: 102}
			- {ancestor_group_id: 103, child_group_id: 103}
			- {ancestor_group_id: 104, child_group_id: 104}
			- {ancestor_group_id: 105, child_group_id: 105}
			- {ancestor_group_id: 106, child_group_id: 106}
		groups_contest_items:
			- {group_id: 102, item_id: 12} # not started
			- {group_id: 104, item_id: 14, additional_time: 00:01:00} # ok
			- {group_id: 105, item_id: 15}  # ok with team mode
			- {group_id: 106, item_id: 14, additional_time: 00:01:00} # multiple
			- {group_id: 106, item_id: 15, additional_time: 00:01:00} # multiple
		groups_attempts:
			- {group_id: 103, item_id: 13, entered_at: 2019-03-22 08:44:55, order: 1} # finished
			- {group_id: 104, item_id: 14, entered_at: 2019-03-22 08:44:55, order: 1} # ok
			- {group_id: 105, item_id: 15, entered_at: 2019-04-22 08:44:55, order: 1}  # ok with team mode
			- {group_id: 106, item_id: 14, entered_at: 2019-03-22 08:44:55, order: 1} # multiple
			- {group_id: 106, item_id: 15, entered_at: 2019-03-22 08:43:55, order: 1} # multiple
		groups_groups:
			- {parent_group_id: 300, child_group_id: 103, expires_at: 2019-03-22 09:44:55}
			- {parent_group_id: 400, child_group_id: 104}
			- {parent_group_id: 500, child_group_id: 105}
			- {parent_group_id: 400, child_group_id: 106}
			- {parent_group_id: 500, child_group_id: 106}`)
	defer func() { _ = db.Close() }()

	tests := []struct {
		name   string
		userID int64
		want   *int64
	}{
		{name: "no item", userID: 101, want: nil},
		{name: "not started", userID: 102, want: nil},
		{name: "finished", userID: 103, want: nil},
		{name: "ok", userID: 104, want: ptrInt64(14)},
		{name: "ok with team mode", userID: 105, want: ptrInt64(15)},
		{name: "ok with multiple active contests", userID: 106, want: ptrInt64(14)},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			store := database.NewDataStore(db)
			user := &database.User{}
			assert.NoError(t, user.LoadByID(store, test.userID))

			got := store.Items().GetActiveContestItemIDForUser(user)
			assert.Equal(t, test.want, got)
		})
	}
}

type itemsTest struct {
	name       string
	ids        []int64
	userID     int64
	wantResult bool
}

func TestItemStore_CanGrantViewContentOnAll(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		items: [{id: 11}, {id: 12}, {id: 13}]
		groups: [{id: 10}, {id: 11}, {id: 40}, {id: 100}, {id: 110}, {id: 400}]
		users: [{login: 1, group_id: 100}, {login: 2, group_id: 110}]
		groups_groups:
			- {parent_group_id: 400, child_group_id: 100}
		groups_ancestors:
			- {ancestor_group_id: 100, child_group_id: 100}
			- {ancestor_group_id: 110, child_group_id: 110}
			- {ancestor_group_id: 400, child_group_id: 100}
			- {ancestor_group_id: 400, child_group_id: 400}
		permissions_generated:
			- {group_id: 400, item_id: 11, can_grant_view_generated: content}
			- {group_id: 100, item_id: 11, can_grant_view_generated: transfer}
			- {group_id: 100, item_id: 12}
			- {group_id: 100, item_id: 13}
			- {group_id: 110, item_id: 12, can_grant_view_generated: transfer}
			- {group_id: 110, item_id: 13, can_grant_view_generated: content}`)

	tests := []itemsTest{
		{name: "two permissions_granted rows for one item", ids: []int64{11}, userID: 100, wantResult: true},
		{name: "cannot grant view", ids: []int64{12}, userID: 100, wantResult: false},
		{name: "can grant view for a part of items", ids: []int64{11, 12}, userID: 100, wantResult: false},
		{name: "another user cannot grant view", ids: []int64{11}, userID: 110, wantResult: false},
		{name: "can_grant_view_generated = transfer", ids: []int64{12}, userID: 110, wantResult: true},
		{name: "can_grant_view_generated = content", ids: []int64{13}, userID: 110, wantResult: true},
		{name: "two items", ids: []int64{12, 13}, userID: 110, wantResult: true},
		{name: "two items (not unique)", ids: []int64{12, 13, 12, 13}, userID: 110, wantResult: true},
		{name: "empty ids list", ids: []int64{}, userID: 110, wantResult: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				user := &database.User{}
				assert.NoError(t, user.LoadByID(store, test.userID))
				canGrant, err := store.Items().CanGrantViewContentOnAll(user, test.ids...)
				assert.NoError(t, err)
				assert.Equal(t, test.wantResult, canGrant)
				return nil
			}))
		})
	}
}

func TestItemStore_AreAllVisible(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		items: [{id: 11}, {id: 12}, {id: 13}]
		groups: [{id: 10}, {id: 11}, {id: 40}, {id: 100}, {id: 110}, {id: 400}]
		users: [{login: 1, group_id: 100}, {login: 2, group_id: 110}]
		groups_groups:
			- {parent_group_id: 400, child_group_id: 100}
		groups_ancestors:
			- {ancestor_group_id: 100, child_group_id: 100}
			- {ancestor_group_id: 110, child_group_id: 110}
			- {ancestor_group_id: 400, child_group_id: 100}
			- {ancestor_group_id: 400, child_group_id: 400}
		permissions_generated:
			- {group_id: 400, item_id: 11, can_view_generated: info}
			- {group_id: 100, item_id: 11, can_view_generated: content}
			- {group_id: 100, item_id: 12}
			- {group_id: 100, item_id: 13}
			- {group_id: 110, item_id: 12, can_view_generated: content_with_descendants}
			- {group_id: 110, item_id: 13, can_view_generated: solution}`)

	tests := []itemsTest{
		{name: "two permissions_granted rows for one item", ids: []int64{11}, userID: 100, wantResult: true},
		{name: "not visible", ids: []int64{12}, userID: 100, wantResult: false},
		{name: "one of two items is not visible", ids: []int64{11, 12}, userID: 100, wantResult: false},
		{name: "not visible for another user", ids: []int64{11}, userID: 110, wantResult: false},
		{name: "can_view_generated = content_with_descendants", ids: []int64{12}, userID: 110, wantResult: true},
		{name: "can_view_generated = solution", ids: []int64{13}, userID: 110, wantResult: true},
		{name: "empty ids list", ids: []int64{}, userID: 110, wantResult: true},
		{name: "two items", ids: []int64{12, 13}, userID: 110, wantResult: true},
		{name: "two items (not unique)", ids: []int64{12, 13, 12, 13}, userID: 110, wantResult: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				user := &database.User{}
				assert.NoError(t, user.LoadByID(store, test.userID))
				allAreVisible, err := store.Items().AreAllVisible(user, test.ids...)
				assert.Equal(t, test.wantResult, allAreVisible)
				assert.NoError(t, err)
				return nil
			}))
		})
	}
}

func TestItemStore_GetAccessDetailsForIDs(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		items: [{id: 11}, {id: 12}, {id: 13}]
		groups: [{id: 10}, {id: 11}, {id: 40}, {id: 100}, {id: 110}, {id: 400}]
		users: [{login: 1, group_id: 100}, {login: 2, group_id: 110}]
		groups_groups:
			- {parent_group_id: 400, child_group_id: 100}
		groups_ancestors:
			- {ancestor_group_id: 100, child_group_id: 100}
			- {ancestor_group_id: 110, child_group_id: 110}
			- {ancestor_group_id: 400, child_group_id: 100}
			- {ancestor_group_id: 400, child_group_id: 400}
		permissions_generated:
			- {group_id: 400, item_id: 11, can_view_generated: info}
			- {group_id: 100, item_id: 11, can_view_generated: content}
			- {group_id: 100, item_id: 12}
			- {group_id: 100, item_id: 13}
			- {group_id: 110, item_id: 12, can_view_generated: content_with_descendants}
			- {group_id: 110, item_id: 13, can_view_generated: solution}`)

	tests := []struct {
		name       string
		ids        []int64
		userID     int64
		wantResult []database.ItemAccessDetailsWithID
	}{
		{name: "two permissions_granted rows for one item", ids: []int64{11}, userID: 100,
			wantResult: []database.ItemAccessDetailsWithID{{
				ItemID: 11, ItemAccessDetails: database.ItemAccessDetails{CanView: "content"},
			}}},
		{name: "not visible", ids: []int64{12}, userID: 100,
			wantResult: []database.ItemAccessDetailsWithID{}},
		{name: "one of two items is not visible", ids: []int64{11, 12}, userID: 100,
			wantResult: []database.ItemAccessDetailsWithID{
				{ItemID: 11, ItemAccessDetails: database.ItemAccessDetails{CanView: "content"}},
			}},
		{name: "no permissions_generated row", ids: []int64{11}, userID: 110, wantResult: []database.ItemAccessDetailsWithID{}},
		{name: "can_view_generated = content_with_descendants", ids: []int64{12}, userID: 110,
			wantResult: []database.ItemAccessDetailsWithID{{
				ItemID: 12, ItemAccessDetails: database.ItemAccessDetails{CanView: "content_with_descendants"},
			}}},
		{name: "can_view_generated = solution", ids: []int64{13}, userID: 110,
			wantResult: []database.ItemAccessDetailsWithID{{
				ItemID: 13, ItemAccessDetails: database.ItemAccessDetails{CanView: "solution"},
			}}},
		{name: "empty ids list", ids: []int64{}, userID: 110, wantResult: []database.ItemAccessDetailsWithID{}},
		{name: "two items", ids: []int64{12, 13}, userID: 110,
			wantResult: []database.ItemAccessDetailsWithID{
				{ItemID: 12, ItemAccessDetails: database.ItemAccessDetails{CanView: "content_with_descendants"}},
				{ItemID: 13, ItemAccessDetails: database.ItemAccessDetails{CanView: "solution"}},
			}},
		{name: "two items (not unique)", ids: []int64{12, 13, 12, 13}, userID: 110,
			wantResult: []database.ItemAccessDetailsWithID{
				{ItemID: 12, ItemAccessDetails: database.ItemAccessDetails{CanView: "content_with_descendants"}},
				{ItemID: 13, ItemAccessDetails: database.ItemAccessDetails{CanView: "solution"}},
			}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			store := database.NewDataStore(db)
			user := &database.User{}
			assert.NoError(t, user.LoadByID(store, test.userID))
			accessDetails, err := store.Items().GetAccessDetailsForIDs(user, test.ids)
			assert.Equal(t, test.wantResult, accessDetails)
			assert.NoError(t, err)
		})
	}
}
