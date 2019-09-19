package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupRequestsViewResponseRow
type groupRequestsViewResponseRow struct {
	// `groups_groups.id`
	// required: true
	ID int64 `json:"id,string"`
	// Nullable
	// required: true
	StatusDate *database.Time `json:"status_date"`
	// `groups_groups.type`
	// enum: invitationSent,requestSent,invitationRefused,requestRefused
	// required: true
	Type string `json:"type"`

	// Nullable
	// required: true
	JoiningUser *struct {
		// `users.id`
		// required: true
		ID *int64 `json:"id,string"`
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
		// `users.id`
		// required: true
		ID *int64 `json:"id,string"`
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

// swagger:operation GET /groups/{group_id}/requests groups users groupRequestsView
// ---
// summary: List pending requests and invitations for a group
// description: >
//
//   Returns a list of group requests and invitations
//   (rows from the `groups_groups` table with `group_parent_id` = `group_id` and
//   `type` = "invitationSent"/"requestSent"/"invitationRefused"/"requestRefused")
//   with basic info on joining (invited/requesting) users and inviting users.
//
//
//   When `old_rejections_weeks` is given, only those rejected invitations/requests
//   (`groups_groups.type` is "invitationRefused" or "requestRefused") are shown
//   whose `status_date` has changed in the last `old_rejections_weeks` weeks.
//   Otherwise all rejected invitations/requests are shown.
//
//
//   Invited user’s `first_name` and `last_name` are nulls
//   if `groups_groups.type` = "invitationSent" or "invitationRefused".
//
//
//   The authenticated user should be an owner of `group_id`, otherwise the 'forbidden' error is returned.
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
//   default: [-status_date,id]
//   type: array
//   items:
//     type: string
//     enum: [status_date,-status_date,joining_user.login,-joining_user.login,type,-type,id,-id]
// - name: from.status_date
//   description: Start the page from the request/invitation next to the request/invitation with
//                `groups_groups.status_date` = `from.status_date`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.joining_user.login
//   description: Start the page from the request/invitation next to the request/invitation
//                whose joining user's login is `from.joining_user.login`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.type
//   description: Start the page from the request/invitation next to the request/invitation with
//                `groups_groups.type` = `from.type`, sorted numerically.
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the request/invitation next to the request/invitation with `groups_groups.id`=`from.id`
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

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.GroupGroups().
		Select(`
			groups_groups.id,
			groups_groups.status_date,
			groups_groups.type,
			joining_user.id AS joining_user__id,
			joining_user.login AS joining_user__login,
			IF(groups_groups.type IN ('invitationSent', 'invitationRefused'), NULL, joining_user.first_name) AS joining_user__first_name,
			IF(groups_groups.type IN ('invitationSent', 'invitationRefused'), NULL, joining_user.last_name) AS joining_user__last_name,
			joining_user.grade AS joining_user__grade,
			inviting_user.id AS inviting_user__id,
			inviting_user.login AS inviting_user__login,
			inviting_user.first_name AS inviting_user__first_name,
			inviting_user.last_name AS inviting_user__last_name`).
		Joins("LEFT JOIN users AS inviting_user ON inviting_user.id = groups_groups.user_inviting_id").
		Joins("LEFT JOIN users AS joining_user ON joining_user.group_self_id = groups_groups.group_child_id").
		Where("groups_groups.type IN ('invitationSent', 'requestSent', 'invitationRefused', 'requestRefused')").
		Where("groups_groups.group_parent_id = ?", groupID)

	if len(r.URL.Query()["rejections_within_weeks"]) > 0 {
		oldRejectionsWeeks, err := service.ResolveURLQueryGetInt64Field(r, "rejections_within_weeks")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
		query = query.Where(`
			groups_groups.type IN ('invitationSent', 'requestSent') OR
			NOW() - INTERVAL ? WEEK < groups_groups.status_date`, oldRejectionsWeeks)
	}

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"type":               {ColumnName: "groups_groups.type"},
			"joining_user.login": {ColumnName: "joining_user.login"},
			"status_date":        {ColumnName: "groups_groups.status_date", FieldType: "time"},
			"id":                 {ColumnName: "groups_groups.id", FieldType: "int64"}},
		"-status_date")

	if apiError != service.NoError {
		return apiError
	}

	var result []groupRequestsViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())
	for index := range result {
		if result[index].InvitingUser.ID == nil {
			result[index].InvitingUser = nil
		}
		if result[index].JoiningUser.ID == nil {
			result[index].JoiningUser = nil
		}
	}

	render.Respond(w, r, result)
	return service.NoError
}
