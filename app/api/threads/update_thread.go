package threads

import (
	"errors"
	"net/http"

	"github.com/France-ioi/validator"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// threadUpdateFields represents the fields of a thread in database.
type threadUpdateFields struct {
	// Optional
	// enum: waiting_for_participant,waiting_for_trainer,closed
	Status string `json:"status" validate:"helper_group_id_set_if_non_open_to_open_status,oneof=waiting_for_participant waiting_for_trainer closed"` //nolint
	// Optional
	HelperGroupID *int64 `json:"helper_group_id" validate:"helper_group_id_not_set_when_set_or_keep_closed,group_visible_by=user_id,group_visible_by=participant_id,can_request_help_to_when_own_thread"` //nolint
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

// swagger:operation PUT /items/{item_id}/participant/{participant_id}/thread threads threadUpdate
//
//	---
//	summary: Update a thread
//	description: >
//
//		Updates a thread with the given information.
//
//
//		If the thread doesn't exist, it is created.
//
//
//		Once a thread has been created, it cannot be deleted or set back to `not_started`.
//
//
//		Validations and restrictions:
//			* the current user should have `can_view` >= content permission on the item in order to have the right to write to the thread.
//			* if `status` is given:
//				- The participant of a thread can always switch the thread from open to any another other status.
//					He can only switch it from non-open to an open status if he is allowed to request help on this item.
//				- A user who has `can_watch>=answer` on the item AND `can_watch_members` on the participant: can always switch
//					a thread to any open status (i.e. he can always open it but not close it)
//				- A user who `can write` on the thread can switch from an open status to another open status.
//			* if `status` is already "closed" and not changing status OR if switching to status "closed":
//				`helper_group_id` must not be given
//			* if switching to an open status from a non-open status: `helper_group_id` must be given
//			* if given, the `helper_group_id` must be visible to the current-user and to participant.
//			* if participant is the current user and `helper_group_id` given, `helper_group_id` must be a descendants
//				(including self) of one of the group he `can_request_help_to`.
//			* if `helper_group_id` or `message_count` or `message_count_increment` is given: the current-user must be allowed
//				to write (see doc) (if `status` is given, checks related to status supersede this one).
//			* at most one of `message_count_increment`, `message_count` must be given
//
//
//		A forbidden error is returned when:
//			* Trying to write to a thread, including changing its status, without having the rights to do so.
//			* The provided `helper_group_id` doesn't match the restrictions.
//
//
//		If the parameters given are inconsistent, a bad request error is returned with the reason of the error.
//
//
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: participant_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- in: body
//			name: data
//			required: true
//			description: New thread property values
//			schema:
//				"$ref": "#/definitions/threadEditRequest"
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
func (srv *Service) updateThread(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	participantID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "participant_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	rawRequestData, err := service.ResolveJSONBodyIntoMap(httpRequest)
	service.MustNotBeError(err)

	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	userCanViewItemContent, err := store.Permissions().MatchingUserAncestors(user).
		Where("permissions.item_id = ?", itemID).
		WherePermissionIsAtLeast("view", "content").
		HasRows()
	service.MustNotBeError(err)
	if !userCanViewItemContent {
		return service.ErrAPIInsufficientAccessRights
	}

	err = store.InTransaction(func(store *database.DataStore) error {
		var oldThreadInfo threadInfo
		err = database.NewDataStore(constructThreadInfoQuery(store, user, itemID, participantID)).
			WithCustomWriteLocks(golang.NewSet[string](), golang.NewSet[string]("threads")).
			Take(&oldThreadInfo).Error()
		if !gorm.IsRecordNotFoundError(err) {
			service.MustNotBeError(err)
		}

		input := updateThreadRequest{}
		formData := formdata.NewFormData(&input)

		formData.RegisterValidation("exclude_increment_if_message_count_set", excludeIncrementIfMessageCountSetValidator)
		formData.RegisterTranslation("exclude_increment_if_message_count_set",
			"cannot have both message_count and message_count_increment set")
		formData.RegisterValidation("helper_group_id_set_if_non_open_to_open_status",
			constructHelperGroupIDSetIfNonOpenToOpenStatus(oldThreadInfo.ThreadStatus))
		formData.RegisterTranslation("helper_group_id_set_if_non_open_to_open_status",
			"the helper_group_id must be set to switch from a non-open to an open status")
		formData.RegisterValidation("helper_group_id_not_set_when_set_or_keep_closed",
			constructHelperGroupIDNotSetWhenSetOrKeepClosed(oldThreadInfo.ThreadStatus))
		formData.RegisterTranslation("helper_group_id_not_set_when_set_or_keep_closed",
			"the helper_group_id must not be given when setting or keeping status to closed")
		formData.RegisterValidation("group_visible_by", constructValidateGroupVisibleBy(srv, httpRequest))
		formData.RegisterTranslation("group_visible_by", "the group must be visible to the current-user and the participant")
		formData.RegisterValidation("can_request_help_to_when_own_thread",
			constructValidateCanRequestHelpToWhenOwnThread(user, store, participantID, itemID))
		formData.RegisterTranslation("can_request_help_to_when_own_thread",
			"the group must be descendant of a group the participant can request help to")

		err = formData.ParseMapData(rawRequestData)
		if err != nil {
			return service.ErrInvalidRequest(err) // rollback
		}

		err = checkUpdateThreadPermissions(user, oldThreadInfo.ThreadStatus, input, participantID, &oldThreadInfo)
		service.MustNotBeError(err)

		threadData := computeNewThreadData(
			formData, oldThreadInfo.ThreadMessageCount, oldThreadInfo.ThreadHelperGroupID, input, itemID, participantID)
		service.MustNotBeError(store.Threads().InsertOrUpdateMap(threadData, nil))

		return nil
	})
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.UpdateSuccess[*struct{}](nil)))
	return nil
}

