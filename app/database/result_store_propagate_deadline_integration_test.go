//go:build !unit

package database_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestResultStore_Propagate_SoftDeadlineAlreadyPassed(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixture(testhelpers.CreateTestContext(), "results_propagation/_common")
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	require.NoError(t, dataStore.Exec(`
		UPDATE results_propagate SET state = 'to_be_recomputed'`).Error())

	database.SetPropagationSoftDeadline(db, time.Now().Add(-time.Second))

	err := runResultsPropagation(dataStore)
	require.NoError(t, err)

	assertResultsMarkedAsChanged(t, dataStore, "results_propagate_internal", []resultPrimaryKeyAndState{
		{ResultPrimaryKey: ResultPrimaryKey{ParticipantID: 101, AttemptID: 1, ItemID: 1}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{ParticipantID: 102, AttemptID: 1, ItemID: 2}, State: "to_be_recomputed"},
	})

	expectedLatestActivityAt1 := database.Time(time.Date(2019, 5, 29, 11, 0, 0, 0, time.UTC))
	expectedLatestActivityAt2 := database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))
	assertAggregatesEqual(t, dataStore.Results(), []aggregatesResultRow{
		{ParticipantID: 101, AttemptID: 1, ItemID: 1, State: "to_be_recomputed", LatestActivityAt: expectedLatestActivityAt1},
		{ParticipantID: 101, AttemptID: 1, ItemID: 2, State: "done", LatestActivityAt: expectedLatestActivityAt1},
		{ParticipantID: 102, AttemptID: 1, ItemID: 2, State: "to_be_recomputed", LatestActivityAt: expectedLatestActivityAt2},
	})
}

func TestResultStore_Propagate_SoftDeadlineTripsMidLoopThenResumes(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixture(testhelpers.CreateTestContext(), "results_propagation/_common")
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	require.NoError(t, dataStore.Results().
		Where("participant_id = 101 AND attempt_id = 1 AND item_id = 1").
		UpdateColumns(map[string]interface{}{
			"tasks_tried":    1,
			"score_computed": 10,
		}).Error())

	database.SetPropagationSoftDeadline(db, time.Now().Add(time.Hour))

	oldHook := database.GetBeforePropagationStepHook()
	defer database.SetBeforePropagationStepHook(oldHook)
	var mainStepCalls int
	database.SetBeforePropagationStepHook(func(step database.PropagationStep) {
		if step == database.PropagationStepResultsInsideNamedLockMain {
			mainStepCalls++
			if mainStepCalls == 1 {
				database.SetPropagationSoftDeadline(db, time.Now().Add(-time.Second))
			}
		}
	})

	err := runResultsPropagation(dataStore)
	require.NoError(t, err)
	require.Equal(t, 1, mainStepCalls)

	var remaining int64
	require.NoError(t, dataStore.Table("results_propagate_internal").Count(&remaining).Error())
	require.Positive(t, remaining, "expected queued work after the soft deadline stopped the run")

	// Resume on a fresh DataStore: InTransaction replaces the embedded *DB with a clone that still
	// carries the expired soft deadline from the interrupted run.
	database.SetPropagationSoftDeadline(db, time.Now().Add(time.Hour))
	database.SetBeforePropagationStepHook(oldHook)
	dataStore = database.NewDataStore(db)

	err = dataStore.InTransaction(func(s *database.DataStore) error {
		return s.Results().Propagate()
	})
	require.NoError(t, err)

	expectedLatestActivityAt1 := database.Time(time.Date(2019, 5, 29, 11, 0, 0, 0, time.UTC))
	expectedLatestActivityAt2 := database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))
	assertAggregatesEqual(t, dataStore.Results(), []aggregatesResultRow{
		{
			ParticipantID: 101, AttemptID: 1, ItemID: 1, TasksTried: 1, ScoreComputed: 10, State: "done",
			LatestActivityAt: expectedLatestActivityAt1,
		},
		{
			ParticipantID: 101, AttemptID: 1, ItemID: 2, TasksTried: 1, ScoreComputed: 10, State: "done",
			LatestActivityAt: expectedLatestActivityAt1,
		},
		{ParticipantID: 102, AttemptID: 1, ItemID: 2, State: "done", LatestActivityAt: expectedLatestActivityAt2},
	})
	assertResultsMarkedAsChanged(t, dataStore, "results_propagate", nil)
	assertResultsMarkedAsChanged(t, dataStore, "results_propagate_internal", nil)
}

