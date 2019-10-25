package items

import (
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
	Order int32 `json:"order"`
	// required: true
	Score float32 `json:"score"`
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

// swagger:operation GET /items/{item_id}/attempts groups users attempts items itemAttemptsView
// ---
// summary: List attempts for a task
// description: Returns attempts made by the current user (if `items.has_attempts` = 0) or his
//              teams (if `items.has_attempts` = 1) while solving the task given in `item_id`.
//
//
//              The task item should be visible to the current user, otherwise an empty list is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: sort
//   in: query
//   default: [order,id]
//   type: array
//   items:
//     type: string
//     enum: [order,-order,id,-id]
// - name: from.order
//   description: Start the page from the attempt next to the attempt with `groups_attempts.order` = `from.order` and
//                `groups_attempts.id` = `from.id` (`from.id` is required when `from.order` is present)
//   in: query
//   type: integer
//   format: int32
// - name: from.id
//   description: Start the page from the attempt next to the attempt with `groups_attempts.order` = `from.order` and
//                `groups_attempts.id` = `from.id` (`from.order` is required when `from.id` is present)
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
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getAttempts(w http.ResponseWriter, r *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	user := srv.GetUser(r)
	query := srv.Store.GroupAttempts().VisibleAndByItemID(user, itemID).
		Joins("LEFT JOIN users AS creators ON creators.group_id = groups_attempts.creator_user_group_id").
		Select(`
			groups_attempts.id, groups_attempts.order, groups_attempts.score, groups_attempts.validated,
			groups_attempts.started_at, creators.login AS user_creator__login,
			creators.first_name AS user_creator__first_name, creators.last_name AS user_creator__last_name,
			creators.group_id AS user_creator__group_id`)
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query, map[string]*service.FieldSortingParams{
		"order": {ColumnName: "groups_attempts.order", FieldType: "int64"},
		"id":    {ColumnName: "groups_attempts.id", FieldType: "int64"},
	}, "order")
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
