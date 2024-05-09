package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/group-invitations group-memberships invitationsView
//
//	---
//	summary: List current invitations and requests to groups
//	description:
//		Returns the list of invitations that the current user received with `group_membership_changes.at`.
//	parameters:
//		- name: sort
//			in: query
//			default: [-at,group_id]
//			type: array
//			items:
//				type: string
//				enum: [at,-at,group_id,-group_id]
//		- name: from.at
//			description: Start the page from the request/invitation next to one with `at` = `{from.at}`
//							 and `group_membership_changes.group_id` = `{from.group_id}`
//							 (`{from.group_id}` is required when `{from.at}` is present)
//			in: query
//			type: string
//		- name: from.group_id
//			description: Start the page from the request/invitation next to one with `at`=`{from.at}`
//							 and `group_id`=`{from.group_id}`
//							 (`{from.at}` is required when `{from.group_id}` is present)
//			in: query
//			type: integer
//		- name: limit
//			description: Display the first N requests/invitations
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
//					"$ref": "#/definitions/invitationsViewResponseRow"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupInvitations(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	query := srv.GetStore(r).GroupMembershipChanges().
		Select(`
			group_membership_changes.group_id,
			group_membership_changes.at,
			users.group_id AS inviting_user__id,
			users.login AS inviting_user__login,
			users.first_name AS inviting_user__first_name,
			users.last_name AS inviting_user__last_name,
			groups.id AS group__id,
			groups.name AS group__name,
			groups.description AS group__description,
			groups.type AS group__type`).
		Joins("JOIN users ON users.group_id = initiator_id AND action = 'invitation_created'").
		Joins("JOIN `groups` ON `groups`.id = group_membership_changes.group_id").
		Joins(`
			JOIN group_pending_requests
				ON group_pending_requests.group_id = group_membership_changes.group_id AND
					group_pending_requests.member_id = group_membership_changes.member_id AND
					(SELECT MAX(latest_change.at) FROM group_membership_changes AS latest_change
					 WHERE latest_change.group_id = group_pending_requests.group_id AND
						latest_change.member_id = group_pending_requests.member_id AND
						latest_change.action = group_membership_changes.action) = group_membership_changes.at`).
		Where("group_membership_changes.member_id = ?", user.GroupID)

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
