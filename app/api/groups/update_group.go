package groups

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/validator"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Information of the group to be modified
// swagger:model
type groupUpdateInput struct {
	Name  string `json:"name"`
	Grade int32  `json:"grade"`
	// Nullable
	Description *string `json:"description"`
	IsOpen      bool    `json:"is_open"`
	// If changed from true to false, is automatically switches all requests to join this group from requestSent to requestRefused
	IsPublic bool `json:"is_public"`
	// Duration after the first use of the code when it will expire
	// Nullable
	// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
	// example: 838:59:59
	CodeLifetime *string `json:"code_lifetime" validate:"omitempty,duration"`
	// Nullable
	CodeExpiresAt *time.Time `json:"code_expires_at"`
	// Nullable
	ActivityID *int64 `json:"activity_id"`
	// Can be set only if activity_id is set and
	// the current user has the 'can_make_session_official' permission on the activity item
	IsOfficialSession bool `json:"is_official_session"`
	OpenContest       bool `json:"open_contest"`

	RequireMembersToJoinParent bool `json:"require_members_to_join_parent"`

	// Can be changed only from false to true
	// (changing auto-rejects all pending join/leave requests and withdraws all pending invitations)
	FrozenMembership bool `json:"frozen_membership" validate:"frozen_membership"`

	// Nullable
	Organizer *string `json:"organizer"`
	// Nullable
	AddressLine1 *string `json:"address_line1"`
	// Nullable
	AddressLine2 *string `json:"address_line2"`
	// Nullable
	AddressPostcode *string `json:"address_postcode"`
	// Nullable
	AddressCity *string `json:"address_city"`
	// Nullable
	AddressCountry *string `json:"address_country"`
	// Nullable
	ExpectedStart *time.Time `json:"expected_start"`
}

type currentGroupDataType struct {
	IsPublic          bool
	ActivityID        *int64
	IsOfficialSession bool
	FrozenMembership  bool
}

// swagger:operation PUT /groups/{group_id} groups groupEdit
// ---
// summary: Update group information
// description: Edit group information.
//   Requires the user to be a manager of the group, otherwise the 'forbidden' error is returned.
//
//
//   If the `activity_id` item is provided and is not null, the user should have at least
//  'can_view:info' permission on it, otherwise the 'forbidden' error is returned.
//
//
//   If `is_official_session` is being changed to true, the user should have at least
//  'can_make_session_official' permission on the activity item, otherwise the 'forbidden' error is returned.
//
//
//   Setting `is_official_session` to true while keeping `activity_id` not set or setting `activity_id` to null for
//   an official session, will cause the "bad request" error.
// parameters:
// - name: group_id
//   in: path
//   required: true
//   type: integer
// - name: group information
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/groupUpdateInput"
// responses:
//   "200":
//     "$ref": "#/responses/updatedResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) updateGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	apiErr := service.NoError

	err = srv.Store.InTransaction(func(s *database.DataStore) error {
		groupStore := s.Groups()

		var currentGroupData currentGroupDataType

		err = groupStore.ManagedBy(user).
			Select("groups.is_public, groups.activity_id, groups.is_official_session, groups.frozen_membership").WithWriteLock().
			Where("groups.id = ?", groupID).Limit(1).Scan(&currentGroupData).Error()
		if gorm.IsRecordNotFoundError(err) {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}
		service.MustNotBeError(err)

		var formData *formdata.FormData
		formData, err = validateUpdateGroupInput(r, &currentGroupData)
		if err != nil {
			apiErr = service.ErrInvalidRequest(err)
			return apiErr.Error // rollback
		}

		dbMap := formData.ConstructMapForDB()

		apiErr = validateActivityIDAndIsOfficial(s, user, currentGroupData.ActivityID, currentGroupData.IsOfficialSession, dbMap)
		if apiErr != service.NoError {
			return apiErr.Error // rollback
		}

		service.MustNotBeError(refuseSentGroupRequestsIfNeeded(
			groupStore, groupID, user.GroupID, dbMap, currentGroupData.IsPublic, currentGroupData.FrozenMembership))

		// update the group
		service.MustNotBeError(groupStore.Where("id = ?", groupID).Updates(dbMap).Error())

		return nil // commit
	})

	if apiErr != service.NoError {
		return apiErr
	}
	service.MustNotBeError(err)

	response := service.Response{Success: true, Message: "updated"}
	render.Respond(w, r, &response)

	return service.NoError
}

