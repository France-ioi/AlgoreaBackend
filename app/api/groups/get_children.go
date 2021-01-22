package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

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
	// Nullable
	// required:true
	Code *string `json:"code"`
	// The number of descendant users
	// required:true
	UserCount int32 `json:"user_count"`
	// required:true
	// enum: none,memberships,memberships_and_group
	CanManage string `json:"can_manage"`
	// required:true
	CanGrantGroupAccess bool `json:"can_grant_group_access"`
	// required:true
	CanWatchMembers bool `json:"can_watch_members"`

	CanManageValue int `json:"-"`
}

// swagger:operation GET /groups/{group_id}/children group-memberships groupChildrenView
// ---
// summary: List group's children
// description: Returns children of the group having types
//   specified by `types_include` and `types_exclude` parameters.
//
//   * The authenticated user should be a manager of the parent group.
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

	if apiError := checkThatUserCanManageTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.Groups().
		Select(`
			groups.id as id, groups.name, groups.type, groups.grade,
			groups.is_open, groups.is_public, groups.code,
			(
				SELECT COUNT(DISTINCT users.group_id) FROM users
				JOIN groups_groups_active ON groups_groups_active.child_group_id = users.group_id
				JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.parent_group_id
				WHERE groups_ancestors_active.ancestor_group_id = groups.id
			) AS user_count,
			can_manage_value, can_grant_group_access, can_watch_members`).
		Joins(`
			JOIN LATERAL (
				SELECT IFNULL(MAX(can_manage_value), 1) AS can_manage_value,
				       IFNULL(MAX(can_grant_group_access), 0) AS can_grant_group_access,
				       IFNULL(MAX(can_watch_members), 0) AS can_watch_members
				FROM group_managers
				JOIN groups_ancestors_active AS manager_ancestors
					ON manager_ancestors.ancestor_group_id = group_managers.manager_id AND manager_ancestors.child_group_id = ?
				JOIN groups_ancestors_active AS group_ancestors
					ON group_ancestors.ancestor_group_id = group_managers.group_id AND group_ancestors.child_group_id = groups.id
			) AS manager_permissions`, user.GroupID).
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
		result[index].CanManage = groupManagerStore.CanManageNameByIndex(result[index].CanManageValue)
	}

	render.Respond(w, r, result)
	return service.NoError
}
