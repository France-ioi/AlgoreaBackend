// +build !unit

package items

import "github.com/France-ioi/AlgoreaBackend/app/database"

func GetDataForResultPathStart(store *database.DataStore, participantID int64, ids []int64) []map[string]interface{} {
	return getDataForResultPathStart(store, participantID, ids)
}
