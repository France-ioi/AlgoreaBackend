// +build !unit

package database_test

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

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

			user := &database.User{GroupID: 11, OwnedGroupID: ptrInt64(12), DefaultLanguageID: 2}
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

func TestItemStore_AccessRights(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &database.User{GroupID: 2, OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery("^" + regexp.QuoteMeta(
		"SELECT item_id, MAX(can_view_generated_value) AS can_view_generated_value "+
			"FROM `permissions_generated` "+
			"JOIN ("+
			"SELECT * FROM `groups_ancestors` "+
			"WHERE (`groups_ancestors`.child_group_id = ?) AND (NOW() < `groups_ancestors`.expires_at)"+
			") AS ancestors "+
			"ON permissions_generated.group_id = ancestors.ancestor_group_id GROUP BY item_id") + "$").
		WithArgs(2).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := database.NewDataStore(db).Items().AccessRights(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
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
		userGroupID   int64
		wantHasAccess bool
		wantReason    error
		initFunc      func(*database.DB) error
	}{
		{name: "no items", itemID: 404, userGroupID: 11, wantHasAccess: true, wantReason: nil},
		{name: "user has no active contest", itemID: 14, userGroupID: 11, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active team contest has expired", itemID: 14, userGroupID: 12, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active team contest has expired (again)", itemID: 14, userGroupID: 12, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active contest has expired", itemID: 15, userGroupID: 13, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active contest has expired (again)", itemID: 15, userGroupID: 13, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active contest is OK and it is from another competition, but the user has full access to the time-limited chapter",
			initFunc: func(db *database.DB) error {
				return database.NewDataStore(db).GroupAttempts().InsertMap(
					map[string]interface{}{
						"item_id":    500, // chapter
						"group_id":   14,
						"entered_at": database.Now(),
						"order":      1,
					})
			},
			itemID: 15, userGroupID: 14, wantHasAccess: true, wantReason: nil},
		{name: "user's active contest is OK and it is the task's time-limited chapter",
			initFunc: func(db *database.DB) error {
				return database.NewDataStore(db).GroupAttempts().
					InsertMap(map[string]interface{}{
						"item_id":    115,
						"group_id":   15,
						"entered_at": database.Now(),
						"order":      1,
					})
			},
			itemID: 15, userGroupID: 15, wantHasAccess: true, wantReason: nil},
		{name: "user's active contest is OK, but it is not an ancestor of the task and the user doesn't have full access to the task's chapter",
			initFunc: func(db *database.DB) error {
				return database.NewDataStore(db).GroupAttempts().
					InsertMap(map[string]interface{}{
						"item_id":    114,
						"group_id":   17,
						"entered_at": database.Now(),
						"order":      1,
					})
			},
			itemID: 15, userGroupID: 17, wantHasAccess: false,
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
				assert.NoError(t, user.LoadByGroupID(store, test.userGroupID))

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
		groups: [{id: 101}, {id: 102}, {id: 103}, {id: 104}, {id: 105}, {id: 106}]
		users:
			- {login: 1, group_id: 101}
			- {login: 2, group_id: 102}
			- {login: 3, group_id: 103}
			- {login: 4, group_id: 104}
			- {login: 5, group_id: 105}
			- {login: 6, group_id: 106}
		items: [{id: 12}, {id: 13}, {id: 14, duration: 10:00:00}, {id: 15}]
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
			- {group_id: 103, item_id: 13, entered_at: 2019-03-22 08:44:55, finished_at: 2019-03-22 09:44:55, order: 1} # finished
			- {group_id: 104, item_id: 14, entered_at: 2019-03-22 08:44:55, order: 1} # ok
			- {group_id: 105, item_id: 15, entered_at: 2019-04-22 08:44:55, order: 1}  # ok with team mode
			- {group_id: 106, item_id: 14, entered_at: 2019-03-22 08:44:55, order: 1} # multiple
			- {group_id: 106, item_id: 15, entered_at: 2019-03-22 08:43:55, order: 1} # multiple`)
	defer func() { _ = db.Close() }()

	tests := []struct {
		name        string
		userGroupID int64
		want        *database.ActiveContestInfo
	}{
		{name: "no item", userGroupID: 101, want: nil},
		{name: "not started", userGroupID: 102, want: nil},
		{name: "finished", userGroupID: 103, want: nil},
		{name: "ok", userGroupID: 104, want: &database.ActiveContestInfo{
			ItemID:                   14,
			UserGroupID:              104,
			DurationInSeconds:        36060,
			EndTime:                  time.Date(2019, 3, 22, 18, 45, 55, 0, time.UTC),
			StartTime:                time.Date(2019, 3, 22, 8, 44, 55, 0, time.UTC),
			ContestEnteringCondition: "None",
		}},
		{name: "ok with team mode", userGroupID: 105, want: &database.ActiveContestInfo{
			ItemID:                   15,
			UserGroupID:              105,
			DurationInSeconds:        0,
			EndTime:                  time.Date(2019, 4, 22, 8, 44, 55, 0, time.UTC),
			StartTime:                time.Date(2019, 4, 22, 8, 44, 55, 0, time.UTC),
			ContestEnteringCondition: "None",
		}},
		{
			name: "ok with multiple active contests", userGroupID: 106, want: &database.ActiveContestInfo{
				ItemID:                   14,
				UserGroupID:              106,
				DurationInSeconds:        36060,
				EndTime:                  time.Date(2019, 3, 22, 18, 45, 55, 0, time.UTC),
				StartTime:                time.Date(2019, 3, 22, 8, 44, 55, 0, time.UTC),
				ContestEnteringCondition: "None",
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			store := database.NewDataStore(db)
			user := &database.User{}
			assert.NoError(t, user.LoadByGroupID(store, test.userGroupID))

			got := store.Items().GetActiveContestInfoForUser(user)
			if got != nil && test.want != nil {
				assert.True(t, time.Since(got.Now) < 3*time.Second)
				assert.True(t, time.Since(got.Now) > -3*time.Second)
				test.want.Now = time.Now().UTC()
				got.Now = test.want.Now
			}
			assert.Equal(t, test.want, got)
		})
	}
}

func TestItemStore_CloseContest(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 20}, {id: 21}]
		users: [{login: 1, group_id: 20}]
		items: [{id: 11}, {id: 12}, {id: 13}, {id: 14}, {id: 15}, {id: 16}, {id: 17}]
		items_ancestors:
			- {ancestor_item_id: 11, child_item_id: 12}
			- {ancestor_item_id: 11, child_item_id: 13}
			- {ancestor_item_id: 11, child_item_id: 14}
			- {ancestor_item_id: 11, child_item_id: 15}
			- {ancestor_item_id: 11, child_item_id: 16}
			- {ancestor_item_id: 11, child_item_id: 17}
		groups_attempts:
			- {group_id: 20, item_id: 11, entered_at: 2018-03-22 08:44:55, order: 1}
			- {group_id: 20, item_id: 12, entered_at: 2018-03-22 08:44:55, order: 1}
			- {group_id: 21, item_id: 11, entered_at: 2018-03-22 08:44:55, order: 1}
		permissions_generated:
			- {group_id: 20, item_id: 11, can_view_generated: content}
			- {group_id: 20, item_id: 12, can_view_generated: info}
			- {group_id: 20, item_id: 13, can_view_generated: content}
			- {group_id: 20, item_id: 14, can_view_generated: content_with_descendants} # full access
			- {group_id: 20, item_id: 15, can_view_generated: content, is_owner_generated: 1}
			- {group_id: 20, item_id: 16, can_view_generated: solution}
			- {group_id: 20, item_id: 17, can_view_generated: content}
			- {group_id: 21, item_id: 12}
		permissions_granted:
			- {group_id: 20, item_id: 11, can_view: content, can_edit: children, giver_group_id: -1}
			- {group_id: 20, item_id: 12, can_view: info, giver_group_id: -1}
			- {group_id: 20, item_id: 13, can_view: content, giver_group_id: -1}
			- {group_id: 20, item_id: 14, can_view: content_with_descendants, giver_group_id: -1}
			- {group_id: 20, item_id: 15, can_view: content, is_owner: 1, giver_group_id: -1}
			- {group_id: 20, item_id: 16, can_view: solution, giver_group_id: -1}
			- {group_id: 20, item_id: 17, can_view: content, giver_group_id: 20}
			- {group_id: 21, item_id: 12, giver_group_id: -1}`)
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		user := &database.User{}
		assert.NoError(t, user.LoadByGroupID(store, 20))
		store.Items().CloseContest(11, user)
		return nil
	}))

	type groupParticipationsInfo struct {
		GroupID       int64
		ItemID        int64
		FinishedAtSet bool
	}
	var participations []groupParticipationsInfo
	store := database.NewDataStore(db)
	assert.NoError(t, store.GroupAttempts().
		Select("group_id, item_id, (finished_at IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, finished_at, NOW())) < 3) AS finished_at_set").
		Where("entered_at IS NOT NULL").
		Order("group_id, item_id").
		Scan(&participations).Error())
	assert.Equal(t, []groupParticipationsInfo{
		{GroupID: 20, ItemID: 11, FinishedAtSet: true},
		{GroupID: 20, ItemID: 12, FinishedAtSet: false},
		{GroupID: 21, ItemID: 11, FinishedAtSet: false},
	}, participations)

	type groupItemInfo struct {
		GroupID int64
		ItemID  int64
	}
	var groupItems []groupItemInfo
	assert.NoError(t, store.PermissionsGranted().Select("group_id, item_id").
		Order("group_id, item_id").
		Scan(&groupItems).Error())
	assert.Equal(t, []groupItemInfo{
		{GroupID: 20, ItemID: 11},
		{GroupID: 20, ItemID: 14},
		{GroupID: 20, ItemID: 15},
		{GroupID: 20, ItemID: 16},
		{GroupID: 20, ItemID: 17},
		{GroupID: 21, ItemID: 12},
	}, groupItems)
}

func TestItemStore_CloseTeamContest(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 10}, {id: 20}, {id: 30}, {id: 40, team_item_id: 11, type: Team}, {id: 50}]
		users:
			- {login: 1, group_id: 10}
			- {login: 2, group_id: 20}
			- {login: 3, group_id: 30}
			- {login: 4, group_id: 50}
		groups_groups:
			- {parent_group_id: 40, child_group_id: 10, type: invitationAccepted}
			- {parent_group_id: 40, child_group_id: 20, type: requestRefused}
			- {parent_group_id: 40, child_group_id: 30, type: requestAccepted}
			- {parent_group_id: 40, child_group_id: 50, type: joinedByCode}
		groups_ancestors:
			- {ancestor_group_id: 10, child_group_id: 10}
			- {ancestor_group_id: 20, child_group_id: 20}
			- {ancestor_group_id: 30, child_group_id: 30}
			- {ancestor_group_id: 40, child_group_id: 10}
			- {ancestor_group_id: 40, child_group_id: 30}
			- {ancestor_group_id: 40, child_group_id: 50}
		items: [{id: 11}, {id: 12}, {id: 13}]
		items_ancestors:
			- {ancestor_item_id: 11, child_item_id: 12}
			- {ancestor_item_id: 11, child_item_id: 13}
		groups_attempts:
			- {group_id: 40, item_id: 11, entered_at: 2018-03-22 08:44:55, order: 1}
			- {group_id: 40, item_id: 12, entered_at: 2018-03-22 08:44:55, order: 1}
		permissions_generated:
			- {group_id: 20, item_id: 11, can_view_generated: content}
			- {group_id: 40, item_id: 11, can_view_generated: content}
			- {group_id: 20, item_id: 12, can_view_generated: content}
			- {group_id: 40, item_id: 12, can_view_generated: content}
			- {group_id: 50, item_id: 11, can_view_generated: content}
			- {group_id: 50, item_id: 12, can_view_generated: content}
		permissions_granted:
			- {group_id: 20, item_id: 11, can_view: content, giver_group_id: -1}
			- {group_id: 40, item_id: 11, can_view: content, giver_group_id: -1}
			- {group_id: 20, item_id: 12, can_view: content, giver_group_id: -1}
			- {group_id: 40, item_id: 12, can_view: content, giver_group_id: -1}
			- {group_id: 50, item_id: 11, can_view: content, giver_group_id: -1}
			- {group_id: 50, item_id: 12, can_view: content, giver_group_id: -1}`)
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		user := &database.User{GroupID: 10}
		store.Items().CloseTeamContest(11, user)
		return nil
	}))

	type contestParticipationsInfo struct {
		GroupID       int64
		ItemID        int64
		FinishedAtSet bool
	}
	var participations []contestParticipationsInfo
	store := database.NewDataStore(db)
	assert.NoError(t, store.GroupAttempts().
		Select("group_id, item_id, (finished_at IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, finished_at, NOW())) < 3) as finished_at_set").
		Where("entered_at IS NOT NULL").
		Order("group_id, item_id").
		Scan(&participations).Error())
	assert.Equal(t, []contestParticipationsInfo{
		{GroupID: 40, ItemID: 11, FinishedAtSet: true},
		{GroupID: 40, ItemID: 12, FinishedAtSet: false},
	}, participations)

	type permissionsInfo struct {
		GroupID int64
		ItemID  int64
		CanView string
	}
	var permissions []permissionsInfo
	assert.NoError(t, store.PermissionsGranted().
		Select("group_id, item_id, can_view").
		Order("group_id, item_id").
		Scan(&permissions).Error())
	assert.Equal(t, []permissionsInfo{
		{GroupID: 20, ItemID: 11, CanView: "content"},
		{GroupID: 20, ItemID: 12, CanView: "content"},
		{GroupID: 40, ItemID: 11, CanView: "none"},
		{GroupID: 40, ItemID: 12, CanView: "content"},
		{GroupID: 50, ItemID: 11, CanView: "content"},
		{GroupID: 50, ItemID: 12, CanView: "content"},
	}, permissions)
	assert.NoError(t, store.PermissionsGenerated().
		Select("group_id, item_id, can_view_generated AS can_view").
		Order("group_id, item_id").
		Scan(&permissions).Error())
	assert.Equal(t, []permissionsInfo{
		{GroupID: 20, ItemID: 11, CanView: "content"},
		{GroupID: 20, ItemID: 12, CanView: "content"},
		{GroupID: 40, ItemID: 11, CanView: "none"},
		{GroupID: 40, ItemID: 12, CanView: "content"},
		{GroupID: 50, ItemID: 11, CanView: "content"},
		{GroupID: 50, ItemID: 12, CanView: "content"},
	}, permissions)
}
