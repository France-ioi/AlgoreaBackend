package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id}/user-descendants groups users groupUserDescendantView
// ---
// summary: List user descendants of the group
// description: Return all users (`sType` = "UserSelf") among the descendants of the given group
//
//   * The authenticated user should own the parent group.
// parameters:
// - name: group_id
//   in: path
//   required: true
//   type: integer
// - name: from.name
//   description: Start the page from the user next to the user with self group's `sName` = `from.name` and `ID` = `from.id`
//                (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the user next to the user with self group's `sName`=`from.name` and `ID`=`from.id`
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
//     description: OK. Success response with an array of users
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/userDescendant"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getUserDescendants(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.Groups().
		Select(`
			groups.ID, groups.sName,
			users.ID AS idUser, users.sFirstName, users.sLastName, users.sLogin, users.iGrade`).
		Joins(`
			JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID AND
				groups_ancestors.idGroupAncestor != groups_ancestors.idGroupChild AND
				groups_ancestors.idGroupAncestor = ?`, groupID).
		Joins("JOIN users ON users.idGroupSelf = groups.ID").
		Where("groups.sType = 'UserSelf'")
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"name": {ColumnName: "groups.sName", FieldType: "string"},
			"id":   {ColumnName: "groups.ID", FieldType: "int64"}},
		"name")
	if apiError != service.NoError {
		return apiError
	}

	var result []userDescendant
	service.MustNotBeError(query.Scan(&result).Error())

	groupIDs := make([]int64, 0, len(result))
	resultMap := make(map[int64]*userDescendant, len(result))
	for index, groupRow := range result {
		groupIDs = append(groupIDs, groupRow.ID)
		resultMap[groupRow.ID] = &result[index]
	}

	var parentsResult []descendantParent
	service.MustNotBeError(srv.Store.Groups().
		Select("parent_links.idGroupChild AS idLinkedGroup, groups.ID, groups.sName").
		Joins(`
			JOIN groups_groups AS parent_links ON parent_links.idGroupParent = groups.ID AND
				parent_links.sType`+database.GroupRelationIsActiveCondition+` AND
				parent_links.idGroupChild IN (?)`, groupIDs).
		Joins(`
			JOIN groups_ancestors AS parent_ancestors ON parent_ancestors.idGroupChild = groups.ID AND
				parent_ancestors.idGroupAncestor = ?`, groupID).
		Order("groups.ID").
		Scan(&parentsResult).Error())

	for _, parentsRow := range parentsResult {
		resultMap[parentsRow.LinkedGroupID].Parents = append(resultMap[parentsRow.LinkedGroupID].Parents, parentsRow)
	}

	render.Respond(w, r, result)
	return service.NoError
}

type userDescendantUser struct {
	// The user's `users.ID`
	// required:true
	ID int64 `sql:"column:idUser" json:"id,string"`
	// Nullable
	// required:true
	FirstName *string `sql:"column:sFirstName" json:"first_name"`
	// Nullable
	// required:true
	LastName *string `sql:"column:sLastName" json:"last_name"`
	// required:true
	Login string `sql:"column:sLogin" json:"login"`
	// Nullable
	// required:true
	Grade *int32 `sql:"column:iGrade" json:"grade"`
}

// swagger:model
type userDescendant struct {
	// The user's self `groups.ID`
	// required:true
	ID int64 `sql:"column:ID" json:"id,string"`
	// The user's self `groups.sName`
	// required:true
	Name string `sql:"column:sName" json:"name"`
	// required:true
	User userDescendantUser `json:"user" gorm:"embedded"`

	// User's parent groups among the input group's descendants
	// required:true
	Parents []descendantParent `sql:"-" json:"parents"`
}
