package database

import "time"

// Test-only exports for package database_test (integration tests).

func SetRetriesTimeBudgetForTests(d time.Duration) {
	setRetriesTimeBudget(d)
}

func ResetRetriesTimeBudgetForTests() {
	resetRetriesTimeBudget()
}

func GetSessionParamsForTests() map[string]string {
	return getSessionParams()
}

func SetResultsPropagationSlowChunkThresholdForTests(d time.Duration) (restore func()) {
	old := getResultsPropagationSlowChunkThreshold()
	setResultsPropagationSlowChunkThreshold(d)
	return func() { setResultsPropagationSlowChunkThreshold(old) }
}

func SetPropagationLogChunkCountersForTests(enabled bool) (restore func()) {
	old := getPropagationLogChunkCounters()
	SetPropagationLogChunkCounters(enabled)
	resetPropagationObservabilityState()
	return func() {
		SetPropagationLogChunkCounters(old)
		resetPropagationObservabilityState()
	}
}

func ResetPropagationObservabilityStateForTests() {
	resetPropagationObservabilityState()
}

func SetResultsPropagateDrainChunkSizeForTests(size int64) (restore func()) {
	old := getResultsPropagateDrainChunkSize()
	setResultsPropagateDrainChunkSize(size)
	return func() { setResultsPropagateDrainChunkSize(old) }
}

// ResultsPropagateDrainChunkPhase is the test-visible name for resultsPropagateDrainChunkPhase.
type ResultsPropagateDrainChunkPhase = resultsPropagateDrainChunkPhase

const (
	ResultsPropagateDrainChunkAfterInsert = resultsPropagateDrainChunkAfterInsert
	ResultsPropagateDrainChunkAfterDelete = resultsPropagateDrainChunkAfterDelete
	ResultsPropagateDrainChunkSelectEmpty = resultsPropagateDrainChunkSelectEmpty
)

func SetResultsPropagateDrainInsideChunkHookForTests(
	hook func(store *DataStore, fromID, toID int64, phase ResultsPropagateDrainChunkPhase),
) (restore func()) {
	old := getResultsPropagateDrainInsideChunkHook()
	setResultsPropagateDrainInsideChunkHook(hook)
	return func() { setResultsPropagateDrainInsideChunkHook(old) }
}
