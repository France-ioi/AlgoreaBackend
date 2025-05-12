package groups

import (
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/France-ioi/validator"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

const (
	enumNone = "none"
	enumView = "view"
	enumEdit = "edit"
)

// Information of the group to be modified
// swagger:model
type groupUpdateInput struct {
	Name        string  `json:"name" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	Grade       int32   `json:"grade" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	Description *string `json:"description" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	IsOpen      bool    `json:"is_open" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	// If changed from true to false, is automatically switches all requests to join this group from requestSent to requestRefused
	IsPublic bool `json:"is_public" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	// Duration after the first use of the code when it will expire (in seconds)
	CodeLifetime   *int32         `json:"code_lifetime" validate:"changing_requires_can_manage_at_least=memberships,null|gte=0"`
	CodeExpiresAt  *database.Time `json:"code_expires_at" validate:"changing_requires_can_manage_at_least=memberships"`
	RootActivityID *int64         `json:"root_activity_id" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	RootSkillID    *int64         `json:"root_skill_id" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	// Can be set only if root_activity_id is set and
	// the current user has the 'can_make_session_official' permission on the activity item
	IsOfficialSession       bool `json:"is_official_session" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	OpenActivityWhenJoining bool `json:"open_activity_when_joining" validate:"changing_requires_can_manage_at_least=memberships_and_group"`

	// Can be changed only from false to true
	// (changing auto-rejects all pending join/leave requests and withdraws all pending invitations)
	FrozenMembership bool `json:"frozen_membership"  validate:"changing_requires_can_manage_at_least=memberships,frozen_membership"`
	// Cannot be set to null when enforce_max_participant is true
	MaxParticipants *int `json:"max_participants" validate:"changing_requires_can_manage_at_least=memberships,max_participants"`
	// Cannot be set to true when max_participants is null
	EnforceMaxParticipants bool `json:"enforce_max_participants" validate:"changing_requires_can_manage_at_least=memberships,enforce_max_participants"` //nolint:lll

	// Strengthened if the new value is `view` and the old value is `none`, or if the new value is `edit` and
	// the old value is either `view` or `none`.
	//
	// Not considered strengthened if the group doesn't have any participants.
	//
	// If it is strengthened, `approval_change_action` should be set.
	//
	// enum: none,view,edit
	RequirePersonalInfoAccessApproval string `json:"require_personal_info_access_approval" validate:"changing_requires_can_manage_at_least=memberships_and_group,oneof=none view edit,strengthening_requires_approval_change_action"` //nolint:lll
	// Strengthened if the new value is > `NOW()` and > the old value, or if the new value is > `NOW()` and the old value is `null`.
	//
	// Not considered strengthened if the group doesn't have any participants.
	//
	// If it is strengthened, `approval_change_action` must be set.
	RequireLockMembershipApprovalUntil *database.Time `json:"require_lock_membership_approval_until" validate:"changing_requires_can_manage_at_least=memberships_and_group,strengthening_requires_approval_change_action"` //nolint:lll
	// Strengthened if the new value is `true` and the old value is `false`.
	//
	// Not considered strengthened if the group doesn't have any participants.
	//
	// If it is strengthened, `approval_change_action` must be set.
	RequireWatchApproval       bool `json:"require_watch_approval" validate:"changing_requires_can_manage_at_least=memberships_and_group,strengthening_requires_approval_change_action"` //nolint:lll
	RequireMembersToJoinParent bool `json:"require_members_to_join_parent" validate:"changing_requires_can_manage_at_least=memberships_and_group"`                                       //nolint:lll

	// Must be present only if a `require_*` field is strengthened.
	//
	// If `empty`, we remove all participants from the group,
	// and we remove all the pending requests to the group.
	//
	// If `reinvite`, we remove all participants from the group,
	// we remove all the pending requests to the group,
	// and we invite all the participants again (`invitation` in `groups_pending_requests`).
	//
	// Additionally, if `require_lock_membership_approval_until` is strengthened,
	// all pending leave requests are removed.
	//
	// enum: empty,reinvite
	ApprovalChangeAction string `json:"approval_change_action" validate:"omitempty,oneof=empty reinvite,not_set_when_no_field_strengthened"`

	Organizer       *string        `json:"organizer" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	AddressLine1    *string        `json:"address_line1" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	AddressLine2    *string        `json:"address_line2" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	AddressPostcode *string        `json:"address_postcode" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	AddressCity     *string        `json:"address_city" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	AddressCountry  *string        `json:"address_country" validate:"changing_requires_can_manage_at_least=memberships_and_group"`
	ExpectedStart   *database.Time `json:"expected_start" validate:"changing_requires_can_manage_at_least=memberships_and_group"`

	CanManageValue int `json:"-"`
}

// swagger:operation PUT /groups/{group_id} groups groupUpdate
//
//	---
//	summary: Update a group
//	description: Updates group information.
//
//		Requires the user to be a manager of the group, otherwise the 'forbidden' error is returned.
//
//
//		The user should have `can_manage` >= 'memberships' in order to modify `code_expires_at`, `code_lifetime`,
//		`frozen_membership`, `max_participants`, or `enforce_max_participants`, otherwise the "bad request" error is returned.
//		In order to modify the values of other fields, the user should have `can_manage` >= 'memberships_and_group'.
//
//
//		If the `root_activity_id` item is provided and is not null, the item should not be a skill, and
//		the user should have at least 'can_view:info' permission on it, otherwise the 'forbidden' error is returned.
//
//
//		If the `root_skill_id` item is provided and is not null, the item should be a skill, and the user should have at least
//		'can_view:info' permission on it, otherwise the 'forbidden' error is returned.
//
//
//		 If `is_official_session` is being changed to true, the user should have at least
//		'can_make_session_official' permission on the activity item, otherwise the 'forbidden' error is returned.
//
//
//		Setting `is_official_session` to true while keeping `root_activity_id` not set or setting `root_activity_id` to null for
//		an official session will cause the "bad request" error.
//
//
//		Setting `enforce_max_participants` to true while keeping `max_participants` null or setting `max_participants` to null
//		while keeping `enforce_max_participants` = true will cause the "bad request" error.
//	parameters:
//		- name: group_id
//			in: path
//			required: true
//			type: integer
//			format: int64
//		- name: group information
//			in: body
//			required: true
//			schema:
//				"$ref": "#/definitions/groupUpdateInput"
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
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
func (srv *Service) updateGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	rawRequestData, apiErr := service.ResolveJSONBodyIntoMap(r)
	service.MustBeNoError(apiErr)

	err = srv.GetStore(r).InTransaction(func(s *database.DataStore) error {
		groupStore := s.Groups()

		var currentGroupData groupUpdateInput

		err = groupStore.ManagedBy(user).
			Select(`
				groups.name, groups.grade, groups.description, groups.is_open, groups.is_public,
				groups.code_lifetime, groups.code_expires_at, groups.root_activity_id, groups.root_skill_id,
				groups.is_official_session, groups.open_activity_when_joining, groups.require_members_to_join_parent,
				groups.frozen_membership, groups.organizer, groups.address_line1, groups.address_line2,
				groups.address_postcode, groups.address_city, groups.address_country, groups.expected_start,
				groups.max_participants, groups.enforce_max_participants, groups.require_personal_info_access_approval,
				groups.require_lock_membership_approval_until, groups.require_watch_approval,
				groups.require_members_to_join_parent,
				MAX(can_manage_value) AS can_manage_value`).WithExclusiveWriteLock().
			Where("groups.id = ?", groupID).Group("groups.id").Scan(&currentGroupData).Error()
		if gorm.IsRecordNotFoundError(err) {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}
		service.MustNotBeError(err)

		var groupHasParticipants bool
		groupHasParticipants, err = groupStore.HasParticipants(groupID)
		service.MustNotBeError(err)

		var formData *formdata.FormData
		formData, err = validateUpdateGroupInput(rawRequestData, groupHasParticipants, &currentGroupData, s)
		if err != nil {
			apiErr = service.ErrInvalidRequest(err)
			return apiErr.Error // rollback
		}

		dbMap := formData.ConstructMapForDB()

		var approvalChangeAction string
		if _, ok := dbMap["approval_change_action"]; ok {
			approvalChangeAction = dbMap["approval_change_action"].(string)
			delete(dbMap, "approval_change_action")
		}

		apiErr = validateRootActivityIDAndIsOfficial(s, user, currentGroupData.RootActivityID, currentGroupData.IsOfficialSession, dbMap)
		if apiErr != service.NoError {
			return apiErr.Error // rollback
		}
		apiErr = validateRootSkillID(s, user, currentGroupData.RootSkillID, dbMap)
		if apiErr != service.NoError {
			return apiErr.Error // rollback
		}

		if approvalChangeAction != "" {
			participantIDs := s.Groups().GetDirectParticipantIDsOf(groupID)

			// If the approval_change_action is 'reinvite', we need to reinvite the participants.
			if approvalChangeAction == "reinvite" {
				_, _, err = s.GroupGroups().Transition(database.AdminStrengthensApprovalWithReinvite, groupID, participantIDs, nil, user.GroupID)
				service.MustNotBeError(err)
			} else {
				_, _, err = s.GroupGroups().Transition(database.AdminStrengthensApprovalWithEmpty, groupID, participantIDs, nil, user.GroupID)
				service.MustNotBeError(err)
			}
		}

		service.MustNotBeError(refuseSentGroupRequestsIfNeeded(
			groupStore, groupID, user.GroupID, dbMap, &currentGroupData, approvalChangeAction))

		// update the group
		service.MustNotBeError(groupStore.Where("id = ?", groupID).UpdateColumns(dbMap).Error())

		return nil // commit
	})

	if apiErr != service.NoError {
		return apiErr
	}
	service.MustNotBeError(err)

	response := service.Response[*struct{}]{Success: true, Message: "updated"}
	render.Respond(w, r, &response)

	return service.NoError
}

func validateRootActivityIDAndIsOfficial(
	store *database.DataStore, user *database.User, oldRootActivityID *int64, oldIsOfficialSession bool,
	dbMap map[string]interface{},
) service.APIError {
	rootActivityIDToCheck := oldRootActivityID
	rootActivityID, rootActivityIDSet := dbMap["root_activity_id"]
	rootActivityIDChanged := rootActivityIDSet && !int64PtrEqualValues(oldRootActivityID, rootActivityID.(*int64))
	if rootActivityIDChanged {
		rootActivityIDToCheck = rootActivityID.(*int64)
		if rootActivityIDToCheck != nil {
			apiError := validateRootActivityID(store, user, rootActivityIDToCheck)
			if apiError != service.NoError {
				return apiError
			}
		}
	}

	if isTryingToChangeOfficialSessionActivity(dbMap, oldIsOfficialSession, rootActivityIDChanged) {
		if rootActivityIDToCheck == nil {
			return service.ErrInvalidRequest(errors.New("the root_activity_id should be set for official sessions"))
		}
		found, err := store.PermissionsGranted().WithExclusiveWriteLock().
			Joins(`
				JOIN groups_ancestors_active ON groups_ancestors_active.ancestor_group_id = permissions_granted.group_id AND
				     groups_ancestors_active.child_group_id = ?`, user.GroupID).
			Where("permissions_granted.can_make_session_official OR permissions_granted.is_owner").
			Where("permissions_granted.item_id = ?", rootActivityIDToCheck).
			HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrForbidden(errors.New("not enough permissions for attaching the group to the activity as an official session"))
		}
	}
	return service.NoError
}

func validateRootActivityID(store *database.DataStore, user *database.User, rootActivityIDToCheck *int64) service.APIError {
	if rootActivityIDToCheck != nil {
		found, errorInTransaction := store.Items().ByID(*rootActivityIDToCheck).Where("type != 'Skill'").WithExclusiveWriteLock().
			WhereUserHasViewPermissionOnItems(user, "info").HasRows()
		service.MustNotBeError(errorInTransaction)
		if !found {
			return service.ErrForbidden(errors.New("no access to the root activity or it is a skill"))
		}
	}
	return service.NoError
}

func validateRootSkillID(store *database.DataStore, user *database.User, oldRootSkillID *int64,
	dbMap map[string]interface{},
) service.APIError {
	newRootSkillIDInterface, newRootSkillIDSet := dbMap["root_skill_id"]
	if newRootSkillIDSet && newRootSkillIDInterface != nil {
		newRootSkillID := newRootSkillIDInterface.(*int64)

		rootSkillIDChanged := !int64PtrEqualValues(oldRootSkillID, newRootSkillID)
		if rootSkillIDChanged && newRootSkillID != nil {
			found, errorInTransaction := store.Items().
				ByID(*newRootSkillID).
				Where("type = 'Skill'").
				WithExclusiveWriteLock().
				WhereUserHasViewPermissionOnItems(user, "info").
				HasRows()
			service.MustNotBeError(errorInTransaction)
			if !found {
				return service.ErrForbidden(errors.New("no access to the root skill or it is not a skill"))
			}
		}
	}

	return service.NoError
}

// refuseSentGroupRequestsIfNeeded automatically refuses all requests to join this group
// (removes them from group_pending_requests and inserts appropriate group_membership_changes
// with `action` = 'join_request_refused') if is_public is changed from true to false.
func refuseSentGroupRequestsIfNeeded(
	store *database.GroupStore, groupID, initiatorID int64, dbMap map[string]interface{},
	currentGroupData *groupUpdateInput, approvalChangeAction string,
) error {
	if shouldRefuseGroupPendingRequests(dbMap, currentGroupData, approvalChangeAction) {
		pendingTypesToRefuse := getGroupPendingRequestTypesToRefuse(dbMap, currentGroupData, approvalChangeAction)

		service.MustNotBeError(store.Exec(`
			INSERT INTO group_membership_changes (group_id, member_id, action, at, initiator_id)
			SELECT group_id, member_id,
				CASE type
					WHEN 'join_request' THEN 'join_request_refused'
					WHEN 'leave_request' THEN 'leave_request_refused'
					WHEN 'invitation' THEN 'invitation_withdrawn'
				END,
				NOW(3), ?
			FROM group_pending_requests
			WHERE group_id = ? AND type IN (?)
			FOR UPDATE`, initiatorID, groupID, pendingTypesToRefuse).Error())

		// refuse sent group requests
		return store.GroupPendingRequests().
			Where("type IN (?)", pendingTypesToRefuse).
			Where("group_id = ?", groupID).
			Delete().Error()
	}
	return nil
}

func getGroupPendingRequestTypesToRefuse(
	dbMap map[string]interface{},
	currentGroupData *groupUpdateInput,
	approvalChangeAction string,
) []string {
	pendingTypesToHandle := []string{"join_request"}

	if newFrozenMembership, ok := dbMap["frozen_membership"]; ok && newFrozenMembership.(bool) && !currentGroupData.FrozenMembership {
		pendingTypesToHandle = append(pendingTypesToHandle, "leave_request", "invitation")
	}

	// If a require_* fields is strengthened, we want to refuse all pending join requests.
	if approvalChangeAction != "" {
		if _, ok := dbMap["require_lock_membership_approval_until"]; ok {
			newRequireLockMembershipApprovalUntil := dbMap["require_lock_membership_approval_until"].(*database.Time)

			// We can pass "true" for "groupHasParticipants" because we know there are participants, since approvalChangeAction is not nil.
			if requireLockMembershipApprovalUntilIsStrengthened(
				true,
				currentGroupData.RequireLockMembershipApprovalUntil,
				newRequireLockMembershipApprovalUntil,
			) {
				pendingTypesToHandle = append(pendingTypesToHandle, "leave_request")
			}
		}
	}
	return pendingTypesToHandle
}

func shouldRefuseGroupPendingRequests(
	dbMap map[string]interface{},
	currentGroupData *groupUpdateInput,
	approvalChangeAction string,
) bool {
	// If is_public is going to be changed from true to false
	if newIsPublic, ok := dbMap["is_public"]; ok && !newIsPublic.(bool) && currentGroupData.IsPublic {
		return true
	}
	if newFrozenMembership, ok := dbMap["frozen_membership"]; ok && newFrozenMembership.(bool) && !currentGroupData.FrozenMembership {
		return true
	}
	if approvalChangeAction != "" {
		return true
	}

	return false
}

func validateUpdateGroupInput(
	rawRequestData map[string]interface{}, groupHasParticipants bool, currentGroupData *groupUpdateInput, store *database.DataStore,
) (*formdata.FormData, error) {
	input := &groupUpdateInput{}
	formData := formdata.NewFormData(input)
	formData.SetOldValues(currentGroupData)
	formData.RegisterValidation("frozen_membership", constructFrozenMembershipValidator(formData))
	formData.RegisterTranslation("frozen_membership", "can only be changed from false to true")

	formData.RegisterValidation("changing_requires_can_manage_at_least",
		constructChangingRequiresCanManageAtLeastValidator(formData, store, currentGroupData))
	formData.RegisterTranslation("changing_requires_can_manage_at_least", "only managers with 'can_manage' >= '%[3]s' can modify this field")

	formData.RegisterValidation("max_participants", constructMaxParticipantsValidator(formData, currentGroupData))
	formData.RegisterTranslation("max_participants", "cannot be set to null when 'enforce_max_participants' is true")

	formData.RegisterValidation("enforce_max_participants", constructEnforceMaxParticipantsValidator(formData, currentGroupData))
	formData.RegisterTranslation("enforce_max_participants", "cannot be set to true when 'max_participants' is null")

	formData.RegisterValidation(
		"strengthening_requires_approval_change_action",
		constructStrengtheningRequiresFieldValidator(formData, groupHasParticipants, currentGroupData),
	)
	formData.RegisterTranslation("strengthening_requires_approval_change_action", "Strengthening requires parameter approval_change_action")

	formData.RegisterValidation(
		"not_set_when_no_field_strengthened",
		constructNotSetWhenNoFieldStrengthenedValidator(groupHasParticipants, currentGroupData),
	)
	formData.RegisterTranslation("not_set_when_no_field_strengthened", "must be present only if a 'require_*' field is strengthened")

	formData.RegisterTranslation("null|gte=0", "can be null or an integer between 0 and 2147483647 inclusively")

	err := formData.ParseMapData(rawRequestData)
	return formData, err
}

func int64PtrEqualValues(a, b *int64) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

func isTryingToChangeOfficialSessionActivity(dbMap map[string]interface{}, oldIsOfficialSession, rootActivityIDChanged bool) bool {
	isOfficialSession, isOfficialSessionSet := dbMap["is_official_session"]
	isOfficialSessionChanged := isOfficialSessionSet && oldIsOfficialSession != isOfficialSession.(bool)
	return isOfficialSessionChanged && isOfficialSession.(bool) ||
		!isOfficialSessionChanged && oldIsOfficialSession && rootActivityIDChanged
}

func constructFrozenMembershipValidator(formData *formdata.FormData) validator.Func {
	return formData.ValidatorSkippingUnchangedFields(func(fl validator.FieldLevel) bool {
		// return false if frozen_membership is going to be changed from true to false
		return fl.Field().Bool()
	})
}

func constructChangingRequiresCanManageAtLeastValidator(formData *formdata.FormData, store *database.DataStore,
	currentGroupData *groupUpdateInput,
) validator.Func {
	return formData.ValidatorSkippingUnchangedFields(func(fl validator.FieldLevel) bool {
		return currentGroupData.CanManageValue >= store.GroupManagers().CanManageIndexByName(fl.Param())
	})
}

func constructMaxParticipantsValidator(formData *formdata.FormData, currentGroupData *groupUpdateInput) validator.Func {
	return formData.ValidatorSkippingUnchangedFields(func(fl validator.FieldLevel) bool {
		if fl.Field().Kind() == reflect.Ptr {
			enforceMaxParticipants := currentGroupData.EnforceMaxParticipants
			if formData.IsSet("enforce_max_participants") {
				enforceMaxParticipants = fl.Top().Elem().FieldByName("EnforceMaxParticipants").Interface().(bool)
			}
			if enforceMaxParticipants {
				return false
			}
		}
		return true
	})
}

func constructEnforceMaxParticipantsValidator(formData *formdata.FormData, currentGroupData *groupUpdateInput) validator.Func {
	return formData.ValidatorSkippingUnchangedFields(func(fl validator.FieldLevel) bool {
		if fl.Field().Interface().(bool) {
			maxParticipants := currentGroupData.MaxParticipants
			if formData.IsSet("max_participants") {
				maxParticipants = fl.Top().Elem().FieldByName("MaxParticipants").Interface().(*int)
			}
			if maxParticipants == nil {
				return false
			}
		}
		return true
	})
}

// requirePersonalInfoAccessApprovalIsStrengthened checks whether the field `require_personal_info_access_approval`
// is strengthened.
func requirePersonalInfoAccessApprovalIsStrengthened(groupHasParticipants bool, oldValue, newValue string) bool {
	if !groupHasParticipants {
		return false
	}

	// If the field is empty, the value is not changed, so it is not strengthened.
	if newValue == "" {
		return false
	}

	switch oldValue {
	case enumNone:
		return newValue != enumNone
	case enumView:
		return newValue == enumEdit
	}

	return false
}

// requireLockMembershipApprovalUntilIsStrengthened checks whether the field `require_lock_membership_approval_until`
// is strengthened.
func requireLockMembershipApprovalUntilIsStrengthened(groupHasParticipants bool, oldValue, newValue *database.Time) bool {
	if !groupHasParticipants {
		return false
	}

	if oldValue == nil {
		if newValue == nil {
			return false
		}

		newValueDate := (*time.Time)(newValue)

		// The field is considered strengthened only if the new value is > NOW().
		return newValueDate.After(time.Now())
	}

	newValueDate := (*time.Time)(newValue)

	// The field is not considered strengthened if the new value is <= NOW().
	if newValueDate != nil && newValueDate.Before(time.Now().Add(time.Second)) {
		return false
	}

	oldValueDate := (*time.Time)(oldValue)

	return newValue != nil && newValueDate.Compare(*oldValueDate) == 1
}

// requireWatchApprovalIsStrengthened checks whether the field `require_watch_approval` is strengthened.
func requireWatchApprovalIsStrengthened(groupHasParticipants, oldValue, newValue bool) bool {
	if !groupHasParticipants {
		return false
	}

	return !oldValue && newValue
}

func fieldIsStrengthened(fl validator.FieldLevel, groupHasParticipants bool, currentGroupData *groupUpdateInput) bool {
	switch fl.FieldName() {
	case "require_personal_info_access_approval":
		newValue := fl.Field().String()

		return requirePersonalInfoAccessApprovalIsStrengthened(
			groupHasParticipants,
			currentGroupData.RequirePersonalInfoAccessApproval,
			newValue,
		)
	case "require_lock_membership_approval_until":
		newValue := fl.Top().Elem().FieldByName("RequireLockMembershipApprovalUntil").Interface().(*database.Time)

		return requireLockMembershipApprovalUntilIsStrengthened(
			groupHasParticipants,
			currentGroupData.RequireLockMembershipApprovalUntil,
			newValue,
		)
	case "require_watch_approval":
		newValue := fl.Field().Bool()

		return requireWatchApprovalIsStrengthened(
			groupHasParticipants,
			currentGroupData.RequireWatchApproval,
			newValue,
		)
	}

	return false
}

func constructStrengtheningRequiresFieldValidator(
	formData *formdata.FormData,
	groupHasParticipants bool,
	currentGroupData *groupUpdateInput,
) validator.Func {
	return formData.ValidatorSkippingUnchangedFields(func(fl validator.FieldLevel) bool {
		if !fieldIsStrengthened(fl, groupHasParticipants, currentGroupData) {
			return true
		}

		approvalChangeAction := fl.Top().Elem().FieldByName("ApprovalChangeAction").String()
		return approvalChangeAction != ""
	})
}

func constructNotSetWhenNoFieldStrengthenedValidator(groupHasParticipants bool, currentGroupData *groupUpdateInput) validator.Func {
	return func(fl validator.FieldLevel) bool {
		newRequirePersonalInfoAccessApproval := fl.Top().Elem().FieldByName("RequirePersonalInfoAccessApproval").Interface().(string)
		newRequireLockMembershipApprovalUntil := fl.Top().Elem().FieldByName("RequireLockMembershipApprovalUntil").Interface().(*database.Time)
		newRequireWatchApproval := fl.Top().Elem().FieldByName("RequireWatchApproval").Interface().(bool)

		// We don't need to check that approval_change_action is set,
		// because this validator is called only if it is (it has omitempty).

		// There must be no require_* fields strengthened.
		return requirePersonalInfoAccessApprovalIsStrengthened(
			groupHasParticipants,
			currentGroupData.RequirePersonalInfoAccessApproval,
			newRequirePersonalInfoAccessApproval,
		) ||
			requireLockMembershipApprovalUntilIsStrengthened(
				groupHasParticipants,
				currentGroupData.RequireLockMembershipApprovalUntil,
				newRequireLockMembershipApprovalUntil,
			) ||
			requireWatchApprovalIsStrengthened(
				groupHasParticipants,
				currentGroupData.RequireWatchApproval,
				newRequireWatchApproval,
			)
	}
}
