package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model groupRootsViewResponseRow
type groupRootsViewResponseRow struct {
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	Name string `json:"name"`
	// required:true
	// enum: Class,Team,Club,Friends,Other,Session
	Type string `json:"type"`
	// whether the user is a member of this group or one of its descendants
	// required:true
	// enum: none,direct,descendant
	CurrentUserMembership string `json:"current_user_membership"`
	// whether the user (or its ancestor) is a manager of this group,
	// or a manager of one of this group's ancestors (so is implicitly manager of this group) or,
	// a manager of one of this group's non-user descendants, or none of above
	// required: true
	// enum: none,direct,ancestor,descendant
	CurrentUserManagership string `json:"current_user_managership"`
}

// swagger:operation GET /groups/roots group-memberships groupRootsView
//
//	---
//	summary: List root groups
//	description: >
//		Returns groups which are ancestors of joined groups or managed non-user groups
//		and do not have parents (except for parents of type "Base").
//		Groups of type "Base", "ContestParticipants" are ignored.
//
//
//		(Note that it's impossible for the service to return groups of type "User" because a user group cannot be joined,
//		and managed user groups are skipped as well.)
//	responses:
//		"200":
//			description: OK. Success response with an array of root groups
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/groupRootsViewResponseRow"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getRoots(w http.ResponseWriter, r *http.Request) *service.APIError {
	user := srv.GetUser(r)
	store := srv.GetStore(r)

	const columns = "ancestor_group.id, ancestor_group.type, ancestor_group.name"
	matchingGroupsQuery := store.Raw(`
		SELECT id, MAX(type) AS type, MAX(name) AS name,
		       MAX(is_ancestor_of_joined) AS is_ancestor_of_joined,
		       MAX(is_direct_parent) AS is_direct_parent,
		       MAX(is_ancestor_of_managed) AS is_ancestor_of_managed,
		       MAX(is_managed_directly) AS is_managed_directly,
		       MAX(is_managed_via_ancestor) AS is_managed_via_ancestor
		FROM ? AS matching_groups
		GROUP BY id`,
		ancestorsOfJoinedGroupsQuery(store, user).
			Select(columns+
				", 1 AS is_ancestor_of_joined, is_direct_parent, 0 AS is_ancestor_of_managed, 0 AS is_managed_directly, 0 AS is_managed_via_ancestor").
			UnionAll(ancestorsOfManagedGroupsQuery(store, user).
				Select(columns+
					", 0 AS is_ancestor_of_joined, 0 AS is_direct_parent, 1 AS is_ancestor_of_managed, is_managed_directly, is_managed_via_ancestor")).
			SubQuery())

	query := store.Raw(`
		SELECT groups.id, groups.type, groups.name,
		       IF(
			       is_ancestor_of_joined,
			       IF(is_direct_parent, 'direct', 'descendant'),
			       'none'
		       ) AS 'current_user_membership',
		       IF(
		         NOT is_ancestor_of_managed,
		         'none',
		         IF(
		           is_managed_directly,
		           'direct',
		           IF(is_managed_via_ancestor, 'ancestor', 'descendant')
		         )
		       ) AS 'current_user_managership'
		FROM ? AS `+"`groups`"+`
		WHERE
			type != 'Base' AND
			NOT EXISTS(
				SELECT 1 FROM `+"`groups`"+` AS parent_group
				JOIN groups_groups_active
					ON groups_groups_active.parent_group_id = parent_group.id AND
					   groups_groups_active.child_group_id = groups.id
				WHERE parent_group.type != 'Base'
			)
		ORDER BY groups.name`, matchingGroupsQuery.SubQuery())

	var result []groupRootsViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}

// ancestorsOfJoinedGroupsQuery returns a query selecting all ancestors of groups joined by the given user
// (excluding ContestParticipants). Additionally, it returns whether the ancestor is a direct parent
// of the given user (`is_direct_parent` column).
func ancestorsOfJoinedGroupsQuery(store *database.DataStore, user *database.User) *database.DB {
	distinctAncestorsOfJoinedGroupsQuery := store.ActiveGroupGroups().
		Where("groups_groups_active.child_group_id = ?", user.GroupID).
		Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.parent_group_id").
		Select("groups_ancestors_active.ancestor_group_id AS id, MAX(groups_ancestors_active.is_self) AS is_direct_parent").
		Group("groups_ancestors_active.ancestor_group_id")

	return store.Table("`groups` AS ancestor_group").
		Joins("JOIN ? AS distinct_ancestors ON ancestor_group.id = distinct_ancestors.id",
			distinctAncestorsOfJoinedGroupsQuery.SubQuery()).
		Where("type != 'ContestParticipants'")
}

// ancestorsOfManagedGroupsQuery returns a query to get the ancestors (excluding ContestParticipants)
// of the groups (excluding users) managed by the given user (as ancestor_group.*).
func ancestorsOfManagedGroupsQuery(store *database.DataStore, user *database.User) *database.DB {
	distinctGroupsManagedByUserQuery := store.ActiveGroupAncestors().ManagedByUser(user).
		Select("groups_ancestors_active.child_group_id AS id, MAX(groups_ancestors_active.is_self) AS is_managed_directly").
		Group("groups_ancestors_active.child_group_id")

	distinctManagedNonUserGroupsQuery := store.Groups().
		Joins("JOIN ? AS distinct_managed ON distinct_managed.id = groups.id", distinctGroupsManagedByUserQuery.SubQuery()).
		Where("groups.type != 'User'").
		Select("groups.id, is_managed_directly")

	distinctAncestorsOfManagedNonUserGroupsQuery := store.ActiveGroupAncestors().
		Joins("JOIN ? AS distinct_managed_non_user_groups ON distinct_managed_non_user_groups.id = child_group_id",
			distinctManagedNonUserGroupsQuery.SubQuery()).
		Select(`
			ancestor_group_id AS id,
			MAX(is_managed_directly AND is_self) AS is_managed_directly,
			MAX(NOT is_managed_directly AND is_self) AS is_managed_via_ancestor`).
		Group("ancestor_group_id")

	return store.Table("`groups` AS ancestor_group").
		Joins("JOIN ? AS distinct_ancestors_of_managed ON distinct_ancestors_of_managed.id = ancestor_group.id",
			distinctAncestorsOfManagedNonUserGroupsQuery.SubQuery()).
		Where("type != 'ContestParticipants'")
}
