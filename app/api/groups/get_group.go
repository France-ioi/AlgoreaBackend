package groups

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// GroupGetResponseCodePart contains fields related to the group's code.
// These fields are only displayed if the current user is a manager of the group.
// swagger:ignore
type GroupGetResponseCodePart struct {
	// Returned only if the current user is a manager
	Code *string `json:"code"`
	// Returned only if the current user is a manager
	CodeLifetime *int32 `json:"code_lifetime"`
	// Returned only if the current user is a manager
	CodeExpiresAt *database.Time `json:"code_expires_at"`
}

// ManagerPermissionsPart contains fields related to permissions for managing the group.
// These fields are only displayed if the current user is a manager of the group.
// swagger:ignore
type ManagerPermissionsPart struct {
	CurrentUserCanManageValue int `json:"-"`
	// returned only if the current user is a manager
	// enum: none,memberships,memberships_and_group
	CurrentUserCanManage string `json:"current_user_can_manage"`
	// returned only if the current user is a manager
	CurrentUserCanGrantGroupAccess bool `json:"current_user_can_grant_group_access"`
	// returned only if the current user is a manager
	CurrentUserCanWatchMembers bool `json:"current_user_can_watch_members"`
}

// swagger:model groupGetResponse
type groupGetResponse struct {
	// required:true
	Grade int32 `json:"grade"`
	// required:true
	Description *string `json:"description"`
	// required:true
	CreatedAt *database.Time `json:"created_at"`
	// required:true
	// enum: Class,Team,Club,Friends,Other,Session,Base
	Type string `json:"type"`
	// required:true
	RootActivityID *int64 `json:"root_activity_id,string"`
	// required:true
	RootSkillID *int64 `json:"root_skill_id,string"`
	// required:true
	IsOpen bool `json:"is_open"`
	// required:true
	IsPublic bool `json:"is_public"`
	// required:true
	OpenActivityWhenJoining bool `json:"open_activity_when_joining"`
	// whether the user is a member of this group or one of its descendants
	// required:true
	// enum: none,direct,descendant
	CurrentUserMembership string `json:"current_user_membership"`
	// whether the user (or its ancestor) is a manager of this group,
	// or a manager of one of this group's ancestors (so is implicitly manager of this group) or,
	// a manager of one of this group's non-user descendants, or none of above
	// required: true
	// enum: none,direct,ancestor,descendant
	CurrentUserManagership string `json:"current_user_managership"`
	// list of descendant (excluding the group itself) groups that the current user is member of
	// required:true
	DescendantsCurrentUserIsMemberOf []structures.GroupShortInfo `json:"descendants_current_user_is_member_of"`
	// list of ancestor (excluding the group itself) groups that the current user (or his ancestor groups) is manager of
	// required:true
	AncestorsCurrentUserIsManagerOf []structures.GroupShortInfo `json:"ancestors_current_user_is_manager_of"`
	// list of descendant (excluding the group itself) non-user groups that the current user (or his ancestor groups) is manager of
	// required:true
	DescendantsCurrentUserIsManagerOf []structures.GroupShortInfo `json:"descendants_current_user_is_manager_of"`

	*structures.GroupShortInfo
	*GroupGetResponseCodePart
	*ManagerPermissionsPart

	// required: true
	IsMembershipLocked bool `json:"is_membership_locked"`
	// Only for joined teams
	// enum: frozen_membership,would_break_entry_conditions,free_to_leave
	CanLeaveTeam string `json:"can_leave_team,omitempty"`
	// required: true
	// enum: none,view,edit
	RequirePersonalInfoAccessApproval string `json:"require_personal_info_access_approval"`
	// required: true
	RequireLockMembershipApprovalUntil *database.Time `json:"require_lock_membership_approval_until"`
	// required: true
	RequireWatchApproval bool `json:"require_watch_approval"`
	// Only for joined groups
	CurrentUserHasPendingLeaveRequest *bool `json:"current_user_has_pending_leave_request,omitempty"`
	// Only for non-joined groups
	CurrentUserHasPendingInvitation *bool `json:"current_user_has_pending_invitation,omitempty"`
	// Only for non-joined groups
	CurrentUserHasPendingJoinRequest *bool `json:"current_user_has_pending_join_request,omitempty"`
}

