package groups

import (
	"bytes"
	"io"
	"math"
	"net/http"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/servicetest"
)

func TestErrIfProgressZIPEntriesExceedLimit(t *testing.T) {
	old := maxProgressZIPEntries
	t.Cleanup(func() { maxProgressZIPEntries = old })
	maxProgressZIPEntries = 100

	require.NoError(t, errIfProgressZIPEntriesExceedLimit(0, 1000))
	require.NoError(t, errIfProgressZIPEntriesExceedLimit(10, 10))
	require.NoError(t, errIfProgressZIPEntriesExceedLimit(100, 1))
	err := errIfProgressZIPEntriesExceedLimit(10, 11)
	require.ErrorIs(t, err, ErrTooManyProgressZIPEntries)
	assert.Contains(t, err.Error(), "100000")
}

func TestMaxVisibleItemsForProgressZIP(t *testing.T) {
	old := maxProgressZIPEntries
	t.Cleanup(func() { maxProgressZIPEntries = old })
	maxProgressZIPEntries = 100000

	assert.Equal(t, int64(math.MaxInt64), maxVisibleItemsForProgressZIP(0))
	assert.Equal(t, int64(100000), maxVisibleItemsForProgressZIP(1))
	assert.Equal(t, int64(50000), maxVisibleItemsForProgressZIP(2))
	assert.Equal(t, int64(33333), maxVisibleItemsForProgressZIP(3))
}

func TestProgressZIPLimitError(t *testing.T) {
	old := maxProgressZIPEntries
	t.Cleanup(func() { maxProgressZIPEntries = old })
	maxProgressZIPEntries = 100

	// Zero users: product is always 0, even with a huge item count.
	require.NoError(t, progressZIPLimitError(0, 50))
	require.NoError(t, progressZIPLimitError(0, 101))
	require.NoError(t, progressZIPLimitError(0, 1_000_000))
	require.NoError(t, progressZIPLimitError(10, 10))
	require.ErrorIs(t, progressZIPLimitError(10, 11), ErrTooManyProgressZIPEntries)
	maxProgressZIPEntries = 10
	require.ErrorIs(t, errIfProgressZIPEntriesExceedLimit(3, 4), ErrTooManyProgressZIPEntries)
}

func TestCheckGroupProgressZIPLimits_EmptyParents(t *testing.T) {
	require.NoError(t, checkGroupProgressZIPLimits(nil, nil, 1, nil))
	require.NoError(t, checkGroupProgressZIPLimits(nil, nil, 1, []int64{}))
}

func TestCheckGroupProgressZIPLimits_Exceeds(t *testing.T) {
	oldUsers, oldBuild := getProgressZIPUsersImpl, buildProgressZIPVisibleSubtreeImpl
	t.Cleanup(func() {
		getProgressZIPUsersImpl, buildProgressZIPVisibleSubtreeImpl = oldUsers, oldBuild
	})
	getProgressZIPUsersImpl = func(*database.DataStore, int64) []progressZIPUser {
		return []progressZIPUser{{GroupID: 1}, {GroupID: 2}}
	}
	buildProgressZIPVisibleSubtreeImpl = func(
		*database.DataStore, *database.User, []int64, int64,
	) ([]progressZIPSubtreeItem, int) {
		return nil, 100000 // 2×100000 > limit
	}

	err := checkGroupProgressZIPLimits(nil, &database.User{}, 11, []int64{6000})
	require.ErrorIs(t, err, ErrTooManyProgressZIPEntries)
}

