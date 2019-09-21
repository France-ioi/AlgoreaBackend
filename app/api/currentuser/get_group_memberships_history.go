package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/group-memberships-history groups users groupsMembershipHistory
// ---
// summary: Get a history of invitations/requests for the current user
// description:
//   Returns the records from `groups_groups` having `status_date` >= `users.notification_read_date`
//   and any user-related type (`type` != "direct") with the corresponding `groups` for the current user.
// parameters:
// - name: sort
//   in: query
//   default: [-status_date,id]
//   type: array
//   items:
//     type: string
//     enum: [status_date,-status_date,id,-id]
// - name: from.status_date
//   description: Start the page from the invitation/request next to one with `status_date` = `from.status_date`
//                and `groups_groups.id` = `from.id`
//                (`from.id` is required when `from.status_date` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the invitation/request next to one with `status_date`=`from.status_date`
//                and `groups_groups.id`=`from.id`
//                (`from.status_date` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Return the first N invitations/requests
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
//         "$ref": "#/definitions/groupsMembershipHistoryResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupMembershipsHistory(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	query := srv.Store.GroupGroups().
		Select(`
			groups_groups.id,
			groups_groups.status_date,
			groups_groups.type,
			groups.name AS group__name,
			groups.type AS group__type`).
		Joins("JOIN `groups` ON `groups`.id = groups_groups.parent_group_id").
		Where("groups_groups.type != 'direct'").
		Where("groups_groups.child_group_id = ?", user.SelfGroupID)
	if user.NotificationReadDate != nil {
		query = query.Where("groups_groups.status_date >= ?", user.NotificationReadDate)
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
