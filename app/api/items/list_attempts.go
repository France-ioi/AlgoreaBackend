package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
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
	// Nullable
	// required: true
	StartedAt *database.Time `json:"started_at"`
	// Nullable
	// required: true
	EndedAt *database.Time `json:"ended_at"`
	// required: true
	AllowsSubmissionsUntil database.Time `json:"allows_submissions_until"`
	// required: true
	LatestActivityAt database.Time `json:"latest_activity_at"`
	UserCreator      *struct {
		// required: true
		Login string `json:"login"`
		// Nullable
		// required: true
		FirstName *string `json:"first_name"`
		// Nullable
		// required: true
		LastName *string `json:"last_name"`
		// required: true
		GroupID *int64 `json:"group_id,string"`
	} `json:"user_creator" gorm:"embedded;embedded_prefix:user_creator__"`
}

// swagger:operation GET /items/{item_id}/attempts items attemptsList
// ---
// summary: List attempts/results for an item
// description: Returns attempts with their results of the current participant (the current user or `as_team_id` team)
//              for the given item within the parent attempt.
//
//
//              Restrictions:
//                * the list of item IDs should be a valid path from a root item
//                  (`items.is_root`=1 or `items.id`=`groups.root_activity_id|root_skill_id` for one of the participant's ancestor groups),
//                * `as_team_id` (if given) should be the current user's team,
//                * the participant should have at least 'content' access on each listed item through that path,
//                * all the results within the ancestry of `attempt_id`/`parent_attempt_id` on the items path
//                  except for the last item should be started (`started_at` is not null),
//
//              otherwise the 'forbidden' error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: parent_attempt_id
//   description: "`id` of an attempt for the second to the last item in the path.
//                 This parameter is incompatible with `attempt_id`."
//   in: query
//   type: integer
// - name: attempt_id
//   description: "`id` of an attempt for the last item in the path.
//                 This parameter is incompatible with `parent_attempt_id`."
//   in: query
//   type: integer
// - name: as_team_id
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [id]
//   type: array
//   items:
//     type: string
//     enum: [id,-id]
// - name: from.id
//   description: Start the page from the attempt next to the attempt with `results.attempt_id` = `from.id`
//   in: query
//   type: integer
//   format: int64
// - name: limit
//   description: Display first N attempts
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with an array of attempts
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/attemptsListResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) listAttempts(w http.ResponseWriter, r *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	attemptID, parentAttemptID, attemptIDSet, apiError := srv.attemptIDOrParentAttemptID(r)
	if apiError != service.NoError {
		return apiError
	}

	user := srv.GetUser(r)
	groupID := user.GroupID
	if len(r.URL.Query()["as_team_id"]) != 0 {
		groupID, err = service.ResolveURLQueryGetInt64Field(r, "as_team_id")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}

		var found bool
		found, err = srv.Store.Groups().TeamGroupForUser(groupID, user).HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrForbidden(errors.New("can't use given as_team_id as a user's team"))
		}
	}

	if attemptIDSet {
		err := srv.Store.Attempts().
			Where("attempts.participant_id = ? AND attempts.id = ?", groupID, attemptID).
			Joins("JOIN attempts ON attempts.participant_id = results.participant_id AND attempts.id = results.attempt_id").
			Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = attempts.participant_id").
			Joins("LEFT JOIN items_items ON items_items.child_item_id = results.item_id").
			Joins(`
				LEFT JOIN results AS parent_result
					ON parent_result.participant_id = attempts.participant_id AND
						 parent_result.attempt_id = IF(attempts.attempt.root_item_id = results.item_id, attempts.parent_attempt_id, attempts.id) AND
						 parent_result.item_id = items_items.parent_item_id`).
			Joins(`
				LEFT JOIN permissions_generated
					ON permissions_generated.group_id = groups_ancestors_active.ancestor_group_id AND
						 permissions_generated.item_id = items_items.parent_item_id AND
						 permissions_generated.can_view_generated_value >= ?`, srv.Store.PermissionsGranted().ViewIndexByName("content")).
			Where("NOT attempts.attempt.root_item_id = results.item_id OR permissions_generated.item_id").
			PluckFirst("parent_attempt.id", &parentAttemptID).Error()
		if gorm.IsRecordNotFoundError(err) {
			return service.InsufficientAccessRightsError
		}
		service.MustNotBeError(err)
	}

	found, err := srv.Store.Permissions().WithViewPermissionForGroup(groupID, "content").
		Where("item_id = ?", itemID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	query := srv.Store.Results().Where("results.participant_id = ?", groupID).
		Where("item_id = ?", itemID).
		Joins("JOIN attempts ON attempts.participant_id = results.participant_id AND attempts.id = results.attempt_id").
		Joins("LEFT JOIN users AS creators ON creators.group_id = attempts.creator_id").
		Where("attempts.id = ? OR attempts.parent_attempt_id = ?", parentAttemptID, parentAttemptID).
		Select(`
			attempts.id, attempts.created_at, attempts.allows_submissions_until,
			results.score_computed, results.validated, attempts.ended_at,
			results.started_at, results.latest_activity_at, creators.login AS user_creator__login,
			creators.first_name AS user_creator__first_name, creators.last_name AS user_creator__last_name,
			creators.group_id AS user_creator__group_id`)
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError = service.ApplySortingAndPaging(r, query, map[string]*service.FieldSortingParams{
		"id": {ColumnName: "results.attempt_id", FieldType: "int64"},
	}, "id", []string{"id"}, false)
	if apiError != service.NoError {
		return apiError
	}
	var result []attemptsListResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	for index := range result {
		if result[index].UserCreator.GroupID == nil {
			result[index].UserCreator = nil
		}
	}

	render.Respond(w, r, result)
	return service.NoError
}
