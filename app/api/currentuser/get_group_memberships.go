package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/group-memberships group-memberships membershipsView
// ---
// summary: List current user's groups
// description:
//   Returns the list of groups memberships of the current user.
// parameters:
// - name: sort
//   in: query
//   default: [-member_since,id]
//   type: array
//   items:
//     type: string
//     enum: [member_since,-member_since,id,-id]
// - name: from.member_since
//   description: Start the page from the membership next to one with `member_since` = `from.member_since`
//                and `groups.id` = `from.id`
//                (`from.id` is required when `from.member_since` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the membership next to one with `member_since`=`from.member_since`
//                and `groups.id`=`from.id`
//                (`from.member_since` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display the first N memberships
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with an array of groups memberships
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/membershipsViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupMemberships(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	query := srv.Store.ActiveGroupGroups().
		Select(`
			latest_change.at AS member_since,
			IFNULL(latest_change.action, 'added_directly') AS action,
			groups.id AS group__id,
			groups.name AS group__name,
			groups.description AS group__description,
			groups.type AS group__type`).
		Joins("JOIN `groups` ON `groups`.id = groups_groups_active.parent_group_id").
		Joins(`
			LEFT JOIN LATERAL (
				SELECT at, action FROM group_membership_changes
				WHERE group_id = groups_groups_active.parent_group_id AND member_id = groups_groups_active.child_group_id
				ORDER BY at DESC
				LIMIT 1
			) AS latest_change ON 1`).
		Where("groups_groups_active.child_group_id = ?", user.GroupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"member_since": {ColumnName: "member_since", FieldType: "time"},
			"id":           {ColumnName: "groups.id", FieldType: "int64"}},
		"-member_since,id", "id", false)
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}
