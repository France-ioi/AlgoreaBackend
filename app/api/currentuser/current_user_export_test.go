package currentuser

import (
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

func CheckPreconditionsForGroupRequests(store *database.DataStore, user *database.User,
	groupID int64, action userGroupRelationAction,
) *service.APIError {
	return checkPreconditionsForGroupRequests(store, user, groupID, action)
}
