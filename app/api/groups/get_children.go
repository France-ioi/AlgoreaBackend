package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// UserCountPart contains the number of descendant users for a group.
// This field is only displayed if the current user is a manager of the group.
// swagger:ignore
type UserCountPart struct {
	// The number of descendant users (returned only if the current user is a manager)
	UserCount int32 `json:"user_count"`
}

// swagger:model groupChildrenViewResponseRow
type groupChildrenViewResponseRow struct {
	// The sub-group's `groups.id`
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	Name string `json:"name"`
	// required:true
	// enum: Class,Team,Club,Friends,Other,User,Session,Base
	Type string `json:"type"`
	// required:true
	Grade int32 `json:"grade"`
	// required:true
	IsOpen bool `json:"is_open"`
	// required:true
	IsPublic bool `json:"is_public"`
	// required:true
	CurrentUserIsManager bool `json:"current_user_is_manager"`
	*ManagerPermissionsPart
	*UserCountPart
}

// swagger:operation GET /groups/{group_id}/children group-memberships groupChildrenView
// ---
// summary: List group's children
// description: >
//   Returns visible children of the group having types
//   specified by `types_include` and `types_exclude` parameters.
//
//
//   A group is visible if it is either
//   1) an ancestor of a group the current user joined, or 2) an ancestor of a non-user group he manages, or
//   3) a descendant of a group he manages, or 4) a public group.
//
//
//   * The `group_id` should be visible to the current user.
//
//
//   Note: `user_count` and `current_user_can_*` fields are omitted when the user is not a manager of the group.
// parameters:
// - name: group_id
//   in: path
//   required: true
//   type: integer
// - name: types_include
//   in: query
//   default: [Class,Team,Club,Friends,Other,User,Session,Base]
//   type: array
//   items:
//     type: string
//     enum: [Class,Team,Club,Friends,Other,User,Session,Base]
// - name: types_exclude
//   in: query
//   default: []
//   type: array
//   items:
//     type: string
//     enum: [Class,Team,Club,Friends,Other,User,Session,Base]
// - name: from.name
//   description: Start the page from the sub-group next to the sub-group with `name` = `from.name` and `id` = `from.id`
//                (`from.id` is required when `from.name` is present,
//                some other 'from.*' parameters may be required too depending on the `sort`)
//   in: query
//   type: string
// - name: from.type
//   description: Start the page from the sub-group next to the sub-group with `type` = `from.type` and `id` = `from.id`
//                (`from.id` is required when `from.type` is present,
//                some other 'from.*' parameters may be required too depending on the `sort`)
//   in: query
//   type: string
// - name: from.grade
//   description: Start the page from the sub-group next to the sub-group with `grade` = `from.grade` and `id` = `from.id`
//                (`from.id` is required when `from.grade` is present,
//                some other 'from.*' parameters may be required too depending on the `sort`)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the sub-group next to the sub-group with `id`=`from.id`
//                (if at least one of other 'from.*' parameters is present, `sort.id` is required)
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [name,id]
//   type: array
//   items:
//     type: string
//     enum: [name,-name,type,-type,grade,-grade,id,-id]
// - name: limit
//   description: Display the first N sub-groups
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with an array of group's children
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/groupChildrenViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getChildren(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	typesList, err := service.ResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters(r, "types",
		map[string]bool{
			"Base": true, "Class": true, "Team": true, "Club": true, "Friends": true,
			"Other": true, "User": true, "Session": true,
		})
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	found, err := pickVisibleGroups(srv.Store.Groups().ByID(groupID), user).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	query := pickVisibleGroups(srv.Store.Groups().DB, user).
		Select(`
			groups.id as id, groups.name, groups.type, groups.grade,
			groups.is_open, groups.is_public,
			IF(manager_permissions.found,
				(SELECT COUNT(DISTINCT users.group_id) FROM users
				JOIN groups_groups_active ON groups_groups_active.child_group_id = users.group_id
				JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.parent_group_id
				WHERE groups_ancestors_active.ancestor_group_id = groups.id),
				0
			) AS user_count,
			manager_permissions.found AS current_user_is_manager,
			current_user_can_manage_value, current_user_can_grant_group_access, current_user_can_watch_members`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT COUNT(*) > 0 AS found,
				       IFNULL(MAX(can_manage_value), 1) AS current_user_can_manage_value,
				       IFNULL(MAX(can_grant_group_access), 0) AS current_user_can_grant_group_access,
				       IFNULL(MAX(can_watch_members), 0) AS current_user_can_watch_members
				FROM group_managers
				JOIN groups_ancestors_active AS manager_ancestors
					ON manager_ancestors.ancestor_group_id = group_managers.manager_id AND manager_ancestors.child_group_id = ?
				JOIN groups_ancestors_active AS group_ancestors
					ON group_ancestors.ancestor_group_id = group_managers.group_id AND group_ancestors.child_group_id = groups.id
			) AS manager_permissions ON 1`, user.GroupID).
		Where("groups.id IN(?)",
			srv.Store.ActiveGroupGroups().
				Select("child_group_id").Where("parent_group_id = ?", groupID).QueryExpr()).
		Where("groups.type IN (?)", typesList)
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"name":  {ColumnName: "groups.name", FieldType: "string"},
			"type":  {ColumnName: "groups.type", FieldType: "string"},
			"grade": {ColumnName: "groups.grade", FieldType: "int64"},
			"id":    {ColumnName: "groups.id", FieldType: "int64"}},
		"name,id", []string{"id"}, false)
	if apiError != service.NoError {
		return apiError
	}

	var result []groupChildrenViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())
	groupManagerStore := srv.Store.GroupManagers()
	for index := range result {
		if !result[index].CurrentUserIsManager {
			result[index].ManagerPermissionsPart = nil
			result[index].UserCountPart = nil
		} else {
			result[index].ManagerPermissionsPart.CurrentUserCanManage =
				groupManagerStore.CanManageNameByIndex(result[index].CurrentUserCanManageValue)
		}
	}

	render.Respond(w, r, result)
	return service.NoError
}