func computeNewThreadData(
	formData *formdata.FormData,
	oldMessageCount int,
	oldHelperGroupID int64,
	input updateThreadRequest,
	itemID int64,
	participantID int64,
) map[string]interface{} {
	threadData := formData.ConstructPartialMapForDB("threadUpdateFields")
	if formData.IsSet("message_count_increment") && input.MessageCountIncrement != nil {
		threadData["message_count"] = newMessageCountWithIncrement(oldMessageCount, *input.MessageCountIncrement)
	}
	if input.HelperGroupID != nil {
		threadData["helper_group_id"] = *input.HelperGroupID
	}

	if len(threadData) > 0 {
		threadData["item_id"] = itemID
		threadData["participant_id"] = participantID
		threadData["latest_update_at"] = database.Now()

		if _, ok := threadData["helper_group_id"]; !ok {
			threadData["helper_group_id"] = oldHelperGroupID
		}
	}

	return threadData
}

func checkUpdateThreadPermissions(
	user *database.User,
	oldThreadStatus string,
	input updateThreadRequest,
	participantID int64,
	threadInfo *threadInfo,
) error {
	if input.Status == "" {
		if input.HelperGroupID == nil && input.MessageCount == nil && input.MessageCountIncrement == nil {
			return service.ErrInvalidRequest(
				errors.New("either status, helper_group_id, message_count or message_count_increment must be given"))
		}

		// the current-user must be allowed to write
		if !userCanWriteInThread(user, participantID, threadInfo) {
			return service.ErrAPIInsufficientAccessRights
		}
	} else if !userCanChangeThreadStatus(user, oldThreadStatus, input.Status, participantID, threadInfo) {
		return service.ErrAPIInsufficientAccessRights
	}

	return nil
}

func newMessageCountWithIncrement(oldMessageCount, increment int) int {
	// if the thread doesn't exist, oldThread.MessageCount = 0
	newMessageCount := oldMessageCount + increment
	if newMessageCount < 0 {
		newMessageCount = 0
	}

	return newMessageCount
}

func constructValidateCanRequestHelpToWhenOwnThread(user *database.User, store *database.DataStore,
	participantID, itemID int64,
) validator.Func {
	return func(fl validator.FieldLevel) bool {
		if user.GroupID != participantID {
			return true
		}

		helperGroupIDPtr := fl.Top().Elem().FieldByName("HelperGroupID").Interface().(*int64)
		return user.CanRequestHelpTo(store, itemID, *helperGroupIDPtr)
	}
}