// swagger:operation GET /groups/{group_id} groups groupGet
//
//	---
//	summary: Get a group
//	description: >
//
//		Returns the group identified by the given `group_id`.
//
//
//		The `group_id` group should be visible to the current user, so it should be either
//		an ancestor of a group he joined, or an ancestor of a non-user group he manages, or
//		a descendant of a group he manages, or a public group,
//		otherwise the 'forbidden' error is returned. If the group is a user or a contest participants group,
//		the 'forbidden' error is returned as well.
//
//
//		Note: `code*` and `current_user_can_*` fields are omitted when the user is not a manager of the group.
//	parameters:
//		- name: group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//	responses:
//		"200":
//			description: OK. The group info
//			schema:
//				"$ref": "#/definitions/groupGetResponse"
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
func (srv *Service) getGroup(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	groupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	query := store.Groups().PickVisibleGroups(store.Groups().DB, user).
		With("user_ancestors", ancestorsOfUserQuery(store, user)).
		Select(`
			groups.id, groups.type, groups.name,
			`+currentUserMembershipSQLColumn(user)+`,
			`+currentUserManagershipSQLColumn+`,
			groups.grade, groups.description, groups.created_at,
			groups.root_activity_id, groups.root_skill_id, groups.is_open, groups.is_public,
			groups.require_personal_info_access_approval, groups.require_lock_membership_approval_until, groups.require_watch_approval,
			IF(manager_access.found, groups.code, NULL) AS code,
			IF(manager_access.found, groups.code_lifetime, NULL) AS code_lifetime,
			IF(manager_access.found, groups.code_expires_at, NULL) AS code_expires_at,
			groups.open_activity_when_joining,
			IF(manager_access.found, manager_access.can_manage_value, 0) AS current_user_can_manage_value,
			IF(manager_access.found, manager_access.can_grant_group_access, 0) AS current_user_can_grant_group_access,
			IF(manager_access.found, manager_access.can_watch_members, 0) AS current_user_can_watch_members,
			groups_groups_active.lock_membership_approved AND NOW() < groups.require_lock_membership_approval_until AS is_membership_locked,
			IF(parent_group_id IS NOT NULL AND groups.type = 'Team',
				IF(groups.frozen_membership,
					'frozen_membership',
					IF(?,
						'would_break_entry_conditions',
						'free_to_leave'
					)
				),
				NULL
			) AS can_leave_team,
			`+currentUserHasPendingRequestSQLColumn("leave_request", user)+`,
			`+currentUserHasPendingRequestSQLColumn("invitation", user)+`,
			`+currentUserHasPendingRequestSQLColumn("join_request", user),
			store.Groups().GenerateQueryCheckingIfActionBreaksEntryConditionsForActiveParticipations(
				gorm.Expr("groups.id"), user.GroupID, false, false).SubQuery()).
		Joins(`
			LEFT JOIN ? AS manager_access ON child_group_id = groups.id`,
			store.GroupAncestors().ManagedByUser(user).
				Select(`
					1 AS found,
					MAX(can_manage_value) AS can_manage_value,
					MAX(can_grant_group_access) AS can_grant_group_access,
					MAX(can_watch_members) AS can_watch_members,
					groups_ancestors.child_group_id`).
				Where("groups_ancestors.child_group_id = ?", groupID).
				Group("groups_ancestors.child_group_id").SubQuery()).
		Joins(`
			LEFT JOIN groups_groups_active
				ON groups_groups_active.parent_group_id = groups.id AND groups_groups_active.child_group_id = ?`, user.GroupID).
		Where("groups.id = ?", groupID).
		Where("groups.type != 'User'").
		Where("groups.type != 'ContestParticipants'").
		Limit(1)

	var result groupGetResponse
	err = query.Scan(&result).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.ErrAPIInsufficientAccessRights
	}
	service.MustNotBeError(err)

	isManager := map[string]bool{"direct": true, "ancestor": true}[result.CurrentUserManagership]
	if !isManager {
		result.GroupGetResponseCodePart = nil
		result.ManagerPermissionsPart = nil
	} else {
		result.ManagerPermissionsPart.CurrentUserCanManage = store.GroupManagers().CanManageNameByIndex(result.CurrentUserCanManageValue)
	}

	if result.CurrentUserMembership != "none" {
		service.MustNotBeError(store.Groups().
			Joins(`
				JOIN groups_ancestors_active
					ON groups_ancestors_active.child_group_id = groups.id AND
					   NOT groups_ancestors_active.is_self AND
					   groups_ancestors_active.ancestor_group_id = ?`, groupID).
			Joins(`
				JOIN groups_groups_active
					ON groups_groups_active.parent_group_id = groups_ancestors_active.child_group_id AND
					   groups_groups_active.child_group_id = ?`, user.GroupID).
			Order("groups.name").
			Group("groups.id").
			Select("groups.id, groups.name").
			Scan(&result.DescendantsCurrentUserIsMemberOf).Error())
	}
	if result.DescendantsCurrentUserIsMemberOf == nil {
		result.DescendantsCurrentUserIsMemberOf = make([]structures.GroupShortInfo, 0)
	}
	if isManager {
		service.MustNotBeError(store.Groups().ManagedBy(user).
			Joins(`
				JOIN groups_ancestors_active AS groups_ancestors
					ON groups_ancestors.ancestor_group_id = groups.id AND
					   NOT groups_ancestors.is_self AND
					   groups_ancestors.child_group_id = ?`, groupID).
			Group("groups.id").
			Order("groups.name").
			Select("groups.id, groups.name").
			Scan(&result.AncestorsCurrentUserIsManagerOf).Error())
	}
	if result.AncestorsCurrentUserIsManagerOf == nil {
		result.AncestorsCurrentUserIsManagerOf = make([]structures.GroupShortInfo, 0)
	}
	if result.CurrentUserManagership != none {
		service.MustNotBeError(store.Groups().ManagedBy(user).
			Joins(`
				JOIN groups_ancestors_active AS groups_ancestors
					ON groups_ancestors.child_group_id = groups.id AND
					   NOT groups_ancestors.is_self AND
					   groups_ancestors.ancestor_group_id = ?`, groupID).
			Where("groups.type != 'User'").
			Group("groups.id").
			Order("groups.name").
			Select("groups.id, groups.name").
			Scan(&result.DescendantsCurrentUserIsManagerOf).Error())
	}
	if result.DescendantsCurrentUserIsManagerOf == nil {
		result.DescendantsCurrentUserIsManagerOf = make([]structures.GroupShortInfo, 0)
	}
	render.Respond(responseWriter, httpRequest, result)

	return nil
}

