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

func TestPropagationStore_AsyncPropagationDone_TriggerNeverCalledTwiceWithoutPropagationDone(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
propagations:
	- {propagation_id: 1}
	`)

	propagationStore := database.NewDataStore(db).Propagations()

	// No propagation scheduled at start.
	assert.False(t, propagationStore.AsyncPropagationScheduled())
	nbScheduled := propagationStore.GetScheduledCounter()

	propagationStore.ScheduleAsyncPropagation()
	expectedNbScheduled := nbScheduled + 1

	// The trigger should have been called and incremented the counter.
	assert.Equal(t, expectedNbScheduled, propagationStore.GetScheduledCounter())
	assert.True(t, propagationStore.AsyncPropagationScheduled())

	propagationStore.ScheduleAsyncPropagation()
	propagationStore.ScheduleAsyncPropagation()

	// The trigger shouldn't have been called again because the propagation wasn't done yet, but it is still scheduled.
	assert.Equal(t, expectedNbScheduled, propagationStore.GetScheduledCounter())
	assert.True(t, propagationStore.AsyncPropagationScheduled())

	propagationStore.AsyncPropagationDone()

	// The counter should be the same until we schedule the next propagation because the trigger hasn't been called.
	// Propagation has been done and is not scheduled anymore.
	assert.Equal(t, expectedNbScheduled, propagationStore.GetScheduledCounter())
	assert.False(t, propagationStore.AsyncPropagationScheduled())

	propagationStore.ScheduleAsyncPropagation()
	expectedNbScheduled++

	// The trigger was called because a propagation was scheduled after it was done.
	assert.Equal(t, expectedNbScheduled, propagationStore.GetScheduledCounter())
}
