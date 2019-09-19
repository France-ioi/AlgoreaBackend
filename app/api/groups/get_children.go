package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

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
//     "$ref": "#/responses/groupChildrenViewResponse"
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
				ON groups_ancestors.group_child_id = user_groups.id AND
					groups_ancestors.group_ancestor_id != groups_ancestors.group_child_id
				WHERE user_groups.type = 'UserSelf' AND groups_ancestors.group_ancestor_id = groups.id
			) AS user_count`).
		Joins(`
			JOIN groups_groups ON groups.id = groups_groups.group_child_id AND
				groups_groups.type`+database.GroupRelationIsActiveCondition+` AND
				groups_groups.group_parent_id = ?`, groupID).
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

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}
