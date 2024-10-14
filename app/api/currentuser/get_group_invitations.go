package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model
type invitationsViewResponseRow struct {
	// `group_membership_changes.group_id`
	// required: true
	GroupID int64 `json:"group_id,string"`

	// `groups_groups.type_changed_at`
	// required: true
	At database.Time `json:"at"`

	// the user who invited
	// required: true
	InvitingUser *invitingUser `json:"inviting_user" gorm:"embedded;embedded_prefix:inviting_user__"`

	// required: true
	Group groupWithApprovals `json:"group" gorm:"embedded;embedded_prefix:group__"`
}

type invitingUser struct {
	// `users.group_id`
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	Login string `json:"login"`
	// required: true
	FirstName *string `json:"first_name"`
	// required: true
	LastName *string `json:"last_name"`
}

type groupWithApprovals struct {
	// `groups.id`
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	Name string `json:"name"`
	// required: true
	Description *string `json:"description"`
	// required: true
	// enum: Class,Team,Club,Friends,Other,Session,Base
	Type string `json:"type"`
	// enum: none,view,edit
	// required: true
	RequirePersonalInfoAccessApproval string `json:"require_personal_info_access_approval"`
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
	store := srv.GetStore(r)

	query := store.GroupPendingRequests().
		Select(`
			group_pending_requests.group_id,
			IFNULL(latest_change.at, group_pending_requests.at) AS at,
			users.group_id AS inviting_user__id,
			users.login AS inviting_user__login,
			users.first_name AS inviting_user__first_name,
			users.last_name AS inviting_user__last_name,
			groups.id AS group__id,
			groups.name AS group__name,
			groups.description AS group__description,
			groups.type AS group__type,
			groups.require_personal_info_access_approval AS group__require_personal_info_access_approval,
			groups.require_lock_membership_approval_until AS group__require_lock_membership_approval_until,
			groups.require_watch_approval AS group__require_watch_approval
		`).
		// 'LEFT JOIN' as there can be no corresponding membership change
		Joins(`
			LEFT JOIN LATERAL (?) AS latest_change
				ON latest_change.action = 'invitation_created'`,
			store.GroupMembershipChanges().
				Select("initiator_id, action, at").
				Where("group_membership_changes.group_id = group_pending_requests.group_id").
				Where("group_membership_changes.member_id = group_pending_requests.member_id").
				Order("group_membership_changes.at DESC").
				Limit(1).QueryExpr()).
		// 'LEFT JOIN' as there can be no corresponding membership change and
		// 'initiator_id' can be NULL even if there is an invitation_created action
		Joins("LEFT JOIN users ON users.group_id = initiator_id").
		Joins("JOIN `groups` ON `groups`.id = group_pending_requests.group_id").
		Where("group_pending_requests.member_id = ?", user.GroupID).
		Where("group_pending_requests.type='invitation'")

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(
		r, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"at":       {ColumnName: "IFNULL(latest_change.at, group_pending_requests.at)"},
				"group_id": {ColumnName: "group_pending_requests.group_id"},
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

	for index := range result {
		if result[index].InvitingUser.ID == 0 {
			result[index].InvitingUser = nil
		}
	}

	render.Respond(w, r, result)
	return service.NoError
}
