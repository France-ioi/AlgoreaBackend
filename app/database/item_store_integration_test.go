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
		"SELECT item_id, MIN(cached_full_access_since) <= NOW() AS full_access, "+
			"MIN(cached_partial_access_since) <= NOW() AS partial_access, "+
			"MIN(cached_grayed_access_since) <= NOW() AS grayed_access, "+
			"MIN(cached_solutions_access_since) <= NOW() AS access_solutions "+
			"FROM `groups_items` "+
			"JOIN ("+
			"SELECT * FROM `groups_ancestors` "+
			"WHERE (`groups_ancestors`.child_group_id = ?) AND (NOW() < `groups_ancestors`.expires_at)"+
			") AS ancestors "+
			"ON groups_items.group_id = ancestors.ancestor_group_id GROUP BY item_id") + "$").
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
		groups: [{id: 20}]
		users: [{login: 1, group_id: 20}]
		items: [{id: 11}, {id: 12}, {id: 13}, {id: 14}, {id: 15}, {id: 16}]
		items_ancestors:
			- {ancestor_item_id: 11, child_item_id: 12}
			- {ancestor_item_id: 11, child_item_id: 13}
			- {ancestor_item_id: 11, child_item_id: 14}
			- {ancestor_item_id: 11, child_item_id: 15}
			- {ancestor_item_id: 11, child_item_id: 16}
		groups_attempts:
			- {group_id: 20, item_id: 11, entered_at: 2018-03-22 08:44:55, order: 1}
			- {group_id: 20, item_id: 12, entered_at: 2018-03-22 08:44:55, order: 1}
			- {group_id: 21, item_id: 11, entered_at: 2018-03-22 08:44:55, order: 1}
		groups_items:
			- {group_id: 20, item_id: 11}
			- {group_id: 20, item_id: 12}
			- {group_id: 20, item_id: 13, cached_full_access_since: 2030-03-22 08:44:55} # no full access
			- {group_id: 20, item_id: 14, cached_full_access_since: 2018-03-22 08:44:55} # full access
			- {group_id: 20, item_id: 15, owner_access: 1}
			- {group_id: 20, item_id: 16, manager_access: 1}
			- {group_id: 21, item_id: 12}`)
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
	assert.NoError(t, store.GroupItems().Select("group_id, item_id").
		Order("group_id, item_id").
		Scan(&groupItems).Error())
	assert.Equal(t, []groupItemInfo{
		{GroupID: 20, ItemID: 11},
		{GroupID: 20, ItemID: 14},
		{GroupID: 20, ItemID: 15},
		{GroupID: 20, ItemID: 16},
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
		groups_items:
			- {group_id: 20, item_id: 11, cached_partial_access_since: 2018-03-22 08:44:55,
				partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1}
			- {group_id: 40, item_id: 11, cached_partial_access_since: 2018-03-22 08:44:55,
				partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1}
			- {group_id: 20, item_id: 12, cached_partial_access_since: 2018-03-22 08:44:55,
				partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1}
			- {group_id: 40, item_id: 12, cached_partial_access_since: 2018-03-22 08:44:55,
				partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1}
			- {group_id: 50, item_id: 11, cached_partial_access_since: 2018-03-22 08:44:55,
				partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1}
			- {group_id: 50, item_id: 12, cached_partial_access_since: 2018-03-22 08:44:55,
			   partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1}`)
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

	type groupItemInfo struct {
		GroupID                  int64
		ItemID                   int64
		PartialAccessSince       *database.Time
		CachedPartialAccessSince *database.Time
		CachedPartialAccess      bool
	}
	var groupItems []groupItemInfo
	assert.NoError(t, store.GroupItems().
		Select("group_id, item_id, partial_access_since, cached_partial_access_since, cached_partial_access").
		Order("group_id, item_id").
		Scan(&groupItems).Error())
	expectedDate := (*database.Time)(ptrTime(time.Date(2018, 3, 22, 8, 44, 55, 0, time.UTC)))
	assert.Equal(t, []groupItemInfo{
		{GroupID: 20, ItemID: 11, PartialAccessSince: expectedDate, CachedPartialAccessSince: expectedDate, CachedPartialAccess: true},
		{GroupID: 20, ItemID: 12, PartialAccessSince: expectedDate, CachedPartialAccessSince: expectedDate, CachedPartialAccess: true},
		{GroupID: 40, ItemID: 11, PartialAccessSince: nil, CachedPartialAccessSince: nil, CachedPartialAccess: false},
		{GroupID: 40, ItemID: 12, PartialAccessSince: expectedDate, CachedPartialAccessSince: expectedDate, CachedPartialAccess: true},
		{GroupID: 50, ItemID: 11, PartialAccessSince: expectedDate, CachedPartialAccessSince: expectedDate, CachedPartialAccess: true},
		{GroupID: 50, ItemID: 12, PartialAccessSince: expectedDate, CachedPartialAccessSince: expectedDate, CachedPartialAccess: true},
	}, groupItems)
}

