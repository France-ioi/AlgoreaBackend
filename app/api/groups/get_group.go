package groups

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// GroupViewResponseCodePart in order to make groupViewResponse work
// swagger:ignore
type GroupViewResponseCodePart struct {
	// Nullable
	Code *string `json:"code"`
	// Nullable
	CodeLifetime *string `json:"code_lifetime"`
	// Nullable
	CodeExpiresAt *database.Time `json:"code_expires_at"`
}

// swagger:model groupViewResponse
type groupViewResponse struct {
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
	// enum: Class,Team,Club,Friends,Other,Session
	Type string `json:"type"`
	// Nullable
	// required:true
	ActivityID *int64 `json:"activity_id,string"`
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

	*GroupViewResponseCodePart
}

// swagger:operation GET /groups/{group_id} groups groupView
// ---
// summary: Get group info
// description: >
//
//   Returns general information about the group from the `groups` table.
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
//       "$ref": "#/definitions/groupViewResponse"
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
				Select("1 AS found, groups_ancestors.child_group_id").SubQuery()).
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
			groups.type, groups.activity_id, groups.is_open, groups.is_public,
			IF(manager_access.found, groups.code, NULL) AS code,
			IF(manager_access.found, groups.code_lifetime, NULL) AS code_lifetime,
			IF(manager_access.found, groups.code_expires_at, NULL) AS code_expires_at,
			groups.open_activity_when_joining,
			manager_access.found AS current_user_is_manager,
			groups_groups_active.parent_group_id IS NOT NULL AS current_user_is_member`).
		Limit(1)

	var result groupViewResponse
	err = query.Scan(&result).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if !result.CurrentUserIsManager {
		result.GroupViewResponseCodePart = nil
	}

	render.Respond(w, r, result)

	return service.NoError
}
