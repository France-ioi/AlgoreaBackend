package groups

import (
	"net/http"
	"time"

	"github.com/France-ioi/validator"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// Access rights to be set
// swagger:model
type updatePermissionsInput struct {
	// Can be changed to true only if the current user is an owner of the item
	IsOwner bool `json:"is_owner" validate:"is_owner"`
	// The current user should have `permissions_generated.can_grant_view_generated` >= the new value (or 'info' if the new value is 'enter')
	// in order to increase level of this permission
	CanView string `json:"can_view" validate:"oneof=none info content content_with_descendants solution,can_view"`
	// The granted `can_view` should be >= the new value
	// (or 'info' if the new value is 'enter', or 'solution' if the new value is 'solution_with_grant') and
	// the current user should have `permissions_generated.can_grant_view_generated` = 'solution_with_grant'
	// in order to increase level of this permission. Only owners can increase it to 'solution_with_grant'.
	CanGrantView string `json:"can_grant_view" validate:"oneof=none enter content content_with_descendants solution solution_with_grant,can_grant_view"` //nolint:lll
	// The granted `can_view` should be >= 'content' and
	// the current user should have `permissions_generated.can_watch_generated` = 'answer_with_grant'
	// in order to increase level of this permission. Only owners can increase it to 'answer_with_grant'.
	CanWatch string `json:"can_watch" validate:"oneof=none result answer answer_with_grant,can_watch"`
	// The granted `can_view` should be >= 'content' and
	// the current user should have permissions_generated.can_edit_generated = 'all_with_grant'
	// in order to increase level of this permission. Only owners can increase it to 'all_with_grant'.
	CanEdit string `json:"can_edit" validate:"oneof=none children all all_with_grant,can_edit"`
	// The granted `can_view` should be >= 'info' and
	// the current user should be an owner in order to change this field to true.
	CanMakeSessionOfficial bool `json:"can_make_session_official" validate:"can_make_session_official"`
	// The current user should have permissions_generated.can_grant_view_generated >= 'enter' in order to
	// decrease this field's value.
	CanEnterFrom time.Time `json:"can_enter_from" validate:"can_enter_from"`
	// The current user should have permissions_generated.can_grant_view_generated >= 'enter' in order to
	// increase this field's value.
	CanEnterUntil time.Time `json:"can_enter_until" validate:"can_enter_until"`
	// Optional
	// The current user should have `permissions_generated.can_grant_view` >= 'content',
	// in order to set this field's value.
	CanRequestHelpTo setCanRequestHelpTo `json:"can_request_help_to" validate:"can_request_help_to_set,can_request_help_to_consistent,can_request_help_to_can_grant_view_content,can_request_help_to_visible"` //nolint:lll
}

// Cannot set `id` and `is_all_users_group` at the same time,
// but one of them have to be set if `can_request_help_to` is provided.
type setCanRequestHelpTo struct {
	// The given group must be visible by both the current user and `group_id`.
	// Can be set to `null` to set the helper group to no group.
	// Optional
	ID *int64 `json:"id" sql:"column:can_request_help_to"`
	// Optional
	IsAllUsersGroup bool `json:"is_all_users_group"`
}

type userPermissions struct {
	CanViewValue           int
	CanGrantViewValue      int
	CanWatchValue          int
	CanEditValue           int
	CanEnterFrom           database.Time
	CanEnterUntil          database.Time
	CanMakeSessionOfficial bool
	IsOwner                bool
	CanRequestHelpTo       *int64
}

type managerGeneratedPermissions struct {
	CanGrantViewGeneratedValue int
	CanWatchGeneratedValue     int
	CanEditGeneratedValue      int
	IsOwnerGenerated           bool
}

// swagger:operation PUT /groups/{source_group_id}/permissions/{group_id}/{item_id} groups updatePermissions
//
//	---
//	summary: Update permissions
//	description: Let a manager of a group give/withdraw permissions on an item to a group.
//
//		* The user giving the access must be a manager (with `can_grant_group_access` permission)
//			of `{source_group_id}` which should be an ancestor of the `{group_id}`.
//
//		* The user giving the access must have `permissions_generated.can_grant_view` >= given `can_view`
//			for the item.
//
//		* The group must already have access to one of the parents of the item or the item itself. If it does not,
//			the item must be a root activity/skill for an ancestor of the group.
//	parameters:
//		- name: group_id
//			in: path
//			required: true
//			type: integer
//		- name: source_group_id
//			in: path
//			required: true
//			type: integer
//		- name: item_id
//			in: path
//			required: true
//			type: integer
//		- name: access rights information
//			in: body
//			required: true
//			schema:
//				"$ref": "#/definitions/updatePermissionsInput"
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) updatePermissions(w http.ResponseWriter, r *http.Request) service.APIError {
	sourceGroupID, err := service.ResolveURLQueryPathInt64Field(r, "source_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	rawRequestData, apiErr := service.ResolveJSONBodyIntoMap(r)
	service.MustBeNoError(apiErr)

	err = srv.GetStore(r).InTransaction(func(s *database.DataStore) error {
		apiErr = checkIfUserIsManagerAllowedToGrantPermissionsOnItem(s, user, sourceGroupID, groupID, itemID)
		if apiErr != service.NoError {
			return apiErr.Error
		}

		// Even if we select a single row from permissions_granted,
		// MAX() is used so that a row with the default values is returned even if no row exists.
		var currentPermissions userPermissions
		err = s.PermissionsGranted().
			WithExclusiveWriteLock().
			Where("group_id = ?", groupID).
			Where("item_id = ?", itemID).
			Where("source_group_id = ?", sourceGroupID).
			Where("origin = 'group_membership'").
			Select(`
				IFNULL(MAX(can_view_value), 1) AS can_view_value,
				IFNULL(MAX(can_grant_view_value), 1) AS can_grant_view_value,
				IFNULL(MAX(can_watch_value), 1) AS can_watch_value,
				IFNULL(MAX(can_edit_value), 1) AS can_edit_value,
				IFNULL(MAX(can_make_session_official), 0) AS can_make_session_official,
				IFNULL(MAX(can_enter_from), '9999-12-31 23:59:59') AS can_enter_from,
				IFNULL(MAX(can_enter_until), '9999-12-31 23:59:59') AS can_enter_until,
				MAX(can_request_help_to) AS can_request_help_to,
				IFNULL(MAX(is_owner), 0) AS is_owner`).
			Scan(&currentPermissions).Error()

		if !gorm.IsRecordNotFoundError(err) {
			service.MustNotBeError(err)
		}

		var managerPermissions managerGeneratedPermissions
		err = s.Permissions().Select(`
			MAX(can_grant_view_generated_value) AS can_grant_view_generated_value,
			MAX(can_watch_generated_value) AS can_watch_generated_value,
			MAX(can_edit_generated_value) AS can_edit_generated_value,
			MAX(is_owner_generated) AS is_owner_generated`).
			Joins("JOIN groups_ancestors_active AS ancestors ON ancestors.ancestor_group_id = permissions.group_id").
			Where("ancestors.child_group_id = ?", user.GroupID).
			Where("permissions.item_id = ?", itemID).
			Scan(&managerPermissions).Error()
		service.MustNotBeError(err)

		var dataMap map[string]interface{}
		var modified bool
		dataMap, modified, apiErr = parsePermissionsInputData(s, &managerPermissions, &currentPermissions, user, groupID, rawRequestData)
		if apiErr != service.NoError {
			return apiErr.Error // rollback
		}

		if modified {
			allUsersGroupID := domain.ConfigFromContext(r.Context()).AllUsersGroupID

			correctPermissionsDataMap(s, dataMap, &currentPermissions)
			if _, ok := dataMap["is_all_users_group"]; ok {
				delete(dataMap, "is_all_users_group")
				dataMap["can_request_help_to"] = &allUsersGroupID
			}

			savePermissionsIntoDB(groupID, itemID, sourceGroupID, dataMap, s)
		}
		return nil
	})

	if apiErr != service.NoError {
		return apiErr
	}

	service.MustNotBeError(err)

	response := service.Response[*struct{}]{Success: true, Message: "updated"}
	render.Respond(w, r, &response)

	return service.NoError
}

func registerOptionalValidator(data *formdata.FormData, tag, message string, validatorFunc func(fl validator.FieldLevel) bool) {
	data.RegisterValidation(tag, data.ValidatorSkippingUnsetFields(validatorFunc))
	data.RegisterTranslation(tag, message)
}

func parsePermissionsInputData(
	s *database.DataStore,
	managerPermissions *managerGeneratedPermissions,
	currentPermissions *userPermissions,
	user *database.User,
	groupID int64,
	rawRequestData map[string]interface{},
) (
	dataMap map[string]interface{}, modified bool, apiError service.APIError,
) {
	data := formdata.NewFormData(&updatePermissionsInput{})
	modifiedPtr := registerPermissionsValidators(s, managerPermissions, currentPermissions, user, groupID, data)

	err := data.ParseMapData(rawRequestData)
	if err != nil {
		return nil, false, service.ErrInvalidRequest(err)
	}
	return data.ConstructMapForDB(), *modifiedPtr, service.NoError
}

func registerPermissionsValidators(
	s *database.DataStore,
	managerPermissions *managerGeneratedPermissions,
	currentPermissions *userPermissions,
	user *database.User,
	groupID int64,
	data *formdata.FormData,
) *bool {
	var modified bool
	registerIsOwnerValidator(data, managerPermissions, currentPermissions, &modified)
	registerCanViewValidator(data, managerPermissions, currentPermissions, &modified, s)
	registerCanGrantViewValidator(data, managerPermissions, currentPermissions, &modified, s)
	registerCanWatchValidator(data, managerPermissions, currentPermissions, &modified, s)
	registerCanEditValidator(data, managerPermissions, currentPermissions, &modified, s)
	registerCanMakeSessionOfficialValidator(data, managerPermissions, currentPermissions, &modified, s)
	registerCanEnterFromValidator(data, managerPermissions, currentPermissions, &modified, s)
	registerCanEnterUntilValidator(data, managerPermissions, currentPermissions, &modified, s)
	registerCanRequestHelpToSetValidator(data, currentPermissions, &modified)
	registerCanRequestHelpToConsistentValidator(data)
	registerCanRequestHelpToVisibleValidator(data, currentPermissions, user, groupID, s)
	registerCanRequestHelpToCanGrantViewContent(data, managerPermissions, s)
	return &modified
}

func registerIsOwnerValidator(data *formdata.FormData, managerPermissions *managerGeneratedPermissions,
	currentPermissions *userPermissions, modified *bool,
) {
	registerOptionalValidator(data, "is_owner", "the value is not permitted", func(fl validator.FieldLevel) bool {
		newValue := fl.Field().Bool()
		if newValue && newValue != currentPermissions.IsOwner && !managerPermissions.IsOwnerGenerated {
			return false
		}
		if currentPermissions.IsOwner != newValue {
			*modified = true
		}
		return true
	})
}

func registerCanViewValidator(data *formdata.FormData, managerPermissions *managerGeneratedPermissions, currentPermissions *userPermissions,
	modified *bool, s *database.DataStore,
) {
	registerOptionalValidator(data, "can_view", "the value is not permitted", func(fl validator.FieldLevel) bool {
		newValue := fl.Field().String()
		if !checkIfPossibleToModifyCanView(newValue, currentPermissions, managerPermissions, s) {
			return false
		}
		newValueIndex := s.PermissionsGranted().ViewIndexByName(newValue)
		if currentPermissions.CanViewValue != newValueIndex {
			currentPermissions.CanViewValue = newValueIndex
			*modified = true
		}
		return true
	})
}

func registerCanGrantViewValidator(
	data *formdata.FormData, managerPermissions *managerGeneratedPermissions, currentPermissions *userPermissions,
	modified *bool, s *database.DataStore,
) {
	registerOptionalValidator(data, "can_grant_view", "the value is not permitted", func(fl validator.FieldLevel) bool {
		newValue := fl.Field().String()
		if !checkIfPossibleToModifyCanGrantView(newValue, currentPermissions, managerPermissions, s) {
			return false
		}
		newValueIndex := s.PermissionsGranted().GrantViewIndexByName(newValue)
		if currentPermissions.CanGrantViewValue != newValueIndex {
			currentPermissions.CanGrantViewValue = newValueIndex
			*modified = true
		}
		return true
	})
}

func registerCanWatchValidator(
	data *formdata.FormData, managerPermissions *managerGeneratedPermissions, currentPermissions *userPermissions,
	modified *bool, s *database.DataStore,
) {
	registerOptionalValidator(data, "can_watch", "the value is not permitted", func(fl validator.FieldLevel) bool {
		newValue := fl.Field().String()
		if !checkIfPossibleToModifyCanWatch(newValue, currentPermissions, managerPermissions, s) {
			return false
		}
		newValueIndex := s.PermissionsGranted().WatchIndexByName(newValue)
		if currentPermissions.CanWatchValue != newValueIndex {
			currentPermissions.CanWatchValue = newValueIndex
			*modified = true
		}
		return true
	})
}

func registerCanEditValidator(
	data *formdata.FormData, managerPermissions *managerGeneratedPermissions, currentPermissions *userPermissions,
	modified *bool, s *database.DataStore,
) {
	registerOptionalValidator(data, "can_edit", "the value is not permitted", func(fl validator.FieldLevel) bool {
		newValue := fl.Field().String()
		if !checkIfPossibleToModifyCanEdit(newValue, currentPermissions, managerPermissions, s) {
			return false
		}
		newValueIndex := s.PermissionsGranted().EditIndexByName(newValue)
		if currentPermissions.CanEditValue != newValueIndex {
			currentPermissions.CanEditValue = newValueIndex
			*modified = true
		}
		return true
	})
}

func registerCanMakeSessionOfficialValidator(data *formdata.FormData, managerPermissions *managerGeneratedPermissions,
	currentPermissions *userPermissions, modified *bool, s *database.DataStore,
) {
	registerOptionalValidator(data, "can_make_session_official", "the value is not permitted", func(fl validator.FieldLevel) bool {
		newValue := fl.Field().Bool()
		if !checkIfPossibleToModifyCanMakeSessionOfficial(newValue, currentPermissions, managerPermissions, s) {
			return false
		}
		if currentPermissions.CanMakeSessionOfficial != newValue {
			currentPermissions.CanMakeSessionOfficial = newValue
			*modified = true
		}
		return true
	})
}

func registerCanEnterFromValidator(data *formdata.FormData, managerPermissions *managerGeneratedPermissions,
	currentPermissions *userPermissions, modified *bool, s *database.DataStore,
) {
	registerOptionalValidator(data, "can_enter_from", "the value is not permitted", func(fl validator.FieldLevel) bool {
		newValue := fl.Field().Interface().(time.Time)
		if !checkIfPossibleToModifyCanEnterFrom(newValue, currentPermissions, managerPermissions, s) {
			return false
		}
		if !newValue.Equal(time.Time(currentPermissions.CanEnterFrom)) {
			*modified = true
		}
		return true
	})
}

func registerCanEnterUntilValidator(
	data *formdata.FormData, managerPermissions *managerGeneratedPermissions, currentPermissions *userPermissions,
	modified *bool, s *database.DataStore,
) {
	registerOptionalValidator(data, "can_enter_until", "the value is not permitted", func(fl validator.FieldLevel) bool {
		newValue := fl.Field().Interface().(time.Time)
		if !checkIfPossibleToModifyCanEnterUntil(newValue, currentPermissions, managerPermissions, s) {
			return false
		}
		if !newValue.Equal(time.Time(currentPermissions.CanEnterUntil)) {
			*modified = true
		}
		return true
	})
}

func registerCanRequestHelpToSetValidator(data *formdata.FormData, currentPermissions *userPermissions, modified *bool) {
	registerOptionalValidator(
		data,
		"can_request_help_to_set",
		"can_request_help_to_set",
		func(fl validator.FieldLevel) bool {
			value := fl.Field().Interface().(setCanRequestHelpTo)

			if value.IsAllUsersGroup || canRequestHelpToIsModified(value.ID, currentPermissions.CanRequestHelpTo) {
				*modified = true
			}

			return true
		})
}

func registerCanRequestHelpToConsistentValidator(data *formdata.FormData) {
	registerOptionalValidator(
		data,
		"can_request_help_to_consistent",
		"cannot set can_request_help_to id and is_all_users_group at the same time",
		func(fl validator.FieldLevel) bool {
			value := fl.Field().Interface().(setCanRequestHelpTo)

			// Cannot set can_request_help_to and can_request_help_to_all_users at the same time.
			if value.ID != nil && value.IsAllUsersGroup {
				return false
			}

			return true
		})
}

func registerCanRequestHelpToVisibleValidator(
	data *formdata.FormData,
	currentPermissions *userPermissions,
	user *database.User,
	groupID int64,
	s *database.DataStore,
) {
	registerOptionalValidator(
		data,
		"can_request_help_to_visible",
		"can_request_help_to is not visible either by the current-user or the groupID",
		func(fl validator.FieldLevel) bool {
			value := fl.Field().Interface().(setCanRequestHelpTo)

			if value.ID != nil {
				// If the same value as before is provided, we don't check for visibility.
				if currentPermissions.CanRequestHelpTo != nil && *value.ID == *currentPermissions.CanRequestHelpTo {
					return true
				}

				// There is only one case for which the can_request_help_to group is visible by the receiving group but not the current-user.
				//
				// When the can_request_help_to group is visible to the receiving group, this can be because:
				// - The group is public.
				//   It is then also visible to the current-user.
				// - It is an ancestor of a group joined by the receiving group (or the joined group itself).
				//   In this case, the group is an ancestor of a group the current-user can manage,
				//   and is therefore visible too.
				// - It is managed by the receiving group, or an ancestor of a managed group.
				//   In this case only, can_request_help_to might not be visible by the current-user.
				if !s.Groups().IsVisibleFor(*value.ID, user) || !s.Groups().IsVisibleForGroup(*value.ID, groupID) {
					return false
				}
			}

			return true
		})
}

func registerCanRequestHelpToCanGrantViewContent(
	data *formdata.FormData,
	managerPermissions *managerGeneratedPermissions,
	s *database.DataStore,
) {
	registerOptionalValidator(
		data,
		"can_request_help_to_can_grant_view_content",
		"the current user doesn't have the right to update can_request_help_to",
		func(fl validator.FieldLevel) bool {
			// The current user must have can_grant_view >= content.

			return managerPermissions.CanGrantViewGeneratedValue >= s.PermissionsGranted().GrantViewIndexByName(content)
		})
}

func canRequestHelpToIsModified(newValue, oldValue *int64) bool {
	// Modified from one group to another group.
	if newValue != nil && oldValue != nil && *newValue != *oldValue {
		return true
	}

	// Modified from NULL to a group, or from a group to NULL.
	if (newValue != nil && oldValue == nil) || (newValue == nil && oldValue != nil) {
		return true
	}

	return false
}

const (
	none              = "none"
	enter             = "enter"
	info              = "info"
	content           = "content"
	solution          = "solution"
	solutionWithGrant = "solution_with_grant"
	answer            = "answer" //nolint:unused
	answerWithGrant   = "answer_with_grant"
	allWithGrant      = "all_with_grant"
)

func checkIfUserIsManagerAllowedToGrantPermissionsOnItem(s *database.DataStore, user *database.User,
	sourceGroupID, groupID, itemID int64,
) service.APIError {
	apiError := checkIfUserIsManagerAllowedToGrantPermissionsToGroupID(s, user, sourceGroupID, groupID)
	if apiError != service.NoError {
		return apiError
	}

	return checkIfItemOrOneOfItsParentsIsVisibleToGroupOrItemIsRoot(s, groupID, itemID)
}

func checkIfUserIsManagerAllowedToGrantPermissionsToGroupID(
	s *database.DataStore, user *database.User, sourceGroupID, groupID int64,
) service.APIError {
	// the authorized user should be a manager of the sourceGroupID with `can_grant_group_access' permission and
	// the 'sourceGroupID' should be an ancestor of 'groupID'
	found, err := s.Groups().ManagedBy(user).Where("groups.id = ?", sourceGroupID).
		Where("groups.type != 'User'").
		Joins(`
				JOIN groups_ancestors_active AS descendants
					ON descendants.ancestor_group_id = groups.id AND descendants.child_group_id = ?`, groupID).
		Where("group_managers.can_grant_group_access").
		HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}
	return service.NoError
}

