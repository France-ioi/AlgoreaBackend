//go:build !unit

package items

import "github.com/France-ioi/AlgoreaBackend/v2/app/database"

func GetDataForResultPathStart(store *database.DataStore, participantID int64, ids []int64) []map[string]interface{} {
	return getDataForResultPathStart(store, participantID, ids)
}
