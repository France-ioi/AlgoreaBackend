package groups

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

// swagger:model groupUserRequestsViewResponseRow
type groupUserRequestsViewResponseRow struct {
	// required: true
	At *database.Time `json:"at"`

	// required: true
	// enum: join_request,leave_request
	Type string `json:"type"`

	// required: true
	Group struct {
		// required: true
		ID int64 `json:"id,string"`
		// required: true
		Name string `json:"name"`
	} `json:"group" gorm:"embedded;embedded_prefix:group__"`

	// required: true
	User struct {
		// `users.group_id`
		// required: true
		GroupID *int64 `json:"group_id,string"`
		// required: true
		Login string `json:"login"`

		*structures.UserPersonalInfo
		ShowPersonalInfo bool `json:"-"`

		// required: true
		Grade *int32 `json:"grade"`
	} `json:"user" gorm:"embedded;embedded_prefix:user__"`
}

// swagger:operation GET /groups/user-requests group-memberships groupUserRequestsView
//
//	---
//	summary: List pending requests for managed groups
//	description: >
//
//		Returns a list of group pending requests created by users with types listed in `{types}`
//		(rows from the `group_pending_requests` table) with basic info on joining/leaving users
//		for the group (if `{group_id}` is given) and
//		its descendants (if `{group_id}` is given and `{include_descendant_groups}` is 1)
//		or for all groups the current user can manage
//		(`can_manage` >= 'memberships') (if `{group_id}` is not given).
//
//
//		`first_name` and `last_name` are only shown for users whose personal info is visible to the current user.
//		A user can see personal info of his own and of those members/candidates of his managed groups
//		who have provided view access to their personal data.
//
//
//		If `{group_id}` is given, the authenticated user should be a manager of `group_id` with `can_manage` >= 'memberships',
//		otherwise the 'forbidden' error is returned. If the group is a user, the 'forbidden' error is returned as well.
//	parameters:
//		- name: group_id
//			in: query
//			type: integer
//			format: int64
//		- name: include_descendant_groups
//			in: query
//			type: integer
//			enum: [0,1]
//			default: 0
//		- name: types
//			in: query
//			default: [join_request]
//			type: array
//			items:
//				type: string
//				enum: [join_request,leave_request]
//		- name: sort
//			in: query
//			default: [group.id,-at,user.group_id]
//			type: array
//			items:
//				type: string
//				enum: [at,-at,user.login,-user.login,group.name,-group.name,user.group_id,-user.group_id,group.id,-group.id]
//		- name: from.group.id
//			description: Start the page from the request next to the request with
//							 `group_pending_requests.group_id`=`{from.group.id}`
//							 (only if `{group_id}` is not given; `{from.user.group_id}` is also required when `{from.group.id}` is given)
//			in: query
//			type: integer
//			format: int64
//		- name: from.user.group_id
//			description: Start the page from the request next to the request with
//							 `group_pending_requests.member_id`=`{from.user.group_id}`
//							 (`{from.group.id}` is also required if `{from.user.group_id}` is given and
//								either `{group_id}` is not given or descendants are included)
//			in: query
//			type: integer
//			format: int64
//		- name: limit
//			description: Display the first N requests
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. The array of pending group requests
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/groupUserRequestsViewResponseRow"
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
func (srv *Service) getUserRequests(w http.ResponseWriter, r *http.Request) error {
	store := srv.GetStore(r)
	groupID, groupIDSet, includeDescendantGroups, types, err := srv.resolveParametersForGetUserRequests(store, r)
	service.MustNotBeError(err)

	user := srv.GetUser(r)
	query := store.GroupPendingRequests().
		Select(`
			group_pending_requests.at,
			group_pending_requests.type,
			group.id AS group__id,
			group.name AS group__name,
			user.group_id AS user__group_id,
			user.login AS user__login,
			users_with_approval.group_id IS NOT NULL AS user__show_personal_info,
			IF(users_with_approval.group_id IS NOT NULL, user.first_name, NULL) AS user__first_name,
			IF(users_with_approval.group_id IS NOT NULL, user.last_name, NULL) AS user__last_name,
			user.grade AS user__grade`).
		Joins("JOIN `groups` AS `group` ON group.id = group_pending_requests.group_id").
		Joins(`LEFT JOIN users AS user ON user.group_id = member_id`).
		Joins(`LEFT JOIN users_with_approval ON users_with_approval.group_id = user.group_id`).
		Where("group_pending_requests.type IN (?)", types)
	tieBreakers := service.SortingAndPagingTieBreakers{
		"group.id":      service.FieldTypeInt64,
		"user.group_id": service.FieldTypeInt64,
	}
	if groupIDSet {
		if includeDescendantGroups {
			query = query.Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = group_pending_requests.group_id").
				Where("groups_ancestors_active.ancestor_group_id = ?", groupID)
		} else {
			query = query.Where("group_pending_requests.group_id = ?", groupID)
			tieBreakers = service.SortingAndPagingTieBreakers{"user.group_id": service.FieldTypeInt64}
		}
	} else {
		query = query.Where("group_pending_requests.group_id IN ?",
			store.ActiveGroupAncestors().ManagedByUser(user).Where("can_manage != 'none'").
				Select("groups_ancestors_active.child_group_id").SubQuery())
	}

	query = service.NewQueryLimiter().Apply(r, query)
	query, err = service.ApplySortingAndPaging(
		r, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"user.login":    {ColumnName: "user.login"},
				"user.group_id": {ColumnName: "group_pending_requests.member_id"},
				"at":            {ColumnName: "group_pending_requests.at"},
				"group.name":    {ColumnName: "group.name"},
				"group.id":      {ColumnName: "group_pending_requests.group_id"},
			},
			DefaultRules: "group.id,-at,user.group_id",
			TieBreakers:  tieBreakers,
		})
	service.MustNotBeError(err)

	query = attachUsersWithApproval(query, user)
	var result []groupUserRequestsViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	for index := range result {
		if !result[index].User.ShowPersonalInfo {
			result[index].User.UserPersonalInfo = nil
		}
	}

	render.Respond(w, r, result)
	return nil
}

