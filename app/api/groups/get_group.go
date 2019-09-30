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
	// enum: Class,Team,Club,Friends,Other,UserSelf
	Type string `json:"type"`
	// Nullable
	// required:true
	RedirectPath *string `json:"redirect_path"`
	// required:true
	Opened bool `json:"opened"`
	// required:true
	FreeAccess bool `json:"free_access"`
	// required:true
	OpenContest bool `json:"open_contest"`
	// required:true
	CurrentUserIsOwner bool `json:"current_user_is_owner"`
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
//   The authenticated user should be an owner of `group_id` OR a descendant of the group OR  the group's `free_access`=1,
//   otherwise the 'forbidden' error is returned.
//
//
//   Note: `code*` fields are omitted when the user is not an owner of the group.
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
			LEFT JOIN groups_ancestors
				ON groups_ancestors.child_group_id = groups.id AND groups_ancestors.ancestor_group_id = ?`, user.OwnedGroupID).
		Joins(`
			LEFT JOIN groups_ancestors AS groups_descendants
				ON groups_descendants.ancestor_group_id = groups.id AND groups_descendants.child_group_id = ?`, user.SelfGroupID).
		Joins(`
			LEFT JOIN groups_groups
				ON groups_groups.type `+database.GroupRelationIsActiveCondition+` AND
					groups_groups.parent_group_id = groups.id AND groups_groups.child_group_id = ?`, user.SelfGroupID).
		Where("groups_ancestors.id IS NOT NULL OR groups_descendants.id IS NOT NULL OR groups.free_access").
		Where("groups.id = ?", groupID).Select(
		`groups.id, groups.name, groups.grade, groups.description, groups.created_at,
			groups.type, groups.redirect_path, groups.opened, groups.free_access,
			IF(groups_ancestors.id IS NOT NULL, groups.code, NULL) AS code,
			IF(groups_ancestors.id IS NOT NULL, groups.code_lifetime, NULL) AS code_lifetime,
			IF(groups_ancestors.id IS NOT NULL, groups.code_expires_at, NULL) AS code_expires_at,
			groups.open_contest,
			groups_ancestors.id IS NOT NULL AS current_user_is_owner,
			groups_groups.id IS NOT NULL AS current_user_is_member`).Limit(1)

	var result groupViewResponse
	err = query.Scan(&result).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if !result.CurrentUserIsOwner {
		result.GroupViewResponseCodePart = nil
	}

	render.Respond(w, r, result)

	return service.NoError
}