func checkIfItemOrOneOfItsParentsIsVisibleToGroupOrItemIsRoot(s *database.DataStore, groupID, itemID int64) service.APIError {
	// at least one of the item's parents should be visible to the group
	found, err := s.Permissions().MatchingGroupAncestors(groupID).
		WherePermissionIsAtLeast("view", info).
		Joins("JOIN items_items ON items_items.parent_item_id = permissions.item_id").
		Where("items_items.child_item_id = ?", itemID).
		HasRows()
	service.MustNotBeError(err)
	if !found {
		// if not, the item itself should be visible to the group
		found, err = s.Permissions().MatchingGroupAncestors(groupID).WherePermissionIsAtLeast("view", info).
			Where("item_id = ?", itemID).HasRows()
		service.MustNotBeError(err)
		if !found {
			// if not, the item should be a root item for one of the group's ancestors
			found, err = s.Groups().
				Joins("JOIN groups_ancestors_active ON groups_ancestors_active.ancestor_group_id = groups.id").
				Where("groups_ancestors_active.child_group_id = ?", groupID).
				Where("root_activity_id = ? OR root_skill_id = ?", itemID, itemID).HasRows()
			service.MustNotBeError(err)
			if !found {
				return service.InsufficientAccessRightsError
			}
		}
	}
	return service.NoError
}

