package database

import (
	"net/http"
	"strings"
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

const PropagationEndpointTimeout = 3 * time.Second

// SchedulePropagation schedules asynchronous propagation of the given types.
// If endpoint is an empty string, it will be done synchronously.
func (s *DataStore) SchedulePropagation(endpoint string, types []string) {
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
		if s.IsInTransaction() {
			s.ScheduleItemsAncestorsPropagation()
			s.SchedulePermissionsPropagation()
			s.ScheduleResultsPropagation()
		} else {
			err := s.InTransaction(func(store *DataStore) error {
				store.ScheduleItemsAncestorsPropagation()
				store.SchedulePermissionsPropagation()
				store.ScheduleResultsPropagation()

				return nil
			})
			mustNotBeError(err)
		}
	}
}
