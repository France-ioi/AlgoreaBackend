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

			user := &database.User{ID: 1, SelfGroupID: ptrInt64(11), OwnedGroupID: ptrInt64(12), DefaultLanguageID: 2}
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

	mockUser := &database.User{ID: 1, SelfGroupID: ptrInt64(2), OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

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
	user := &database.User{ID: 1, SelfGroupID: ptrInt64(10)}

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
		userID        int64
		wantHasAccess bool
		wantReason    error
		initFunc      func(*database.DB) error
	}{
		{name: "no items", itemID: 404, userID: 1, wantHasAccess: true, wantReason: nil},
		{name: "user has no active contest", itemID: 14, userID: 1, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active team contest has expired", itemID: 14, userID: 2, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active team contest has expired (again)", itemID: 14, userID: 2, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active contest has expired", itemID: 15, userID: 3, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active contest has expired (again)", itemID: 15, userID: 3, wantHasAccess: false,
			wantReason: errors.New("the contest has not started yet or has already finished")},
		{name: "user's active contest is OK and it is from another competition, but the user has full access to the time-limited chapter",
			initFunc: func(db *database.DB) error {
				return database.NewDataStore(db).ContestParticipations().InsertMap(
					map[string]interface{}{
						"item_id":    500, // chapter
						"group_id":   14,
						"entered_at": database.Now(),
					})
			},
			itemID: 15, userID: 4, wantHasAccess: true, wantReason: nil},
		{name: "user's active contest is OK and it is the task's time-limited chapter",
			initFunc: func(db *database.DB) error {
				return database.NewDataStore(db).ContestParticipations().
					InsertMap(map[string]interface{}{
						"item_id":    115,
						"group_id":   15,
						"entered_at": database.Now(),
					})
			},
			itemID: 15, userID: 5, wantHasAccess: true, wantReason: nil},
		{name: "user's active contest is OK, but it is not an ancestor of the task and the user doesn't have full access to the task's chapter",
			initFunc: func(db *database.DB) error {
				return database.NewDataStore(db).ContestParticipations().
					InsertMap(map[string]interface{}{
						"item_id":    114,
						"group_id":   17,
						"entered_at": database.Now(),
					})
			},
			itemID: 15, userID: 7, wantHasAccess: false,
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
		users:
			- {id: 1, login: 1, self_group_id: 101}
			- {id: 2, login: 2, self_group_id: 102}
			- {id: 3, login: 3, self_group_id: 103}
			- {id: 4, login: 4, self_group_id: 104}
			- {id: 5, login: 5, self_group_id: 105}
			- {id: 6, login: 6, self_group_id: 106}
		items: [{id: 12}, {id: 13}, {id: 14, duration: 10:00:00}, {id: 15}]
		groups_ancestors:
			- {ancestor_group_id: 101, child_group_id: 101}
			- {ancestor_group_id: 102, child_group_id: 102}
			- {ancestor_group_id: 103, child_group_id: 103}
			- {ancestor_group_id: 104, child_group_id: 104}
			- {ancestor_group_id: 105, child_group_id: 105}
			- {ancestor_group_id: 106, child_group_id: 106}
		users_items:
			- {user_id: 2, item_id: 12}
			- {user_id: 3, item_id: 13, finished_at: 2019-03-23 08:44:55} #finished
			- {user_id: 4, item_id: 14} # ok
			- {user_id: 5, item_id: 15} # ok with team mode
			- {user_id: 6, item_id: 14} # multiple
			- {user_id: 6, item_id: 15} # multiple
		groups_contest_items:
			- {group_id: 102, item_id: 12} # not started
			- {group_id: 104, item_id: 14, additional_time: 00:01:00} # ok
			- {group_id: 105, item_id: 15}  # ok with team mode
			- {group_id: 106, item_id: 14, additional_time: 00:01:00} # multiple
			- {group_id: 106, item_id: 15, additional_time: 00:01:00} # multiple
		contest_participations:
			- {group_id: 104, item_id: 14, entered_at: 2019-03-22 08:44:55} # ok
			- {group_id: 105, item_id: 15, entered_at: 2019-04-22 08:44:55}  # ok with team mode
			- {group_id: 106, item_id: 14, entered_at: 2019-03-22 08:44:55} # multiple
			- {group_id: 106, item_id: 15, entered_at: 2019-03-22 08:43:55} # multiple`)
	defer func() { _ = db.Close() }()

	tests := []struct {
		name   string
		userID int64
		want   *database.ActiveContestInfo
	}{
		{name: "no item", userID: 1, want: nil},
		{name: "not started", userID: 2, want: nil},
		{name: "finished", userID: 3, want: nil},
		{name: "ok", userID: 4, want: &database.ActiveContestInfo{
			ItemID:                   14,
			UserID:                   4,
			DurationInSeconds:        36060,
			EndTime:                  time.Date(2019, 3, 22, 18, 45, 55, 0, time.UTC),
			StartTime:                time.Date(2019, 3, 22, 8, 44, 55, 0, time.UTC),
			ContestEnteringCondition: "None",
		}},
		{name: "ok with team mode", userID: 5, want: &database.ActiveContestInfo{
			ItemID:                   15,
			UserID:                   5,
			DurationInSeconds:        0,
			EndTime:                  time.Date(2019, 4, 22, 8, 44, 55, 0, time.UTC),
			StartTime:                time.Date(2019, 4, 22, 8, 44, 55, 0, time.UTC),
			ContestEnteringCondition: "None",
		}},
		{
			name: "ok with multiple active contests", userID: 6, want: &database.ActiveContestInfo{
				ItemID:                   14,
				UserID:                   6,
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
			assert.NoError(t, user.LoadByID(store, test.userID))

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
		users: [{id: 1, login: 1, self_group_id: 20}]
		groups: [{id: 20}]
		items: [{id: 11}, {id: 12}, {id: 13}, {id: 14}, {id: 15}, {id: 16}]
		items_ancestors:
			- {ancestor_item_id: 11, child_item_id: 12}
			- {ancestor_item_id: 11, child_item_id: 13}
			- {ancestor_item_id: 11, child_item_id: 14}
			- {ancestor_item_id: 11, child_item_id: 15}
			- {ancestor_item_id: 11, child_item_id: 16}
		users_items: [{user_id: 1, item_id: 11}, {user_id: 1, item_id: 12}, {user_id: 2, item_id: 11}]
		groups_items:
			- {group_id: 20, item_id: 11, creator_user_id: 1}
			- {group_id: 20, item_id: 12, creator_user_id: 1}
			- {group_id: 20, item_id: 13, cached_full_access_since: 2030-03-22 08:44:55, creator_user_id: 1} # no full access
			- {group_id: 20, item_id: 14, cached_full_access_since: 2018-03-22 08:44:55, creator_user_id: 1} # full access
			- {group_id: 20, item_id: 15, owner_access: 1, creator_user_id: 1}
			- {group_id: 20, item_id: 16, manager_access: 1, creator_user_id: 1}
			- {group_id: 21, item_id: 12, creator_user_id: 1}`)
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		user := &database.User{}
		assert.NoError(t, user.LoadByID(store, 1))
		store.Items().CloseContest(11, user)
		return nil
	}))

	type userItemInfo struct {
		UserID        int64
		ItemID        int64
		FinishedAtSet bool
	}
	var userItems []userItemInfo
	store := database.NewDataStore(db)
	assert.NoError(t, store.UserItems().
		Select("user_id, item_id, (finished_at IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, finished_at, NOW())) < 3) AS finished_at_set").
		Order("user_id, item_id").
		Scan(&userItems).Error())
	assert.Equal(t, []userItemInfo{
		{UserID: 1, ItemID: 11, FinishedAtSet: true},
		{UserID: 1, ItemID: 12, FinishedAtSet: false},
		{UserID: 2, ItemID: 11, FinishedAtSet: false},
	}, userItems)

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
		users:
			- {id: 1, login: 1, self_group_id: 10}
			- {id: 2, login: 2, self_group_id: 20}
			- {id: 3, login: 3, self_group_id: 30}
			- {id: 4, login: 4, self_group_id: 50}
		groups: [{id: 10}, {id: 20}, {id: 30}, {id: 40, team_item_id: 11, type: Team}, {id: 50}]
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
		users_items:
			- {user_id: 1, item_id: 11}
			- {user_id: 1, item_id: 12}
			- {user_id: 2, item_id: 11}
			- {user_id: 3, item_id: 11}
			- {user_id: 4, item_id: 11}
		groups_items:
			- {group_id: 20, item_id: 11, cached_partial_access_since: 2018-03-22 08:44:55,
				partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1, creator_user_id: 1}
			- {group_id: 40, item_id: 11, cached_partial_access_since: 2018-03-22 08:44:55,
				partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1, creator_user_id: 1}
			- {group_id: 20, item_id: 12, cached_partial_access_since: 2018-03-22 08:44:55,
				partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1, creator_user_id: 1}
			- {group_id: 40, item_id: 12, cached_partial_access_since: 2018-03-22 08:44:55,
				partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1, creator_user_id: 1}
			- {group_id: 50, item_id: 11, cached_partial_access_since: 2018-03-22 08:44:55,
				partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1, creator_user_id: 1}
			- {group_id: 50, item_id: 12, cached_partial_access_since: 2018-03-22 08:44:55,
			   partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1, creator_user_id: 1}`)
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		user := &database.User{ID: 1, SelfGroupID: ptrInt64(10)}
		store.Items().CloseTeamContest(11, user)
		return nil
	}))

	type userItemInfo struct {
		UserID        int64
		ItemID        int64
		FinishedAtSet bool
	}
	var userItems []userItemInfo
	store := database.NewDataStore(db)
	assert.NoError(t, store.UserItems().
		Select("user_id, item_id, (finished_at IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, finished_at, NOW())) < 3) as finished_at_set").
		Order("user_id, item_id").
		Scan(&userItems).Error())
	assert.Equal(t, []userItemInfo{
		{UserID: 1, ItemID: 11, FinishedAtSet: true},
		{UserID: 1, ItemID: 12, FinishedAtSet: false},
		{UserID: 2, ItemID: 11, FinishedAtSet: false},
		{UserID: 3, ItemID: 11, FinishedAtSet: true},
		{UserID: 4, ItemID: 11, FinishedAtSet: true},
	}, userItems)

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
		users: [{id: 1, self_group_id: 10}]
		groups: [{id: 10}, {id: 40}]
		groups_groups:
			- {parent_group_id: 40, child_group_id: 10}
		groups_ancestors:
			- {ancestor_group_id: 10, child_group_id: 10}
			- {ancestor_group_id: 40, child_group_id: 10}
			- {ancestor_group_id: 40, child_group_id: 40}
		groups_items:
			- {group_id: 40, item_id: 11, cached_full_access_since: 2018-03-22 08:44:55, cached_solutions_access_since: 2018-03-22 08:44:55,
		     creator_user_id: 1}
			- {group_id: 10, item_id: 11, cached_full_access_since: 2018-03-22 08:44:55, cached_solutions_access_since: 2019-03-22 08:44:55,
			   creator_user_id: 1}
			- {group_id: 10, item_id: 12, cached_full_access_since: 2018-03-22 08:44:55, cached_solutions_access_since: 2019-04-22 08:44:55,
			   creator_user_id: 1}
			- {group_id: 10, item_id: 13, cached_full_access_since: 2018-03-22 08:44:55, creator_user_id: 1}`)
	type resultType struct {
		ID              int64
		AccessSolutions bool
	}
	var result []resultType

	assert.NoError(t, database.NewDataStore(db).Items().
		Visible(&database.User{ID: 1, SelfGroupID: ptrInt64(10)}).
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
		users: [{id: 1, login: 1, self_group_id: 100}, {id: 2, login: 2, self_group_id: 110}]
		groups: [{id: 10}, {id: 11}, {id: 40}]
		groups_groups:
			- {parent_group_id: 400, child_group_id: 100}
		groups_ancestors:
			- {ancestor_group_id: 100, child_group_id: 100}
			- {ancestor_group_id: 110, child_group_id: 110}
			- {ancestor_group_id: 400, child_group_id: 100}
			- {ancestor_group_id: 400, child_group_id: 400}
		groups_items:
			- {group_id: 400, item_id: 11, cached_manager_access: 1, creator_user_id: 1}
			- {group_id: 100, item_id: 11, owner_access: 1, creator_user_id: 1}
			- {group_id: 100, item_id: 12, creator_user_id: 1}
			- {group_id: 100, item_id: 13, creator_user_id: 1}
			- {group_id: 110, item_id: 12, owner_access: 1, creator_user_id: 1}
			- {group_id: 110, item_id: 13, cached_manager_access: 1, creator_user_id: 1}`)

	tests := []struct {
		name       string
		ids        []int64
		userID     int64
		wantResult bool
	}{
		{name: "two groups_items rows for one item", ids: []int64{11}, userID: 1, wantResult: true},
		{name: "no manager/owner access", ids: []int64{12}, userID: 1, wantResult: false},
		{name: "access to a part of items", ids: []int64{11, 12}, userID: 1, wantResult: false},
		{name: "no manager/owner access for another user", ids: []int64{11}, userID: 2, wantResult: false},
		{name: "owner access", ids: []int64{12}, userID: 2, wantResult: true},
		{name: "manager access", ids: []int64{13}, userID: 2, wantResult: true},
		{name: "two items", ids: []int64{12, 13}, userID: 2, wantResult: true},
		{name: "two items (not unique)", ids: []int64{12, 13, 12, 13}, userID: 2, wantResult: true},
		{name: "empty ids list", ids: []int64{}, userID: 2, wantResult: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				user := &database.User{}
				assert.NoError(t, user.LoadByID(store, test.userID))
				hasAccess, err := store.Items().
					HasManagerAccess(user, test.ids...)
				assert.NoError(t, err)
				assert.Equal(t, test.wantResult, hasAccess)
				return nil
			}))
		})
	}
}