func validateActivityIDAndIsOfficial(
	store *database.DataStore, user *database.User, oldActivityID *int64, oldIsOfficialSession bool,
	dbMap map[string]interface{}) service.APIError {
	activityIDToCheck := oldActivityID
	activityID, activityIDSet := dbMap["activity_id"]
	activityIDChanged := activityIDSet && !int64PtrEqualValues(oldActivityID, activityID.(*int64))
	if activityIDChanged {
		activityIDToCheck = activityID.(*int64)
		if activityIDToCheck != nil {
			apiError := validateActivityID(store, user, activityIDToCheck)
			if apiError != service.NoError {
				return apiError
			}
		}
	}

	if isTryingToChangeOfficialSessionActivity(dbMap, oldIsOfficialSession, activityIDChanged) {
		if activityIDToCheck == nil {
			return service.ErrInvalidRequest(errors.New("the activity_id should be set for official sessions"))
		}
		found, err := store.PermissionsGranted().WithWriteLock().
			Joins(`
				JOIN groups_ancestors ON groups_ancestors.ancestor_group_id = permissions_granted.group_id AND
				     groups_ancestors.child_group_id = ?`, user.GroupID).
			Where("permissions_granted.can_make_session_official").
			Where("permissions_granted.item_id = ?", activityIDToCheck).
			HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrForbidden(errors.New("not enough permissions for attaching the group to the activity as an official session"))
		}
	}
	return service.NoError
}

func validateActivityID(store *database.DataStore, user *database.User, activityIDToCheck *int64) service.APIError {
	if activityIDToCheck != nil {
		found, errorInTransaction := store.Items().ByID(*activityIDToCheck).WithWriteLock().
			WhereUserHasViewPermissionOnItems(user, "info").HasRows()
		service.MustNotBeError(errorInTransaction)
		if !found {
			return service.ErrForbidden(errors.New("no access to the activity"))
		}
	}
	return service.NoError
}

// refuseSentGroupRequestsIfNeeded automatically refuses all requests to join this group
// (removes them from group_pending_requests and inserts appropriate group_membership_changes
// with `action` = 'join_request_refused') if is_public is changed from true to false
func refuseSentGroupRequestsIfNeeded(
	store *database.GroupStore, groupID, initiatorID int64, dbMap map[string]interface{},
	previousIsPublicValue, previousFrozenMembershipValue bool) error {
	var shouldRefusePending bool

	pendingTypesToHandle := make([]string, 0, 3)
	pendingTypesToHandle = append(pendingTypesToHandle, "join_request")

	// if is_public is going to be changed from true to false
	if newIsPublic, ok := dbMap["is_public"]; ok && !newIsPublic.(bool) && previousIsPublicValue {
		shouldRefusePending = true
	}
	if newFrozenMembership, ok := dbMap["frozen_membership"]; ok && newFrozenMembership.(bool) && !previousFrozenMembershipValue {
		shouldRefusePending = true
		pendingTypesToHandle = append(pendingTypesToHandle, "leave_request", "invitation")
	}

	if shouldRefusePending {
		service.MustNotBeError(store.Exec(`
			INSERT INTO group_membership_changes (group_id, member_id, action, at, initiator_id)
			SELECT group_id, member_id,
				CASE type
					WHEN 'join_request' THEN 'join_request_refused'
					WHEN 'leave_request' THEN 'leave_request_refused'
					WHEN 'invitation' THEN 'invitation_withdrawn'
				END,
				NOW(), ?
			FROM group_pending_requests
			WHERE group_id = ? AND type IN (?)
			FOR UPDATE`, initiatorID, groupID, pendingTypesToHandle).Error())
		// refuse sent group requests
		return store.GroupPendingRequests().
			Where("type IN (?)", pendingTypesToHandle).
			Where("group_id = ?", groupID).
			Delete().Error()
	}
	return nil
}

func validateUpdateGroupInput(r *http.Request, currentGroupData *currentGroupDataType) (*formdata.FormData, error) {
	formData := formdata.NewFormData(&groupUpdateInput{})
	formData.RegisterValidation("frozen_membership", func(fl validator.FieldLevel) bool {
		if !formData.IsSet("frozen_membership") || fl.Field().Bool() == currentGroupData.FrozenMembership ||
			!currentGroupData.FrozenMembership {
			return true
		}
		// frozen_membership is going to be changed from true to false
		return false
	})
	formData.RegisterTranslation("frozen_membership", "can only be changed from false to true")
	err := formData.ParseJSONRequestData(r)
	return formData, err
}

func int64PtrEqualValues(a, b *int64) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

func isTryingToChangeOfficialSessionActivity(dbMap map[string]interface{}, oldIsOfficialSession, activityIDChanged bool) bool {
	isOfficialSession, isOfficialSessionSet := dbMap["is_official_session"]
	isOfficialSessionChanged := isOfficialSessionSet && oldIsOfficialSession != isOfficialSession.(bool)
	return isOfficialSessionChanged && isOfficialSession.(bool) ||
		!isOfficialSessionChanged && oldIsOfficialSession && activityIDChanged
}
