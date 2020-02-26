package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupManagersViewResponseRow
type groupManagersViewResponseRow struct {
	// `groups.id`
	// required: true
	ID int64 `json:"id,string"`
	// `groups.name`
	// required: true
	Name string `json:"name"`
	// enum: none,memberships,memberships_and_group
	// required: true
	CanManage string `json:"can_manage"`
	// required: true
	CanGrantGroupAccess bool `json:"can_grant_group_access"`
	// required: true
	CanWatchMembers bool `json:"can_watch_members"`
}

// swagger:operation GET /groups/{group_id}/managers groups groupManagersView
// ---
// summary: List group managers
// description: >
//
//   Returns a list of group managers
//   (rows from the `group_managers` table with `group_id` = `{group_id}`) including managers' group names.
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
//   default: [name,id]
//   type: array
//   items:
//     type: string
//     enum: [name,-name,id,-id]
// - name: from.name
//   description: Start the page from the manager next to the manager with `groups.name` = `from.name`
//                and `groups.id`=`from.id` (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the manager next to the manager with `groups.id`=`from.id`
//                (depending on `sort`, `from.name` may be required when `from.id` is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display the first N managers
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. The array of group managers
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/groupManagersViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getManagers(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserCanManageTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.GroupManagers().Where("group_id = ?", groupID).
		Joins("JOIN `groups` ON groups.id = group_managers.manager_id").
		Select(`groups.id, groups.name, can_manage, can_grant_group_access, can_watch_members`)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"name": {ColumnName: "groups.name"},
			"id":   {ColumnName: "groups.id", FieldType: "int64"}},
		"name,id", []string{"id"}, false)

	if apiError != service.NoError {
		return apiError
	}

	var result []groupManagersViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