func TestItemStore_Visible_ProvidesAccessSolutions(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		items: [{id: 11}, {id: 12}, {id: 13}]
		groups: [{id: 10}, {id: 40}]
		users: [{group_id: 10}]
		groups_groups:
			- {parent_group_id: 40, child_group_id: 10}
		groups_ancestors:
			- {ancestor_group_id: 10, child_group_id: 10}
			- {ancestor_group_id: 40, child_group_id: 10}
			- {ancestor_group_id: 40, child_group_id: 40}
		groups_items:
			- {group_id: 40, item_id: 11, cached_full_access_since: 2018-03-22 08:44:55, cached_solutions_access_since: 2018-03-22 08:44:55}
			- {group_id: 10, item_id: 11, cached_full_access_since: 2018-03-22 08:44:55, cached_solutions_access_since: 2019-03-22 08:44:55}
			- {group_id: 10, item_id: 12, cached_full_access_since: 2018-03-22 08:44:55, cached_solutions_access_since: 2019-04-22 08:44:55}
			- {group_id: 10, item_id: 13, cached_full_access_since: 2018-03-22 08:44:55}`)
	type resultType struct {
		ID              int64
		AccessSolutions bool
	}
	var result []resultType

	assert.NoError(t, database.NewDataStore(db).Items().
		Visible(&database.User{GroupID: 10}).
		Select("id, access_solutions").Order("id").Scan(&result).Error())
	assert.Equal(t, []resultType{
		{ID: 11, AccessSolutions: true},
		{ID: 12, AccessSolutions: true},
		{ID: 13, AccessSolutions: false},
	}, result)
}

func TestItemStore_HasManagerAccess(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		items: [{id: 11}, {id: 12}, {id: 13}]
		groups: [{id: 10}, {id: 11}, {id: 40}, {id: 100}, {id: 110}]
		users: [{login: 1, group_id: 100}, {login: 2, group_id: 110}]
		groups_groups:
			- {parent_group_id: 400, child_group_id: 100}
		groups_ancestors:
			- {ancestor_group_id: 100, child_group_id: 100}
			- {ancestor_group_id: 110, child_group_id: 110}
			- {ancestor_group_id: 400, child_group_id: 100}
			- {ancestor_group_id: 400, child_group_id: 400}
		groups_items:
			- {group_id: 400, item_id: 11, cached_manager_access: 1}
			- {group_id: 100, item_id: 11, owner_access: 1}
			- {group_id: 100, item_id: 12}
			- {group_id: 100, item_id: 13}
			- {group_id: 110, item_id: 12, owner_access: 1}
			- {group_id: 110, item_id: 13, cached_manager_access: 1}`)

	tests := []struct {
		name        string
		ids         []int64
		userGroupID int64
		wantResult  bool
	}{
		{name: "two groups_items rows for one item", ids: []int64{11}, userGroupID: 100, wantResult: true},
		{name: "no manager/owner access", ids: []int64{12}, userGroupID: 100, wantResult: false},
		{name: "access to a part of items", ids: []int64{11, 12}, userGroupID: 100, wantResult: false},
		{name: "no manager/owner access for another user", ids: []int64{11}, userGroupID: 110, wantResult: false},
		{name: "owner access", ids: []int64{12}, userGroupID: 110, wantResult: true},
		{name: "manager access", ids: []int64{13}, userGroupID: 110, wantResult: true},
		{name: "two items", ids: []int64{12, 13}, userGroupID: 110, wantResult: true},
		{name: "two items (not unique)", ids: []int64{12, 13, 12, 13}, userGroupID: 110, wantResult: true},
		{name: "empty ids list", ids: []int64{}, userGroupID: 110, wantResult: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				user := &database.User{}
				assert.NoError(t, user.LoadByGroupID(store, test.userGroupID))
				hasAccess, err := store.Items().
					HasManagerAccess(user, test.ids...)
				assert.NoError(t, err)
				assert.Equal(t, test.wantResult, hasAccess)
				return nil
			}))
		})
	}
}
