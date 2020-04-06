package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model itemAttemptsViewResponseRow
type itemAttemptsViewResponseRow struct {
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	ScoreComputed float32 `json:"score_computed"`
	// required: true
	Validated bool `json:"validated"`
	// Nullable
	// required: true
	StartedAt   *database.Time `json:"started_at"`
	UserCreator *struct {
		// required: true
		Login string `json:"login"`
		// Nullable
		// required: true
		FirstName *string `json:"first_name"`
		// Nullable
		// required: true
		LastName *string `json:"last_name"`

		GroupID *int64 `json:"-"`
	} `json:"user_creator" gorm:"embedded;embedded_prefix:user_creator__"`
}

// swagger:operation GET /items/{item_id}/attempts items itemAttemptsView
// ---
// summary: List attempts for a task
// description: Returns attempts (with results) made by the current user (if `as_team_id` is not given) or
//              `as_team_id` (otherwise) while solving the task given in `item_id`.
//
//
//              * The current user (if no `as_team_id` given) or `as_team_id` (otherwise) should have
//                at least 'content' permission on the task item,
//
//              * the current user must be member of the team if `as_team_id` is provided,
//
//
//              otherwise the 'forbidden' error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
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
//         "$ref": "#/definitions/itemAttemptsViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getAttempts(w http.ResponseWriter, r *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	user := srv.GetUser(r)

	groupID := user.GroupID
	if len(r.URL.Query()["as_team_id"]) != 0 {
		groupID, err = service.ResolveURLQueryGetInt64Field(r, "as_team_id")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}

		var found bool
		found, err = srv.Store.Groups().TeamGroupForUser(groupID, user).Where("groups.id = ?", groupID).HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrForbidden(errors.New("can't use given as_team_id as a user's team"))
		}
	}

	var found bool
	found, err = srv.Store.Items().ByID(itemID).WhereGroupHasViewPermissionOnItems(groupID, "content").HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	query := srv.Store.Results().Where("results.participant_id = ?", groupID).
		Where("item_id = ?", itemID).
		Joins("JOIN attempts ON attempts.participant_id = results.participant_id AND attempts.id = results.attempt_id").
		Joins("LEFT JOIN users AS creators ON creators.group_id = attempts.creator_id").
		Select(`
			attempts.id, results.score_computed, results.validated,
			results.started_at, creators.login AS user_creator__login,
			creators.first_name AS user_creator__first_name, creators.last_name AS user_creator__last_name,
			creators.group_id AS user_creator__group_id`)
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query, map[string]*service.FieldSortingParams{
		"id": {ColumnName: "results.attempt_id", FieldType: "int64"},
	}, "id", []string{"id"}, false)
	if apiError != service.NoError {
		return apiError
	}
	var result []itemAttemptsViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	for index := range result {
		if result[index].UserCreator.GroupID == nil {
			result[index].UserCreator = nil
		}
	}

	render.Respond(w, r, result)
	return service.NoError
}
