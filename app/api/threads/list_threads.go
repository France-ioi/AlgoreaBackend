package threads

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

// swagger:model thread
type thread struct {
	// required:true
	Item item `gorm:"embedded;embedded_prefix:item__" json:"item"`
	// required:true
	Participant participant `gorm:"embedded;embedded_prefix:participant__" json:"participant"`

	// required:true
	// enum: not_started,waiting_for_participant,waiting_for_trainer,closed
	Status string `json:"status"`
	// required:true
	LatestUpdateAt database.Time `json:"latest_update_at"`
	// required:true
	MessageCount int `json:"message_count"`
}

type item struct {
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	// enum: Chapter,Task,Skill
	Type string `json:"type"`
	// required:true
	Title *string `json:"title"`
	// required:true
	LanguageTag string `json:"language_tag"`
}

type participant struct {
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	Login string `json:"login"`

	*structures.UserPersonalInfo
	ShowPersonalInfo bool `json:"-"`
}

type listThreadParameters struct {
	WatchedGroupID int64
	IsMine         bool
	ItemID         int64
	Status         string
	LatestUpdateGt *time.Time
}

// swagger:operation GET /threads threads listThreads
//
//		---
//		summary: List threads
//		description: >
//
//			List threads visible to the current user.
//
//
//			Exactly one of [`watched_group_id`, `is_mine`] should be given.
//				* If `is_mine` = 1, only threads for which the participant is the current user are returned.
//				* If `is_mine` = 0, only threads for which the current user is NOT the participant are returned.
//				* If `watched_group_id` is given, only threads in which the participant is descendant (including self)
//				of the `watched_group_id` are returned.
//				* If `item_id` is given, only threads for which the `item_id` is or is descendant of the given `item_id` are returned.
//				* `first_name` and `last_name` are only returned for the current user, or if the user approved access to their personal
//				info for some group managed by the current user.
//
//
//			The returned threads are those for which the current user has `can_view`>=content permission on item_id and
//	   matching one of these conditions:
//	    	* the participant of the thread is the current user, OR
//	    	* the current user has `can_watch`>=answer permission on the thread's item, OR
//	    	* the current user has `can_watch`=result permission on the thread's item, AND the current user is a descendant
//	      of the group the participant has requested help to, AND the thread is either open or closed for less than 2 weeks,
//	      AND the current user has a validated result on the thread's item.
//
//
//			Validations:
//				* if `watched_group_id` is given: the current user must be (implicitly or explicitly) a manager
//					with `can_watch_members` on `watched_group_id`.
//					Otherwise, the 'forbidden' error is returned.
//				* if `item_id` is given: the current user must have can_view >= 'content' permission on the `item_id`.
//					Otherwise, the 'forbidden' error is returned.
//
//
//			Extra:
//				* By default, ordering is by `latest_update_at` DESC.
//				* Filter by `status` by providing the `status` parameter with the filter value.
//				* Filter by greater than `latest_update_at` by providing the `latest_update_gt` with a datetime in UTC8601 format.
//
//		parameters:
//			- name: watched_group_id
//				in: query
//				type: integer
//				format: int64
//			- name: is_mine
//				in: query
//				type: boolean
//			- name: status
//				description: Filter by status
//				in: query
//				type: string
//				enum: [waiting_for_participant,waiting_for_trainer,closed]
//			- name: latest_update_gt
//				description: Only threads where `latest_update_at`>`latest_update_gt`.
//				in: query
//				type: string
//				format: date-time
//			- name: sort
//				in: query
//				default: [-latest_update_at,item_id,participant_id]
//				type: array
//				items:
//					type: string
//					enum: [latest_update_at,-latest_update_at,item_id,-item_id,participant_id,-participant_id]
//			- name: from.item_id
//				description: >
//					Start the page from the thread next to the thread with `threads.item.id`=`{from.item_id}`.
//					When provided, from.participant_id should be provided too.
//				in: query
//				type: integer
//				format: int64
//			- name: from.participant_id
//				description: >
//					Start the page from the thread next to the thread with `threads.participant.id`=`{from.participant_id}`.
//					When provided, from.item_id should be provided too.
//				in: query
//				type: integer
//				format: int64
//			- name: limit
//				description: Display the first N threads
//				in: query
//				type: integer
//				maximum: 1000
//				default: 500
//		responses:
//			"200":
//				description: OK. Threads data
//				schema:
//					type: array
//					items:
//						"$ref": "#/definitions/thread"
//			"400":
//				"$ref": "#/responses/badRequestResponse"
//			"401":
//				"$ref": "#/responses/unauthorizedResponse"
//			"403":
//				"$ref": "#/responses/forbiddenResponse"
//			"408":
//				"$ref": "#/responses/requestTimeoutResponse"
//			"500":
//				"$ref": "#/responses/internalErrorResponse"
func (srv *Service) listThreads(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	params, err := srv.resolveListThreadParameters(httpRequest)
	service.MustNotBeError(err)

	queryDB, err := srv.constructListThreadsQuery(httpRequest, params)
	service.MustNotBeError(err)

	var threads []thread
	err = queryDB.Scan(&threads).Error()
	service.MustNotBeError(err)

	for index := range threads {
		if !threads[index].Participant.ShowPersonalInfo {
			threads[index].Participant.UserPersonalInfo = nil
		}
	}

	render.Respond(responseWriter, httpRequest, threads)
	return nil
}

