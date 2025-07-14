package service_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thingful/httpmock"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestSchedulePropagation(t *testing.T) {
	type args struct {
		endpoint string
	}
	tests := []struct {
		name                 string
		args                 args
		endpointCallErr      error
		endpointResponseCode int
		loggedError          string
		propagated           bool
	}{
		{
			name: "should propagate sync when endpoint is undefined",
			args: args{
				endpoint: "",
			},
			propagated: true,
		},
		{
			name: "should not propagate sync when the endpoint is defined",
			args: args{
				endpoint: "https://example.com",
			},
			endpointResponseCode: http.StatusOK,
			propagated:           false,
		},
		{
			name: "should not propagate sync when the endpoint is defined, but returns an error",
			args: args{
				endpoint: "https://example.com",
			},
			loggedError:     "Propagation endpoint error: Get \"https://example.com?types=permissions\": error",
			endpointCallErr: fmt.Errorf("error"),
			propagated:      false,
		},
		{
			name: "should not propagate sync when the endpoint is defined, but the response code is not 200",
			args: args{
				endpoint: "https://example.com",
			},
			loggedError:          "Propagation endpoint error: status=500",
			endpointResponseCode: http.StatusInternalServerError,
			propagated:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(`
				groups:
					- {id: 1, type: Class}
				items:
					- {id: 1, type: Task, default_language_tag: fr}
				permissions_granted:
					- {group_id: 1, item_id: 1, source_group_id: 1, origin: "group_membership", can_view: "content"}
			`)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			logHook, restoreFunc := logging.MockSharedLoggerHook()
			defer restoreFunc()

			if tt.args.endpoint != "" {
				if tt.endpointCallErr == nil {
					httpmock.RegisterStubRequests(
						httpmock.NewStubRequest(
							"GET",
							tt.args.endpoint+"?types=permissions",
							httpmock.NewStringResponder(tt.endpointResponseCode, ""),
						),
					)
				} else {
					httpmock.RegisterStubRequests(
						httpmock.NewStubRequest(
							"GET",
							tt.args.endpoint+"?types=permissions",
							func(*http.Request) (*http.Response, error) {
								return nil, tt.endpointCallErr
							},
						),
					)
				}
			}

			service.SchedulePropagation(store, tt.args.endpoint, []string{"permissions"})

			exists, err := store.Permissions().Where("item_id = 1").HasRows()
			require.NoError(t, err)
			assert.Equal(t, tt.propagated, exists)

			// Verify that all stubs were called.
			if err := httpmock.AllStubsCalled(); err != nil {
				t.Errorf("Not all stubs were called: %s", err)
			}

			// Verify logs.
			if tt.loggedError != "" {
				logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
				assert.Regexp(t, "level=error .* "+regexp.QuoteMeta(fmt.Sprintf("msg=%q", tt.loggedError)), logs)
			}
		})
	}
}
