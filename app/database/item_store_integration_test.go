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
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/loggingtest"
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
		{methodToCall: "Visible", column: "ID", expected: []int64{190, 191, 192, 1900, 1901, 1902, 19000, 19001, 19002}},
		{methodToCall: "VisibleByID", args: []interface{}{int64(191)}, column: "ID", expected: []int64{191}},
		{methodToCall: "VisibleChildrenOfID", args: []interface{}{int64(190)}, column: "items.ID", expected: []int64{1900, 1901, 1902}},
		{methodToCall: "VisibleGrandChildrenOfID", args: []interface{}{int64(190)}, column: "items.ID", expected: []int64{19000, 19001, 19002}},
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
		"SELECT idItem, MIN(sCachedFullAccessDate) <= NOW() AS fullAccess, "+
			"MIN(sCachedPartialAccessDate) <= NOW() AS partialAccess, "+
			"MIN(sCachedGrayedAccessDate) <= NOW() AS grayedAccess, "+
			"MIN(sCachedAccessSolutionsDate) <= NOW() AS accessSolutions "+
			"FROM `groups_items` "+
			"JOIN (SELECT * FROM `groups_ancestors` WHERE (groups_ancestors.idGroupChild = ?)) AS ancestors "+
			"ON groups_items.idGroup = ancestors.idGroupAncestor GROUP BY idItem") + "$").
		WithArgs(2).
		WillReturnRows(mock.NewRows([]string{"ID"}))

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
				return database.NewDataStore(db).UserItems().
					Where("idItem = ?", 500). // chapter
					Where("idUser = ?", 4).
					UpdateColumn("sContestStartDate", database.Now()).Error()
			},
			itemID: 15, userID: 4, wantHasAccess: true, wantReason: nil},
		{name: "user's active contest is OK and it is the task's time-limited chapter",
			initFunc: func(db *database.DB) error {
				return database.NewDataStore(db).UserItems().
					Where("idItem = ?", 115). // chapter
					Where("idUser = ?", 5).
					UpdateColumn("sContestStartDate", database.Now()).Error()
			},
			itemID: 15, userID: 5, wantHasAccess: true, wantReason: nil},
		{name: "user's active contest is OK, but it is not an ancestor of the task and the user doesn't have full access to the task's chapter",
			initFunc: func(db *database.DB) error {
				return database.NewDataStore(db).UserItems().
					Where("idItem = ?", 114). // chapter
					Where("idUser = ?", 7).
					UpdateColumn("sContestStartDate", database.Now()).Error()
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
			- {ID: 1, sLogin: 1, idGroupSelf: 101}
			- {ID: 2, sLogin: 2, idGroupSelf: 102}
			- {ID: 3, sLogin: 3, idGroupSelf: 103}
			- {ID: 4, sLogin: 4, idGroupSelf: 104}
			- {ID: 5, sLogin: 5, idGroupSelf: 105}
			- {ID: 6, sLogin: 6, idGroupSelf: 106}
		items: [{ID: 12}, {ID: 13}, {ID: 14, sDuration: 10:00:00}, {ID: 15, sTeamMode: "None"}]
		groups_ancestors:
			- {idGroupAncestor: 101, idGroupChild: 101}
			- {idGroupAncestor: 102, idGroupChild: 102}
			- {idGroupAncestor: 103, idGroupChild: 103}
			- {idGroupAncestor: 104, idGroupChild: 104}
			- {idGroupAncestor: 105, idGroupChild: 105}
			- {idGroupAncestor: 106, idGroupChild: 106}
		users_items:
			- {idUser: 2, idItem: 12} # not started
			- {idUser: 3, idItem: 13, sContestStartDate: 2019-03-22 08:44:55, sFinishDate: 2019-03-23 08:44:55} #finished
			- {idUser: 4, idItem: 14, sContestStartDate: 2019-03-22 08:44:55} # ok
			- {idUser: 5, idItem: 15, sContestStartDate: 2019-04-22 08:44:55} # ok with team mode
			- {idUser: 6, idItem: 14, sContestStartDate: 2019-03-22 08:44:55} # multiple
			- {idUser: 6, idItem: 15, sContestStartDate: 2019-03-22 08:43:55} # multiple
		groups_items:
			- {idGroup: 102, idItem: 12, idUserCreated: 1}
			- {idGroup: 103, idItem: 13, idUserCreated: 1}
			- {idGroup: 104, idItem: 14, sAdditionalTime: 0000-00-00 00:01:00, idUserCreated: 1}
			- {idGroup: 105, idItem: 15, idUserCreated: 1}
			- {idGroup: 106, idItem: 14, sAdditionalTime: 0000-00-00 00:01:00, idUserCreated: 1}
			- {idGroup: 106, idItem: 15, sAdditionalTime: 0000-00-00 00:01:00, idUserCreated: 1}`)
	defer func() { _ = db.Close() }()

	tests := []struct {
		name    string
		userID  int64
		want    *database.ActiveContestInfo
		wantLog string
	}{
		{name: "no item", userID: 1, want: nil},
		{name: "not started", userID: 2, want: nil},
		{name: "finished", userID: 3, want: nil},
		{name: "ok", userID: 4, want: &database.ActiveContestInfo{
			ItemID:            14,
			UserID:            4,
			DurationInSeconds: 36060,
			EndTime:           time.Date(2019, 3, 22, 18, 45, 55, 0, time.UTC),
			StartTime:         time.Date(2019, 3, 22, 8, 44, 55, 0, time.UTC),
		}},
		{name: "ok with team mode", userID: 5, want: &database.ActiveContestInfo{
			ItemID:            15,
			UserID:            5,
			DurationInSeconds: 0,
			EndTime:           time.Date(2019, 4, 22, 8, 44, 55, 0, time.UTC),
			StartTime:         time.Date(2019, 4, 22, 8, 44, 55, 0, time.UTC),
			TeamMode:          ptrString("None"),
		}},
		{
			name: "ok with multiple active contests", userID: 6, want: &database.ActiveContestInfo{
				ItemID:            14,
				UserID:            6,
				DurationInSeconds: 36060,
				EndTime:           time.Date(2019, 3, 22, 18, 45, 55, 0, time.UTC),
				StartTime:         time.Date(2019, 3, 22, 8, 44, 55, 0, time.UTC),
			},
			wantLog: "User with ID = 6 has 2 (>1) active contests",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			store := database.NewDataStore(db)
			user := &database.User{}
			assert.NoError(t, user.LoadByID(store, test.userID))
			hook, restoreLogFunc := logging.MockSharedLoggerHook()
			defer restoreLogFunc()

			got := store.Items().GetActiveContestInfoForUser(user)
			if got != nil && test.want != nil {
				assert.True(t, time.Since(got.Now) < 3*time.Second)
				assert.True(t, time.Since(got.Now) > -3*time.Second)
				test.want.Now = time.Now().UTC()
				got.Now = test.want.Now
			}
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantLog, (&loggingtest.Hook{Hook: hook}).GetAllLogs())
		})
	}
}

func TestItemStore_CloseContest(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		users: [{ID: 1, sLogin: 1, idGroupSelf: 20}]
		groups: [{ID: 20}]
		items: [{ID: 11}, {ID: 12}, {ID: 13}, {ID: 14}, {ID: 15}, {ID: 16}]
		items_ancestors:
			- {idItemAncestor: 11, idItemChild: 12}
			- {idItemAncestor: 11, idItemChild: 13}
			- {idItemAncestor: 11, idItemChild: 14}
			- {idItemAncestor: 11, idItemChild: 15}
			- {idItemAncestor: 11, idItemChild: 16}
		users_items: [{idUser: 1, idItem: 11}, {idUser: 1, idItem: 12}, {idUser: 2, idItem: 11}]
		groups_items:
			- {idGroup: 20, idItem: 11, idUserCreated: 1}
			- {idGroup: 20, idItem: 12, idUserCreated: 1}
			- {idGroup: 20, idItem: 13, sCachedFullAccessDate: 2030-03-22 08:44:55, idUserCreated: 1} # no full access
			- {idGroup: 20, idItem: 14, sCachedFullAccessDate: 2018-03-22 08:44:55, idUserCreated: 1} # full access
			- {idGroup: 20, idItem: 15, bOwnerAccess: 1, idUserCreated: 1}
			- {idGroup: 20, idItem: 16, bManagerAccess: 1, idUserCreated: 1}
			- {idGroup: 21, idItem: 12, idUserCreated: 1}`)
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		user := &database.User{}
		assert.NoError(t, user.LoadByID(store, 1))
		store.Items().CloseContest(11, user)
		return nil
	}))

	type userItemInfo struct {
		UserID        int64 `gorm:"column:idUser"`
		ItemID        int64 `gorm:"column:idItem"`
		FinishDateSet bool  `gorm:"column:finishDateSet"`
	}
	var userItems []userItemInfo
	store := database.NewDataStore(db)
	assert.NoError(t, store.UserItems().
		Select("idUser, idItem, (sFinishDate IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, sFinishDate, NOW())) < 3) AS finishDateSet").
		Order("idUser, idItem").
		Scan(&userItems).Error())
	assert.Equal(t, []userItemInfo{
		{UserID: 1, ItemID: 11, FinishDateSet: true},
		{UserID: 1, ItemID: 12, FinishDateSet: false},
		{UserID: 2, ItemID: 11, FinishDateSet: false},
	}, userItems)

	type groupItemInfo struct {
		GroupID int64 `gorm:"column:idGroup"`
		ItemID  int64 `gorm:"column:idItem"`
	}
	var groupItems []groupItemInfo
	assert.NoError(t, store.GroupItems().Select("idGroup, idItem").
		Order("idGroup, idItem").
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
			- {ID: 1, sLogin: 1, idGroupSelf: 10}
			- {ID: 2, sLogin: 2, idGroupSelf: 20}
			- {ID: 3, sLogin: 3, idGroupSelf: 30}
			- {ID: 4, sLogin: 4, idGroupSelf: 50}
		groups: [{ID: 10}, {ID: 20}, {ID: 30}, {ID: 40, idTeamItem: 11, sType: Team}, {ID: 50}]
		groups_groups:
			- {idGroupParent: 40, idGroupChild: 10, sType: invitationAccepted}
			- {idGroupParent: 40, idGroupChild: 20, sType: requestRefused}
			- {idGroupParent: 40, idGroupChild: 30, sType: requestAccepted}
			- {idGroupParent: 40, idGroupChild: 50, sType: joinedByCode}
		groups_ancestors:
			- {idGroupAncestor: 10, idGroupChild: 10}
			- {idGroupAncestor: 20, idGroupChild: 20}
			- {idGroupAncestor: 30, idGroupChild: 30}
			- {idGroupAncestor: 40, idGroupChild: 10}
			- {idGroupAncestor: 40, idGroupChild: 30}
			- {idGroupAncestor: 40, idGroupChild: 50}
		items: [{ID: 11}, {ID: 12}, {ID: 13}]
		items_ancestors:
			- {idItemAncestor: 11, idItemChild: 12}
			- {idItemAncestor: 11, idItemChild: 13}
		users_items:
			- {idUser: 1, idItem: 11}
			- {idUser: 1, idItem: 12}
			- {idUser: 2, idItem: 11}
			- {idUser: 3, idItem: 11}
			- {idUser: 4, idItem: 11}
		groups_items:
			- {idGroup: 20, idItem: 11, sCachedPartialAccessDate: 2018-03-22 08:44:55,
				sPartialAccessDate: 2018-03-22 08:44:55, bCachedPartialAccess: 1, idUserCreated: 1}
			- {idGroup: 40, idItem: 11, sCachedPartialAccessDate: 2018-03-22 08:44:55,
				sPartialAccessDate: 2018-03-22 08:44:55, bCachedPartialAccess: 1, idUserCreated: 1}
			- {idGroup: 20, idItem: 12, sCachedPartialAccessDate: 2018-03-22 08:44:55,
				sPartialAccessDate: 2018-03-22 08:44:55, bCachedPartialAccess: 1, idUserCreated: 1}
			- {idGroup: 40, idItem: 12, sCachedPartialAccessDate: 2018-03-22 08:44:55,
				sPartialAccessDate: 2018-03-22 08:44:55, bCachedPartialAccess: 1, idUserCreated: 1}
			- {idGroup: 50, idItem: 11, sCachedPartialAccessDate: 2018-03-22 08:44:55,
				sPartialAccessDate: 2018-03-22 08:44:55, bCachedPartialAccess: 1, idUserCreated: 1}
			- {idGroup: 50, idItem: 12, sCachedPartialAccessDate: 2018-03-22 08:44:55,
			   sPartialAccessDate: 2018-03-22 08:44:55, bCachedPartialAccess: 1, idUserCreated: 1}`)
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		user := &database.User{ID: 1, SelfGroupID: ptrInt64(10)}
		store.Items().CloseTeamContest(11, user)
		return nil
	}))

	type userItemInfo struct {
		UserID        int64 `gorm:"column:idUser"`
		ItemID        int64 `gorm:"column:idItem"`
		FinishDateSet bool  `gorm:"column:finishDateSet"`
	}
	var userItems []userItemInfo
	store := database.NewDataStore(db)
	assert.NoError(t, store.UserItems().
		Select("idUser, idItem, (sFinishDate IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, sFinishDate, NOW())) < 3) as finishDateSet").
		Order("idUser, idItem").
		Scan(&userItems).Error())
	assert.Equal(t, []userItemInfo{
		{UserID: 1, ItemID: 11, FinishDateSet: true},
		{UserID: 1, ItemID: 12, FinishDateSet: false},
		{UserID: 2, ItemID: 11, FinishDateSet: false},
		{UserID: 3, ItemID: 11, FinishDateSet: true},
		{UserID: 4, ItemID: 11, FinishDateSet: true},
	}, userItems)

	type groupItemInfo struct {
		GroupID                 int64          `gorm:"column:idGroup"`
		ItemID                  int64          `gorm:"column:idItem"`
		PartialAccessDate       *database.Time `gorm:"column:sPartialAccessDate"`
		CachedPartialAccessDate *database.Time `gorm:"column:sCachedPartialAccessDate"`
		CachedPartialAccess     bool           `gorm:"column:bCachedPartialAccess"`
	}
	var groupItems []groupItemInfo
	assert.NoError(t, store.GroupItems().
		Select("idGroup, idItem, sPartialAccessDate, sCachedPartialAccessDate, bCachedPartialAccess").
		Order("idGroup, idItem").
		Scan(&groupItems).Error())
	expectedDate := (*database.Time)(ptrTime(time.Date(2018, 3, 22, 8, 44, 55, 0, time.UTC)))
	assert.Equal(t, []groupItemInfo{
		{GroupID: 20, ItemID: 11, PartialAccessDate: expectedDate, CachedPartialAccessDate: expectedDate, CachedPartialAccess: true},
		{GroupID: 20, ItemID: 12, PartialAccessDate: expectedDate, CachedPartialAccessDate: expectedDate, CachedPartialAccess: true},
		{GroupID: 40, ItemID: 11, PartialAccessDate: nil, CachedPartialAccessDate: nil, CachedPartialAccess: false},
		{GroupID: 40, ItemID: 12, PartialAccessDate: expectedDate, CachedPartialAccessDate: expectedDate, CachedPartialAccess: true},
		{GroupID: 50, ItemID: 11, PartialAccessDate: expectedDate, CachedPartialAccessDate: expectedDate, CachedPartialAccess: true},
		{GroupID: 50, ItemID: 12, PartialAccessDate: expectedDate, CachedPartialAccessDate: expectedDate, CachedPartialAccess: true},
	}, groupItems)
}

func TestItemStore_Visible_ProvidesAccessSolutions(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		items: [{ID: 11}, {ID: 12}, {ID: 13}]
		users: [{ID: 1, idGroupSelf: 10}]
		groups: [{ID: 10}, {ID: 40}]
		groups_groups:
			- {idGroupParent: 40, idGroupChild: 10}
		groups_ancestors:
			- {idGroupAncestor: 10, idGroupChild: 10}
			- {idGroupAncestor: 40, idGroupChild: 10}
			- {idGroupAncestor: 40, idGroupChild: 40}
		groups_items:
			- {idGroup: 40, idItem: 11, sCachedFullAccessDate: 2018-03-22 08:44:55, sCachedAccessSolutionsDate: 2018-03-22 08:44:55,
		     idUserCreated: 1}
			- {idGroup: 10, idItem: 11, sCachedFullAccessDate: 2018-03-22 08:44:55, sCachedAccessSolutionsDate: 2019-03-22 08:44:55,
			   idUserCreated: 1}
			- {idGroup: 10, idItem: 12, sCachedFullAccessDate: 2018-03-22 08:44:55, sCachedAccessSolutionsDate: 2019-04-22 08:44:55,
			   idUserCreated: 1}
			- {idGroup: 10, idItem: 13, sCachedFullAccessDate: 2018-03-22 08:44:55, idUserCreated: 1}`)
	type resultType struct {
		ID              int64 `gorm:"column:ID"`
		AccessSolutions bool  `gorm:"column:accessSolutions"`
	}
	var result []resultType

	assert.NoError(t, database.NewDataStore(db).Items().
		Visible(&database.User{ID: 1, SelfGroupID: ptrInt64(10)}).
		Select("ID, accessSolutions").Order("ID").Scan(&result).Error())
	assert.Equal(t, []resultType{
		{ID: 11, AccessSolutions: true},
		{ID: 12, AccessSolutions: true},
		{ID: 13, AccessSolutions: false},
	}, result)
}

