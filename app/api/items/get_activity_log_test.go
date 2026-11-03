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
	sjcFuncCalled, sql := callConstructActivityLogQueryForTestsOfIndexForcing(t, db, true, true)

	require.True(t, sjcFuncCalled)
	require.NoError(t, mock.ExpectationsWereMet())

	assert.Contains(t, sql, "MYTESTCONDITION")
	assert.Contains(t, sql, "FORCE INDEX ("+answersCreatedAtActivityLogIndex+")")
	assert.NotContains(t, sql, "IGNORE INDEX ("+answersCreatedAtActivityLogIndex+")")
}

func Test_constructActivityLogQuery_DoesNotForceIndexWhenConditionsSatisfiedButTheFlagIsFalse(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	sjcFuncCalled, sql := callConstructActivityLogQueryForTestsOfIndexForcing(t, db, false, false)

	require.False(t, sjcFuncCalled)
	require.NoError(t, mock.ExpectationsWereMet())
	assert.NotContains(t, sql, "MYTESTCONDITION")
	assert.NotContains(t, sql, "FORCE INDEX ")
	assert.NotContains(t, sql, "IGNORE INDEX ")
}

func Test_constructActivityLogQuery_IgnoresCreatedAtIndexOnParticipantOnlySmallPath(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := constructDBMockCheckingAnswersCountQueryInActivityLog(t, itemActivityLogStraightJoinBoundary)
	defer func() { _ = db.Close() }()

	sjcFuncCalled, sql := callConstructActivityLogQueryForTestsOfIndexForcing(t, db, true, true)

	require.False(t, sjcFuncCalled)
	require.NoError(t, mock.ExpectationsWereMet())
	assert.NotContains(t, sql, "MYTESTCONDITION")
	assert.NotContains(t, sql, "FORCE INDEX ("+answersCreatedAtActivityLogIndex+")")
	assert.Contains(t, sql, "IGNORE INDEX ("+answersCreatedAtActivityLogIndex+")")
}

func Test_constructActivityLogQuery_DoesNotIgnoreCreatedAtIndexOnWatchedGroupSmallPath(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := constructDBMockCheckingAnswersCountQueryInActivityLog(t, itemActivityLogStraightJoinBoundary)
	defer func() { _ = db.Close() }()

	sjcFuncCalled, sql := callConstructActivityLogQueryForTestsOfIndexForcing(t, db, true, false)

	require.False(t, sjcFuncCalled)
	require.NoError(t, mock.ExpectationsWereMet())
	assert.NotContains(t, sql, "MYTESTCONDITION")
	assert.NotContains(t, sql, "FORCE INDEX ")
	assert.NotContains(t, sql, "IGNORE INDEX ")
}

func constructDBMockCheckingAnswersCountQueryInActivityLog(t *testing.T, expectedCount int) (*database.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock := database.NewDBMock()
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT count(*) AS cnt FROM `answers`")).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(expectedCount))

	return db, mock
}

func callConstructActivityLogQueryForTestsOfIndexForcing(
	t *testing.T, db *database.DB, doForceIndex, ignoreCreatedAtIndexOnSmallPath bool,
) (sjcFuncCalled bool, sql string) {
	t.Helper()

	query := constructActivityLogQuery(database.NewDataStore(db), &http.Request{URL: &url.URL{}}, &database.User{GroupID: 1}, nil,
		func(q *database.DB) *database.DB { return q }, func(q *database.DB) *database.DB { return q },
		func(q *database.DB) *database.DB { sjcFuncCalled = true; return q.Where("MYTESTCONDITION") },
		func(q *database.DB) *database.DB { return q }, func(q *database.DB) *database.DB { return q },
		func(q *database.DB) *database.DB { return q },
		"NULL AS can_watch_answer", "-at", doForceIndex, ignoreCreatedAtIndexOnSmallPath)
	sql = sqlFromQueryExpr(t, query)

	return sjcFuncCalled, sql
}

