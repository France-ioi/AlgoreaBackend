package threads

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

const threadTokenLifetime = 2 * time.Hour

// swagger:model threadGetResponse
type threadGetResponse struct {
	// required:true
	ParticipantID int64 `json:"participant_id,string"`
	// required:true
	ItemID int64 `json:"item_id,string"`
	// required:true
	// enum: not_started,waiting_for_participant,waiting_for_trainer,closed
	Status string `json:"status"`
	// The ThreadToken
	// required:true
	ThreadToken string `json:"token"`
	// This field is not really present, it is here only to document the content of token.
	// required:false
	TokenForDoc *payloads.ThreadToken `json:"token_not_present_only_for_doc,omitempty"`
}

// swagger:operation GET /items/{item_id}/participant/{participant_id}/thread threads threadGet
//
//	---
//	summary: Retrieve a thread information
//	description: >
//		Retrieve a thread information.
//
//
//		The `status` is `not_started` if the thread hasn't been started
//
//
//		Restrictions:
//			* the current user should have `can_view` >= content permission on the item AND
//			* one of these conditions must match:
//				- the current user should be the thread's participant OR
//				- the current user should have `can_watch` >= answer permission on the item OR
//				- all the following rules should be satisfied:
//					* the current user should have `can_watch` >= result permission on the item AND
//					* the current user should be a descendant of the thread helper_group AND
//					* the thread should be either open (=waiting_for_participant or =waiting_for_trainer), or closed for less than 2 weeks AND
//					* the current user should have a validated result on the item.
//
//			Otherwise, a forbidden error is returned.
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
//	responses:
//		"200":
//			description: OK. Success response with thread data
//			schema:
//				"$ref": "#/definitions/threadGetResponse"
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
func (srv *Service) getThread(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	participantID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "participant_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	threadGetResponse := new(threadGetResponse)
	threadGetResponse.ItemID = itemID
	threadGetResponse.ParticipantID = participantID

	// check if the current-user has "can_view >= content" on the item
	currentUserCanViewContentSubQuery := store.Permissions().MatchingUserAncestors(user).
		Where("permissions.item_id = ?", itemID).
		WherePermissionIsAtLeast("view", "content").
		Select("1").
		Limit(1).
		SubQuery()

	var threadInfo threadInfo
	err = constructThreadInfoQuery(store, user, itemID, participantID).
		Where("?", currentUserCanViewContentSubQuery).
		Having(`
			(? = ?) OR
			user_can_watch_answer OR (
				(thread_is_open OR thread_was_updated_recently) AND
				user_can_watch_result AND user_is_descendant_of_helper_group AND user_has_validated_result_on_item
			)`,
			user.GroupID, participantID).
		Take(&threadInfo).Error()

	if gorm.IsRecordNotFoundError(err) {
		return service.ErrAPIInsufficientAccessRights
	}
	service.MustNotBeError(err)

	threadGetResponse.Status = threadInfo.ThreadStatus

	threadGetResponse.ThreadToken, err = srv.generateThreadToken(itemID, participantID, &threadInfo, user)
	service.MustNotBeError(err)

	render.Respond(responseWriter, httpRequest, threadGetResponse)

	return nil
}

func (srv *Service) generateThreadToken(itemID, participantID int64, threadInfo *threadInfo, user *database.User) (string, error) {
	expirationTime := time.Now().Add(threadTokenLifetime)

	threadToken, err := (&token.Token[payloads.ThreadToken]{Payload: payloads.ThreadToken{
		ItemID:        strconv.FormatInt(itemID, 10),
		ParticipantID: strconv.FormatInt(participantID, 10),
		UserID:        strconv.FormatInt(user.GroupID, 10),
		IsMine:        participantID == user.GroupID,
		CanWatch:      userCanWatchForThread(threadInfo),
		CanWrite:      userCanWriteInThread(user, participantID, threadInfo),
		Exp:           strconv.FormatInt(expirationTime.Unix(), 10),
	}}).Sign(srv.TokenConfig.PrivateKey)

	return threadToken, err
}