func checkIfPossibleToModifyCanView(viewPermissionToSet string, currentPermissions *userPermissions,
	managerPermissions *managerGeneratedPermissions, store *database.DataStore,
) bool {
	permissionGrantedStore := store.PermissionsGranted()
	if permissionGrantedStore.ViewIndexByName(viewPermissionToSet) <= currentPermissions.CanViewValue {
		return true
	}

	requiredGrantViewPermission := viewPermissionToSet
	if requiredGrantViewPermission == info { // no "info" in can_grant_view
		requiredGrantViewPermission = enter
	}

	// permissions_generated.can_grant_view_generated should be >= data["can_view"] for the manager
	return permissionGrantedStore.GrantViewIndexByName(requiredGrantViewPermission) <= managerPermissions.CanGrantViewGeneratedValue
}

func checkIfPossibleToModifyCanGrantView(grantViewPermissionToSet string, currentPermissions *userPermissions,
	managerPermissions *managerGeneratedPermissions, store *database.DataStore,
) bool {
	permissionGrantedStore := store.PermissionsGranted()
	if permissionGrantedStore.GrantViewIndexByName(grantViewPermissionToSet) <= currentPermissions.CanGrantViewValue {
		return true
	}

	requiredUserViewPermission := grantViewPermissionToSet
	switch requiredUserViewPermission {
	case enter: // no "enter" in can_view
		requiredUserViewPermission = info
	case solutionWithGrant:
		requiredUserViewPermission = solution
	}

	// permissions_granted.can_view should be >= requiredUserViewPermission for the user
	if currentPermissions.CanViewValue < permissionGrantedStore.ViewIndexByName(requiredUserViewPermission) {
		return false
	}

	if grantViewPermissionToSet == solutionWithGrant && !managerPermissions.IsOwnerGenerated {
		return false
	}

	return managerPermissions.CanGrantViewGeneratedValue == permissionGrantedStore.GrantViewIndexByName(solutionWithGrant)
}

