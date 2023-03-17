package threads

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model threadInfo
type threadInfo struct {
	// required: true
	ParticipantID int64 `json:"participant_id"`
	// required: true
	ItemID int64 `json:"item_id"`
	// required: true
	// enum: not_started,waiting_for_participant,waiting_for_trainer,closed
	Status string `json:"status"`
}

// swagger:operation GET /items/{item_id}/participant/{participant_id}/thread threads getThread
// ---
// summary: Retrieve a thread information
// description: >
//   Retrieve a thread information.
//
//   The `status` is `not_started` if the thread hasn't been started
//
//   Restrictions:
//     * one of these conditions matches:
//       - the current-user is the thread participant and allowed to "can_view >= content" the item
//       - the current-user has the "can_watch >= answer" permission on the item
//       - the following rules all matches:
//         * the current-user is descendant of the thread helper_group
//         * the thread is either open (=waiting_for_participant or =waiting_for_trainer), or closed for less than 2 weeks
//         * the current-user has validated the item
//
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: participant_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// responses:
//   "200":
//     description: OK. Success response with thread data
//     schema:
//       "$ref": "#/definitions/threadInfo"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getThread(rw http.ResponseWriter, r *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	participantID, err := service.ResolveURLQueryPathInt64Field(r, "participant_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	store := srv.GetStore(r)

	canRetrieveThread := store.Threads().CanRetrieveThread(user, participantID, itemID)
	if !canRetrieveThread {
		return service.InsufficientAccessRightsError
	}

	threadInfo := new(threadInfo)
	threadInfo.ItemID = itemID
	threadInfo.ParticipantID = participantID
	threadInfo.Status = store.Threads().GetThreadStatus(participantID, itemID)

	render.Respond(rw, r, threadInfo)
	return service.NoError
}
