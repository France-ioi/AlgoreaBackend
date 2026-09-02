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
