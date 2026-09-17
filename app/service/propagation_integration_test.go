//go:build !unit

package service_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/event"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestSchedulePropagation(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	tests := []struct {
		name           string
		async          bool
		withDispatcher bool
		withNoop       bool
		loggedError    string
		propagated     bool
		expectEvent    bool
	}{
		{
			name:       "async=false propagates rows and dispatches no events",
			async:      false,
			propagated: true,
		},
		{
			name:           "async=true with dispatcher dispatches propagation_requested and skips DB propagation",
			async:          true,
			withDispatcher: true,
			propagated:     false,
			expectEvent:    true,
		},
		{
			name:        "async=true with NoopDispatcher logs an error and skips DB propagation",
			async:       true,
			withNoop:    true,
			propagated:  false,
			loggedError: "propagation.async is enabled but no event dispatcher is configured: propagation not scheduled",
		},
	}

	ctx := testhelpers.CreateTestContext()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			logger, logHook := logging.NewMockLogger()
			ctx := logging.ContextWithLogger(ctx, logger)

			var mockDispatcher *event.MockDispatcher
			switch {
			case tt.withDispatcher:
				mockDispatcher = event.NewMockDispatcher()
				ctx = event.ContextWithDispatcher(ctx, mockDispatcher)
			case tt.withNoop:
				// Mirrors production: NewDispatcherFromConfig returns NoopDispatcher when unset.
				ctx = event.ContextWithDispatcher(ctx, &event.NoopDispatcher{})
			}

			db := testhelpers.SetupDBWithFixtureString(ctx, `
				groups:
					- {id: 1, type: Class}
				items:
					- {id: 1, type: Task, default_language_tag: fr}
				permissions_granted:
					- {group_id: 1, item_id: 1, source_group_id: 1, origin: "group_membership", can_view: "content"}
			`)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)

			service.SchedulePropagation(store, tt.async, []string{"permissions"})

			exists, err := store.Permissions().Where("item_id = 1").HasRows()
			require.NoError(t, err)
			assert.Equal(t, tt.propagated, exists)

			if tt.expectEvent {
				require.NotNil(t, mockDispatcher)
				events := mockDispatcher.GetEvents()
				require.Len(t, events, 1)
				assert.Equal(t, event.TypePropagationRequested, events[0].Type)
				assert.Equal(t, map[string]interface{}{"types": []string{"permissions"}}, events[0].Payload)
			} else if mockDispatcher != nil {
				assert.Empty(t, mockDispatcher.GetEvents())
			}

			if tt.loggedError != "" {
				logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
				assert.Regexp(t, "level=error .* "+regexp.QuoteMeta("msg=\""+tt.loggedError+"\""), logs)
			}
		})
	}
}
