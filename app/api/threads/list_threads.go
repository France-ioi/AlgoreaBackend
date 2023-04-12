package threads

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model thread
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

// swagger:operation GET /thread threads listThreads
//
//	---
//	summary: Service to list the visible threads for a user.
//	description: >
//
//		Service to list the visible threads for a user.
//
//		Exactly one of [`watched_group_id`, `is_mine`] is given
//
//		* If `is_mine` = 1, only threads for which the participant is the current-user and for which the current-user has
//			`can_view >= content` on the item are returned.
//		* If `is_mine` = 0, only threads that the current-user
//			[https://france-ioi.github.io/algorea-devdoc/forum/forum-perm/#listing--reading-a-thread](can list) and for which
//			the current-user is NOT the participant are returned.
//		* If `watched_group_id` is given, only threads in which the participant is descendant (including self)
//			of the `watched_group_id` are returned.
//		* `first_name` and `last_name` are only returned for the current user or if the user approved access to their personal
//			info for some group managed by the current user.
//
//		Validations:
//			* if `watched_group_id` is given: the current-user must be (implicitly or explicitly) a manager
//				with `can_watch_members` on `watched_group_id`.
//
//	parameters:
//		- name: watched_group_id
//			in: query
//			type: integer
//			format: int64
//		- name: is_mine
//			in: query
//			type: boolean
//
//	responses:
//
//		"200":
//			description: OK. Threads data
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/thread"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) listThreads(rw http.ResponseWriter, r *http.Request) service.APIError {
	watchedGroupID, isMine, apiError := srv.resolveListThreadParameters(r)
	if apiError != service.NoError {
		return apiError
	}

	user := srv.GetUser(r)
	store := srv.GetStore(r)

	var threads []thread
	query := store.Threads().
		JoinsItem().
		JoinsUserParticipant()

	switch {
	case watchedGroupID != 0:
		query = query.WhereParticipantIsInGroup(watchedGroupID)
	case isMine:
		query = query.NewThreadStore(query.
			WhereGroupHasPermissionOnItems(user.GroupID, "view", "content").
			Where("threads.participant_id = ?", user.GroupID),
		)

	case !isMine:
		// The user needs to:
		// - be allowed to view the item, AND
		// - not be the participant of the thread, AND
		// - one of the following conditions:
		//		* [canWatchAnswerQuery] have can_watch>=answer permission on the item, OR:
		//		* [userCanHelpQuery] The conditions of "WhereUserCanHelp"

		// It doesn't seem to be very efficient to do this. We could try to leverage the fact that MatchingGroupAncestors
		// is used in both canWatchAnswerQuery and WhereItemsAreVisible, if we measure perf issues.
		canWatchAnswerQuery := store.Threads().
			Select("items.id").
			WhereUserHasPermissionOnItems(user, "watch", "answer").
			SubQuery()

		userCanHelpQuery := store.Threads().
			WhereUserCanHelp(user).
			Select("threads.item_id, threads.participant_id").
			SubQuery()

		query = query.NewThreadStore(query.
			WhereItemsAreVisible(user.GroupID).
			Where("threads.participant_id != ?", user.GroupID).
			Where("threads.item_id IN (?) OR (threads.item_id, threads.participant_id) IN (?)", canWatchAnswerQuery, userCanHelpQuery),
		)
	}

	err := query.
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

func (srv *Service) resolveListThreadParameters(r *http.Request) (watchedGroupID int64, isMine bool, apiError service.APIError) {
	var watchedGroupOK bool
	watchedGroupID, watchedGroupOK, apiError = srv.ResolveWatchedGroupID(r)
	if apiError != service.NoError {
		return 0, false, apiError
	}

	var isMineError error
	isMine, isMineError = service.ResolveURLQueryGetBoolField(r, "is_mine")

	if watchedGroupOK && isMineError == nil {
		return 0, false, service.ErrInvalidRequest(errors.New("must not provide watched_group_id and is_mine at the same time"))
	}
	if !watchedGroupOK && isMineError != nil {
		return 0, false, service.ErrInvalidRequest(errors.New("one of watched_group_id or is_mine must be given"))
	}

	return watchedGroupID, isMine, service.NoError
}
