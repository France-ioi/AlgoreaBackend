package items

import (
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
	Status string `json:"status"`
	// Optional
	HelperGroupID int64 `json:"helper_group_id"`
	// Optional
	MessageCount int `json:"message_count" validate:"omitempty,gte=0,exclude_increment_if_set"`
}

// updateThreadRequest is the expected input for thread updating
// swagger:model threadEditRequest
type updateThreadRequest struct {
	threadUpdateFields `json:"thread,squash"`

	// Used to increment the message count when we are not sure of the exact total message count. Can be negative.
	// Optional
	MessageCountIncrement int `json:"message_count_increment"`
}

// swagger:operation PUT /items/{item_id}/participant/{participant_id}/thread items threadUpdate
// ---
// summary: Update a thread
// description: >
//
//	  Service to update thread information.
//
//	  If the thread doesn't exist, it is created.
//
//		 Once a thread has been created, it cannot be deleted or set back to `not_started`.
//
//	  Validations and restrictions:
//	    * if `status` is given:
//		     - The participant of a thread can always switch the thread from open to any another other status. He can only switch it from non-open to an open status if he is allowed to request help on this item (see “specific permission” above)
//	      - A user who has `can_watch>=answer` on the item AND `can_watch_members` on the participant: can always switch a thread to any open status (i.e. he can always open it but not close it)
//	      - A user who `can write` on the thread can switch from an open status to another open status.
//	    * if `status` is already "closed" and not changing status OR if switching to status "closed": `helper_group_id` must not be given
//	    * if switching to an open status from a non-closed status: `helper_group_id` must be given
//	    * if given, the `helper_group_id` must be visible to the current-user and to participant.
//	    * if participant is the current user and `helper_group_id` given, `helper_group_id` must be a descendants (including self) of one of the group he `can_request_help_to`.
//	    * if `helper_group_id` or `message_count` or `message_count_increment` is given: the current-user must be allowed to write (see doc) (if `status` is given, checks related to status supersede this one).
//	    * at most one of `message_count_increment`, `message_count` must be given
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

	input := updateThreadRequest{}
	formData := formdata.NewFormData(&input)

	formData.RegisterValidation("exclude_increment_if_set", excludeIncrementIfSetValidator)
	formData.RegisterTranslation("exclude_increment_if_set",
		"cannot have both message_count and message_count_increment set")

	err = formData.ParseJSONRequestData(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	apiError := service.NoError
	err = srv.GetStore(r).InTransaction(func(store *database.DataStore) error {
		threadRecordQuery := store.Threads().
			Where("threads.participant_id = ?", participantID).
			Where("threads.item_id = ?", itemID).
			Limit(1)

		var oldThread threadUpdateFields
		err = threadRecordQuery.WithWriteLock().Take(&oldThread).Error()
		if gorm.IsRecordNotFoundError(err) {
			// Create
		}
		service.MustNotBeError(err)

		if input.Status == "" {
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
			canChangeStatus, err := store.Threads().UserCanChangeStatus(srv.GetUser(r), oldThread.Status, input.Status, participantID, itemID)
			if err != nil {
				return err
			}
			if !canChangeStatus {
				apiError = service.InsufficientAccessRightsError
				return apiError.Error
			}
		}

		threadData := formData.ConstructPartialMapForDB("threadUpdateFields")

		if formData.IsSet("message_count_increment") {
			// if the thread doesn't exist, oldThread.MessageCount = 0
			newMessageCount := oldThread.MessageCount + input.MessageCountIncrement
			if newMessageCount < 0 {
				newMessageCount = 0
			}

			threadData["message_count"] = newMessageCount
		}

		if len(threadData) > 0 {
			threadData["item_id"] = itemID
			threadData["participant_id"] = participantID
			threadData["latest_update_at"] = time.Now()
			threadData["helper_group_id"] = oldThread.HelperGroupID

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

// excludeIncrementIfSetValidator validates that message_count and message_count_increment are not both set
func excludeIncrementIfSetValidator(messageCountField validator.FieldLevel) bool {
	return true
	//return messageCountField.Field().Interface() == nil ||
	//	messageCountField.Top().Elem().FieldByName("MessageCountIncrement").Interface() == nil
}