func TestGenerateGroupProgressWithAnswersZIP_Exceeds(t *testing.T) {
	oldUsers, oldBuild := getProgressZIPUsersImpl, buildProgressZIPVisibleSubtreeImpl
	t.Cleanup(func() {
		getProgressZIPUsersImpl, buildProgressZIPVisibleSubtreeImpl = oldUsers, oldBuild
	})
	getProgressZIPUsersImpl = func(*database.DataStore, int64) []progressZIPUser {
		return []progressZIPUser{{GroupID: 1}}
	}
	buildProgressZIPVisibleSubtreeImpl = func(
		*database.DataStore, *database.User, []int64, int64,
	) ([]progressZIPSubtreeItem, int) {
		return nil, int(maxProgressZIPEntries) + 1
	}

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	mock.ExpectQuery("^SELECT (.+) FROM `groups`").
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Our Class"))

	_, err := GenerateGroupProgressWithAnswersZIP(
		bytes.NewBuffer(nil), database.NewDataStore(db), &database.User{GroupID: 21}, 11, []int64{6000},
	)
	require.ErrorIs(t, err, ErrTooManyProgressZIPEntries)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBuildProgressZIPVisibleSubtree_StopsWhenOverMaxItems(t *testing.T) {
	oldLoad := loadProgressZIPChildrenByParentImpl
	t.Cleanup(func() { loadProgressZIPChildrenByParentImpl = oldLoad })
	loadProgressZIPChildrenByParentImpl = func(
		*database.DataStore, []int64, interface{}, int64,
	) (map[int64][]progressZIPItemChild, int) {
		return map[int64][]progressZIPItemChild{}, 3
	}

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	// Permissions subquery builds SQL only; enum load may still hit the DB.
	expectPermissionEnumLoads(mock)

	subtree, count := buildProgressZIPVisibleSubtree(
		database.NewDataStore(db), &database.User{GroupID: 21}, []int64{6000}, 2,
	)
	assert.Nil(t, subtree)
	assert.Equal(t, 3, count)
}

func TestLoadProgressZIPChildrenByParent_StopsWhenOverMaxItems(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	// Enums may already be cached from earlier tests; only the items_items query is required.
	mock.ExpectQuery("^SELECT (.+) FROM `items_items`").
		WillReturnRows(sqlmock.NewRows([]string{"parent_item_id", "child_item_id", "child_order"}).
			AddRow(int64(6000), int64(6001), int32(1)))

	store := database.NewDataStore(db)
	// Build a lightweight subquery without triggering permission enum SQL when enums are warm.
	permissionsSubQuery := store.Table("permissions_generated").Select("item_id").SubQuery()

	_, count := loadProgressZIPChildrenByParent(store, []int64{6000}, permissionsSubQuery, 1)
	assert.Equal(t, 2, count) // parent + new child exceeds maxItems=1
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGroupProgressWithAnswersZIP_TooManyEntries(t *testing.T) {
	monkey.Patch((*database.User).CanWatchGroupMembers, func(*database.User, *database.DataStore, int64) bool {
		return true
	})
	monkey.Patch(resolveAndCheckParentIDs,
		func(*database.DataStore, *http.Request, *database.User, string) ([]int64, error) {
			return []int64{6000}, nil
		})
	monkey.Patch(GenerateGroupProgressWithAnswersZIP,
		func(io.Writer, *database.DataStore, *database.User, int64, []int64) (GroupProgressZIPMeta, error) {
			return GroupProgressZIPMeta{}, ErrTooManyProgressZIPEntries
		})
	defer monkey.UnpatchAll()

	response, mock, logs, err := servicetest.GetResponseForRouteWithMockedDBAndUser(
		"GET", "/groups/11/group-progress-with-answers-zip?parent_item_ids=6000", "",
		&database.User{GroupID: 21},
		func(sqlmock.Sqlmock) {},
		func(router *chi.Mux, baseService *service.Base) {
			srv := &Service{Base: baseService}
			router.Get("/groups/{group_id}/group-progress-with-answers-zip",
				service.AppHandler(srv.getGroupProgressWithAnswersZIP).ServeHTTP)
		},
	)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, response.StatusCode, logs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGenerateGroupResultsToken_TooManyEntries(t *testing.T) {
	monkey.Patch((*database.User).CanWatchGroupMembers, func(*database.User, *database.DataStore, int64) bool {
		return true
	})
	monkey.Patch(resolveAndCheckParentIDs,
		func(*database.DataStore, *http.Request, *database.User, string) ([]int64, error) {
			return []int64{6000}, nil
		})
	monkey.Patch(checkGroupProgressZIPLimits,
		func(*database.DataStore, *database.User, int64, []int64) error {
			return ErrTooManyProgressZIPEntries
		})
	defer monkey.UnpatchAll()

	response, mock, logs, err := servicetest.GetResponseForRouteWithMockedDBAndUser(
		"POST", "/groups/11/group-results-token?parent_item_ids=6000", "",
		&database.User{GroupID: 21},
		func(sqlmock.Sqlmock) {},
		func(router *chi.Mux, baseService *service.Base) {
			srv := &Service{Base: baseService}
			router.Post("/groups/{group_id}/group-results-token",
				service.AppHandler(srv.generateGroupResultsToken).ServeHTTP)
		},
	)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, response.StatusCode, logs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func expectPermissionEnumLoads(mock sqlmock.Sqlmock) {
	// First WherePermissionIsAtLeast loads all DB enums from information_schema.
	for range 7 {
		mock.ExpectQuery("^SELECT SUBSTRING\\(COLUMN_TYPE").
			WillReturnRows(sqlmock.NewRows([]string{"v"}).
				AddRow("'none','info','content','content_with_descendants','solution'"))
	}
}