// currentUserMembershipSQLColumn returns an SQL column expression to get the current user membership
// (direct/descendant/none) in the group. The column name is `current_user_membership`.
func currentUserMembershipSQLColumn(currentUser *database.User) string {
	return fmt.Sprintf(`
		IF(
			EXISTS(
				SELECT 1 FROM groups_groups_active
				WHERE groups_groups_active.parent_group_id = groups.id AND
				      groups_groups_active.child_group_id = %d
			),
			'direct',
			IF(
				EXISTS(
					SELECT 1 FROM groups_groups_active
					JOIN groups_ancestors_active AS group_descendants
						ON group_descendants.ancestor_group_id = groups.id AND
						   group_descendants.child_group_id = groups_groups_active.parent_group_id
					WHERE groups_groups_active.child_group_id = %d
				),
				'descendant',
				'none'
			)
		) AS 'current_user_membership'`, currentUser.GroupID, currentUser.GroupID)
}

// currentUserManagershipSQLColumn is an SQL column expression to get the current user managership
// (direct/ancestor/descendant/none) of the group. The column name is `current_user_managership`.
const currentUserManagershipSQLColumn = `
		IF(
			EXISTS(
				SELECT 1 FROM user_ancestors
				JOIN group_managers
					ON group_managers.group_id = groups.id AND
					   group_managers.manager_id = user_ancestors.ancestor_group_id
			),
			'direct',
			IF(
				EXISTS(
					SELECT 1 FROM user_ancestors
					JOIN groups_ancestors_active AS group_ancestors ON group_ancestors.child_group_id = groups.id
					JOIN group_managers
						ON group_managers.group_id = group_ancestors.ancestor_group_id AND
						   group_managers.manager_id = user_ancestors.ancestor_group_id
				),
				'ancestor',
				IF(
					EXISTS(
						SELECT 1 FROM user_ancestors
						JOIN group_managers ON group_managers.manager_id = user_ancestors.ancestor_group_id
						JOIN groups_ancestors_active AS managed_groups
							ON managed_groups.ancestor_group_id = group_managers.group_id AND managed_groups.child_group_type != 'User'
						JOIN groups_ancestors_active AS group_descendants
							ON group_descendants.ancestor_group_id = groups.id AND
							   group_descendants.child_group_id = managed_groups.child_group_id
					),
					'descendant',
					'none'
				)
			)
		) AS 'current_user_managership'`

// ancestorsOfUserQuery returns a query to get the ancestors of the given user (as ancestor_group_id).
func ancestorsOfUserQuery(store *database.DataStore, user *database.User) *database.DB {
	return store.ActiveGroupAncestors().Where("child_group_id = ?", user.GroupID).Select("ancestor_group_id")
}

func currentUserHasPendingRequestSQLColumn(requestType string, user *database.User) string {
	return fmt.Sprintf(`
		IF(parent_group_id IS %sNULL,
			EXISTS(
				SELECT 1 FROM group_pending_requests
				WHERE group_pending_requests.group_id = groups.id AND
							group_pending_requests.member_id = %d AND
							group_pending_requests.type = '%s'
			),
			NULL
		) AS 'current_user_has_pending_%s'`,
		golang.If(requestType == "leave_request", "NOT "),
		user.GroupID, requestType, requestType)
}