func TestResultStore_Propagate_SoftDeadlineInTheFuture(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixture(testhelpers.CreateTestContext(), "results_propagation/_common")
	defer func() { _ = db.Close() }()

	database.SetPropagationSoftDeadline(db, time.Now().Add(time.Hour))
	resultStore := database.NewDataStore(db).Results()
	err := runResultsPropagation(resultStore.DataStore)
	require.NoError(t, err)

	expectedLatestActivityAt1 := database.Time(time.Date(2019, 5, 29, 11, 0, 0, 0, time.UTC))
	expectedLatestActivityAt2 := database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))
	assertAggregatesEqual(t, resultStore, []aggregatesResultRow{
		{ParticipantID: 101, AttemptID: 1, ItemID: 1, State: "done", LatestActivityAt: expectedLatestActivityAt1},
		{ParticipantID: 101, AttemptID: 1, ItemID: 2, State: "done", LatestActivityAt: expectedLatestActivityAt1},
		{ParticipantID: 102, AttemptID: 1, ItemID: 2, State: "done", LatestActivityAt: expectedLatestActivityAt2},
	})
	assertResultsMarkedAsChanged(t, resultStore.DataStore, "results_propagate", nil)
}

func TestResultStore_Propagate_SoftDeadlineStopsBeforeRecursionAfterUnlocks(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixture(testhelpers.CreateTestContext(),
		"results_propagation/_common", "results_propagation/unlocks")
	defer func() { _ = db.Close() }()

	prepareDependencies(t, db)
	database.SetPropagationSoftDeadline(db, time.Now().Add(time.Hour))

	oldHook := database.GetBeforePropagationStepHook()
	defer database.SetBeforePropagationStepHook(oldHook)
	database.SetBeforePropagationStepHook(func(step database.PropagationStep) {
		if step == database.PropagationStepResultsInsideNamedLockItemUnlocking {
			// Expire the main budget after the outer-loop check passed. Post-unlock
			// computeAllAccess still gets its own bounded budget so permissions are generated;
			// recursion then stops on the main deadline.
			database.SetPropagationSoftDeadline(db, time.Now().Add(-time.Second))
		}
	})

	dataStore := database.NewDataStore(db)
	err := runResultsPropagation(dataStore)
	require.NoError(t, err)

	var generatedCount int64
	require.NoError(t, dataStore.Permissions().
		Where("group_id = 101 AND item_id IN (1001, 1002, 2001, 2002, 4001, 4002)").
		Count(&generatedCount).Error())
	assert.Positive(t, generatedCount, "post-unlock computeAllAccess must still populate permissions_generated")

	var remaining int64
	require.NoError(t, dataStore.Table("results_propagate_internal").Count(&remaining).Error())
	assert.Positive(t, remaining, "recursion must stop with work still queued")
}

func TestResultStore_ScheduledPropagation_SoftDeadlineEmitsWarning(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, logHook := logging.NewMockLogger()
	ctx := testhelpers.CreateTestContextWithLogger(logger)
	db := testhelpers.SetupDBWithFixture(ctx, "results_propagation/_common")
	defer func() { _ = db.Close() }()

	database.SetPropagationSoftDeadline(db, time.Now().Add(-time.Second))
	dataStore := database.NewDataStore(db)

	require.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
		store.ScheduleResultsPropagation()
		return nil
	}))

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllLogs()
	assert.Contains(t, logs, "propagation stopped early: soft deadline exceeded")
	assert.Contains(t, logs, "results_propagate_internal=")

	var remaining int64
	require.NoError(t, dataStore.Table("results_propagate_internal").Count(&remaining).Error())
	assert.Positive(t, remaining)
}

func TestResultStore_ScheduledPropagation_SoftDeadlineNoWarningWhenQueueEmpty(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, logHook := logging.NewMockLogger()
	ctx := testhelpers.CreateTestContextWithLogger(logger)
	db := testhelpers.SetupDBWithFixture(ctx, "results_propagation/_common")
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	require.NoError(t, dataStore.Exec("DELETE FROM results_propagate").Error())
	database.SetPropagationSoftDeadline(db, time.Now().Add(-time.Second))

	require.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
		store.ScheduleResultsPropagation()
		return nil
	}))

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllLogs()
	assert.NotContains(t, logs, "propagation stopped early")
}

