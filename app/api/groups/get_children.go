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
	// enum: Class,Team,Club,Friends,Other,UserSelf,UserAdmin,Base
	Type string `json:"type"`
	// required:true
	Grade int32 `json:"grade"`
	// required:true
	Opened bool `json:"opened"`
	// required:true
	FreeAccess bool `json:"free_access"`
	// Nullable
	// required:true
	Code *string `json:"code"`
	// The number of descendant users
	// required:true
	UserCount int32 `json:"user_count"`
}

// swagger:operation GET /groups/{group_id}/children groups groupChildrenView
// ---
// summary: List group's children
// description: Returns children of the group having types
//   specified by `types_include` and `types_exclude` parameters.
//
//   * The authenticated user should own the parent group.
// parameters:
// - name: group_id
//   in: path
//   required: true
//   type: integer
// - name: types_include
//   in: query
//   default: [Class,Team,Club,Friends,Other,UserSelf,UserAdmin,Base]
//   type: array
//   items:
//     type: string
//     enum: [Class,Team,Club,Friends,Other,UserSelf,UserAdmin,Base]
// - name: types_exclude
//   in: query
//   default: []
//   type: array
//   items:
//     type: string
//     enum: [Class,Team,Club,Friends,Other,UserSelf,UserAdmin,Base]
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
			"Other": true, "UserSelf": true, "UserAdmin": true,
		})
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.Groups().
		Select(`
			groups.id as id, groups.name, groups.type, groups.grade,
			groups.opened, groups.free_access, groups.code,
			(
				SELECT COUNT(*) FROM `+"`groups`"+` AS user_groups
				JOIN groups_ancestors
				ON groups_ancestors.child_group_id = user_groups.id AND
					groups_ancestors.ancestor_group_id != groups_ancestors.child_group_id AND
					NOW() < groups_ancestors.expires_at
				WHERE user_groups.type = 'UserSelf' AND groups_ancestors.ancestor_group_id = groups.id
			) AS user_count`).
		Where("groups.id IN(?)",
			srv.Store.GroupGroups().WhereGroupRelationIsActive().Table("groups_groups USE INDEX(parent_type)").
				Select("child_group_id").Where("parent_group_id = ?", groupID).QueryExpr()).
		Where("groups.type IN (?)", typesList)
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"name":  {ColumnName: "groups.name", FieldType: "string"},
			"type":  {ColumnName: "groups.type", FieldType: "string"},
			"grade": {ColumnName: "groups.grade", FieldType: "int64"},
			"id":    {ColumnName: "groups.id", FieldType: "int64"}},
		"name")
	if apiError != service.NoError {
		return apiError
	}

	var result []groupChildrenViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
