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

			groupID := int64(11)
			dataStore := database.NewDataStore(db)
			itemStore := dataStore.Items()

			var result []int64
			parameters := make([]reflect.Value, 0, len(testCase.args)+1)
			parameters = append(parameters, reflect.ValueOf(groupID))
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

func TestItemStore_IsValidHierarchy(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		items:
			- {id: 1, default_language_tag: fr}
			- {id: 2, default_language_tag: fr, is_root: 1}
			- {id: 3, default_language_tag: fr}
			- {id: 4, default_language_tag: fr}
			- {id: 5, default_language_tag: fr}
			- {id: 6, default_language_tag: fr}
			- {id: 7, default_language_tag: fr}
			- {id: 8, default_language_tag: fr}
		items_items:
			- {parent_item_id: 2, child_item_id: 4, child_order: 1}
			- {parent_item_id: 4, child_item_id: 6, child_order: 1}
			- {parent_item_id: 6, child_item_id: 8, child_order: 1}`)
	defer func() { _ = db.Close() }()

	tests := []struct {
		name     string
		ids      []int64
		expected bool
	}{
		{name: "empty list", ids: []int64{}, expected: false},
		{name: "the first item does not exist", ids: []int64{404}, expected: false},
		{name: "the first item is not a root item", ids: []int64{1}, expected: false},
		{name: "only the root item", ids: []int64{2}, expected: true},
		{name: "the second item is not a child of the root item", ids: []int64{2, 3}, expected: false},
		{name: "the third item is not a child of the second item", ids: []int64{2, 4, 5}, expected: false},
		{name: "the fourth item is not a child of the third item", ids: []int64{2, 4, 6, 7}, expected: false},
		{name: "the correct hierarchy of two items", ids: []int64{2, 4}, expected: true},
		{name: "the correct hierarchy of three items", ids: []int64{2, 4, 6}, expected: true},
		{name: "the correct hierarchy of four items", ids: []int64{2, 4, 6, 8}, expected: true},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			itemStore := database.NewDataStore(db).Items()

			valid, err := itemStore.IsValidHierarchy(testCase.ids)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, valid)
		})
	}
}

func TestItemStore_CheckSubmissionRights(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("item_store/check_submission_rights")
	defer func() { _ = db.Close() }()

	tests := []struct {
		name          string
		participantID int64
		attemptID     int64
		itemID        int64
		wantHasAccess bool
		wantReason    error
		wantError     error
	}{
		{name: "normal", participantID: 10, attemptID: 1, itemID: 13, wantHasAccess: true, wantReason: nil, wantError: nil},
		{name: "read-only", participantID: 10, attemptID: 2, itemID: 12, wantHasAccess: false,
			wantReason: errors.New("item is read-only"), wantError: nil},
		{name: "no access", participantID: 11, attemptID: 1, itemID: 10, wantHasAccess: false,
			wantReason: errors.New("no access to the task item"), wantError: nil},
		{name: "info access", participantID: 11, attemptID: 2, itemID: 10, wantHasAccess: false,
			wantReason: errors.New("no access to the task item"), wantError: nil},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				hasAccess, reason, err := store.Items().CheckSubmissionRights(test.participantID, test.itemID)
				assert.Equal(t, test.wantHasAccess, hasAccess)
				assert.Equal(t, test.wantReason, reason)
				assert.Equal(t, test.wantError, err)
				assert.NoError(t, err)
				return nil
			}))
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
		items: [{id: 11, default_language_tag: fr}, {id: 12, default_language_tag: fr}, {id: 13, default_language_tag: fr}]
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
			- {group_id: 100, item_id: 11, can_grant_view_generated: solution_with_grant}
			- {group_id: 100, item_id: 12}
			- {group_id: 100, item_id: 13}
			- {group_id: 110, item_id: 12, can_grant_view_generated: solution_with_grant}
			- {group_id: 110, item_id: 13, can_grant_view_generated: content}`)
	defer func() { _ = db.Close() }()

	tests := []itemsTest{
		{name: "two permissions_granted rows for one item", ids: []int64{11}, userID: 100, wantResult: true},
		{name: "cannot grant view", ids: []int64{12}, userID: 100, wantResult: false},
		{name: "can grant view for a part of items", ids: []int64{11, 12}, userID: 100, wantResult: false},
		{name: "another user cannot grant view", ids: []int64{11}, userID: 110, wantResult: false},
		{name: "can_grant_view_generated = solution_with_grant", ids: []int64{12}, userID: 110, wantResult: true},
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
		items: [{id: 11, default_language_tag: fr}, {id: 12, default_language_tag: fr}, {id: 13, default_language_tag: fr}]
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
	defer func() { _ = db.Close() }()

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

func TestItemStore_IsValidParticipationHierarchy(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		items:
			- {id: 1, default_language_tag: fr}
			- {id: 2, default_language_tag: fr, is_root: 1}
			- {id: 3, default_language_tag: fr}
			- {id: 4, default_language_tag: fr}
			- {id: 5, default_language_tag: fr}
			- {id: 6, default_language_tag: fr}
			- {id: 7, default_language_tag: fr}
			- {id: 8, default_language_tag: fr}
		items_items:
			- {parent_item_id: 2, child_item_id: 4, child_order: 1}
			- {parent_item_id: 4, child_item_id: 6, child_order: 1}
			- {parent_item_id: 6, child_item_id: 8, child_order: 1}
		groups:
			- {id: 50, activity_id: 4}
			- {id: 100}
			- {id: 101, activity_id: 4}
			- {id: 102}
			- {id: 103}
			- {id: 104}
			- {id: 105}
			- {id: 106}
			- {id: 107}
			- {id: 108}
			- {id: 109}
			- {id: 110}
			- {id: 111}
			- {id: 112}
			- {id: 113}
			- {id: 114}
			- {id: 115}
			- {id: 116}
			- {id: 117}
			- {id: 118}
		groups_groups:
			- {parent_group_id: 50, child_group_id: 102}
		permissions_generated:
			- {group_id: 50, item_id: 4, can_view_generated: content}
			- {group_id: 100, item_id: 2, can_view_generated: content}
			- {group_id: 100, item_id: 4, can_view_generated: content}
			- {group_id: 101, item_id: 4, can_view_generated: content}
			- {group_id: 101, item_id: 6, can_view_generated: content}
			- {group_id: 101, item_id: 8, can_view_generated: content}
			- {group_id: 102, item_id: 6, can_view_generated: content}
			- {group_id: 103, item_id: 2, can_view_generated: content}
			- {group_id: 103, item_id: 4, can_view_generated: info}
			- {group_id: 104, item_id: 2, can_view_generated: content}
			- {group_id: 104, item_id: 4, can_view_generated: none}
			- {group_id: 105, item_id: 2, can_view_generated: content}
			- {group_id: 105, item_id: 4, can_view_generated: info}
			- {group_id: 105, item_id: 6, can_view_generated: content}
			- {group_id: 106, item_id: 2, can_view_generated: info}
			- {group_id: 106, item_id: 4, can_view_generated: content}
			- {group_id: 106, item_id: 6, can_view_generated: content}
			- {group_id: 107, item_id: 2, can_view_generated: content}
			- {group_id: 107, item_id: 4, can_view_generated: content}
			- {group_id: 107, item_id: 6, can_view_generated: content}
			- {group_id: 108, item_id: 2, can_view_generated: content}
			- {group_id: 108, item_id: 4, can_view_generated: content}
			- {group_id: 108, item_id: 6, can_view_generated: content}
			- {group_id: 109, item_id: 2, can_view_generated: content}
			- {group_id: 109, item_id: 4, can_view_generated: content}
			- {group_id: 109, item_id: 6, can_view_generated: content}
			- {group_id: 110, item_id: 2, can_view_generated: content}
			- {group_id: 110, item_id: 4, can_view_generated: content}
			- {group_id: 110, item_id: 6, can_view_generated: content}
			- {group_id: 111, item_id: 2, can_view_generated: content}
			- {group_id: 111, item_id: 4, can_view_generated: content}
			- {group_id: 111, item_id: 6, can_view_generated: content}
			- {group_id: 112, item_id: 2, can_view_generated: content}
			- {group_id: 112, item_id: 4, can_view_generated: content}
			- {group_id: 112, item_id: 6, can_view_generated: content}
			- {group_id: 113, item_id: 2, can_view_generated: content}
			- {group_id: 113, item_id: 4, can_view_generated: content}
			- {group_id: 113, item_id: 6, can_view_generated: content}
			- {group_id: 114, item_id: 2, can_view_generated: content}
			- {group_id: 114, item_id: 4, can_view_generated: content}
			- {group_id: 114, item_id: 6, can_view_generated: content}
			- {group_id: 114, item_id: 8, can_view_generated: content}
			- {group_id: 115, item_id: 2, can_view_generated: content}
			- {group_id: 115, item_id: 4, can_view_generated: content}
			- {group_id: 115, item_id: 6, can_view_generated: content}
			- {group_id: 115, item_id: 8, can_view_generated: content}
			- {group_id: 116, item_id: 2, can_view_generated: content}
			- {group_id: 116, item_id: 4, can_view_generated: content}
			- {group_id: 116, item_id: 6, can_view_generated: content}
			- {group_id: 116, item_id: 8, can_view_generated: content}
			- {group_id: 117, item_id: 2, can_view_generated: content}
			- {group_id: 117, item_id: 4, can_view_generated: content}
			- {group_id: 117, item_id: 6, can_view_generated: content}
			- {group_id: 117, item_id: 8, can_view_generated: content}
			- {group_id: 118, item_id: 2, can_view_generated: content}
			- {group_id: 118, item_id: 4, can_view_generated: content}
			- {group_id: 118, item_id: 6, can_view_generated: content}
			- {group_id: 118, item_id: 8, can_view_generated: content}
		attempts:
			- {participant_id: 100, id: 0}
			- {participant_id: 100, id: 200, root_item_id: 2, parent_attempt_id: 0}
			- {participant_id: 101, id: 200, root_item_id: 4, allows_submissions_until: 3019-06-30 11:00:00}
			- {participant_id: 101, id: 201, root_item_id: 6, parent_attempt_id: 200, allows_submissions_until: 3019-06-30 11:00:00}
			- {participant_id: 102, id: 200, root_item_id: 4}
			- {participant_id: 103, id: 200, root_item_id: 2}
			- {participant_id: 104, id: 200, root_item_id: 2}
			- {participant_id: 105, id: 200, root_item_id: 2}
			- {participant_id: 105, id: 201, root_item_id: 4, parent_attempt_id: 200}
			- {participant_id: 106, id: 200, root_item_id: 2}
			- {participant_id: 106, id: 201, root_item_id: 4, parent_attempt_id: 200}
			- {participant_id: 107, id: 200, root_item_id: 2}
			- {participant_id: 107, id: 201, root_item_id: 4, parent_attempt_id: 200}
			- {participant_id: 108, id: 200, root_item_id: 2}
			- {participant_id: 108, id: 201, root_item_id: 4, parent_attempt_id: 200}
			- {participant_id: 109, id: 200, root_item_id: 2, allows_submissions_until: 2019-06-30 11:00:00}
			- {participant_id: 109, id: 201, root_item_id: 4, parent_attempt_id: 200}
			- {participant_id: 110, id: 200, root_item_id: 2}
			- {participant_id: 110, id: 201, root_item_id: 4, parent_attempt_id: 200, allows_submissions_until: 2019-06-30 11:00:00}
			- {participant_id: 111, id: 200, root_item_id: 2, ended_at: 2019-06-30 11:00:00}
			- {participant_id: 111, id: 201, root_item_id: 4, parent_attempt_id: 200}
			- {participant_id: 112, id: 200, root_item_id: 2}
			- {participant_id: 112, id: 201, root_item_id: 4, parent_attempt_id: 200, ended_at: 2019-06-30 11:00:00}
			- {participant_id: 113, id: 200}
			- {participant_id: 114, id: 0}
			- {participant_id: 114, id: 200, root_item_id: 2}
			- {participant_id: 114, id: 201, root_item_id: 4, parent_attempt_id: 0}
			- {participant_id: 115, id: 0}
			- {participant_id: 115, id: 200, parent_attempt_id: 0}
			- {participant_id: 116, id: 0}
			- {participant_id: 116, id: 200}
			- {participant_id: 116, id: 201, root_item_id: 6, parent_attempt_id: 0}
			- {participant_id: 117, id: 0}
			- {participant_id: 117, id: 200, parent_attempt_id: 0}
			- {participant_id: 118, id: 0}
			- {participant_id: 118, id: 200, root_item_id: 6, parent_attempt_id: 0}
		results:
			- {participant_id: 100, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 101, attempt_id: 200, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 101, attempt_id: 201, item_id: 6, started_at: 2019-05-30 11:00:00}
			- {participant_id: 102, attempt_id: 200, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 103, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 104, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 105, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 105, attempt_id: 201, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 106, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 106, attempt_id: 201, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 107, attempt_id: 200, item_id: 2, started_at: null}
			- {participant_id: 107, attempt_id: 201, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 108, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 108, attempt_id: 201, item_id: 4, started_at: null}
			- {participant_id: 109, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 109, attempt_id: 201, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 110, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 110, attempt_id: 201, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 111, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 111, attempt_id: 201, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 112, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 112, attempt_id: 201, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 113, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 113, attempt_id: 200, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 114, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 114, attempt_id: 201, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 114, attempt_id: 201, item_id: 6, started_at: 2019-05-30 11:00:00}
			- {participant_id: 115, attempt_id: 0, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 115, attempt_id: 200, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 115, attempt_id: 200, item_id: 6, started_at: 2019-05-30 11:00:00}
			- {participant_id: 116, attempt_id: 200, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 116, attempt_id: 200, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 116, attempt_id: 201, item_id: 6, started_at: 2019-05-30 11:00:00}
			- {participant_id: 117, attempt_id: 0, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 117, attempt_id: 0, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 117, attempt_id: 200, item_id: 6, started_at: 2019-05-30 11:00:00}
			- {participant_id: 118, attempt_id: 0, item_id: 2, started_at: 2019-05-30 11:00:00}
			- {participant_id: 118, attempt_id: 0, item_id: 4, started_at: 2019-05-30 11:00:00}
			- {participant_id: 118, attempt_id: 200, item_id: 6, started_at: 2019-05-30 11:00:00}
	`)
	defer func() { _ = db.Close() }()

	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.GroupGroups().CreateNewAncestors()
		return nil
	}))

	type args struct {
		ids                               []int64
		groupID                           int64
		attemptID                         int64
		requireContentAccessToTheLastItem bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "empty list of ids", args: args{ids: []int64{}, groupID: 100, attemptID: 0}},
		{name: "one item, but attemptID != 0", args: args{ids: []int64{2}, groupID: 100, attemptID: 200}},
		{name: "wrong attempt_id", args: args{ids: []int64{2, 4}, groupID: 100, attemptID: 0}},
		{name: "first item is the group's activity", args: args{ids: []int64{4, 6}, groupID: 101, attemptID: 200}, want: true},
		{name: "first item is an activity of the group's ancestor", args: args{ids: []int64{4, 6}, groupID: 102, attemptID: 200}, want: true},
		{name: "first item is a root item", args: args{ids: []int64{2, 4}, groupID: 100, attemptID: 200}, want: true},
		{name: "first item is neither a root item nor the group's activity", args: args{ids: []int64{6, 8}, groupID: 101, attemptID: 201}},
		{
			name: "no content access to the last item when requireContentAccessToTheLastItem = true",
			args: args{ids: []int64{2, 4}, groupID: 103, attemptID: 200, requireContentAccessToTheLastItem: true},
		},
		{
			name: "no access to the last item when requireContentAccessToTheLastItem = false",
			args: args{ids: []int64{2, 4}, groupID: 104, attemptID: 200, requireContentAccessToTheLastItem: false},
		},
		{
			name: "content access to the last item when requireContentAccessToTheLastItem = true",
			args: args{ids: []int64{4, 6}, groupID: 101, attemptID: 200, requireContentAccessToTheLastItem: true},
			want: true,
		},
		{
			name: "info access to the last item when requireContentAccessToTheLastItem = false",
			args: args{ids: []int64{2, 4}, groupID: 103, attemptID: 200, requireContentAccessToTheLastItem: false},
			want: true,
		},
		{
			name: "no access to the last item when requireContentAccessToTheLastItem = true",
			args: args{ids: []int64{2, 4}, groupID: 104, attemptID: 200, requireContentAccessToTheLastItem: true},
		},
		{name: "no content access to the second to the last item", args: args{ids: []int64{2, 4, 6}, groupID: 105, attemptID: 201}},
		{name: "no content access to the first item", args: args{ids: []int64{2, 4, 6}, groupID: 106, attemptID: 201}},
		{name: "result of the first item is not started", args: args{ids: []int64{2, 4, 6}, groupID: 107, attemptID: 201}},
		{name: "result of the second to the last item is not started", args: args{ids: []int64{2, 4, 6}, groupID: 108, attemptID: 201}},
		{name: "attempt of the first item is expired", args: args{ids: []int64{2, 4, 6}, groupID: 109, attemptID: 201}},
		{name: "attempt of the second to the last item is expired", args: args{ids: []int64{2, 4, 6}, groupID: 110, attemptID: 201}},
		{name: "attempt of the first item is ended", args: args{ids: []int64{2, 4, 6}, groupID: 111, attemptID: 201}},
		{name: "attempt of the second to the last item is ended", args: args{ids: []int64{2, 4, 6}, groupID: 112, attemptID: 201}},
		{name: "the first item is not a parent of the second item", args: args{ids: []int64{4, 4, 6}, groupID: 113, attemptID: 200}},
		{name: "the second to the last item is not a parent of the last item", args: args{ids: []int64{2, 4, 4}, groupID: 113, attemptID: 200}},
		{
			name: "the first item's attempt is not a parent for the second items's attempt while the second item's attempt root_item_id is set",
			args: args{ids: []int64{2, 4, 6, 8}, groupID: 114, attemptID: 201},
		},
		{
			name: "the first item's attempt is not the same as the the second items's attempt " +
				"while the second item's attempt root_item_id is not set",
			args: args{ids: []int64{2, 4, 6, 8}, groupID: 115, attemptID: 200},
		},
		{
			name: "the third from the end item's attempt is not a parent for the second to the last items's attempt " +
				"while the second to the last item's attempt root_item_id is set",
			args: args{ids: []int64{2, 4, 6, 8}, groupID: 116, attemptID: 201},
		},
		{
			name: "the third from the end item's attempt is not the same as the second to the last items's attempt " +
				"while the second to the last item's attempt root_item_id is not set",
			args: args{ids: []int64{2, 4, 6, 8}, groupID: 117, attemptID: 200},
		},
		{name: "everything is okay (1 item)", args: args{ids: []int64{4}, groupID: 101, attemptID: 0}, want: true},
		{name: "everything is okay (4 items)", args: args{ids: []int64{2, 4, 6, 8}, groupID: 118, attemptID: 200}, want: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				got, err := store.Items().IsValidParticipationHierarchy(
					tt.args.ids, tt.args.groupID, tt.args.attemptID, tt.args.requireContentAccessToTheLastItem)
				assert.Equal(t, tt.want, got)
				assert.NoError(t, err)
				return nil
			}))
		})
	}
}

func TestItemStore_GetAccessDetailsForIDs(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		items: [{id: 11, default_language_tag: fr}, {id: 12, default_language_tag: fr}, {id: 13, default_language_tag: fr}]
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
	defer func() { _ = db.Close() }()

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

func TestItemStore_TriggerBeforeInsert_SetsPlatformID(t *testing.T) {
	tests := []struct {
		name           string
		url            *string
		wantPlatformID *int64
	}{
		{name: "url is null", url: nil, wantPlatformID: nil},
		{name: "chooses a platform with higher priority", url: ptrString("1234"), wantPlatformID: ptrInt64(2)},
		{name: "url doesn't match any regexp", url: ptrString("34"), wantPlatformID: nil},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				platforms:
					- {id: 3, regexp: "^1.*", priority: 1}
					- {id: 4, regexp: "^2.*", priority: 2}
					- {id: 2, regexp: "^1.*", priority: 3}
					- {id: 1, regexp: "^4.*", priority: 4}
				languages: [{tag: fr}]`)
			defer func() { _ = db.Close() }()

			itemStore := database.NewDataStore(db).Items()
			assert.NoError(t, itemStore.WithForeignKeyChecksDisabled(func(store *database.DataStore) error {
				return store.Items().InsertMap(map[string]interface{}{
					"url":                  test.url,
					"default_language_tag": "fr",
				})
			}))
			var platformID *int64
			assert.NoError(t, itemStore.PluckFirst("platform_id", &platformID).Error())
			if test.wantPlatformID == nil {
				assert.Nil(t, platformID)
			} else {
				assert.NotNil(t, platformID)
				if platformID != nil {
					assert.Equal(t, *test.wantPlatformID, *platformID)
				}
			}
		})
	}
}

