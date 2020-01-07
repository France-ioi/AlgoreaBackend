package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupsMembersViewResponseRow
type groupsMembersViewResponseRow struct {
	// `groups_groups.id`
	// required: true
	ID int64 `json:"id,string"`
	// Nullable
	// required: true
	MemberSince *database.Time `json:"member_since"`
	// the latest `group_membership_changes.action`
	// enum: invitation_accepted,join_request_accepted,joined_by_code,added_directly
	// required: true
	Action string `json:"action"`
	// Nullable
	// required: true
	User *struct {
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
	} `json:"user" gorm:"embedded;embedded_prefix:user__"`
}

// swagger:operation GET /groups/{group_id}/members group-memberships groupsMembersView
// ---
// summary: List group members
// description: >
//
//   Returns a list of group members
//   (rows from the `groups_groups` table with `parent_group_id` = `group_id` and NOW() < `groups_groups.expires_at`).
//   Rows related to users contain basic user info (`first_name` and `last_name` are only shown if the user
//   approved access to their personal info for some group managed by the authenticated user).
//
//
//   The authenticated user should be a manager of `group_id`, otherwise the 'forbidden' error is returned.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: sort
//   in: query
//   default: [-member_since,id]
//   type: array
//   items:
//     type: string
//     enum: [member_since,-member_since,user.login,-user.login,user.grade,-user.grade,id,-id]
// - name: from.member_since
//   description: Start the page from the member next to the member with `groups_groups.member_since` = `from.member_since`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.user.login
//   description: Start the page from the member next to the member with `users.login` = `from.user.login`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.user.grade
//   description: Start the page from the member next to the member with `users.grade` = `from.user.grade`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: integer
// - name: from.id
//   description: Start the page from the member next to the member with `groups_groups.id`=`from.id`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: integer
// - name: limit
//   description: Display the first N members
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. The array of group members
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/groupsMembersViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getMembers(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserCanManageTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.GroupGroups().
		Select(`
			groups_groups.id,
			latest_change.at AS member_since,
			latest_change.action,
			users.group_id AS user__group_id,
			users.login AS user__login,
			IF(managed_groups_with_approval.id IS NOT NULL, users.first_name, NULL) AS user__first_name,
			IF(managed_groups_with_approval.id IS NOT NULL, users.last_name, NULL) AS user__last_name,
			users.grade AS user__grade`).
		Joins("LEFT JOIN users ON users.group_id = groups_groups.child_group_id").
		Joins("LEFT JOIN LATERAL ? AS managed_groups_with_approval ON 1",
			srv.Store.ActiveGroupAncestors().ManagedByUser(user).
				Joins(`
					JOIN groups_groups_active
						ON groups_groups_active.parent_group_id = groups_ancestors_active.child_group_id AND
						   groups_groups_active.personal_info_view_approved`).
				Where("groups_groups_active.child_group_id = users.group_id").
				Select("groups_groups_active.child_group_id AS id").
				Limit(1).
				SubQuery()).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT at, action FROM group_membership_changes USE INDEX(group_id_member_id_at_desc)
				WHERE group_membership_changes.group_id = groups_groups.parent_group_id AND
					group_membership_changes.member_id = groups_groups.child_group_id
				ORDER BY group_membership_changes.group_id, group_membership_changes.member_id, at DESC
				LIMIT 1
			) AS latest_change ON 1`).
		WhereGroupRelationIsActual().
		Where("groups_groups.parent_group_id = ?", groupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"user.login":   {ColumnName: "users.login"},
			"user.grade":   {ColumnName: "users.grade"},
			"member_since": {ColumnName: "member_since", FieldType: "time"},
			"id":           {ColumnName: "groups_groups.id", FieldType: "int64"}},
		"-member_since,id", "id", false)

	if apiError != service.NoError {
		return apiError
	}

	var result []groupsMembersViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())
	for index := range result {
		if result[index].User.GroupID == nil {
			result[index].User = nil
		}
	}

	render.Respond(w, r, result)
	return service.NoError
}
