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
	t *testing.T, db *database.DB, doForceIndex, ignoreCreatedAtIndexOnAnswersWhenNotForced bool,
) (sjcFuncCalled bool, sql string) {
	t.Helper()

	query := constructActivityLogQuery(database.NewDataStore(db), &http.Request{URL: &url.URL{}}, &database.User{GroupID: 1}, nil,
		func(q *database.DB) *database.DB { return q }, func(q *database.DB) *database.DB { return q },
		func(q *database.DB) *database.DB { sjcFuncCalled = true; return q.Where("MYTESTCONDITION") },
		func(q *database.DB) *database.DB { return q }, func(q *database.DB) *database.DB { return q },
		func(q *database.DB) *database.DB { return q },
		"NULL AS can_watch_answer", "-at", doForceIndex, ignoreCreatedAtIndexOnAnswersWhenNotForced)
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

	filters := constructItemActivityLogFilters(participantID, store.Raw("SELECT 100 AS id"), nil, &itemID, nil)

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

	filters := constructItemActivityLogFilters(participantID, store.Raw("SELECT 100 AS id"), nil, nil, nil)

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
		participantID, store.Raw("SELECT 100 AS id"), store.Raw("SELECT 1 AS id"), nil, nil)

	answersSQL := sqlFromQueryExpr(t, filters.answersMandatory(store.Answers().DB))
	assert.Contains(t, answersSQL, "answers.participant_id IN (SELECT id FROM participants)")
	assert.Contains(t, answersSQL, "answers.item_id IN (SELECT id FROM items_to_show)")
	assert.NotContains(t, answersSQL, "answers.participant_id =")
	assert.NotContains(t, answersSQL, "JOIN items_to_show")

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_constructItemActivityLogFilters_AttemptIDAddsEquality(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	store := database.NewDataStore(db)
	participantID := int64(42)
	itemID := int64(100)
	attemptID := int64(7)

	filters := constructItemActivityLogFilters(
		participantID, store.Raw("SELECT 100 AS id"), nil, &itemID, &attemptID)
	_ = filters.addWithTables(store.Raw("SELECT 1"))

	for _, sql := range []string{
		sqlFromQueryExpr(t, filters.answersMandatory(store.Answers().DB)),
		sqlFromQueryExpr(t, filters.answersStraightJoin(store.Answers().DB)),
		sqlFromQueryExpr(t, filters.startedResults(store.Table("results AS started_results"))),
		sqlFromQueryExpr(t, filters.validatedResults(store.Table("results AS validated_results"))),
	} {
		assert.Contains(t, sql, "attempt_id =")
	}

	multiItemFilters := constructItemActivityLogFilters(
		participantID, store.Raw("SELECT 100 AS id"), nil, nil, &attemptID)
	multiStraightJoinSQL := sqlFromQueryExpr(t, multiItemFilters.answersStraightJoin(store.Answers().DB))
	assert.Contains(t, multiStraightJoinSQL, "answers.attempt_id =")
	assert.Contains(t, multiStraightJoinSQL, "answers.item_id <=")

	watchedFilters := constructItemActivityLogFilters(
		participantID, store.Raw("SELECT 100 AS id"), store.Raw("SELECT 1 AS id"), nil, &attemptID)
	_ = watchedFilters.addWithTables(store.Raw("SELECT 1"))
	watchedSQL := sqlFromQueryExpr(t, watchedFilters.answersMandatory(store.Answers().DB))
	assert.Contains(t, watchedSQL, "answers.attempt_id =")
	assert.Contains(t, sqlFromQueryExpr(t, watchedFilters.answersStraightJoin(store.Answers().DB)), "answers.attempt_id =")
	assert.Contains(t, sqlFromQueryExpr(t, watchedFilters.startedResults(store.Table("results AS started_results"))),
		"started_results.attempt_id =")
	assert.Contains(t, sqlFromQueryExpr(t, watchedFilters.validatedResults(store.Table("results AS validated_results"))),
		"validated_results.attempt_id =")

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_parseOptionalTaskAttemptID(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	t.Run("no item", func(t *testing.T) {
		testoutput.SuppressIfPasses(t)
		db, mock := database.NewDBMock()
		defer func() { _ = db.Close() }()
		req, err := http.NewRequest(http.MethodGet, "/items/log?attempt_id=1", http.NoBody)
		require.NoError(t, err)

		attemptID, isSet, err := parseOptionalTaskAttemptID(database.NewDataStore(db), req, nil)
		require.NoError(t, err)
		assert.False(t, isSet)
		assert.Zero(t, attemptID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("absent", func(t *testing.T) {
		testoutput.SuppressIfPasses(t)
		db, mock := database.NewDBMock()
		defer func() { _ = db.Close() }()
		itemID := int64(500)
		req, err := http.NewRequest(http.MethodGet, "/items/500/log", http.NoBody)
		require.NoError(t, err)

		attemptID, isSet, err := parseOptionalTaskAttemptID(database.NewDataStore(db), req, &itemID)
		require.NoError(t, err)
		assert.False(t, isSet)
		assert.Zero(t, attemptID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid", func(t *testing.T) {
		testoutput.SuppressIfPasses(t)
		db, mock := database.NewDBMock()
		defer func() { _ = db.Close() }()
		itemID := int64(500)
		req, err := http.NewRequest(http.MethodGet, "/items/500/log?attempt_id=abc", http.NoBody)
		require.NoError(t, err)

		attemptID, isSet, err := parseOptionalTaskAttemptID(database.NewDataStore(db), req, &itemID)
		require.Error(t, err)
		assert.False(t, isSet)
		assert.Zero(t, attemptID)
		assert.Contains(t, err.Error(), "wrong value for attempt_id")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("task", func(t *testing.T) {
		testoutput.SuppressIfPasses(t)
		db, mock := database.NewDBMock()
		defer func() { _ = db.Close() }()
		itemID := int64(500)
		mock.ExpectQuery(regexp.QuoteMeta(
			"SELECT 1 FROM `items`  WHERE (items.id = ?) AND (items.type = 'Task') LIMIT 1")).
			WithArgs(itemID).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
		req, err := http.NewRequest(http.MethodGet, "/items/500/log?attempt_id=3", http.NoBody)
		require.NoError(t, err)

		attemptID, isSet, err := parseOptionalTaskAttemptID(database.NewDataStore(db), req, &itemID)
		require.NoError(t, err)
		assert.True(t, isSet)
		assert.Equal(t, int64(3), attemptID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("non-task", func(t *testing.T) {
		testoutput.SuppressIfPasses(t)
		db, mock := database.NewDBMock()
		defer func() { _ = db.Close() }()
		itemID := int64(600)
		mock.ExpectQuery(regexp.QuoteMeta(
			"SELECT 1 FROM `items`  WHERE (items.id = ?) AND (items.type = 'Task') LIMIT 1")).
			WithArgs(itemID).
			WillReturnRows(sqlmock.NewRows([]string{"1"}))
		req, err := http.NewRequest(http.MethodGet, "/items/600/log?attempt_id=1", http.NoBody)
		require.NoError(t, err)

		attemptID, isSet, err := parseOptionalTaskAttemptID(database.NewDataStore(db), req, &itemID)
		require.Error(t, err)
		assert.False(t, isSet)
		assert.Zero(t, attemptID)
		assert.Contains(t, err.Error(), "attempt_id is only allowed when the item is a task")
		require.NoError(t, mock.ExpectationsWereMet())
	})
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