func TestItemStore_TriggerBeforeUpdate_SetsPlatformID(t *testing.T) {
	tests := []struct {
		name           string
		updateMap      map[string]interface{}
		wantPlatformID *int64
	}{
		{name: "url is unchanged", updateMap: map[string]interface{}{"type": "Chapter"}, wantPlatformID: ptrInt64(4)},
		{name: "new url is null", updateMap: map[string]interface{}{"url": nil}, wantPlatformID: nil},
		{name: "chooses a platform with higher priority", updateMap: map[string]interface{}{"url": ptrString("12345")},
			wantPlatformID: ptrInt64(2)},
		{name: "new url doesn't match any regexp", updateMap: map[string]interface{}{"url": ptrString("34")}, wantPlatformID: nil},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				platforms:
					- {id: 3, regexp: "^1.*", priority: 1}
					- {id: 4, regexp: "^2.*", priority: 2}
					- {id: 2, regexp: "^1.*", priority: 3}
					- {id: 1, regexp: "^4.*", priority: 4}
				languages: [{tag: fr}]
				items:
					- {id: 1, platform_id: 4, url: 1234, default_language_tag: fr}`)
			defer func() { _ = db.Close() }()

			itemStore := database.NewDataStore(db).Items()
			assert.NoError(t, itemStore.ByID(1).UpdateColumn("platform_id", 4).Error())
			assert.NoError(t, itemStore.UpdateColumn(test.updateMap).Error())
			var platformID *int64
			assert.NoError(t, itemStore.ByID(1).PluckFirst("platform_id", &platformID).Error())
			if test.wantPlatformID == nil {
				assert.Nil(t, platformID)
			} else {
				assert.NotNil(t, platformID)
				if platformID != nil {
					assert.Equal(t, *test.wantPlatformID, *platformID)
				}
			}
		})
	}
}

func TestItemStore_PlatformsTriggerAfterInsert_SetsPlatformID(t *testing.T) {
	tests := []struct {
		name            string
		regexp          string
		priority        int
		wantPlatformIDs []*int64
	}{
		{name: "recalculates items linked to platforms with lower priority or no platform",
			regexp: "1", priority: 3,
			wantPlatformIDs: []*int64{ptrInt64(2), ptrInt64(1), ptrInt64(2), ptrInt64(2)},
		},
		{name: "recalculates items linked to platforms with lower priority or no platform (higher priority)",
			regexp: "1", priority: 6,
			wantPlatformIDs: []*int64{ptrInt64(5), ptrInt64(4), ptrInt64(5), nil},
		},
		{name: "recalculates only item without a platform when the new platform has the lowest priority",
			regexp: "1", priority: -1,
			wantPlatformIDs: []*int64{ptrInt64(4), ptrInt64(1), ptrInt64(2), ptrInt64(2)},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				platforms:
					- {id: 3, regexp: "^1.*", priority: 1}
					- {id: 4, regexp: "^2.*", priority: 2}
					- {id: 2, regexp: "^1.*", priority: 4}
					- {id: 1, regexp: "^4.*", priority: 5}
				languages: [{tag: fr}]
				items:
					- {id: 1, platform_id: 4, url: 1234, default_language_tag: fr}
					- {id: 2, platform_id: 1, url: 234, default_language_tag: fr}
					- {id: 3, platform_id: null, url: 123, default_language_tag: fr}
					- {id: 4, platform_id: 2, url: "987", default_language_tag: fr}`)
			defer func() { _ = db.Close() }()

			itemStore := database.NewDataStore(db).Items()
			assert.NoError(t, itemStore.ByID(1).UpdateColumn("platform_id", 4).Error())
			assert.NoError(t, itemStore.ByID(2).UpdateColumn("platform_id", 1).Error())
			assert.NoError(t, itemStore.ByID(3).UpdateColumn("platform_id", nil).Error())
			assert.NoError(t, itemStore.ByID(4).UpdateColumn("platform_id", 2).Error())
			assert.NoError(t, itemStore.
				Exec("INSERT platforms (id, `regexp`, priority) VALUES (5, ?, ?)", test.regexp, test.priority).Error())
			var platformIDs []*int64
			assert.NoError(t, itemStore.Order("id").Pluck("platform_id", &platformIDs).Error())
			assert.Equal(t, test.wantPlatformIDs, platformIDs)
		})
	}
}

