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
//   Returns the records from `groups_groups` having `type_changed_at` >= `users.notifications_read_at`
//   and any user-related type (`type` != "direct") with the corresponding `groups` for the current user.
// parameters:
// - name: sort
//   in: query
//   default: [-type_changed_at,id]
//   type: array
//   items:
//     type: string
//     enum: [type_changed_at,-type_changed_at,id,-id]
// - name: from.type_changed_at
//   description: Start the page from the invitation/request next to one with `type_changed_at` = `from.type_changed_at`
//                and `groups_groups.id` = `from.id`
//                (`from.id` is required when `from.type_changed_at` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the invitation/request next to one with `type_changed_at`=`from.type_changed_at`
//                and `groups_groups.id`=`from.id`
//                (`from.type_changed_at` is required when from.id is present)
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
			groups_groups.type_changed_at,
			groups_groups.type,
			groups.name AS group__name,
			groups.type AS group__type`).
		Joins("JOIN `groups` ON `groups`.id = groups_groups.parent_group_id").
		Where("groups_groups.type != 'direct'").
		Where("groups_groups.child_group_id = ?", user.SelfGroupID)

	if user.NotificationsReadAt != nil {
		query = query.Where("groups_groups.type_changed_at >= ?", user.NotificationsReadAt)
	}

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"type_changed_at": {ColumnName: "groups_groups.type_changed_at", FieldType: "time"},
			"id":              {ColumnName: "groups_groups.id", FieldType: "int64"}},
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