func checkIfPossibleToModifyCanWatch(watchPermissionToSet string, currentPermissions *userPermissions,
	managerPermissions *managerGeneratedPermissions, store *database.DataStore,
) bool {
	permissionGrantedStore := store.PermissionsGranted()
	if permissionGrantedStore.WatchIndexByName(watchPermissionToSet) <= currentPermissions.CanWatchValue {
		return true
	}

	if currentPermissions.CanViewValue < permissionGrantedStore.ViewIndexByName(content) {
		return false
	}

	if watchPermissionToSet == answerWithGrant && !managerPermissions.IsOwnerGenerated {
		return false
	}

	return managerPermissions.CanWatchGeneratedValue == permissionGrantedStore.WatchIndexByName(answerWithGrant)
}

func checkIfPossibleToModifyCanEdit(editPermissionToSet string, currentPermissions *userPermissions,
	managerPermissions *managerGeneratedPermissions, store *database.DataStore,
) bool {
	permissionGrantedStore := store.PermissionsGranted()
	if permissionGrantedStore.EditIndexByName(editPermissionToSet) <= currentPermissions.CanEditValue {
		return true
	}

	if currentPermissions.CanViewValue < permissionGrantedStore.ViewIndexByName(content) {
		return false
	}

	if editPermissionToSet == allWithGrant && !managerPermissions.IsOwnerGenerated {
		return false
	}

	return managerPermissions.CanEditGeneratedValue == permissionGrantedStore.EditIndexByName(allWithGrant)
}

