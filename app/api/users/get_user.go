package users

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

const (
	personalInfoAccessApprovalNone = "none"
	personalInfoAccessApprovalView = "view"
	personalInfoAccessApprovalEdit = "edit"
)

// ManagerPermissionsPart contains fields related to permissions for managing the user.
// These fields are only displayed if the current user is a manager of the user.
// swagger:ignore
type ManagerPermissionsPart struct {
	CurrentUserIsManager bool `json:"-"`
	// returned only if the current user is a manager
	CurrentUserCanGrantUserAccess bool `json:"current_user_can_grant_user_access"`
	// returned only if the current user is a manager
	CurrentUserCanWatchUser bool `json:"current_user_can_watch_user"`
	// returned only if the current user is a manager
	// enum: none,view,edit
	PersonalInfoAccessApprovalToCurrentUser string `json:"personal_info_access_approval_to_current_user"`
}

// swagger:model
type userViewResponse struct {
	// required: true
	GroupID int64 `json:"group_id,string"`
	// required: true
	TempUser bool `json:"temp_user"`
	// required: true
	Login string `json:"login"`
	// required: true
	FreeText *string `json:"free_text"`
	// required: true
	WebSite *string `json:"web_site"`

	*structures.UserPersonalInfo
	ShowPersonalInfo bool `json:"-"`

	// list of ancestor (excluding the user himself) groups that the current user (or his ancestor groups) is manager of
	// required:true
	AncestorsCurrentUserIsManagerOf []structures.GroupShortInfo `json:"ancestors_current_user_is_manager_of"`

	*ManagerPermissionsPart

	// required: true
	IsCurrentUser bool `json:"is_current_user"`
}

// swagger:operation GET /users/{user_id} users userViewByID
//
//	---
//	summary: Get profile info for a user by ID
//	description: Returns data from the `users` table for the given `{user_id}`
//              (`first_name` and `last_name` are only shown for the authenticated user or
//               if the user approved access to their personal info for some group
//               managed by the authenticated user) along with some permissions if the current user is a manager.
//	parameters:
//		- name: user_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//	responses:
//		"200":
//			description: OK. Success response with user's data
//			schema:
//				"$ref": "#/definitions/userViewResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"404":
//			"$ref": "#/responses/notFoundResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"

// swagger:operation GET /users/by-login/{login} users userViewByLogin
//
//	---
//	summary: Get profile info for a user by login
//	description: >
//		Returns data from the `users` table for the given `{login}`
//		(`first_name` and `last_name` are only shown for the authenticated user or
//		if the user approved access to their personal info for some group
//		managed by the authenticated user) along with some permissions if the current user is a manager.
//	parameters:
//		- name: login
//			in: path
//			type: string
//			required: true
//	responses:
//		"200":
//			description: OK. Success response with user's data
//			schema:
//				"$ref": "#/definitions/userViewResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"404":
//			"$ref": "#/responses/notFoundResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getUser(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)

	var scope *database.DB
	store := srv.GetStore(httpRequest)
	if userLogin := chi.URLParam(httpRequest, "login"); userLogin != "" {
		scope = store.Users().Where("login = ?", userLogin)
	} else {
		userID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "user_id")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
		scope = store.Users().ByID(userID)
	}

	var userInfo userViewResponse
	err := scope.
		Select(`
			group_id, temp_user, login, free_text, web_site,
			users.group_id = ? OR personal_info_view_approvals.approved AS show_personal_info,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.first_name, NULL) AS first_name,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.last_name, NULL) AS last_name,
			manager_access.found AS current_user_is_manager,
			IF(manager_access.found, manager_access.can_grant_group_access, 0) AS current_user_can_grant_user_access,
			IF(manager_access.found, manager_access.can_watch_members, 0) AS current_user_can_watch_user`,
			user.GroupID, user.GroupID, user.GroupID).
		WithPersonalInfoViewApprovals(user).
		Joins(`
			LEFT JOIN LATERAL ? AS manager_access ON 1`,
			store.GroupAncestors().ManagedByUser(user).
				Select(`
					1 AS found,
					MAX(can_manage_value) AS can_manage_value,
					MAX(can_grant_group_access) AS can_grant_group_access,
					MAX(can_watch_members) AS can_watch_members,
					groups_ancestors.child_group_id`).
				Where("groups_ancestors.child_group_id = users.group_id").
				Group("groups_ancestors.child_group_id").SubQuery()).
		Scan(&userInfo).Error()

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return service.ErrNotFound(errors.New("no such user"))
	}
	service.MustNotBeError(err)

	if !userInfo.ShowPersonalInfo {
		userInfo.UserPersonalInfo = nil
	}

	if userInfo.CurrentUserIsManager {
		setUserInfosForManager(store, user, &userInfo)
	} else {
		userInfo.ManagerPermissionsPart = nil
		userInfo.AncestorsCurrentUserIsManagerOf = make([]structures.GroupShortInfo, 0)
	}

	userInfo.IsCurrentUser = userInfo.GroupID == user.GroupID

	render.Respond(responseWriter, httpRequest, &userInfo)
	return nil
}

type groupInfo struct {
	ID                                int64
	Name                              string
	RequirePersonalInfoAccessApproval string
}

// setUserInfosForManager sets the following fields in the response:
// - AncestorsCurrentUserIsManagerOf
// - PersonalInfoAccessApprovalToCurrentUser.
func setUserInfosForManager(store *database.DataStore, user *database.User, userInfo *userViewResponse) {
	var groupInfos []groupInfo

	service.MustNotBeError(store.Groups().ManagedBy(user).
		Joins(`
				JOIN groups_ancestors_active AS groups_ancestors
					ON groups_ancestors.ancestor_group_id = groups.id AND
						 NOT groups_ancestors.is_self AND
						 groups_ancestors.child_group_id = ?`, userInfo.GroupID).
		Group("groups.id").
		Order("groups.name").
		Select("groups.id, groups.name, groups.require_personal_info_access_approval").
		Scan(&groupInfos).Error())

	userInfo.AncestorsCurrentUserIsManagerOf = getGroupShortInfos(groupInfos)
	userInfo.PersonalInfoAccessApprovalToCurrentUser = computeHighestPersonalInfoAccessApproval(groupInfos)
}

// getGroupShortInfos returns a list of GroupShortInfo from the given groupInfos.
func getGroupShortInfos(groupInfos []groupInfo) []structures.GroupShortInfo {
	ancestorsCurrentUserIsManagerOf := make([]structures.GroupShortInfo, len(groupInfos))
	for i, groupInfoValue := range groupInfos {
		ancestorsCurrentUserIsManagerOf[i] = structures.GroupShortInfo{
			ID:   groupInfoValue.ID,
			Name: groupInfoValue.Name,
		}
	}

	return ancestorsCurrentUserIsManagerOf
}

// computeHighestPersonalInfoAccessApproval computes the highest personal info access approval ("edit" > "view" > "none").
func computeHighestPersonalInfoAccessApproval(groupInfos []groupInfo) string {
	highestPersonalInfoAccessApproval := personalInfoAccessApprovalNone
	for _, group := range groupInfos {
		if group.RequirePersonalInfoAccessApproval == personalInfoAccessApprovalEdit {
			highestPersonalInfoAccessApproval = personalInfoAccessApprovalEdit
			break
		} else if group.RequirePersonalInfoAccessApproval == personalInfoAccessApprovalView &&
			highestPersonalInfoAccessApproval == personalInfoAccessApprovalNone {
			highestPersonalInfoAccessApproval = personalInfoAccessApprovalView
		}
	}

	return highestPersonalInfoAccessApproval
}
