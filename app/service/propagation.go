package service

import (
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/event"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// SchedulePropagation schedules propagation of the given types.
// When async, it dispatches a propagation_requested event (best-effort, picked up by the
// worker via SQS/EventBridge); otherwise propagation runs synchronously in a transaction.
func SchedulePropagation(store *database.DataStore, async bool, types []string) {
	if async {
		ctx := store.GetContext()
		// event.Dispatch is a silent no-op with NoopDispatcher (production default when
		// event.dispatcher is unset): surface the misconfiguration instead of dropping work.
		if !event.HasActiveDispatcher(ctx) {
			logging.EntryFromContext(ctx).Error(
				"propagation.async is enabled but no event dispatcher is configured: propagation not scheduled")
			return
		}
		event.Dispatch(ctx, event.TypePropagationRequested, map[string]interface{}{"types": types})
		return
	}

	// Sync.
	MustNotBeError(store.InTransaction(func(store *database.DataStore) error {
		store.SchedulePermissionsPropagation()
		store.ScheduleResultsPropagation()

		return nil
	}))
}