func checkIfPossibleToModifyCanMakeSessionOfficial(canMakeSessionOfficalToSet bool, currentPermissions *userPermissions,
	managerPermissions *managerGeneratedPermissions, store *database.DataStore,
) bool {
	if !canMakeSessionOfficalToSet || canMakeSessionOfficalToSet == currentPermissions.CanMakeSessionOfficial {
		return true
	}

	if currentPermissions.CanViewValue < store.PermissionsGranted().ViewIndexByName(info) {
		return false
	}

	return managerPermissions.IsOwnerGenerated
}

func checkIfPossibleToModifyCanEnterFrom(canEnterFromToSet time.Time, currentPermissions *userPermissions,
	managerPermissions *managerGeneratedPermissions, store *database.DataStore,
) bool {
	if !time.Time(currentPermissions.CanEnterFrom).After(canEnterFromToSet) {
		return true
	}

	return store.PermissionsGranted().GrantViewIndexByName(enter) <= managerPermissions.CanGrantViewGeneratedValue
}

func checkIfPossibleToModifyCanEnterUntil(canEnterUntilToSet time.Time, currentPermissions *userPermissions,
	managerPermissions *managerGeneratedPermissions, store *database.DataStore,
) bool {
	if !time.Time(currentPermissions.CanEnterUntil).Before(canEnterUntilToSet) {
		return true
	}

	return store.PermissionsGranted().GrantViewIndexByName(enter) <= managerPermissions.CanGrantViewGeneratedValue
}

