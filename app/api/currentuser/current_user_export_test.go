package currentuser

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func CheckPreconditionsForGroupRequests(store *database.DataStore, user *database.User,
	groupID int64, requireFreeAccess bool) service.APIError {
	return checkPreconditionsForGroupRequests(store, user, groupID, requireFreeAccess)
}
