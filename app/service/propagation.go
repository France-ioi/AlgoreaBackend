package service

import (
	"net/http"
	"strings"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// SchedulePropagation schedules asynchronous propagation of the given types.
// If endpoint is an empty string, it will be done synchronously.
func SchedulePropagation(store *database.DataStore, endpoint string, types []string) {
	endpointFailed := false
	if endpoint != "" {
		// Async.
		response, err := http.Get(endpoint + "?types=" + strings.Join(types, ",")) //nolint:bodyclose Closed in defer.
		defer func(response *http.Response) {
			_ = response.Body.Close()
		}(response)

		if err != nil || response.StatusCode != http.StatusOK {
			endpointFailed = true
		}
	}

	if endpoint == "" || endpointFailed {
		// Sync.
		err := store.InTransaction(func(store *database.DataStore) error {
			return store.ItemItems().After()
		})
		MustNotBeError(err)
	}
}
