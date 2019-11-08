package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/group-memberships groups users membershipsView
// ---
// summary: List groups that the current user has joined
// description:
//   Returns the list of groups memberships of the current user
//   (`groups_groups.type` is “requestAccepted”, “invitationAccepted” or “direct”).
// parameters:
// - name: sort
//   in: query
//   default: [-type_changed_at,id]
//   type: array
//   items:
//     type: string
//     enum: [type_changed_at,-type_changed_at,id,-id]
// - name: from.type_changed_at
//   description: Start the page from the membership next to one with `type_changed_at` = `from.type_changed_at`
//                and `groups_groups.id` = `from.id`
//                (`from.id` is required when `from.type_changed_at` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the membership next to one with `type_changed_at`=`from.type_changed_at`
//                and `groups_groups.id`=`from.id`
//                (`from.type_changed_at` is required when from.id is present)
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
			groups_groups_active.id,
			groups_groups_active.type_changed_at,
			groups_groups_active.type,
			groups.id AS group__id,
			groups.name AS group__name,
			groups.description AS group__description,
			groups.type AS group__type`).
		Joins("JOIN `groups` ON `groups`.id = groups_groups_active.parent_group_id").
		Where("groups_groups_active.type IN ('invitationAccepted', 'requestAccepted', 'direct')").
		Where("groups_groups_active.child_group_id = ?", user.GroupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"type_changed_at": {ColumnName: "groups_groups_active.type_changed_at", FieldType: "time"},
			"id":              {ColumnName: "groups_groups_active.id", FieldType: "int64"}},
		"-type_changed_at")
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}
