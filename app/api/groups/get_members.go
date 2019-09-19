package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id}/members groups users groupsMemberView
// ---
// summary: List group members
// description: >
//
//   Returns a list of group members
//   (rows from the `groups_groups` table with `group_parent_id` = `group_id` and
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
//   default: [-status_date,id]
//   type: array
//   items:
//     type: string
//     enum: [status_date,-status_date,user.login,-user.login,user.grade,-user.grade,id,-id]
// - name: from.status_date
//   description: Start the page from the member next to the member with `groups_groups.status_date` = `from.status_date`
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
//         type: object
//         required: [id, status_date, type, user]
//         properties:
//           id:
//             description: "`groups_groups.ID`"
//             type: string
//             format: int64
//           status_date:
//             type: string
//             description: Nullable
//             format: date-time
//           type:
//             type: string
//             description: "`groups_groups.sType`"
//             enum: [invitationAccepted, requestAccepted, joinedByCode, direct]
//           user:
//             type: object
//             description: Nullable
//             required: [id, login, first_name, last_name, grade]
//             properties:
//               id:
//                 description: "`users.ID`"
//                 type: string
//                 format: int64
//               login:
//                 type: string
//               first_name:
//                 description: Nullable
//                 type: string
//               last_name:
//                 description: Nullable
//                 type: string
//               grade:
//                 description: Nullable
//                 type: integer
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
			groups_groups.status_date,
			groups_groups.type,
			users.id AS user__id,
			users.login AS user__login,
			users.first_name AS user__first_name,
			users.last_name AS user__last_name,
			users.grade AS user__grade`).
		Joins("LEFT JOIN users ON users.group_self_id = groups_groups.group_child_id").
		WhereGroupRelationIsActive().
		Where("groups_groups.group_parent_id = ?", groupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"user.login":  {ColumnName: "users.login"},
			"user.grade":  {ColumnName: "users.grade"},
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