func Test_constructItemActivityLogFilters_ParticipantOnlySingletonUsesEqualities(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	store := database.NewDataStore(db)
	participantID := int64(42)
	itemID := int64(100)

	filters := constructItemActivityLogFilters(participantID, store.Raw("SELECT 100 AS id"), nil, &itemID)

	assertParticipantOnlySingletonFilterSQL(t, filters.answersMandatory(store.Answers().DB),
		"answers.participant_id", "answers.item_id")
	assertParticipantOnlySingletonFilterSQL(t, filters.startedResults(store.Table("results AS started_results")),
		"started_results.participant_id", "started_results.item_id")
	assertParticipantOnlySingletonFilterSQL(t, filters.validatedResults(store.Table("results AS validated_results")),
		"validated_results.participant_id", "validated_results.item_id")

	straightJoinSQL := sqlFromQueryExpr(t, filters.answersStraightJoin(store.Answers().DB))
	assert.Contains(t, straightJoinSQL, "answers.participant_id =")
	assert.Contains(t, straightJoinSQL, "answers.item_id =")
	assert.NotContains(t, straightJoinSQL, "IN (SELECT id FROM participants)")
	assert.NotContains(t, straightJoinSQL, "IN (SELECT id FROM items_to_show)")

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_constructItemActivityLogFilters_ParticipantOnlyMultiItemJoinsItemsToShow(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	store := database.NewDataStore(db)
	participantID := int64(42)

	filters := constructItemActivityLogFilters(participantID, store.Raw("SELECT 100 AS id"), nil, nil)

	assertParticipantOnlyMultiItemFilterSQL(t, filters.answersMandatory(store.Answers().DB), "answers")
	assertParticipantOnlyMultiItemFilterSQL(t, filters.startedResults(store.Table("results AS started_results")), "started_results")
	assertParticipantOnlyMultiItemFilterSQL(t, filters.validatedResults(store.Table("results AS validated_results")), "validated_results")

	straightJoinSQL := sqlFromQueryExpr(t, filters.answersStraightJoin(store.Answers().DB))
	assert.Contains(t, straightJoinSQL, "answers.participant_id =")
	assert.Contains(t, straightJoinSQL, "answers.item_id <=")
	assert.Contains(t, straightJoinSQL, "answers.item_id >=")
	assert.NotContains(t, straightJoinSQL, "IN (SELECT id FROM participants)")
	assert.NotContains(t, straightJoinSQL, "JOIN items_to_show")

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_constructItemActivityLogFilters_WatchedGroupKeepsCTEInPredicates(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	store := database.NewDataStore(db)
	participantID := int64(42)

	filters := constructItemActivityLogFilters(
		participantID, store.Raw("SELECT 100 AS id"), store.Raw("SELECT 1 AS id"), nil)

	answersSQL := sqlFromQueryExpr(t, filters.answersMandatory(store.Answers().DB))
	assert.Contains(t, answersSQL, "answers.participant_id IN (SELECT id FROM participants)")
	assert.Contains(t, answersSQL, "answers.item_id IN (SELECT id FROM items_to_show)")
	assert.NotContains(t, answersSQL, "answers.participant_id =")
	assert.NotContains(t, answersSQL, "JOIN items_to_show")

	require.NoError(t, mock.ExpectationsWereMet())
}

func assertParticipantOnlySingletonFilterSQL(t *testing.T, query *database.DB, participantColumn, itemColumn string) {
	t.Helper()
	sql := sqlFromQueryExpr(t, query)
	assert.Contains(t, sql, participantColumn+" =")
	assert.Contains(t, sql, itemColumn+" =")
	assert.NotContains(t, sql, participantColumn+" IN (SELECT id FROM participants)")
	assert.NotContains(t, sql, itemColumn+" IN (SELECT id FROM items_to_show)")
}

func assertParticipantOnlyMultiItemFilterSQL(t *testing.T, query *database.DB, tableAlias string) {
	t.Helper()
	sql := sqlFromQueryExpr(t, query)
	assert.Contains(t, sql, tableAlias+".participant_id =")
	assert.Contains(t, sql, "JOIN items_to_show")
	assert.NotContains(t, sql, tableAlias+".participant_id IN (SELECT id FROM participants)")
	assert.NotContains(t, sql, tableAlias+".item_id IN (SELECT id FROM items_to_show)")
}

func sqlFromQueryExpr(t *testing.T, query *database.DB) string {
	t.Helper()
	queryExpr, ok := query.QueryExpr().(*gorm.SqlExpr)
	require.True(t, ok)
	require.NotNil(t, queryExpr)
	return (*struct{ sql string })(unsafe.Pointer(queryExpr)).sql //nolint:gosec // extract SQL for test verification
}
