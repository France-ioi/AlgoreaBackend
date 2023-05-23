package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/group-memberships-history group-memberships groupsMembershipHistory
//
//	---
//	summary: Get invitations/requests for the current user
//	description:
//		Returns the records from `group_membership_changes` having `at` >= `users.notifications_read_at`
//		and any user-related type (`action` != "added_directly") with the corresponding `groups` for the current user.
//	parameters:
//		- name: sort
//			in: query
//			default: [-at,group_id]
//			type: array
//			items:
//				type: string
//				enum: [at,-at,group_id,-group_id]
//		- name: from.at
//			description: Start the page from the invitation/request next to one with `at` = `{from.at}`
//							 and `group_membership_changes.group_id` = `{from.group_id}`
//							 (`{from.group_id}` is required when `{from.at}` is present)
//			in: query
//			type: string
//		- name: from.group_id
//			description: Start the page from the invitation/request next to one with `at`=`{from.at}`
//							 and `group_membership_changes.group_id`=`{from.group_id}`
//							 (`{from.at}` is required when `{from.group_id}` is present)
//			in: query
//			type: integer
//		- name: limit
//			description: Return the first N invitations/requests
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. Success response with an array of invitations/requests
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/groupsMembershipHistoryResponseRow"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupMembershipsHistory(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	query := srv.GetStore(r).GroupMembershipChanges().
		Select(`
			group_membership_changes.at,
			group_membership_changes.action,
			groups.id AS group__id,
			groups.name AS group__name,
			groups.type AS group__type`).
		Joins("JOIN `groups` ON `groups`.id = group_membership_changes.group_id").
		Where("group_membership_changes.action != 'added_directly'").
		Where("group_membership_changes.member_id = ?", user.GroupID)

	if user.NotificationsReadAt != nil {
		query = query.Where("group_membership_changes.at >= ?", user.NotificationsReadAt)
	}

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(
		r, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"at":       {ColumnName: "group_membership_changes.at"},
				"group_id": {ColumnName: "group_membership_changes.group_id"},
			},
			DefaultRules: "-at,group_id",
			TieBreakers: service.SortingAndPagingTieBreakers{
				"group_id": service.FieldTypeInt64,
				"at":       service.FieldTypeTime,
			},
		})
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}
