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
//		and do not have parents. Groups of type "Base" or "User" are ignored.
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
func (srv *Service) getRoots(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	store := srv.GetStore(r)

	const columns = "ancestor_group.id, ancestor_group.type, ancestor_group.name"
	matchingGroupsQuery := store.Groups().AncestorsOfJoinedGroups(store, user).Select(columns).
		Union(ancestorsOfManagedGroupsQuery(store, user).Select(columns))

	query := store.
		With("matching_groups", matchingGroupsQuery).
		With("user_ancestors", ancestorsOfUserQuery(store, user)).
		Table("matching_groups AS `groups`").
		Select(`
				groups.id, groups.type, groups.name,
				` + currentUserMembershipSQLColumn(user) + `,
				` + currentUserManagershipSQLColumn).
		Where("groups.type != 'Base'").
		Where(`
			NOT EXISTS(
				SELECT 1 FROM ` + "`groups`" + ` AS parent_group
				JOIN groups_groups_active
					ON groups_groups_active.parent_group_id = parent_group.id AND
					   groups_groups_active.child_group_id = groups.id
				WHERE parent_group.type != 'Base'
			)`).
		Order("groups.name")

	var result []groupRootsViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}

// ancestorsOfManagedGroupsQuery returns a query to get the ancestors of the groups (excluding users) managed by
// the given user (as ancestor_group_id).
func ancestorsOfManagedGroupsQuery(store *database.DataStore, user *database.User) *database.DB {
	managedNonUserGroupsQuery := store.ActiveGroupAncestors().ManagedByUser(user).
		Joins("JOIN `groups` ON groups.id = groups_ancestors_active.child_group_id AND groups.type != 'User'").
		Select("DISTINCT groups.id")

	return store.With("managed_non_user_groups", managedNonUserGroupsQuery).
		Table("managed_non_user_groups").
		Joins(`
			JOIN groups_ancestors_active AS ancestors_of_managed
				ON ancestors_of_managed.child_group_id = managed_non_user_groups.id`).
		Joins("JOIN `groups` AS ancestor_group ON ancestor_group.id = ancestors_of_managed.ancestor_group_id").
		Where("ancestor_group.type != 'ContestParticipants'").
		Select("DISTINCT ancestors_of_managed.ancestor_group_id")
}