func (srv *Service) constructListThreadsQuery(httpRequest *http.Request, params listThreadParameters) (*database.DB, error) {
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	query := store.Threads().
		Joins("JOIN items ON	items.id = threads.item_id").
		// actually participants can be teams, but here we filter such threads out :(
		Joins("JOIN users ON	users.group_id = threads.participant_id")

	if params.ItemID != 0 {
		query = query.
			Where("items.id IN (?) OR items.id = ?",
				store.ItemAncestors().Select("child_item_id").Where("ancestor_item_id = ?", params.ItemID).QueryExpr(),
				params.ItemID)
	}

	if params.Status != "" {
		query = query.Where("threads.status = ?", params.Status)
	}

	if params.LatestUpdateGt != nil {
		query = query.Where("threads.latest_update_at > ?", params.LatestUpdateGt)
	}

	switch {
	case params.WatchedGroupID != 0:
		query = query.Joins(`
			JOIN groups_ancestors_active ON
						threads.participant_id = groups_ancestors_active.child_group_id AND
						groups_ancestors_active.ancestor_group_id = ?`, params.WatchedGroupID)

	case params.IsMine:
		query = query.Where("threads.participant_id = ?", user.GroupID)

	case !params.IsMine:
		query = query.Where("threads.participant_id != ?", user.GroupID)
	}

	if !params.IsMine /* watched_group_id is given or is_mine is false */ {
		// check if the current user has `can_watch`>=answer permission on the thread's item
		canWatchAnswerSubQuery := store.Permissions().MatchingUserAncestors(user).
			WherePermissionIsAtLeast("watch", "answer").
			Where("permissions.item_id = threads.item_id").
			Select("1").Limit(1).SubQuery()
		// check if the current user has `can_watch`=result permission on the thread's item
		canWatchResultsSubQuery := store.Permissions().MatchingUserAncestors(user).
			WherePermissionIsAtLeast("watch", "result").
			Where("permissions.item_id = threads.item_id").
			Select("1").Limit(1).SubQuery()

		// check if the current user is a descendant of the group the participant has requested help to
		userIsDescendantOfHelperGroupSubQuery := store.ActiveGroupAncestors().
			Where("groups_ancestors_active.ancestor_group_id = threads.helper_group_id").
			Where("groups_ancestors_active.child_group_id = ?", user.GroupID).
			Select("1").Limit(1).SubQuery()

		// check if the current user has a validated result on the thread's item
		userHasValidatedResultOnItemSubQuery := store.Results().
			Where("results.item_id = threads.item_id").
			Where("results.validated").
			Where("results.participant_id = ?", user.GroupID).
			Select("1").Limit(1).SubQuery()

		query = query.
			Where(`
				threads.participant_id = ? OR
				? OR
				(
					(threads.status IN ('waiting_for_participant', 'waiting_for_trainer') OR threads.latest_update_at > NOW() - INTERVAL 2 WEEK) AND
					? AND
					? AND
					?
				)`,
				user.GroupID,           /* the current user is the participant OR */
				canWatchAnswerSubQuery, /* the current user has `can_watch`>=answer permission on the thread's item, OR */
				/* ( */
				// the thread is either open or closed for less than 2 weeks AND
				canWatchResultsSubQuery,               /* the current user has `can_watch`=result permission on the thread's item, AND */
				userIsDescendantOfHelperGroupSubQuery, /* the current user is a descendant of the group the participant has requested help to, AND */
				userHasValidatedResultOnItemSubQuery,  /* the current user has a validated result on the thread's item. */
				/* ) */
			)
	}

	queryDB := query.
		WhereItemsContentAreVisible(user.GroupID).
		JoinsUserAndDefaultItemStrings(user).
		WithPersonalInfoViewApprovals(user).
		Select(`
			items.id AS item__id,
			items.type AS item__type,
			COALESCE(user_strings.language_tag, default_strings.language_tag) AS item__language_tag,
			COALESCE(user_strings.title, default_strings.title) AS item__title,
			threads.participant_id AS participant__id,
			threads.participant_id = ? OR personal_info_view_approvals.approved AS participant__show_personal_info,
			IF(threads.participant_id = ? OR personal_info_view_approvals.approved, users.first_name, NULL) AS participant__first_name,
			IF(threads.participant_id = ? OR personal_info_view_approvals.approved, users.last_name, NULL) AS participant__last_name,
			users.login AS participant__login,
			threads.status AS status,
			threads.message_count AS message_count,
			threads.latest_update_at AS latest_update_at
		`, user.GroupID, user.GroupID, user.GroupID)

	return applySortingAndPaging(httpRequest, queryDB)
}

