package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/structures"
)

// swagger:operation GET /groups/{group_id}/team-descendants group-memberships groupTeamDescendantView
// ---
// summary: List group's team descendants
// description: Returns all teams (`type` = "Team") among the descendants of the given group
//
//
//   `first_name` and `last_name` of descendant team members are only visible to the members themselves and
//   to managers of those groups to which those members provided view access to personal data.
//
//
//   * The authenticated user should be a manager of the parent group.
// parameters:
// - name: group_id
//   in: path
//   required: true
//   type: integer
// - name: from.name
//   description: Start the page from the team next to the team with `name` = `from.name` and `id` = `from.id`
//                (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the team next to the team with `name`=`from.name` and `id`=`from.id`
//                (`from.name` is required when from.id is present)
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [name,id]
//   type: array
//   items:
//     type: string
//     enum: [name,-name,id,-id]
// - name: limit
//   description: Display the first N teams
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with an array of teams
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/teamDescendant"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getTeamDescendants(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserCanManageTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.Groups().
		Select("groups.id, groups.name, groups.grade").
		Joins(`
			JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups.id AND
				groups_ancestors_active.ancestor_group_id != groups_ancestors_active.child_group_id AND
				groups_ancestors_active.ancestor_group_id = ?`, groupID).
		Where("groups.type = 'Team'")
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"name": {ColumnName: "groups.name", FieldType: "string"},
			"id":   {ColumnName: "groups.id", FieldType: "int64"}},
		"name,id", []string{"id"}, false)
	if apiError != service.NoError {
		return apiError
	}

	var result []teamDescendant
	service.MustNotBeError(query.Scan(&result).Error())

	groupIDs := make([]int64, 0, len(result))
	resultMap := make(map[int64]*teamDescendant, len(result))
	for index, groupRow := range result {
		groupIDs = append(groupIDs, groupRow.ID)
		resultMap[groupRow.ID] = &result[index]
		result[index].Members = []teamDescendantMember{}
	}

	var parentsResult []descendantParent
	service.MustNotBeError(srv.Store.Groups().
		Select("parent_links.child_group_id AS linked_group_id, groups.id, groups.name").
		Joins(`
			JOIN groups_groups_active AS parent_links
			ON parent_links.parent_group_id = groups.id AND
				parent_links.child_group_id IN (?)`, groupIDs).
		Joins(`
			JOIN groups_ancestors_active AS parent_ancestors
			ON parent_ancestors.child_group_id = groups.id AND
				parent_ancestors.ancestor_group_id = ?`, groupID).
		Order("groups.id").
		Scan(&parentsResult).Error())

	for _, parentsRow := range parentsResult {
		resultMap[parentsRow.LinkedGroupID].Parents = append(resultMap[parentsRow.LinkedGroupID].Parents, parentsRow)
	}

	var membersResult []teamDescendantMember
	service.MustNotBeError(srv.Store.Users().
		Select(`
			member_links.parent_group_id AS linked_group_id,
			users.group_id,
			users.group_id = ? OR personal_info_view_approvals.approved AS show_personal_info,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.first_name, NULL) AS first_name,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.last_name, NULL) AS last_name,
			users.login, users.grade`,
			user.GroupID, user.GroupID, user.GroupID).
		Joins(`
			JOIN groups_groups_active AS member_links ON
				member_links.child_group_id = users.group_id AND
				member_links.parent_group_id IN (?)`, groupIDs).
		WithPersonalInfoViewApprovals(user).
		Order("member_links.parent_group_id, member_links.child_group_id").
		Scan(&membersResult).Error())

	for _, membersRow := range membersResult {
		if !membersRow.ShowPersonalInfo {
			membersRow.UserPersonalInfo = nil
		}
		resultMap[membersRow.LinkedGroupID].Members = append(resultMap[membersRow.LinkedGroupID].Members, membersRow)
	}

	render.Respond(w, r, result)
	return service.NoError
}

type teamDescendantMember struct {
	// required:true
	GroupID int64 `json:"group_id"`

	*structures.UserPersonalInfo
	ShowPersonalInfo bool `json:"-"`

	// required:true
	Login string `json:"login"`
	// Nullable
	// required:true
	Grade *int32 `json:"grade"`

	LinkedGroupID int64 `json:"-"`
}

// swagger:model
type teamDescendant struct {
	// The team's `groups.id`
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	Name string `json:"name"`
	// required:true
	Grade int32 `json:"grade"`

	// Team's parent groups among the input group's descendants
	// required:true
	Parents []descendantParent `sql:"-" json:"parents"`
	// Team's member users
	// required:true
	Members []teamDescendantMember `sql:"-" json:"members"`
}
