package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupParentsViewResponseRow
type groupParentsViewResponseRow struct {
	// `groups.id`
	// required: true
	ID int64 `json:"id,string"`
	// `groups.name`
	// required: true
	Name string `json:"name"`
	// required: true
	CurrentUserCanGrantGroupAccess bool `json:"current_user_can_grant_group_access"`
	// required: true
	// enum: Class,Team,Club,Friends,Other,Session,Base
	Type string `json:"type"`
}

// swagger:operation GET /groups/{group_id}/parents groups groupParentsView
// ---
// summary: List group parents
// description: >
//
//   Lists visible parents of the given group.
//
//
//   A group is visible if it is either
//   1) an ancestor of a group the current user joined, or 2) an ancestor of a non-user group he manages, or
//   3) a descendant of a group he manages, or 4) a public group.
//
//
//   Groups with `type`='ContestParticipants' are not displayed.
//
//
//   * The `group_id` should be visible to the current user, otherwise the 'forbidden' error is returned.
//
// parameters:
//   - name: group_id
//     in: path
//     type: integer
//     required: true
//   - name: sort
//     in: query
//     default: [name,id]
//     type: array
//     items:
//       type: string
//       enum: [name,-name,id,-id]
//   - name: from.id
//     description: Start the page from the parent next to the parent with `groups.id`=`{from.id}`
//     in: query
//     type: integer
//   - name: limit
//     description: Display the first N parents
//     in: query
//     type: integer
//     maximum: 1000
//     default: 500
// responses:
//   "200":
//     description: OK. The array of group parents
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/groupParentsViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getParents(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	store := srv.GetStore(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	found, err := store.Groups().PickVisibleGroups(store.Groups().ByID(groupID), user).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	query := store.Groups().PickVisibleGroups(store.Groups().DB, user).
		Select(`
			groups.id, groups.name, groups.type,
			(
				SELECT
					IFNULL(MAX(can_grant_group_access), 0) AS current_user_can_grant_group_access
				FROM group_managers
				JOIN groups_ancestors_active AS manager_ancestors
					ON manager_ancestors.ancestor_group_id = group_managers.manager_id AND manager_ancestors.child_group_id = ?
				JOIN groups_ancestors_active AS group_ancestors
					ON group_ancestors.ancestor_group_id = group_managers.group_id AND group_ancestors.child_group_id = groups.id
			) AS current_user_can_grant_group_access`, user.GroupID).
		Where("groups.id IN(?)",
			store.ActiveGroupGroups().
				Select("parent_group_id").Where("child_group_id = ?", groupID).QueryExpr()).
		Where("groups.type != 'ContestParticipants'")
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(
		r, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"name": {ColumnName: "groups.name"},
				"id":   {ColumnName: "groups.id"},
			},
			DefaultRules: "name,id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	if apiError != service.NoError {
		return apiError
	}

	var result []groupParentsViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
