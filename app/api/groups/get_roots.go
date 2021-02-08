package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupRootsViewResponseRow
type groupRootsViewResponseRow struct {
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	Name string `json:"name"`
	// required:true
	// enum: Class,Team,Club,Friends,Other,User,Session
	Type string `json:"type"`
	// whether the user is a member of this group or one of its descendants
	// required:true
	// enum: none,direct,descendant
	CurrentUserMembership string `json:"current_user_membership"`
	// whether the user (or its ancestor) is a manager of this group,
	// or a manager of one of this group's ancestors (so is implicitly manager of this group) or,
	// a manager of one of this group's descendants, or none of above
	// required: true
	// enum: none,direct,ancestor,descendant
	CurrentUserManagership string `json:"current_user_managership"`
}

// swagger:operation GET /groups/roots group-memberships groupRootsView
// ---
// summary: List root groups
// description: Returns groups which are ancestors of a joined or managed groups
//   and do not have parents, not considering "type='Base'" groups
// responses:
//   "200":
//     description: OK. Success response with an array of root groups
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/groupRootsViewResponseRow"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getRoots(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	innerQuery := srv.Store.Groups().
		Where(`
			groups.id IN(?) OR groups.id IN(?)`,
			ancestorsOfJoinedGroups(srv.Store, user).QueryExpr(), ancestorsOfManagedGroups(srv.Store, user).QueryExpr()).
		Where("groups.type != 'Base'").
		Where("groups.id != ?", user.GroupID).
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
	service.MustNotBeError(selectGroupsDataForMenu(srv.Store, innerQuery, user).Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}

func selectGroupsDataForMenu(store *database.DataStore, db *database.DB, user *database.User) *database.DB {
	usersAncestorsQuery := store.ActiveGroupAncestors().
		Where("child_group_id = ?", user.GroupID).Select("ancestor_group_id")

	db = db.Select(`
		groups.id as id, groups.name, groups.type,
		IF(
			EXISTS(
				SELECT 1 FROM groups_groups_active
				WHERE groups_groups_active.parent_group_id = groups.id AND
				  	  groups_groups_active.child_group_id = ?
			),
			'direct',
			IF(
				EXISTS(
					SELECT 1 FROM groups_groups_active
					JOIN groups_ancestors_active AS group_descendants
						ON group_descendants.ancestor_group_id = groups.id AND
					     group_descendants.child_group_id = groups_groups_active.parent_group_id
					WHERE groups_groups_active.child_group_id = ?
				),
				'descendant',
				'none'
			)
		) AS 'current_user_membership',
		IF(
			EXISTS(
				SELECT 1 FROM user_ancestors
				JOIN group_managers
					ON group_managers.group_id = groups.id AND
				  	 group_managers.manager_id = user_ancestors.ancestor_group_id
			),
			'direct',
			IF(
				EXISTS(
					SELECT 1 FROM user_ancestors
					JOIN groups_ancestors_active AS group_ancestors ON group_ancestors.child_group_id = groups.id
					JOIN group_managers
						ON group_managers.group_id = group_ancestors.ancestor_group_id AND
						   group_managers.manager_id = user_ancestors.ancestor_group_id
				),
				'ancestor',
				IF(
					EXISTS(
						SELECT 1 FROM user_ancestors
						JOIN groups_ancestors_active AS group_descendants ON group_descendants.ancestor_group_id = groups.id
						JOIN group_managers
							ON group_managers.group_id = group_descendants.child_group_id AND
								 group_managers.manager_id = user_ancestors.ancestor_group_id
					),
					'descendant',
					'none'
				)
			)
		) AS 'current_user_managership'`, user.GroupID, user.GroupID)

	return store.Raw("WITH user_ancestors AS ? ?", usersAncestorsQuery.SubQuery(), db.QueryExpr())
}

func ancestorsOfJoinedGroups(store *database.DataStore, user *database.User) *database.DB {
	return store.ActiveGroupGroups().
		Where("groups_groups_active.child_group_id = ?", user.GroupID).
		Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.parent_group_id").
		Select("groups_ancestors_active.ancestor_group_id")
}

func ancestorsOfManagedGroups(store *database.DataStore, user *database.User) *database.DB {
	return store.ActiveGroupAncestors().ManagedByUser(user).
		Joins(`
			JOIN groups_ancestors_active AS ancestors_of_managed
				ON ancestors_of_managed.child_group_id = groups_ancestors_active.child_group_id`).
		Select("ancestors_of_managed.ancestor_group_id")
}
