package groups

import (
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

const (
	visibleJoinedAncestorsCTEName = "visible_joined_ancestors"
	visibleManagedGroupsCTEName   = "visible_managed_groups"

	visibleGroupsWhereSQL = "groups.is_public OR groups.id IN (SELECT ancestor_group_id FROM visible_joined_ancestors) " +
		"OR groups.id IN (SELECT ancestor_group_id FROM visible_managed_groups)"

	hasVisibleChildrenSQLColumn = `
		EXISTS(
			SELECT 1
			FROM groups_groups_active AS child_links
			JOIN ` + "`groups`" + ` AS visible_child
				ON visible_child.id = child_links.child_group_id
			WHERE child_links.parent_group_id = groups.id
				AND visible_child.type != 'User'
				AND (visible_child.is_public OR visible_child.id IN (SELECT ancestor_group_id FROM visible_joined_ancestors)
					OR visible_child.id IN (SELECT ancestor_group_id FROM visible_managed_groups))
		) AS has_visible_children`
)

func visibleJoinedAncestorsQuery(store *database.DataStore, user *database.User) *database.DB {
	return store.Groups().AncestorsOfJoinedGroups(store, user)
}

func visibleManagedGroupsQuery(store *database.DataStore, user *database.User) *database.DB {
	return store.Groups().ManagedUsersAndAncestorsOfManagedGroupsForGroup(store, user.GroupID)
}

func withVisibleGroupCTEs(db *database.DB, store *database.DataStore, user *database.User) *database.DB {
	return db.
		With(visibleJoinedAncestorsCTEName, visibleJoinedAncestorsQuery(store, user)).
		With(visibleManagedGroupsCTEName, visibleManagedGroupsQuery(store, user))
}
