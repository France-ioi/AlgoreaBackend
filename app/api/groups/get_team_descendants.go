package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id}/team-descendants groups groupTeamDescendantView
// ---
// summary: List team descendants of the group
// description: Returns all teams (`sType` = "Team") among the descendants of the given group
//
//   * The authenticated user should own the parent group.
// parameters:
// - name: group_id
//   in: path
//   required: true
//   type: integer
// - name: from.name
//   description: Start the page from the team next to the team with `sName` = `from.name` and `ID` = `from.id`
//                (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the team next to the team with `sName`=`from.name` and `ID`=`from.id`
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

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.Groups().
		Select("groups.ID, groups.sName, groups.iGrade").
		Joins(`
			JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID AND
				groups_ancestors.idGroupAncestor != groups_ancestors.idGroupChild AND
				groups_ancestors.idGroupAncestor = ?`, groupID).
		Where("groups.sType = 'Team'")
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"name": {ColumnName: "groups.sName", FieldType: "string"},
			"id":   {ColumnName: "groups.ID", FieldType: "int64"}},
		"name")
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

	var parentsResult []teamDescendantParent
	service.MustNotBeError(srv.Store.Groups().
		Select("parent_links.idGroupChild AS idLinkedGroup, groups.ID, groups.sName").
		Joins(`
			JOIN groups_groups AS parent_links ON parent_links.idGroupParent = groups.ID AND
				parent_links.sType = 'direct' AND parent_links.idGroupChild IN (?)`, groupIDs).
		Joins(`
			JOIN groups_ancestors AS parent_ancestors ON parent_ancestors.idGroupChild = groups.ID AND
				parent_ancestors.idGroupAncestor = ?`, groupID).
		Order("groups.ID").
		Scan(&parentsResult).Error())

	for _, parentsRow := range parentsResult {
		resultMap[parentsRow.LinkedGroupID].Parents = append(resultMap[parentsRow.LinkedGroupID].Parents, parentsRow)
	}

	var membersResult []teamDescendantMember
	service.MustNotBeError(srv.Store.Users().
		Select(`
			member_links.idGroupParent AS idLinkedGroup,
			users.idGroupSelf, users.ID, users.sFirstName, users.sLastName, users.sLogin, users.iGrade`).
		Joins(`
			JOIN groups_groups AS member_links ON
				member_links.sType IN ('direct', 'invitationAccepted', 'requestAccepted') AND
				member_links.idGroupChild = users.idGroupSelf AND
				member_links.idGroupParent IN (?)`, groupIDs).
		Order("member_links.idGroupParent, member_links.idGroupChild").
		Scan(&membersResult).Error())

	for _, membersRow := range membersResult {
		resultMap[membersRow.LinkedGroupID].Members = append(resultMap[membersRow.LinkedGroupID].Members, membersRow)
	}

	render.Respond(w, r, result)
	return service.NoError
}

type teamDescendantParent struct {
	// required:true
	ID int64 `sql:"column:ID" json:"id,string"`
	// required:true
	Name string `sql:"column:sName" json:"name"`

	LinkedGroupID int64 `sql:"column:idLinkedGroup" json:"-"`
}

type teamDescendantMember struct {
	// User's `ID`
	// required:true
	ID int64 `sql:"column:ID" json:"id,string"`
	// required:true
	SelfGroupID int64 `sql:"column:idGroupSelf" json:"self_group_id"`
	// Nullable
	FirstName *string `sql:"column:sFirstName" json:"first_name"`
	// Nullable
	LastName *string `sql:"column:sLastName" json:"last_name"`
	// required:true
	Login string `sql:"column:sLogin" json:"login"`
	// Nullable
	Grade *int32 `sql:"column:iGrade" json:"grade"`

	LinkedGroupID int64 `sql:"column:idLinkedGroup" json:"-"`
}

// swagger:model
type teamDescendant struct {
	// The team's `groups.ID`
	// required:true
	ID int64 `sql:"column:ID" json:"id,string"`
	// required:true
	Name string `sql:"column:sName" json:"name"`
	// required:true
	Grade int32 `sql:"column:iGrade" json:"grade"`

	// Team's parent groups among the input group's descendants
	// required:true
	Parents []teamDescendantParent `sql:"-" json:"parents"`
	// Team's member users
	// required:true
	Members []teamDescendantMember `sql:"-" json:"members"`
}
