package items

import (
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model thread
type thread struct {
	// required: true
	ParticipantID int64 `json:"participant_id"`
	// required: true
	ItemID int64 `json:"item_id"`
	// required: true
	// enum: not_started,waiting_for_participant,waiting_for_trainer,closed
	Status string `json:"status"`
}

// swagger:operation GET /items/{item_id}/participant/{participant_id}/thread items getThread
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
//   format: int64
//   required: true
// - name: participant_id
//   in: path
//   format: int64
//   required: true
//  responses:
//    "200":
//     description: OK. Success response with thread data
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/thread"
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

	threadInfo := new(thread)
	threadInfo.ItemID = itemID
	threadInfo.ParticipantID = participantID

	// TODO: Try to make the permission checks one query with OR instead of using subqueries.

	// TODO: We need to update GORM for this and use https://gorm.io/docs/advanced_query.html#Group-Conditions

	// we check the permissions first without joining the threads because we need to distinguish between an
	// access error and the non-existence of the thread, which should be reported as status=not_started

	// check if the current-user is the thread participant and allowed to "can_view >= content" the item
	currentUserParticipantCanViewContent, err := store.Permissions().MatchingUserAncestors(user).
		Where("? = ?", user.GroupID, participantID).
		Where("permissions.item_id = ?", itemID).
		WherePermissionIsAtLeast("view", "content").
		Select("1").
		Limit(1).
		HasRows()
	service.MustNotBeError(err)

	// the current-user has the "can_watch >= answer" permission on the item
	currentUserCanWatch, err := store.Permissions().MatchingUserAncestors(user).
		Where("permissions.item_id = ?", itemID).
		WherePermissionIsAtLeast("watch", "answer").
		Select("1").
		Limit(1).
		HasRows()
	service.MustNotBeError(err)

	// the following rules all matches:
	// the current-user is descendant of the thread helper_group
	// the thread is either open (=waiting_for_participant or =waiting_for_trainer), or closed for less than 2 weeks
	// the current-user has validated the item

	// TODO: What if the current-user didn't validate the item but a team did?
	now := time.Now()
	twoWeeksAgo := now.AddDate(0, 0, -14)
	currentUserCanHelp, err := store.Threads().
		Joins("JOIN results ON results.item_id = threads.item_id").
		Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = ?", user.GroupID).
		Where("threads.helper_group_id = groups_ancestors_active.ancestor_group_id").
		Where("threads.item_id = ?", itemID).
		Where("threads.status != 'closed' OR (threads.status = 'closed' AND threads.latest_update_at > ?)", twoWeeksAgo).
		Where("results.participant_id = ?", user.GroupID).
		Where("results.validated").
		Limit(1).
		HasRows()
	service.MustNotBeError(err)

	// TODO: Do we really send 403 if it has been closed for more than 2 weeks? Or do we send status=not_started?
	if !currentUserParticipantCanViewContent && !currentUserCanWatch && !currentUserCanHelp {
		return service.InsufficientAccessRightsError
	}

	err = store.Threads().
		Select("threads.status AS status").
		Where("threads.participant_id = ?", participantID).
		Where("threads.item_id = ?", itemID).
		Limit(1).
		Take(&threadInfo).Error()
	if err != nil {
		threadInfo.Status = "not_started"
	}

	render.Respond(rw, r, threadInfo)
	return service.NoError
}
