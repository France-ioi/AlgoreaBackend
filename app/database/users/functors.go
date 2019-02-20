package users

import "github.com/France-ioi/AlgoreaBackend/app/database"

// OnlyDescendantsOf filters users by joining `groups_ancestors`
// on idGroupAncestor=groupID & idGroupChild=users.idGroupSelf
func OnlyDescendantsOf(groupID int64) database.Functor {
	return func(context *database.Context) {
		context.DB = context.DB.
			Joins("JOIN groups_ancestors ON groups_ancestors.idGroupChild=users.idGroupSelf").
			Where("groups_ancestors.idGroupAncestor = ?", groupID)
	}
}
