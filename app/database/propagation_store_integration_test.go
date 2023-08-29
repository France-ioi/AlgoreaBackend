//go:build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestPropagationStore_AsyncPropagationScheduled_NoPropagationScheduledDefault(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
propagations:
	- {propagation_id: 1}
	`)

	propagationStore := database.NewDataStore(db).Propagations()

	assert.False(t, propagationStore.AsyncPropagationScheduled())
}

func TestPropagationStore_ScheduleAsyncPropagation(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
propagations:
	- {propagation_id: 1}
	`)

	propagationStore := database.NewDataStore(db).Propagations()

	propagationStore.ScheduleAsyncPropagation()
	assert.True(t, propagationStore.AsyncPropagationScheduled())
}