func correctPermissionsDataMap(store *database.DataStore, dataMap map[string]interface{}, currentPermissions *userPermissions) {
	permissionGrantedStore := store.PermissionsGranted()

	currentCanViewValue := currentPermissions.CanViewValue
	if currentPermissions.CanGrantViewValue == permissionGrantedStore.GrantViewIndexByName(solutionWithGrant) &&
		currentCanViewValue != permissionGrantedStore.ViewIndexByName(solution) ||
		currentPermissions.CanGrantViewValue < permissionGrantedStore.GrantViewIndexByName(solutionWithGrant) &&
			currentCanViewValue < currentPermissions.CanGrantViewValue {
		dataMap["can_grant_view"] = permissionGrantedStore.GrantViewNameByIndex(currentCanViewValue)
	}
	if currentCanViewValue < permissionGrantedStore.ViewIndexByName(content) {
		if permissionGrantedStore.WatchIndexByName(none) < currentPermissions.CanWatchValue {
			dataMap["can_watch"] = none
		}
		if permissionGrantedStore.EditIndexByName(none) < currentPermissions.CanEditValue {
			dataMap["can_edit"] = none
		}
		if currentPermissions.CanMakeSessionOfficial && currentCanViewValue < permissionGrantedStore.ViewIndexByName(info) {
			dataMap["can_make_session_official"] = false
		}
	}
}

func savePermissionsIntoDB(groupID, itemID, sourceGroupID int64, dbMap map[string]interface{}, s *database.DataStore) {
	dbMap["latest_update_at"] = database.Now()

	columnsToUpdate := make([]string, 0, len(dbMap))
	for key := range dbMap {
		columnsToUpdate = append(columnsToUpdate, key)
	}

	dbMap["group_id"] = groupID
	dbMap["item_id"] = itemID
	dbMap["source_group_id"] = sourceGroupID
	dbMap["origin"] = "group_membership"

	permissionGrantedStore := s.PermissionsGranted()
	service.MustNotBeError(permissionGrantedStore.InsertOrUpdateMap(dbMap, columnsToUpdate))
	s.SchedulePermissionsPropagation()
	if dbMap["can_view"] != nil && dbMap["can_view"] != none || dbMap["is_owner"] != nil && dbMap["is_owner"].(bool) {
		// the permissions propagation implicitly (via triggers) marks some results as to_be_propagated
		// when an item becomes visible, so we should propagate results here
		s.ScheduleResultsPropagation()
	}
}
