package currentuser

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func CheckPreconditionsForGroupRequests(store *database.DataStore, user *database.User,
	groupID int64, action userGroupRelationAction, approvals database.GroupApprovals) service.APIError {
	return checkPreconditionsForGroupRequests(store, user, groupID, action, approvals)
}