func (srv *Service) resolveParametersForGetUserRequests(store *database.DataStore, r *http.Request) (
	groupID int64, groupIDSet, includeDescendantGroups bool, types []string, err error,
) {
	user := srv.GetUser(r)

	urlQuery := r.URL.Query()
	groupIDSet = len(urlQuery["group_id"]) > 0
	includeDescendantGroupsSet := len(urlQuery["include_descendant_groups"]) > 0
	if groupIDSet {
		groupID, err = service.ResolveURLQueryGetInt64Field(r, "group_id")
		if err != nil {
			return 0, false, false, nil, service.ErrInvalidRequest(err)
		}

		service.MustNotBeError(checkThatUserCanManageTheGroupMemberships(store, user, groupID))

		if includeDescendantGroupsSet {
			if includeDescendantGroups, err = service.ResolveURLQueryGetBoolField(r, "include_descendant_groups"); err != nil {
				return 0, false, false, nil, service.ErrInvalidRequest(err)
			}
		}
	}

	if !groupIDSet && includeDescendantGroupsSet {
		return 0, false, false, nil,
			service.ErrInvalidRequest(errors.New("'include_descendant_groups' should not be given when 'group_id' is not given"))
	}

	types, err = resolveTypesParameterForGetUserRequests(r)
	return groupID, groupIDSet, includeDescendantGroups, types, err
}

func resolveTypesParameterForGetUserRequests(r *http.Request) ([]string, error) {
	types := []string{"join_request"}
	urlQuery := r.URL.Query()
	if len(urlQuery["types"]) > 0 {
		types, _ = service.ResolveURLQueryGetStringSliceField(r, "types")
		for _, typ := range types {
			if !map[string]bool{"join_request": true, "leave_request": true}[typ] {
				return nil, service.ErrInvalidRequest(fmt.Errorf("wrong value in 'types': %q", typ))
			}
		}
	}
	return types, nil
}
