package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/database"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model
type invitationsViewResponseRow struct {
	// `group_membership_changes.group_id`
	// required: true
	GroupID int64 `json:"group_id,string"`
	// `groups_groups.type_changed_at`
	// required: true
	At database.Time `json:"at"`

	// the user that invited
	// required: true
	InvitingUser invitingUser `json:"inviting_user" gorm:"embedded;embedded_prefix:inviting_user__"`

	// required: true
	Group groupWithApprovals `json:"group" gorm:"embedded;embedded_prefix:group__"`
}

type invitingUser struct {
	// `users.group_id`
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	Login string `json:"login"`
	// Nullable
	// required: true
	FirstName string `json:"first_name"`
	// Nullable
	// required: true
	LastName string `json:"last_name"`
}

type groupWithApprovals struct {
	// `groups.id`
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	Name string `json:"name"`
	// Nullable
	// required: true
	Description *string `json:"description"`
	// required: true
	// enum: Class,Team,Club,Friends,Other,Session,Base
	Type string `json:"type"`
	// enum: none,view,edit
	// required: true
	RequirePersonalInfoAccessApproval string `json:"require_personal_info_access_approval"`
	// Nullable
	// required: true
	RequireLockMembershipApprovalUntil *database.Time `json:"require_lock_membership_approval_until"`
	// required: true
	RequireWatchApproval bool `json:"require_watch_approval"`
}

// swagger:operation GET /current-user/group-invitations group-memberships invitationsView
//
//	---
//	summary: List current invitations to groups
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
//			description: OK. Success response with an array of invitations.
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
			action,
			users.group_id AS inviting_user__id,
			users.login AS inviting_user__login,
			users.first_name AS inviting_user__first_name,
			users.last_name AS inviting_user__last_name,
			groups.id AS group__id,
			groups.name AS group__name,
			groups.description AS group__description,
			groups.type AS group__type`).
		Joins("LEFT JOIN users ON users.group_id = initiator_id AND action = 'invitation_created'").
		Joins("JOIN `groups` ON `groups`.id = group_membership_changes.group_id").
		Joins(`
			LEFT JOIN group_pending_requests
				ON group_pending_requests.group_id = group_membership_changes.group_id AND
					group_pending_requests.member_id = group_membership_changes.member_id AND
					IF(group_pending_requests.type = 'invitation', 'invitation_created', 'join_request_created') =
						group_membership_changes.action AND
					(SELECT MAX(latest_change.at) FROM group_membership_changes AS latest_change
					 WHERE latest_change.group_id = group_pending_requests.group_id AND
						latest_change.member_id = group_pending_requests.member_id AND
						latest_change.action = group_membership_changes.action) = group_membership_changes.at`).
		Where("action IN ('invitation_created', 'join_request_created', 'join_request_refused')").
		Where("action = 'join_request_refused' OR group_pending_requests.group_id IS NOT NULL").
		Where("group_membership_changes.member_id = ?", user.GroupID)

	if len(r.URL.Query()["within_weeks"]) > 0 {
		withinWeeks, err := service.ResolveURLQueryGetInt64Field(r, "within_weeks")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
		query = query.Where("NOW() - INTERVAL ? WEEK < group_membership_changes.at", withinWeeks)
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

	var result []invitationsViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