func TestItemStore_PlatformsTriggerAfterUpdate_SetsPlatformID(t *testing.T) {
	tests := []struct {
		name            string
		regexp          string
		priority        int
		wantPlatformIDs []*int64
	}{
		{name: "recalculates items linked to platforms with lower priority or no platform or the modified platform (only priority is changed)",
			regexp: "^1.*", priority: 3,
			wantPlatformIDs: []*int64{ptrInt64(2), ptrInt64(1), ptrInt64(2), nil},
		},
		{name: "recalculates items linked to platforms with lower priority or no platform or the modified platform (only regexp is changed)",
			regexp: "1", priority: 4,
			wantPlatformIDs: []*int64{ptrInt64(2), ptrInt64(1), ptrInt64(2), nil},
		},
		{name: "recalculates items linked to platforms with lower priority or no platform or the modified platform (higher priority)",
			regexp: "1", priority: 6,
			wantPlatformIDs: []*int64{ptrInt64(2), ptrInt64(4), ptrInt64(2), nil},
		},
		{name: "recalculates only item without a platform when the new platform has the lowest priority",
			regexp: "1", priority: -1,
			wantPlatformIDs: []*int64{ptrInt64(4), ptrInt64(1), ptrInt64(3), nil},
		},
		{name: "doesn't recalculate anything when regexp & priority stays unchanged",
			regexp: "^1.*", priority: 4,
			wantPlatformIDs: []*int64{ptrInt64(4), ptrInt64(1), nil, ptrInt64(2)},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				platforms:
					- {id: 3, regexp: "^1.*", priority: 1}
					- {id: 4, regexp: "^2.*", priority: 2}
					- {id: 2, regexp: "^1.*", priority: 4}
					- {id: 1, regexp: "^4.*", priority: 5}
				languages: [{tag: fr}]
				items:
					- {id: 1, platform_id: 4, url: 1234, default_language_tag: fr}
					- {id: 2, platform_id: 1, url: 234, default_language_tag: fr}
					- {id: 3, platform_id: null, url: 123, default_language_tag: fr}
					- {id: 4, platform_id: 2, url: "987", default_language_tag: fr}`)
			defer func() { _ = db.Close() }()

			itemStore := database.NewDataStore(db).Items()
			assert.NoError(t, itemStore.ByID(1).UpdateColumn("platform_id", 4).Error())
			assert.NoError(t, itemStore.ByID(2).UpdateColumn("platform_id", 1).Error())
			assert.NoError(t, itemStore.ByID(3).UpdateColumn("platform_id", nil).Error())
			assert.NoError(t, itemStore.ByID(4).UpdateColumn("platform_id", 2).Error())
			assert.NoError(t, itemStore.Table("platforms").Where("id = ?", 2).
				UpdateColumn(map[string]interface{}{"regexp": test.regexp, "priority": test.priority}).Error())
			var platformIDs []*int64
			assert.NoError(t, itemStore.Order("id").Pluck("platform_id", &platformIDs).Error())
			assert.Equal(t, test.wantPlatformIDs, platformIDs)
		})
	}
}
