package currentuser

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
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
	// required:true
	RootActivityID *int64 `json:"root_activity_id,string"`
	// required:true
	ExpectedStart *database.Time `json:"expected_start"`
	// required:true
	Organizer *string `json:"organizer"`
	// required:true
	Code *string `json:"code"`
	// required:true
	CodeExpiresAt *database.Time `json:"code_expires_at"`

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
//		Returns groups for which the current user is a manager (subgroups are skipped).
//
//
//		Optional `type` keeps only groups of that type.
//
//
//		Optional `root_activity_id_among` (comma-separated item ids) keeps only groups whose
//		`root_activity_id` is one of them. When the parameter is present but empty, no group matches.
//	parameters:
//		- name: type
//			in: query
//			type: string
//			enum: [Class,Team,Club,Friends,Other,Session,Base]
//		- name: root_activity_id_among
//			in: query
//			type: array
//			items:
//				type: integer
//				format: int64
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

	groupType, err := resolveManagedGroupsTypeFilter(httpRequest)
	if err != nil {
		return err
	}
	rootActivityIDs, rootActivityIDAmongSet, err := resolveManagedGroupsRootActivityIDAmongFilter(httpRequest)
	if err != nil {
		return err
	}

	query := store.Groups().
		Joins("JOIN group_managers ON group_managers.group_id = groups.id").
		Joins(`
			JOIN groups_ancestors_active AS user_ancestors
				ON user_ancestors.ancestor_group_id = group_managers.manager_id AND
					user_ancestors.child_group_id = ?`, user.GroupID).
		Select(`
			groups.id, groups.name, groups.type, groups.description,
			groups.root_activity_id, groups.expected_start, groups.organizer, groups.code, groups.code_expires_at,
			MAX(can_manage_value) AS can_manage_value,
			MAX(can_grant_group_access) AS can_grant_group_access,
			MAX(can_watch_members) AS can_watch_members`).
		Group("groups.id")

	if groupType != "" {
		query = query.Where("groups.type = ?", groupType)
	}
	if rootActivityIDAmongSet {
		if len(rootActivityIDs) == 0 {
			query = query.Where("FALSE")
		} else {
			query = query.Where("groups.root_activity_id IN (?)", rootActivityIDs)
		}
	}

	query = service.NewQueryLimiter().Apply(httpRequest, query)
	query, err = service.ApplySortingAndPaging(
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

func resolveManagedGroupsTypeFilter(httpRequest *http.Request) (string, error) {
	if !service.URLQueryPathHasField(httpRequest, "type") {
		return "", nil
	}
	groupType := httpRequest.URL.Query().Get("type")
	if !map[string]bool{
		"Class": true, "Team": true, "Club": true, "Friends": true, "Other": true, "Session": true, "Base": true,
	}[groupType] {
		return "", service.ErrInvalidRequest(fmt.Errorf("wrong value for type: %q", groupType))
	}
	return groupType, nil
}

func resolveManagedGroupsRootActivityIDAmongFilter(httpRequest *http.Request) (ids []int64, set bool, err error) {
	if !service.URLQueryPathHasField(httpRequest, "root_activity_id_among") {
		return nil, false, nil
	}
	ids, err = service.ResolveURLQueryGetInt64SliceField(httpRequest, "root_activity_id_among")
	if err != nil {
		return nil, true, service.ErrInvalidRequest(err)
	}
	return ids, true, nil
}
