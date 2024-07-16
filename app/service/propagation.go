package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

const PropagationEndpointTimeout = 3 * time.Second

// SchedulePropagation schedules asynchronous propagation of the given types.
// If endpoint is an empty string, it will be done synchronously.
func SchedulePropagation(store *database.DataStore, endpoint string, types []string) {
	endpointFailed := false
	if endpoint != "" {
		// Async.
		client := http.Client{
			Timeout: PropagationEndpointTimeout,
		}

		callTime := time.Now()
		response, err := client.Get(endpoint + "?types=" + strings.Join(types, ","))
		logging.Infof("Propagation endpoint called: %v, types=%v, duration=%v", endpoint, types, time.Since(callTime))

		if err != nil {
			logging.Errorf("Propagation endpoint error: %v", err)

			endpointFailed = true
		} else {
			defer func(response *http.Response) {
				_ = response.Body.Close()
			}(response)

			if response.StatusCode != http.StatusOK {
				logging.Errorf("Propagation endpoint error: status=%v", response.StatusCode)

				endpointFailed = true
			}
		}
	}

	if endpoint == "" || endpointFailed {
		// Sync.
		if store.IsInTransaction() {
			store.ScheduleItemsAncestorsPropagation()
			store.SchedulePermissionsPropagation()
			store.ScheduleResultsPropagation()
		} else {
			err := store.InTransaction(func(store *database.DataStore) error {
				store.ScheduleItemsAncestorsPropagation()
				store.SchedulePermissionsPropagation()
				store.ScheduleResultsPropagation()

				return nil
			})
			MustNotBeError(err)
		}
	}
}