func TestResultStore_Propagate_SlowChunkWarning(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, logHook := logging.NewMockLogger()
	ctx := testhelpers.CreateTestContextWithLogger(logger)
	db := testhelpers.SetupDBWithFixture(ctx, "results_propagation/_common")
	defer func() { _ = db.Close() }()

	restoreThreshold := database.SetResultsPropagationSlowChunkThresholdForTests(time.Nanosecond)
	defer restoreThreshold()
	restoreCounters := database.SetPropagationLogChunkCountersForTests(false)
	defer restoreCounters()

	err := runResultsPropagation(database.NewDataStore(db))
	require.NoError(t, err)

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
	assertPropagationStepLoggedAtLevel(t, logs, "Duration of step of results propagation:", "warning")
	// Observability uses the non-logging path; a missing PROCESS grant must not produce
	// type=db Error spam (may Warn once that dumps are disabled).
	for _, line := range strings.Split(logs, "\n") {
		if strings.Contains(line, "type=db") && strings.Contains(line, "level=error") {
			assert.NotContains(t, line, "INNODB_TRX",
				"INNODB_TRX failure must not emit type=db Error lines")
			assert.NotContains(t, line, "PROCESS privilege",
				"PROCESS Access denied must not emit type=db Error lines")
		}
	}
}

func TestPermissionGrantedStore_ComputeAllAccess_SoftDeadlineStopsBetweenChunks(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, logHook := logging.NewMockLogger()
	ctx := testhelpers.CreateTestContextWithLogger(logger)
	db := testhelpers.SetupDBWithFixture(ctx, "permission_granted_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	database.SetPropagationSoftDeadline(db, time.Now().Add(-time.Second))
	dataStore := database.NewDataStore(db)

	require.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
		store.SchedulePermissionsPropagation()
		return nil
	}))

	hasRows, err := dataStore.Table("permissions_propagate").HasRows()
	require.NoError(t, err)
	assert.True(t, hasRows, "permissions propagation must leave queued work when the soft deadline has already passed")

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllLogs()
	assert.Contains(t, logs, "propagation stopped early: soft deadline exceeded")
	assert.Contains(t, logs, "permissions_propagate=")
}

func TestResultStore_ScheduledPropagation_SoftDeadlineStopsResultsRecomputeForItems(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, logHook := logging.NewMockLogger()
	ctx := testhelpers.CreateTestContextWithLogger(logger)
	db := testhelpers.SetupDBWithFixtureString(ctx, `
		groups: [{id: 1}, {id: 2}]
		items:
			- {id: 111, default_language_tag: fr}
			- {id: 222, default_language_tag: fr}
		items_items:
			- {parent_item_id: 111, child_item_id: 222, child_order: 1}
		items_ancestors:
			- {ancestor_item_id: 111, child_item_id: 222}
		results:
			- {participant_id: 1, attempt_id: 1, item_id: 111, latest_activity_at: '2019-05-30 11:00:00'}
			- {participant_id: 2, attempt_id: 1, item_id: 222, latest_activity_at: '2019-05-30 11:00:00'}
		attempts:
			- {participant_id: 1, id: 1}
			- {participant_id: 2, id: 1}
		results_recompute_for_items:
			- {item_id: 111}
			- {item_id: 222}
	`)
	defer func() { _ = db.Close() }()

	database.SetPropagationSoftDeadline(db, time.Now().Add(time.Hour))
	oldHook := database.GetBeforePropagationStepHook()
	defer database.SetBeforePropagationStepHook(oldHook)
	var insertStepCalls int
	database.SetBeforePropagationStepHook(func(step database.PropagationStep) {
		if step == database.PropagationStepResultsInsideNamedLockInsertIntoResultsPropagateInternal {
			insertStepCalls++
			// Call 1 is before setResults...; call 2 is the first chunk inside the drain loop.
			// Expire after that chunk so the next iteration's deadline check stops between chunks.
			if insertStepCalls == 2 {
				database.SetPropagationSoftDeadline(db, time.Now().Add(-time.Second))
			}
		}
	})

	dataStore := database.NewDataStore(db)
	require.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
		store.ScheduleResultsPropagation()
		return nil
	}))

	hasRows, err := dataStore.Table("results_recompute_for_items").HasRows()
	require.NoError(t, err)
	assert.True(t, hasRows,
		"expected results_recompute_for_items left after a between-chunk stop (marking is resumable)")
	require.GreaterOrEqual(t, insertStepCalls, 2)

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllLogs()
	assert.Contains(t, logs, "propagation stopped early: soft deadline exceeded")
	assert.Contains(t, logs, "results_recompute_for_items=")
}

func assertPropagationStepLoggedAtLevel(t *testing.T, logs, msg, level string) {
	t.Helper()
	for _, line := range strings.Split(logs, "\n") {
		if strings.Contains(line, msg) {
			assert.Regexp(t, `level=`+level+`\b`, line)
			return
		}
	}
	t.Fatalf("log line containing %q not found in:\n%s", msg, logs)
}
