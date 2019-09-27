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
//   default: [-status_changed_at,id]
//   type: array
//   items:
//     type: string
//     enum: [status_changed_at,-status_changed_at,id,-id]
// - name: from.status_changed_at
//   description: Start the page from the membership next to one with `status_changed_at` = `from.status_changed_at`
//                and `groups_groups.id` = `from.id`
//                (`from.id` is required when `from.status_changed_at` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the membership next to one with `status_changed_at`=`from.status_changed_at`
//                and `groups_groups.id`=`from.id`
//                (`from.status_changed_at` is required when from.id is present)
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

	query := srv.Store.GroupGroups().
		Select(`
			groups_groups.id,
			groups_groups.status_changed_at,
			groups_groups.type,
			groups.id AS group__id,
			groups.name AS group__name,
			groups.description AS group__description,
			groups.type AS group__type`).
		Joins("JOIN `groups` ON `groups`.id = groups_groups.parent_group_id").
		Where("groups_groups.type IN ('invitationAccepted', 'requestAccepted', 'direct')").
		Where("groups_groups.child_group_id = ?", user.SelfGroupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"status_changed_at": {ColumnName: "groups_groups.status_changed_at", FieldType: "time"},
			"id":                {ColumnName: "groups_groups.id", FieldType: "int64"}},
		"-status_changed_at")
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}
