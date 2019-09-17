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
//   description: Start the page from the sub-group next to the sub-group with `sName` = `from.name` and `ID` = `from.id`
//                (`from.id` is required when `from.name` is present,
//                some other 'from.*' parameters may be required too depending on the `sort`)
//   in: query
//   type: string
// - name: from.type
//   description: Start the page from the sub-group next to the sub-group with `sType` = `from.type` and `ID` = `from.id`
//                (`from.id` is required when `from.type` is present,
//                some other 'from.*' parameters may be required too depending on the `sort`)
//   in: query
//   type: string
// - name: from.grade
//   description: Start the page from the sub-group next to the sub-group with `iGrade` = `from.grade` and `ID` = `from.id`
//                (`from.id` is required when `from.grade` is present,
//                some other 'from.*' parameters may be required too depending on the `sort`)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the sub-group next to the sub-group with `ID`=`from.id`
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
			groups.ID as ID, groups.sName, groups.sType, groups.iGrade,
			groups.bOpened, groups.bFreeAccess, groups.sCode,
			(
				SELECT COUNT(*) FROM `+"`groups`"+` AS user_groups
				JOIN groups_ancestors
				ON groups_ancestors.idGroupChild = user_groups.ID AND
					groups_ancestors.idGroupAncestor != groups_ancestors.idGroupChild
				WHERE user_groups.sType = 'UserSelf' AND groups_ancestors.idGroupAncestor = groups.ID
			) AS iUserCount`).
		Joins(`
			JOIN groups_groups ON groups.ID = groups_groups.idGroupChild AND
				groups_groups.sType`+database.GroupRelationIsActiveCondition+` AND
				groups_groups.idGroupParent = ?`, groupID).
		Where("groups.sType IN (?)", typesList)
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"name":  {ColumnName: "groups.sName", FieldType: "string"},
			"type":  {ColumnName: "groups.sType", FieldType: "string"},
			"grade": {ColumnName: "groups.iGrade", FieldType: "int64"},
			"id":    {ColumnName: "groups.ID", FieldType: "int64"}},
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
