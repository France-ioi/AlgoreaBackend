package items

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

// swagger:model attemptsListResponseRow
type attemptsListResponseRow struct {
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	CreatedAt database.Time `json:"created_at"`
	// required: true
	ScoreComputed float32 `json:"score_computed"`
	// required: true
	Validated bool `json:"validated"`
	// required: true
	StartedAt *database.Time `json:"started_at"`
	// required: true
	EndedAt *database.Time `json:"ended_at"`
	// required: true
	AllowsSubmissionsUntil database.Time `json:"allows_submissions_until"`
	// required: true
	LatestActivityAt database.Time `json:"latest_activity_at"`
	// required: true
	HelpRequested bool `json:"help_requested"`
	// required: true
	UserCreator *struct {
		*structures.UserPersonalInfo

		// required: true
		Login string `json:"login"`

		ShowPersonalInfo bool `json:"-"`

		// required: true
		GroupID *int64 `json:"group_id,string"`
	} `gorm:"embedded;embedded_prefix:user_creator__" json:"user_creator"`
}

// swagger:operation GET /items/{item_id}/attempts items attemptsList
//
//	---
//	summary: List attempts/results for an item
//	description: >
//	 Returns attempts of the current participant (the current user or `{as_team_id}` team) with their results
//	 for the given item within the parent attempt.
//
//
//	 `first_name` and `last_name` of attempt creators are only visible to attempt creators themselves and
//	 to managers of those attempt creators' groups to which they provided view access to personal data.
//
//
//	 Restrictions:
//		 * `{as_team_id}` (if given) should be the current user's team,
//		 * the participant should have at least 'content' access on the item,
//		 * if `{attempt_id}` is given, it should exist for the participant in order to determine `{parent_attempt_id}`
//			 (we assume that the 'zero attempt' always exists and it is its own parent attempt),
//
//	 otherwise the 'forbidden' error is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: parent_attempt_id
//			description: "`id` of a parent attempt. This parameter is incompatible with `attempt_id`."
//			in: query
//			type: integer
//			format: int64
//		- name: attempt_id
//			description: "`id` of an attempt for the `{item_id}`.
//								This parameter is incompatible with `parent_attempt_id`."
//			in: query
//			type: integer
//			format: int64
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//		- name: sort
//			in: query
//			default: [id]
//			type: array
//			items:
//				type: string
//				enum: [id,-id]
//		- name: from.id
//			description: Start the page from the attempt next to the attempt with `results.attempt_id` = `{from.id}`
//			in: query
//			type: integer
//			format: int64
//		- name: limit
//			description: Display first N attempts
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. Success response with an array of attempts
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/attemptsListResponseRow"
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
func (srv *Service) listAttempts(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	itemID, participantID, parentAttemptID, err := srv.resolveParametersForListAttempts(httpRequest)
	service.MustNotBeError(err)

	user := srv.GetUser(httpRequest)

	query := constructQueryForGettingAttemptsList(srv.GetStore(httpRequest), participantID, itemID, user).
		Where("attempts.id = ? OR attempts.parent_attempt_id = ?", parentAttemptID, parentAttemptID)

	query = service.NewQueryLimiter().Apply(httpRequest, query)
	query, err = service.ApplySortingAndPaging(
		httpRequest, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"id": {ColumnName: "results.attempt_id"},
			},
			DefaultRules: "id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	service.MustNotBeError(err)

	var result []attemptsListResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	for index := range result {
		if result[index].UserCreator.GroupID == nil {
			result[index].UserCreator = nil
		} else if !result[index].UserCreator.ShowPersonalInfo {
			result[index].UserCreator.UserPersonalInfo = nil
		}
	}

	render.Respond(responseWriter, httpRequest, result)
	return nil
}

func constructQueryForGettingAttemptsList(store *database.DataStore, participantID, itemID int64, user *database.User) *database.DB {
	return store.Results().Where("results.participant_id = ?", participantID).
		Where("item_id = ?", itemID).
		Joins("JOIN attempts ON attempts.participant_id = results.participant_id AND attempts.id = results.attempt_id").
		Joins("LEFT JOIN users ON users.group_id = attempts.creator_id").
		WithPersonalInfoViewApprovals(user).
		Select(`
			attempts.id, attempts.created_at, attempts.allows_submissions_until,
			results.score_computed, results.validated, attempts.ended_at,
			results.started_at, results.latest_activity_at, results.help_requested,
			users.login AS user_creator__login,
			users.group_id = ? OR personal_info_view_approvals.approved AS user_creator__show_personal_info,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.first_name, NULL) AS user_creator__first_name,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.last_name, NULL) AS user_creator__last_name,
			users.group_id AS user_creator__group_id`, user.GroupID, user.GroupID, user.GroupID)
}

func (srv *Service) resolveParametersForListAttempts(httpRequest *http.Request) (
	itemID, participantID, parentAttemptID int64, err error,
) {
	itemID, err = service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return 0, 0, 0, service.ErrInvalidRequest(err)
	}

	attemptID, parentAttemptID, attemptIDSet, err := attemptIDOrParentAttemptID(httpRequest)
	if err != nil {
		return 0, 0, 0, err
	}

	participantID = service.ParticipantIDFromContext(httpRequest.Context())
	store := srv.GetStore(httpRequest)

	if attemptIDSet {
		if attemptID != 0 {
			var result struct{ ParentAttemptID int64 }
			err = store.Attempts().
				Where("attempts.participant_id = ? AND attempts.id = ?", participantID, attemptID).
				Select("IF(attempts.root_item_id = ?, attempts.parent_attempt_id, attempts.id) AS parent_attempt_id", itemID).
				Take(&result).Error()
			if gorm.IsRecordNotFoundError(err) {
				return 0, 0, 0, service.ErrAPIInsufficientAccessRights
			}
			service.MustNotBeError(err)
			parentAttemptID = result.ParentAttemptID
		}
	}

	found, err := store.Permissions().MatchingGroupAncestors(participantID).
		WherePermissionIsAtLeast("view", "content").
		Where("item_id = ?", itemID).HasRows()

	service.MustNotBeError(err)
	if !found {
		return 0, 0, 0, service.ErrAPIInsufficientAccessRights
	}
	return itemID, participantID, parentAttemptID, nil
}
