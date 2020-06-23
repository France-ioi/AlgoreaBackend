package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupRequestsViewResponseRow
type groupRequestsViewResponseRow struct {
	// `group_membership_changes.member_id`
	// required: true
	MemberID int64 `json:"member_id,string"`
	// Nullable
	// required: true
	At *database.Time `json:"at"`
	// `group_membership_changes.action`
	// enum: invitation_created,join_request_created,invitation_refused,join_request_refused
	// required: true
	Action string `json:"action"`

	// required: true
	JoiningUser struct {
		// `users.group_id`
		// required: true
		GroupID *int64 `json:"group_id,string"`
		// required: true
		Login string `json:"login"`
		// Nullable
		// required: true
		FirstName *string `json:"first_name"`
		// Nullable
		// required: true
		LastName *string `json:"last_name"`
		// Nullable
		// required: true
		Grade *int32 `json:"grade"`
	} `json:"joining_user" gorm:"embedded;embedded_prefix:joining_user__"`

	// Nullable
	// required: true
	InvitingUser *struct {
		// `users.group_id`
		// required: true
		GroupID *int64 `json:"group_id,string"`
		// required: true
		Login string `json:"login"`
		// Nullable
		// required: true
		FirstName *string `json:"first_name"`
		// Nullable
		// required: true
		LastName *string `json:"last_name"`
	} `json:"inviting_user" gorm:"embedded;embedded_prefix:inviting_user__"`
}

// swagger:operation GET /groups/{group_id}/requests group-memberships groupRequestsView
// ---
// summary: List pending requests and invitations for a group
// description: >
//
//   Returns a list of group requests and invitations
//   (rows from the `group_membership_changes` table with `group_id` = `{group_id}` and
//   `action` = "invitation_created"/"join_request_created"/"invitation_refused"/"join_request_refused")
//   with basic info on joining (invited/requesting) users and inviting users.
//
//
//   When `old_rejections_weeks` is given, only those rejected invitations/requests
//   (`group_membership_changes.action` is "invitation_refused" or "join_request_refused") are shown
//   that are created in the last `old_rejections_weeks` weeks.
//   Otherwise all rejected invitations/requests are shown.
//
//
//   Invited user’s `first_name` and `last_name` are nulls
//   if `group_membership_changes.action` = "invitation_created" or "invitation_refused".
//
//
//   The authenticated user should be a manager of `group_id` with `can_manage` >= 'memberships',
//   otherwise the 'forbidden' error is returned. If the group is a user, the 'forbidden' error is returned as well.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: old_rejections_weeks
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [-at,member_id]
//   type: array
//   items:
//     type: string
//     enum: [at,-at,joining_user.login,-joining_user.login,action,-action,member_id,-member_id]
// - name: from.at
//   description: Start the page from the request/invitation next to the request/invitation with
//                `group_membership_changes.at` = `from.at`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.joining_user.login
//   description: Start the page from the request/invitation next to the request/invitation
//                whose joining user's login is `from.joining_user.login`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.action
//   description: Start the page from the request/invitation next to the request/invitation with
//                `group_membership_changes.action` = `from.action`, sorted numerically.
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.member_id
//   description: Start the page from the request/invitation next to the request/invitation with
//                `group_membership_changes.member_id`=`from.member_id`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
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
//     description: OK. The array of group requests/invitations
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/groupRequestsViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserCanManageTheGroupMemberships(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.GroupMembershipChanges().
		Select(`
			group_membership_changes.member_id,
			group_membership_changes.at,
			action,
			joining_user.group_id AS joining_user__group_id,
			joining_user.login AS joining_user__login,
			IF(action IN ('invitation_created', 'invitation_refused'), NULL, joining_user.first_name) AS joining_user__first_name,
			IF(action IN ('invitation_created', 'invitation_refused'), NULL, joining_user.last_name) AS joining_user__last_name,
			joining_user.grade AS joining_user__grade,
			inviting_user.group_id AS inviting_user__group_id,
			inviting_user.login AS inviting_user__login,
			inviting_user.first_name AS inviting_user__first_name,
			inviting_user.last_name AS inviting_user__last_name`).
		Joins(`
			LEFT JOIN users AS inviting_user
				ON inviting_user.group_id = initiator_id AND action = 'invitation_created'`).
		Joins(`JOIN users AS joining_user ON joining_user.group_id = member_id`).
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
		Where("action IN ('join_request_refused', 'invitation_refused') OR group_pending_requests.group_id IS NOT NULL").
		Where("action IN ('invitation_created', 'join_request_created', 'invitation_refused', 'join_request_refused')").
		Where("group_membership_changes.group_id = ?", groupID)

	if len(r.URL.Query()["rejections_within_weeks"]) > 0 {
		oldRejectionsWeeks, err := service.ResolveURLQueryGetInt64Field(r, "rejections_within_weeks")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
		query = query.Where(`
			group_membership_changes.action IN ('invitation_created', 'join_request_created') OR
			NOW() - INTERVAL ? WEEK < group_membership_changes.at`, oldRejectionsWeeks)
	}

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"action":             {ColumnName: "group_membership_changes.action", FieldType: "string"},
			"joining_user.login": {ColumnName: "joining_user.login", FieldType: "string"},
			"at":                 {ColumnName: "group_membership_changes.at", FieldType: "time"},
			"member_id":          {ColumnName: "group_membership_changes.member_id", FieldType: "int64"}},
		"-at,member_id", []string{"member_id"}, false)

	if apiError != service.NoError {
		return apiError
	}

	var result []groupRequestsViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())
	for index := range result {
		if result[index].InvitingUser.GroupID == nil {
			result[index].InvitingUser = nil
		}
	}

	render.Respond(w, r, result)
	return service.NoError
}
