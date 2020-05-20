package groups

import "github.com/France-ioi/AlgoreaBackend/app/database"

func FilterOtherTeamsMembersOutForLogins(store *database.DataStore, parentGroupID int64, groupsToInvite []int64,
	results map[string]string, groupIDToLoginMap map[int64]string) []int64 {
	return filterOtherTeamsMembersOutForLogins(store, parentGroupID, groupsToInvite, results, groupIDToLoginMap)
}
