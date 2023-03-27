//go:build !unit

package items

import "github.com/France-ioi/AlgoreaBackend/app/database"

func FindItemPath(store *database.DataStore, participantID, itemID int64) []string {
	return findItemPath(store, participantID, itemID)
}