func TestItemStore_HasManagerAccess(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		items: [{ID: 11}, {ID: 12}, {ID: 13}]
		users: [{ID: 1, sLogin: 1, idGroupSelf: 100}, {ID: 2, sLogin: 2, idGroupSelf: 110}]
		groups: [{ID: 10}, {ID: 11}, {ID: 40}]
		groups_groups:
			- {idGroupParent: 400, idGroupChild: 100}
		groups_ancestors:
			- {idGroupAncestor: 100, idGroupChild: 100}
			- {idGroupAncestor: 110, idGroupChild: 110}
			- {idGroupAncestor: 400, idGroupChild: 100}
			- {idGroupAncestor: 400, idGroupChild: 400}
		groups_items:
			- {idGroup: 400, idItem: 11, bCachedManagerAccess: 1, idUserCreated: 1}
			- {idGroup: 100, idItem: 11, bOwnerAccess: 1, idUserCreated: 1}
			- {idGroup: 100, idItem: 12, idUserCreated: 1}
			- {idGroup: 100, idItem: 13, idUserCreated: 1}
			- {idGroup: 110, idItem: 12, bOwnerAccess: 1, idUserCreated: 1}
			- {idGroup: 110, idItem: 13, bCachedManagerAccess: 1, idUserCreated: 1}`)

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