func constructValidateGroupVisibleBy(srv *Service, httpRequest *http.Request) validator.Func {
	return func(fieldInfo validator.FieldLevel) bool {
		store := srv.GetStore(httpRequest)
		groupIDPtr := fieldInfo.Top().Elem().FieldByName("HelperGroupID").Interface().(*int64)
		if groupIDPtr == nil {
			return false
		}

		var user *database.User
		param := fieldInfo.Param()
		if param == "user_id" {
			user = srv.GetUser(httpRequest)
		} else {
			var err error
			user = new(database.User)
			user.GroupID, err = service.ResolveURLQueryPathInt64Field(httpRequest, param)
			service.MustNotBeError(err)
		}

		return store.Groups().IsVisibleFor(*groupIDPtr, user)
	}
}

func constructHelperGroupIDNotSetWhenSetOrKeepClosed(oldStatus string) validator.Func {
	// if status is already "closed" and not changing status OR if switching to status "closed": helper_group_id must not be given
	wasOpen := database.IsThreadOpenStatus(oldStatus)

	return func(fl validator.FieldLevel) bool {
		newStatus := fl.Top().Elem().FieldByName("Status").String()
		willBeOpen := database.IsThreadOpenStatus(newStatus)

		helperGroupIDPtr := fl.Top().Elem().FieldByName("HelperGroupID").Interface().(*int64)
		if helperGroupIDPtr != nil {
			if (!wasOpen && newStatus == "") || (!willBeOpen && newStatus != "") {
				return false
			}
		}

		return true
	}
}

func constructHelperGroupIDSetIfNonOpenToOpenStatus(oldStatus string) validator.Func {
	// if switching to an open status from a non-open status: helper_group_id must be given
	wasOpen := database.IsThreadOpenStatus(oldStatus)

	return func(fl validator.FieldLevel) bool {
		newStatus := fl.Field().String()
		willBeOpen := database.IsThreadOpenStatus(newStatus)

		helperGroupIDPtr := fl.Top().Elem().FieldByName("HelperGroupID").Interface().(*int64)
		if !wasOpen && willBeOpen && helperGroupIDPtr == nil {
			return false
		}

		return true
	}
}

// excludeIncrementIfMessageCountSetValidator validates that message_count and message_count_increment are not both set.
func excludeIncrementIfMessageCountSetValidator(messageCountField validator.FieldLevel) bool {
	messageCountPtr := messageCountField.Top().Elem().FieldByName("MessageCount").Interface().(*int)
	messageCountIncrementPtr := messageCountField.Top().Elem().FieldByName("MessageCountIncrement").Interface().(*int)

	return !(messageCountPtr != nil && messageCountIncrementPtr != nil)
}

// userCanChangeThreadStatus checks whether a user can change the status of a thread
//   - The participant of a thread can always switch the thread from open to any another other status.
//     He can only switch it from non-open to an open status if he is allowed to request help on this item
//   - A user who has can_watch>=answer on the item AND can_watch_members on the participant:
//     can always switch a thread to any open status (i.e. he can always open it but not close it)
//   - A user who can write on the thread can switch from an open status to another open status.
func userCanChangeThreadStatus(user *database.User, oldStatus, newStatus string, participantID int64, threadInfo *threadInfo) bool {
	if oldStatus == newStatus {
		return true
	}

	wasOpen := database.IsThreadOpenStatus(oldStatus)
	willBeOpen := database.IsThreadOpenStatus(newStatus)

	if user.GroupID == participantID {
		// * the participant of a thread can always switch the thread from open to any another other status.
		// * he can only switch it from not-open to an open status if he is allowed to request help on this item.
		// -> "allowed request help" have been checked before calling this method, therefore, the user can always
		//     change the status in this situation.
		return true
	} else if willBeOpen {
		// a user who has can_watch>=answer on the item AND can_watch_members on the participant:
		// can always switch a thread to any open status (i.e. he can always open it but not close it)
		if userCanWatchForThread(threadInfo) {
			return true
		} else if wasOpen {
			// a user who can write on the thread can switch from an open status to another open status
			return userCanWriteInThread(user, participantID, threadInfo)
		}
	}

	return false
}
