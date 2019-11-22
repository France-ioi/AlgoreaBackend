package groups

import (
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/render"

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
	Opened      bool    `json:"opened"`
	// If changed from true to false, automatically switch all requests to join this group from requestSent to requestRefused
	FreeAccess bool `json:"free_access"`
	// Duration after the first use of the code when it will expire
	// Nullable
	// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
	// example: 838:59:59
	CodeLifetime *string `json:"code_lifetime" validate:"omitempty,duration"`
	// Nullable
	CodeExpiresAt *time.Time `json:"code_expires_at"`
	// Nullable
	// pattern:  ^(\d+(/\d+))$
	RedirectPath *string `json:"redirect_path" validate:"omitempty,redirect_path"`
	OpenContest  bool    `json:"open_contest"`
}

// swagger:operation PUT /groups/{group_id} groups groupEdit
// ---
// summary: Update group information
// description: Edit group information.
//   Requires the user to be the owner of the group.
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

	formData, err := validateUpdateGroupInput(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	apiErr := service.NoError

	err = srv.Store.InTransaction(func(s *database.DataStore) error {
		groupStore := s.Groups()

		var currentGroupData []struct {
			FreeAccess bool
		}

		if errInTransaction := groupStore.OwnedBy(user).
			Select("groups.free_access").WithWriteLock().
			Where("groups.id = ?", groupID).Limit(1).Scan(&currentGroupData).Error(); errInTransaction != nil {
			return errInTransaction // rollback
		}
		if len(currentGroupData) < 1 {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}

		dbMap := formData.ConstructMapForDB()
		if errInTransaction := refuseSentGroupRequestsIfNeeded(
			groupStore, groupID, user.GroupID, dbMap, currentGroupData[0].FreeAccess); errInTransaction != nil {
			return errInTransaction // rollback
		}

		// update the group
		if errInTransaction := groupStore.Where("id = ?", groupID).Updates(dbMap).Error(); errInTransaction != nil {
			return errInTransaction // rollback
		}

		return nil // commit
	})

	if apiErr != service.NoError {
		return apiErr
	}

	if err != nil {
		return service.ErrUnexpected(err)
	}

	response := service.Response{Success: true, Message: "updated"}
	render.Respond(w, r, &response)

	return service.NoError
}

// refuseSentGroupRequestsIfNeeded automatically refuses all requests to join this group
// (removes them from group_pending_requests and inserts appropriate group_membership_changes
// with `action` = 'join_request_refused') if free_access is changed from true to false
func refuseSentGroupRequestsIfNeeded(
	store *database.GroupStore, groupID, initiatorID int64, dbMap map[string]interface{}, previousFreeAccessValue bool) error {
	// if free_access is going to be changed from true to false
	if newFreeAccess, ok := dbMap["free_access"]; ok && !newFreeAccess.(bool) && previousFreeAccessValue {
		service.MustNotBeError(store.Exec(`
			INSERT INTO group_membership_changes (group_id, member_id, action, at, initiator_id)
			SELECT group_id, member_id, 'join_request_refused', NOW(), ?
			FROM group_pending_requests
			WHERE group_id = ? AND type = 'join_request'
			FOR UPDATE`, initiatorID, groupID).Error())
		// refuse sent group requests
		return store.GroupPendingRequests().
			Where("type = 'join_request'").
			Where("group_id = ?", groupID).
			Delete().Error()
	}
	return nil
}

var redirectPathRegexp = regexp.MustCompile(`^(\d+(/\d+)*)*$`)

func validateUpdateGroupInput(r *http.Request) (*formdata.FormData, error) {
	formData := formdata.NewFormData(&groupUpdateInput{})
	formData.RegisterValidation("redirect_path", validator.Func(func(fl validator.FieldLevel) bool {
		return redirectPathRegexp.MatchString(fl.Field().Interface().(string))
	}))
	formData.RegisterTranslation("redirect_path", "invalid redirect path")
	return formData, formData.ParseJSONRequestData(r)
}
