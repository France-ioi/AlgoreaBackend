package items

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"
	"unsafe"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_constructActivityLogQuery_ForcesIndexConditionsSatisfied(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := constructDBMockCheckingAnswersCountQueryInActivityLog(t, itemActivityLogStraightJoinBoundary+1)
	defer func() { _ = db.Close() }()
	sjcFuncCalled, sql := callConstructActivityLogQueryForTestsOfIndexForcing(t, db, true)

	require.True(t, sjcFuncCalled)
	require.NoError(t, mock.ExpectationsWereMet())

	assert.Contains(t, sql, "MYTESTCONDITION")
	assert.Contains(t, sql, "FORCE INDEX ")
}

func Test_constructActivityLogQuery_DoesNotForceIndexWhenConditionsSatisfiedButTheFlagIsFalse(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	sjcFuncCalled, sql := callConstructActivityLogQueryForTestsOfIndexForcing(t, db, false)

	require.False(t, sjcFuncCalled)
	require.NoError(t, mock.ExpectationsWereMet())
	assert.NotContains(t, sql, "MYTESTCONDITION")
	assert.NotContains(t, sql, "FORCE INDEX ")
}

func Test_constructActivityLogQuery_DoesNotForceIndexWhenConditionsNotSatisfied(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := constructDBMockCheckingAnswersCountQueryInActivityLog(t, itemActivityLogStraightJoinBoundary)
	defer func() { _ = db.Close() }()

	sjcFuncCalled, sql := callConstructActivityLogQueryForTestsOfIndexForcing(t, db, true)

	require.False(t, sjcFuncCalled)
	require.NoError(t, mock.ExpectationsWereMet())
	assert.NotContains(t, sql, "MYTESTCONDITION")
	assert.NotContains(t, sql, "FORCE INDEX ")
}

func constructDBMockCheckingAnswersCountQueryInActivityLog(t *testing.T, expectedCount int) (*database.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock := database.NewDBMock()
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT count(*) AS cnt FROM `answers`")).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(expectedCount))

	return db, mock
}

func callConstructActivityLogQueryForTestsOfIndexForcing(
	t *testing.T, db *database.DB, doForceIndex bool,
) (sjcFuncCalled bool, sql string) {
	t.Helper()

	query := constructActivityLogQuery(database.NewDataStore(db), &http.Request{URL: &url.URL{}}, &database.User{GroupID: 1}, nil,
		func(q *database.DB) *database.DB { return q }, func(q *database.DB) *database.DB { return q },
		func(q *database.DB) *database.DB { sjcFuncCalled = true; return q.Where("MYTESTCONDITION") },
		func(q *database.DB) *database.DB { return q }, func(q *database.DB) *database.DB { return q },
		func(q *database.DB) *database.DB { return q },
		"NULL AS can_watch_answer", "-at", doForceIndex)
	queryExpr, ok := query.QueryExpr().(*gorm.SqlExpr)
	require.True(t, ok)
	require.NotNil(t, queryExpr)
	sql = (*struct{ sql string })(unsafe.Pointer(queryExpr)).sql //nolint:gosec // here we extract the SQL string for test verification

	return sjcFuncCalled, sql
}
