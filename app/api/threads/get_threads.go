package threads

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:response thread
type thread struct {
	Item        item        `json:"item"`
	Participant participant `json:"participant"`

	// enum: not_started,waiting_for_participant,waiting_for_trainer,closed
	Status         string         `json:"status"`
	LatestUpdateAt *database.Time `json:"latest_update_at"`
	MessageCount   int            `json:"message_count"`
}

type item struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	LanguageTag string `json:"language_tag"`
}

type participant struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	FirstName int64  `json:"first_name"`
	LastName  int64  `json:"last_name"`
}

// swagger:operation GET /items/{item_id}/participant/{participant_id}/thread threads getThreads
// ---
// summary: Service to list the visible threads for a user.
// description: >
//
//   Service to list the visible threads for a user.
//
//   Exactly one of [`watched_group_id`, `is_mine`] is given.
//
//   * If `is_mine = 1`, only threads whose the participant is the current-user and whose the item is visible
//     (can_view >= content) by the current-user are returned.
//   * If `is_mine = 0`, only threads that the current-user can list and whose the current-user is NOT the participant
//     are returned
//   * If `watched_group_id` is given, only threads whose the participant is descendant (including self) of the
//     `watched_group_id` are returned
//   * If `item_id` is given, only threads whose the `item_id` is or is descendant of the given `item_id` are returned
//   * `first_name` and `last_name` are only returned for the current user or if the user approved access to their personal
//     info for some group managed by the current user
//
//   Validations:
//     * if `watched_group_id` given: the current-user must be (implicitly or explicitly) a manager with
//       `can_watch_members` of `watched_group_id`
//     * if `item_id` is given, the current user needs `can_view >= content` on it
//
//   Extra:
//     * Use ordering and page limit as usual. By default, ordering should be `latest_update_at desc` (+ tie-breaking rules)
//     * The service should support filtering by `status`
//     * The service should support filtering by `latest_update_at` only return entries whose the `latest_update_at` is
//       greater than a given datetime (given in UTC8601)
//
// parameters:
//   - name: watched_group_id
//     in: query
//     type: integer
//     format: int64
//   - name: is_mine
//     in: query
//     type: bool
//   - name: item_id
//     in: query
//     type: integer
//     format: int64
// responses:
//   "200":
//     description: OK. Threads data
//     schema:
//       type: array
//       items:
//         "$ref": "#/responses/thread"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getThreads(rw http.ResponseWriter, r *http.Request) service.APIError {
	watchedGroupID, ok, apiError := srv.ResolveWatchedGroupID(r)
	if apiError != service.NoError {
		return apiError
	}

	isMine, err := service.ResolveURLQueryGetBoolField(r, "is_mine")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	itemID, err := service.ResolveURLQueryGetInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	fmt.Println(watchedGroupID, ok, isMine, itemID)

	threadItem := thread{}
	threads := []thread{threadItem}

	render.Respond(rw, r, threads)
	return service.NoError
}
