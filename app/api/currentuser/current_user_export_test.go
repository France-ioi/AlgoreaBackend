package currentuser

import (
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func CheckPreconditionsForGroupRequests(store *database.DataStore, user *database.User,
	groupID int64, action userGroupRelationAction,
) error {
	return checkPreconditionsForGroupRequests(store, user, groupID, action)
}
