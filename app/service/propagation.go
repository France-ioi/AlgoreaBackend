package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// PropagationEndpointTimeout is the timeout for the propagation endpoint.
const PropagationEndpointTimeout = 3 * time.Second

// SchedulePropagation schedules asynchronous propagation of the given types.
// If endpoint is an empty string, it will be done synchronously.
func SchedulePropagation(store *database.DataStore, endpoint string, types []string) {
	if endpoint != "" {
		// Async.
		client := http.Client{
			Timeout: PropagationEndpointTimeout,
		}

		callTime := time.Now()
		req, _ := http.NewRequestWithContext(store.GetContext(), http.MethodGet, endpoint+"?types="+strings.Join(types, ","), http.NoBody)
		response, err := client.Do(req)
		logging.EntryFromContext(store.GetContext()).
			Infof("Propagation endpoint called: %v, types=%v, duration=%v", endpoint, types, time.Since(callTime))

		if err != nil {
			logging.EntryFromContext(store.GetContext()).Errorf("Propagation endpoint error: %v", err)
		} else {
			defer func(response *http.Response) {
				_ = response.Body.Close()
			}(response)

			if response.StatusCode != http.StatusOK {
				logging.EntryFromContext(store.GetContext()).
					Errorf("Propagation endpoint error: status=%v", response.StatusCode)
			}
		}
		return
	}

	// Sync.
	MustNotBeError(store.InTransaction(func(store *database.DataStore) error {
		store.SchedulePermissionsPropagation()
		store.ScheduleResultsPropagation()

		return nil
	}))
}
