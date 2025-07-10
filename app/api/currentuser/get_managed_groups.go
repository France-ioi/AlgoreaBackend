package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model managedGroupsGetResponseRow
type managedGroupsGetResponseRow struct {
	// group's `id`
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	Name string `json:"name"`
	// required:true
	Description *string `json:"description"`
	// required:true
	// enum: Class,Team,Club,Friends,Other,Session,Base
	Type string `json:"type"`

	CanManageValue int `json:"-"`
	// required:true
	// enum: none,memberships,memberships_and_group
	CanManage string `json:"can_manage"`
	// required:true
	CanGrantGroupAccess bool `json:"can_grant_group_access"`
	// required:true
	CanWatchMembers bool `json:"can_watch_members"`
}

// swagger:operation GET /current-user/managed-groups groups managedGroupsView
//
//	---
//	summary: List groups managed by the current user
//	description:
//		Returns groups for which the current user is a manager (subgroups are skipped)
//	parameters:
//		- name: sort
//			in: query
//			default: [type,name,id]
//			type: array
//			items:
//				type: string
//				enum: [type,-type,name,-name,id,-id]
//		- name: from.id
//			description: Start the page from the group next to one with `id`=`{from.id}`
//			in: query
//			type: integer
//			format: int64
//		- name: limit
//			description: Display the first N groups
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. Success response with an array of managed groups
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/managedGroupsGetResponseRow"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getManagedGroups(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	query := store.Groups().
		Joins("JOIN group_managers ON group_managers.group_id = groups.id").
		Joins(`
			JOIN groups_ancestors_active AS user_ancestors
				ON user_ancestors.ancestor_group_id = group_managers.manager_id AND
					user_ancestors.child_group_id = ?`, user.GroupID).
		Select(`
			groups.id, groups.name, groups.type, groups.description,
			MAX(can_manage_value) AS can_manage_value,
			MAX(can_grant_group_access) AS can_grant_group_access,
			MAX(can_watch_members) AS can_watch_members`).
		Group("groups.id")

	query = service.NewQueryLimiter().Apply(httpRequest, query)
	query, err := service.ApplySortingAndPaging(
		httpRequest, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"type": {ColumnName: "groups.type"},
				"name": {ColumnName: "groups.name"},
				"id":   {ColumnName: "groups.id"},
			},
			DefaultRules: "type,name,id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	service.MustNotBeError(err)

	var result []managedGroupsGetResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	groupManagerStore := store.GroupManagers()
	for index := range result {
		result[index].CanManage = groupManagerStore.CanManageNameByIndex(result[index].CanManageValue)
	}

	render.Respond(responseWriter, httpRequest, &result)
	return nil
}
