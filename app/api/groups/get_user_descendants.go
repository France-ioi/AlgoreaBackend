package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

// swagger:operation GET /groups/{group_id}/user-descendants group-memberships groupUserDescendantView
//
//	---
//	summary: List group's user descendants
//	description: Return all users (`type` = "User") among the descendants of the given group
//
//
//		`first_name` and `last_name` of descendant users are only visible to the users themselves and
//		to managers of those groups to which those users provided view access to personal data.
//
//
//		* The authenticated user should be a manager of the parent group.
//	parameters:
//		- name: group_id
//			in: path
//			required: true
//			type: integer
//			format: int64
//		- name: from.id
//			description: Start the page from the user next to the user with `group_id`=`{from.id}`
//			in: query
//			type: integer
//			format: int64
//		- name: sort
//			in: query
//			default: [name,id]
//			type: array
//			items:
//				type: string
//				enum: [name,-name,id,-id]
//		- name: limit
//			description: Display the first N teams
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. Success response with an array of users
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/userDescendant"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getUserDescendants(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	groupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	service.MustNotBeError(checkThatUserCanManageTheGroup(store, user, groupID))

	query := store.Groups().
		Select(`
			groups.id, groups.name,
			users.group_id = ? OR MAX(personal_info_view_approvals.approved) AS show_personal_info,
			IF(users.group_id = ? OR MAX(personal_info_view_approvals.approved), users.first_name, NULL) AS first_name,
			IF(users.group_id = ? OR MAX(personal_info_view_approvals.approved), users.last_name, NULL) AS last_name,
			users.login, users.grade`, user.GroupID, user.GroupID, user.GroupID).
		Joins("JOIN groups_groups_active ON groups_groups_active.child_group_id = groups.id").
		Joins(`
			JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.parent_group_id AND
				groups_ancestors_active.ancestor_group_id = ?`, groupID).
		Joins("JOIN users ON users.group_id = groups.id").
		Group("groups.id").
		WithPersonalInfoViewApprovals(user)
	query = service.NewQueryLimiter().Apply(httpRequest, query)
	query, err = service.ApplySortingAndPaging(
		httpRequest, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"name": {ColumnName: "groups.name"},
				"id":   {ColumnName: "groups.id"},
			},
			DefaultRules: "name,id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	service.MustNotBeError(err)

	var result []userDescendant
	service.MustNotBeError(query.Scan(&result).Error())

	groupIDs := make([]int64, 0, len(result))
	resultMap := make(map[int64]*userDescendant, len(result))
	for index, groupRow := range result {
		groupIDs = append(groupIDs, groupRow.ID)
		if !result[index].User.ShowPersonalInfo {
			result[index].User.UserPersonalInfo = nil
		}
		resultMap[groupRow.ID] = &result[index]
	}

	var parentsResult []descendantParent
	service.MustNotBeError(store.Groups().
		Select("parent_links.child_group_id AS linked_group_id, groups.id, groups.name").
		Joins(`
			JOIN groups_groups_active AS parent_links ON parent_links.parent_group_id = groups.id AND
				parent_links.child_group_id IN (?)`, groupIDs).
		Joins(`
			JOIN groups_ancestors_active AS parent_ancestors ON parent_ancestors.child_group_id = groups.id AND
				parent_ancestors.ancestor_group_id = ?`, groupID).
		Order("groups.id").
		Scan(&parentsResult).Error())

	for _, parentsRow := range parentsResult {
		resultMap[parentsRow.LinkedGroupID].Parents = append(resultMap[parentsRow.LinkedGroupID].Parents, parentsRow)
	}

	render.Respond(responseWriter, httpRequest, result)
	return nil
}

type userDescendantUser struct {
	*structures.UserPersonalInfo
	ShowPersonalInfo bool `json:"-"`

	// required:true
	Login string `json:"login"`
	// required:true
	Grade *int32 `json:"grade"`
}

// swagger:model
type userDescendant struct {
	// The user's self `groups.id`
	// required:true
	ID int64 `json:"id,string"`
	// The user's self `groups.name`
	// required:true
	Name string `json:"name"`
	// required:true
	User userDescendantUser `gorm:"embedded" json:"user"`

	// User's parent groups among the input group's descendants
	// required:true
	Parents []descendantParent `json:"parents" sql:"-"`
}