func applySortingAndPaging(r *http.Request, queryDB *database.DB) (*database.DB, error) {
	queryDB = service.NewQueryLimiter().Apply(r, queryDB)

	return service.ApplySortingAndPaging(r, queryDB, &service.SortingAndPagingParameters{
		Fields: service.SortingAndPagingFields{
			"latest_update_at": {ColumnName: "threads.latest_update_at"},
			"item_id":          {ColumnName: "items.id"},
			"participant_id":   {ColumnName: "threads.participant_id"},
		},
		DefaultRules: "-latest_update_at",
		TieBreakers: service.SortingAndPagingTieBreakers{
			"item_id":        service.FieldTypeInt64,
			"participant_id": service.FieldTypeInt64,
		},
	})
}

func (srv *Service) resolveListThreadParameters(httpRequest *http.Request) (params listThreadParameters, err error) {
	var watchedGroupOK bool
	params.WatchedGroupID, watchedGroupOK, err = srv.ResolveWatchedGroupID(httpRequest)
	if err != nil {
		return params, err
	}

	var isMineError error
	params.IsMine, isMineError = service.ResolveURLQueryGetBoolField(httpRequest, "is_mine")

	if watchedGroupOK && isMineError == nil {
		return params, service.ErrInvalidRequest(errors.New("must not provide watched_group_id and is_mine at the same time"))
	}
	if !watchedGroupOK && isMineError != nil {
		return params, service.ErrInvalidRequest(errors.New("one of watched_group_id or is_mine must be given"))
	}

	if service.URLQueryPathHasField(httpRequest, "item_id") {
		params.ItemID, err = service.ResolveURLQueryGetInt64Field(httpRequest, "item_id")
		if err != nil {
			return params, service.ErrInvalidRequest(err)
		}

		user := srv.GetUser(httpRequest)
		store := srv.GetStore(httpRequest)
		found, err := store.Permissions().
			MatchingGroupAncestors(user.GroupID).
			WherePermissionIsAtLeast("view", "content").
			Where("permissions.item_id = ?", params.ItemID).HasRows()
		service.MustNotBeError(err)
		if !found {
			return params, service.ErrForbidden(errors.New("no rights to view content of the item"))
		}
	}

	return resolveFilterParameters(httpRequest, params)
}

func resolveFilterParameters(httpRequest *http.Request, params listThreadParameters) (listThreadParameters, error) {
	var err error

	params.Status, err = service.ResolveURLQueryGetStringField(httpRequest, "status")
	if err != nil {
		params.Status = ""
	}

	if service.URLQueryPathHasField(httpRequest, "latest_update_gt") {
		latestUpdateGt, err := service.ResolveURLQueryGetTimeField(httpRequest, "latest_update_gt")
		if err != nil {
			return params, service.ErrInvalidRequest(err)
		}

		params.LatestUpdateGt = &latestUpdateGt
	}

	return params, nil
}
