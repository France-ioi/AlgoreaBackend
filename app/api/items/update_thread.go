package items

import (
	"errors"
	"net/http"
	"time"

	"github.com/France-ioi/validator"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// threadUpdateFields represents the fields of a thread in database
type threadUpdateFields struct {
	// Optional
	// enum: waiting_for_participant,waiting_for_trainer,closed
	Status string `json:"status" validate:"helper_group_id_set_if_non_open_to_open_status"`
	// Optional
	HelperGroupID *int64 `json:"helper_group_id" validate:"helper_group_id_not_set_when_set_or_keep_closed,group_visible_by=:user_id,group_visible_by=participant_id,can_request_help_to_when_own_thread"`
	// Optional
	MessageCount *int `json:"message_count" validate:"omitempty,gte=0,exclude_increment_if_message_count_set"`
}

// updateThreadRequest is the expected input for thread updating
// swagger:model threadEditRequest
type updateThreadRequest struct {
	threadUpdateFields `json:"thread,squash"`

	// Used to increment the message count when we are not sure of the exact total message count. Can be negative.
	// Optional
	MessageCountIncrement *int `json:"message_count_increment"`
}

// swagger:operation PUT /items/{item_id}/participant/{participant_id}/thread items threadUpdate
// ---
// summary: Update a thread
// description: >
//
//	Service to update thread information.
//
//	If the thread doesn't exist, it is created.
//
//	Once a thread has been created, it cannot be deleted or set back to `not_started`.
//
//	Validations and restrictions:
//	  * if `status` is given:
//	    - The participant of a thread can always switch the thread from open to any another other status. He can only switch it from non-open to an open status if he is allowed to request help on this item (see “specific permission” above)
//	    - A user who has `can_watch>=answer` on the item AND `can_watch_members` on the participant: can always switch a thread to any open status (i.e. he can always open it but not close it)
//	    - A user who `can write` on the thread can switch from an open status to another open status.
//	  * if `status` is already "closed" and not changing status OR if switching to status "closed": `helper_group_id` must not be given
//	  * if switching to an open status from a non-closed status: `helper_group_id` must be given
//	  * if given, the `helper_group_id` must be visible to the current-user and to participant.
//	  * if participant is the current user and `helper_group_id` given, `helper_group_id` must be a descendants (including self) of one of the group he `can_request_help_to`.
//	  * if `helper_group_id` or `message_count` or `message_count_increment` is given: the current-user must be allowed to write (see doc) (if `status` is given, checks related to status supersede this one).
//	  * at most one of `message_count_increment`, `message_count` must be given
//
// parameters:
//   - name: item_id
//     in: path
//     type: integer
//     format: int64
//     required: true
//   - name: participant_id
//     in: path
//     type: integer
//     format: int64
//     required: true
//   - in: body
//     name: data
//     required: true
//     description: New thread property values
//     schema:
//     "$ref": "#/definitions/threadEditRequest"
//
// responses:
//
//	"200":
//	  "$ref": "#/responses/updatedResponse"
//	"400":
//	  "$ref": "#/responses/badRequestResponse"
//	"401":
//	  "$ref": "#/responses/unauthorizedResponse"
//	"403":
//	  "$ref": "#/responses/forbiddenResponse"
//	"500":
//	  "$ref": "#/responses/internalErrorResponse"
func (srv *Service) updateThread(w http.ResponseWriter, r *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	participantID, err := service.ResolveURLQueryPathInt64Field(r, "participant_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	apiError := service.NoError
	err = srv.GetStore(r).InTransaction(func(store *database.DataStore) error {
		threadRecordQuery := store.Threads().
			Where("threads.participant_id = ?", participantID).
			Where("threads.item_id = ?", itemID).
			Limit(1)

		var oldThread struct {
			Status        string
			HelperGroupID int64
			MessageCount  int
		}
		err = threadRecordQuery.WithWriteLock().Take(&oldThread).Error()
		if gorm.IsRecordNotFoundError(err) {
			// Create
		} else if err != nil {
			service.MustNotBeError(err)
		}

		user := srv.GetUser(r)

		input := updateThreadRequest{}
		formData := formdata.NewFormData(&input)

		formData.RegisterValidation("exclude_increment_if_message_count_set", excludeIncrementIfMessageCountSetValidator)
		formData.RegisterTranslation("exclude_increment_if_message_count_set",
			"cannot have both message_count and message_count_increment set")
		formData.RegisterValidation("helper_group_id_set_if_non_open_to_open_status",
			constructHelperGroupIdSetIfNonOpenToOpenStatus(oldThread.Status))
		formData.RegisterTranslation("helper_group_id_set_if_non_open_to_open_status",
			"the helper_group_id must be set to switch from a non-open to an open status")
		formData.RegisterValidation("helper_group_id_not_set_when_set_or_keep_closed",
			constructHelperGroupIdNotSetWhenSetOrKeepClosed(oldThread.Status))
		formData.RegisterTranslation("helper_group_id_not_set_when_set_or_keep_closed",
			"the helper_group_id must not be given when setting or keeping status to closed")
		formData.RegisterValidation("group_visible_by", constructValidateGroupVisibleBy(srv, r))
		formData.RegisterTranslation("group_visible_by", "the group must be visible to the current-user and the participant")
		formData.RegisterValidation("can_request_help_to_when_own_thread",
			constructValidateCanRequestHelpToWhenOwnThread(user, itemID, participantID, store))
		formData.RegisterTranslation("can_request_help_to_when_own_thread",
			"the group must be descendant of a group the participant can request help to")

		err = formData.ParseJSONRequestData(r)
		if err != nil {
			apiError = service.ErrInvalidRequest(err)
			return apiError.Error
		}

		newHelperGroupID := oldThread.HelperGroupID
		if input.HelperGroupID != nil {
			newHelperGroupID = *input.HelperGroupID
		}

		if input.Status == "" {
			if input.HelperGroupID == nil && input.MessageCount == nil && input.MessageCountIncrement == nil {
				apiError = service.ErrInvalidRequest(
					errors.New("either status, helper_group_id, message_count or message_count_increment must be given"))
				return apiError.Error
			}

			// the current-user must be allowed to write
			canWrite, err := store.Threads().UserCanWrite(srv.GetUser(r), participantID, itemID)
			if err != nil {
				return err
			}
			if !canWrite {
				apiError = service.InsufficientAccessRightsError
				return apiError.Error
			}
		} else {
			if !store.Threads().UserCanChangeStatus(srv.GetUser(r), oldThread.Status, input.Status, participantID, itemID, newHelperGroupID) {
				apiError = service.InsufficientAccessRightsError
				return apiError.Error
			}
		}

		threadData := formData.ConstructPartialMapForDB("threadUpdateFields")

		if formData.IsSet("message_count_increment") && input.MessageCountIncrement != nil {
			// if the thread doesn't exist, oldThread.MessageCount = 0
			newMessageCount := oldThread.MessageCount + *input.MessageCountIncrement
			if newMessageCount < 0 {
				newMessageCount = 0
			}

			threadData["message_count"] = newMessageCount
		}

		if len(threadData) > 0 {
			threadData["item_id"] = itemID
			threadData["participant_id"] = participantID
			threadData["latest_update_at"] = time.Now()
			threadData["helper_group_id"] = newHelperGroupID

			service.MustNotBeError(store.Threads().InsertOrUpdateMap(threadData, nil))
		}

		return nil
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(nil)))
	return service.NoError
}

func constructValidateCanRequestHelpToWhenOwnThread(user *database.User, itemID, participantID int64, store *database.DataStore) validator.Func {
	return func(fl validator.FieldLevel) bool {
		if user.GroupID != participantID {
			return true
		}

		helperGroupIDPtr := fl.Top().Elem().FieldByName("HelperGroupID").Interface().(*int64)
		return store.Threads().CanRequestHelpTo(user, itemID, *helperGroupIDPtr)
	}
}

func constructValidateGroupVisibleBy(srv *Service, r *http.Request) validator.Func {
	return func(fl validator.FieldLevel) bool {
		store := srv.GetStore(r)
		groupIDPtr := fl.Top().Elem().FieldByName("HelperGroupID").Interface().(*int64)
		if groupIDPtr == nil {
			return false
		}

		var user *database.User
		param := fl.Param()
		if param == ":user_id" {
			user = srv.GetUser(r)
		} else {
			var err error
			user = new(database.User)
			user.GroupID, err = service.ResolveURLQueryPathInt64Field(r, param)
			if err != nil {
				return false
			}
		}

		return store.Groups().IsVisibleFor(*groupIDPtr, user)
	}
}

func constructHelperGroupIdNotSetWhenSetOrKeepClosed(oldStatus string) validator.Func {
	// if status is already "closed" and not changing status OR if switching to status "closed": helper_group_id must not be given
	wasOpen := oldStatus == "waiting_for_trainer" || oldStatus == "waiting_for_participant"

	return func(fl validator.FieldLevel) bool {
		newStatus := fl.Top().Elem().FieldByName("Status").String()
		willBeOpen := newStatus == "waiting_for_trainer" || newStatus == "waiting_for_participant"

		helperGroupIdPtr := fl.Top().Elem().FieldByName("HelperGroupID").Interface().(*int64)
		if helperGroupIdPtr != nil {
			if (!wasOpen && newStatus == "") || (!willBeOpen && newStatus != "") {
				return false
			}
		}

		return true
	}
}

func constructHelperGroupIdSetIfNonOpenToOpenStatus(oldStatus string) validator.Func {
	// if switching to an open status from a non-open status: helper_group_id must be given
	wasOpen := oldStatus == "waiting_for_trainer" || oldStatus == "waiting_for_participant"

	return func(fl validator.FieldLevel) bool {
		newStatus := fl.Field().String()
		willBeOpen := newStatus == "waiting_for_trainer" || newStatus == "waiting_for_participant"

		helperGroupIdPtr := fl.Top().Elem().FieldByName("HelperGroupID").Interface().(*int64)
		if !wasOpen && willBeOpen && helperGroupIdPtr == nil {
			return false
		}

		return true
	}
}

// excludeIncrementIfMessageCountSetValidator validates that message_count and message_count_increment are not both set
func excludeIncrementIfMessageCountSetValidator(messageCountField validator.FieldLevel) bool {
	messageCountPtr := messageCountField.Top().Elem().FieldByName("MessageCount").Interface().(*int)
	messageCountIncrementPtr := messageCountField.Top().Elem().FieldByName("MessageCountIncrement").Interface().(*int)

	return !(messageCountPtr != nil && messageCountIncrementPtr != nil)
}
