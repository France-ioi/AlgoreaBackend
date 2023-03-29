package threads

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:response thread
type thread struct {
	Item        item        `json:"item" gorm:"embedded;embedded_prefix:item__"`
	Participant participant `json:"participant" gorm:"embedded;embedded_prefix:participant__"`

	// enum: not_started,waiting_for_participant,waiting_for_trainer,closed
	Status         string         `json:"status"`
	LatestUpdateAt *database.Time `json:"latest_update_at"`
	MessageCount   int            `json:"message_count"`
}

type item struct {
	ID          int64  `json:"id,string"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	LanguageTag string `json:"language_tag"`
}

type participant struct {
	ID        int64  `json:"id,string"`
	Login     string `json:"login"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// swagger:operation GET /items/{item_id}/participant/{participant_id}/thread threads listThreads
// ---
// summary: Service to list the visible threads for a user.
// description: >
//
//	Service to list the visible threads for a user.
//
//	* If `watched_group_id` is given, only threads in which the participant is descendant (including self)
//	  of the `watched_group_id` are returned.
//	* `first_name` and `last_name` are only returned for the current user or if the user approved access to their personal
//	  info for some group managed by the current user
//
//	Validations:
//	  * if `watched_group_id` is given: the current-user must be (implicitly or explicitly) a manager
//	    with `can_watch_members` on `watched_group_id`.
//
// parameters:
//   - name: watched_group_id
//     in: query
//     type: integer
//     format: int64
//
// responses:
//
//	"200":
//	  description: OK. Threads data
//	  schema:
//	    type: array
//	    items:
//	      "$ref": "#/responses/thread"
//	"400":
//	  "$ref": "#/responses/badRequestResponse"
//	"401":
//	  "$ref": "#/responses/unauthorizedResponse"
//	"403":
//	  "$ref": "#/responses/forbiddenResponse"
//	"500":
//	  "$ref": "#/responses/internalErrorResponse"
func (srv *Service) listThreads(rw http.ResponseWriter, r *http.Request) service.APIError {
	watchedGroupID, ok, apiError := srv.ResolveWatchedGroupID(r)
	if apiError != service.NoError {
		return apiError
	}
	if !ok {
		return service.ErrInvalidRequest(errors.New("not implemented yet: watchedGroupID must be given"))
	}

	user := srv.GetUser(r)
	store := srv.GetStore(r)

	var threads []thread
	err := store.Threads().
		JoinsItem().
		JoinsUserParticipant().
		WhereParticipantIsInGroup(watchedGroupID).
		JoinsUserAndDefaultItemStrings(user).
		WithPersonalInfoViewApprovals(user).
		Order("items.id ASC"). // Default to make the result deterministic.
		Select(`
			items.id AS item__id,
			items.type AS item__type,
			COALESCE(user_strings.language_tag, default_strings.language_tag) AS item__language_tag,
			COALESCE(user_strings.title, default_strings.title) AS item__title,
			threads.participant_id AS participant__id,
			IF(threads.participant_id = ? OR personal_info_view_approvals.approved, users.first_name, NULL) AS participant__first_name,
			IF(threads.participant_id = ? OR personal_info_view_approvals.approved, users.last_name, NULL) AS participant__last_name,
			users.login AS participant__login,
			threads.status AS status,
			threads.message_count AS message_count,
			threads.latest_update_at AS latest_update_at
		`, user.GroupID, user.GroupID).
		Scan(&threads).
		Error()
	service.MustNotBeError(err)

	render.Respond(rw, r, threads)
	return service.NoError
}
