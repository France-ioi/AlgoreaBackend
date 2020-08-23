package groups

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// GroupGetResponseCodePart contains fields related to the group's code.
// These fields are only displayed if the current user is a manager of the group.
// swagger:ignore
type GroupGetResponseCodePart struct {
	// Nullable
	Code *string `json:"code"`
	// Nullable
	CodeLifetime *string `json:"code_lifetime"`
	// Nullable
	CodeExpiresAt *database.Time `json:"code_expires_at"`
}

// GroupGetResponseManagerPermissionsPart contains fields related to permissions for managing the group.
// These fields are only displayed if the current user is a manager of the group.
// swagger:ignore
type GroupGetResponseManagerPermissionsPart struct {
	CurrentUserCanManageValue int `json:"-"`
	// required:true
	// enum: none,memberships,memberships_and_group
	CurrentUserCanManage string `json:"current_user_can_manage"`
	// required:true
	CurrentUserCanGrantGroupAccess bool `json:"current_user_can_grant_group_access"`
	// required:true
	CurrentUserCanWatchMembers bool `json:"current_user_can_watch_members"`
}

// swagger:model groupGetResponse
type groupGetResponse struct {
	// group's `id`
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	Name string `json:"name"`
	// required:true
	Grade int32 `json:"grade"`
	// Nullable
	// required:true
	Description *string `json:"description"`
	// Nullable
	// required:true
	CreatedAt *database.Time `json:"created_at"`
	// required:true
	// enum: Class,Team,Club,Friends,Other,Session,Base
	Type string `json:"type"`
	// Nullable
	// required:true
	RootActivityID *int64 `json:"root_activity_id,string"`
	// Nullable
	// required:true
	RootSkillID *int64 `json:"root_skill_id,string"`
	// required:true
	IsOpen bool `json:"is_open"`
	// required:true
	IsPublic bool `json:"is_public"`
	// required:true
	OpenActivityWhenJoining bool `json:"open_activity_when_joining"`
	// required:true
	CurrentUserIsManager bool `json:"current_user_is_manager"`
	// `True` when there is an active group->user relation in `groups_groups`
	// required:true
	CurrentUserIsMember bool `json:"current_user_is_member"`

	*GroupGetResponseCodePart
	*GroupGetResponseManagerPermissionsPart
}

// swagger:operation GET /groups/{group_id} groups groupGet
// ---
// summary: Get a group
// description: >
//
//   Returns the group identified by the given `group_id`.
//
//
//   The authenticated user should be a manager of `group_id` OR a descendant of the group OR  the group's `is_public`=1,
//   otherwise the 'forbidden' error is returned. If the group is a user, the 'forbidden' error is returned as well.
//
//
//   Note: `code*` fields are omitted when the user is not a manager of the group.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "200":
//     description: OK. The group info
//     schema:
//       "$ref": "#/definitions/groupGetResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	query := srv.Store.Groups().
		Joins(`
			LEFT JOIN ? AS manager_access ON child_group_id = groups.id`,
			srv.Store.GroupAncestors().ManagedByUser(user).
				Select(`
					1 AS found,
					MAX(can_manage_value) AS can_manage_value,
					MAX(can_grant_group_access) AS can_grant_group_access,
					MAX(can_watch_members) AS can_watch_members,
					groups_ancestors.child_group_id`).
				Where("groups_ancestors.child_group_id = ?", groupID).
				Group("groups_ancestors.child_group_id").SubQuery()).
		Joins(`
			LEFT JOIN groups_ancestors_active AS groups_descendants
				ON groups_descendants.ancestor_group_id = groups.id AND
					groups_descendants.child_group_id = ?`, user.GroupID).
		Joins(`
			LEFT JOIN groups_groups_active
				ON groups_groups_active.parent_group_id = groups.id AND groups_groups_active.child_group_id = ?`, user.GroupID).
		Where("manager_access.found OR groups_descendants.ancestor_group_id IS NOT NULL OR groups.is_public").
		Where("groups.id = ?", groupID).
		Where("groups.type != 'User'").
		Select(
			`groups.id, groups.name, groups.grade, groups.description, groups.created_at,
			groups.type, groups.root_activity_id, groups.root_skill_id, groups.is_open, groups.is_public,
			IF(manager_access.found, groups.code, NULL) AS code,
			IF(manager_access.found, groups.code_lifetime, NULL) AS code_lifetime,
			IF(manager_access.found, groups.code_expires_at, NULL) AS code_expires_at,
			groups.open_activity_when_joining,
			manager_access.found AS current_user_is_manager,
			IF(manager_access.found, manager_access.can_manage_value, 0) AS current_user_can_manage_value,
			IF(manager_access.found, manager_access.can_grant_group_access, 0) AS current_user_can_grant_group_access,
			IF(manager_access.found, manager_access.can_watch_members, 0) AS current_user_can_watch_members,
			groups_groups_active.parent_group_id IS NOT NULL AS current_user_is_member`).
		Limit(1)

	var result groupGetResponse
	err = query.Scan(&result).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if !result.CurrentUserIsManager {
		result.GroupGetResponseCodePart = nil
		result.GroupGetResponseManagerPermissionsPart = nil
	} else {
		result.GroupGetResponseManagerPermissionsPart.CurrentUserCanManage =
			srv.Store.GroupManagers().CanManageNameByIndex(result.CurrentUserCanManageValue)
	}

	render.Respond(w, r, result)

	return service.NoError
}
