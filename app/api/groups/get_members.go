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
	TypeChangedAt *database.Time `json:"type_changed_at"`
	// `groups_groups.type`
	// enum: invitationAccepted,requestAccepted,joinedByCode,direct
	// required: true
	Type string `json:"type"`
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

// swagger:operation GET /groups/{group_id}/members groups users groupsMembersView
// ---
// summary: List group members
// description: >
//
//   Returns a list of group members
//   (rows from the `groups_groups` table with `parent_group_id` = `group_id` and
//   `type` = "invitationAccepted"/"requestAccepted"/"joinedByCode"/"direct").
//   Rows related to users contain basic user info.
//
//
//   The authenticated user should be an owner of `group_id`, otherwise the 'forbidden' error is returned.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: sort
//   in: query
//   default: [-type_changed_at,id]
//   type: array
//   items:
//     type: string
//     enum: [type_changed_at,-type_changed_at,user.login,-user.login,user.grade,-user.grade,id,-id]
// - name: from.type_changed_at
//   description: Start the page from the member next to the member with `groups_groups.type_changed_at` = `from.type_changed_at`
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

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.GroupGroups().
		Select(`
			groups_groups.id,
			groups_groups.type_changed_at,
			groups_groups.type,
			users.group_id AS user__group_id,
			users.login AS user__login,
			users.first_name AS user__first_name,
			users.last_name AS user__last_name,
			users.grade AS user__grade`).
		Joins("LEFT JOIN users ON users.group_id = groups_groups.child_group_id").
		WhereGroupRelationIsActual().
		Where("groups_groups.parent_group_id = ?", groupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"user.login":      {ColumnName: "users.login"},
			"user.grade":      {ColumnName: "users.grade"},
			"type_changed_at": {ColumnName: "groups_groups.type_changed_at", FieldType: "time"},
			"id":              {ColumnName: "groups_groups.id", FieldType: "int64"}},
		"-type_changed_at,id", false)

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
