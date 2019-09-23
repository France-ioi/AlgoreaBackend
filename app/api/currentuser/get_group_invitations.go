package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/group-invitations groups users invitationsView
// ---
// summary: List current invitations and requests to groups
// description:
//   Returns the list of invitations that the current user received and requests sent by him
//   (`groups_groups.type` is “invitationSent” or “requestSent” or “requestRefused”)
//   with `groups_groups.status_date` within `within_weeks` back from now (if `within_weeks` is present).
// parameters:
// - name: within_weeks
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [-status_date,id]
//   type: array
//   items:
//     type: string
//     enum: [status_date,-status_date,id,-id]
// - name: from.status_date
//   description: Start the page from the request/invitation next to one with `status_date` = `from.status_date`
//                and `groups_groups.id` = `from.id`
//                (`from.id` is required when `from.status_date` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the request/invitation next to one with `status_date`=`from.status_date`
//                and `groups_groups.id`=`from.id`
//                (`from.status_date` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display the first N requests/invitations
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with an array of invitations/requests
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/invitationsViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupInvitations(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	query := srv.Store.GroupGroups().
		Select(`
			groups_groups.id,
			groups_groups.status_date,
			groups_groups.type,
			users.id AS inviting_user__id,
			users.login AS inviting_user__login,
			users.first_name AS inviting_user__first_name,
			users.last_name AS inviting_user__last_name,
			groups.id AS group__id,
			groups.name AS group__name,
			groups.description AS group__description,
			groups.type AS group__type`).
		Joins("LEFT JOIN users ON users.id = groups_groups.inviting_user_id").
		Joins("JOIN `groups` ON `groups`.id = groups_groups.parent_group_id").
		Where("groups_groups.type IN ('invitationSent', 'requestSent', 'requestRefused')").
		Where("groups_groups.child_group_id = ?", user.SelfGroupID)

	if len(r.URL.Query()["within_weeks"]) > 0 {
		withinWeeks, err := service.ResolveURLQueryGetInt64Field(r, "within_weeks")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
		query = query.Where("NOW() - INTERVAL ? WEEK < groups_groups.status_date", withinWeeks)
	}

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"status_date": {ColumnName: "groups_groups.status_date", FieldType: "time"},
			"id":          {ColumnName: "groups_groups.id", FieldType: "int64"}},
		"-status_date")
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}
