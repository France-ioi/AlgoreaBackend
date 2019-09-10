package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /items/{item_id}/attempts groups users attempts items itemAttemptsView
// ---
// summary: List attempts for a task
// description: Returns attempts made by the current user (if `items.bHasAttempts` = 0) or his
//              teams (if `items.bHasAttempts` = 1) while solving the task given in `item_id`.
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
//   description: Start the page from the attempt next to the attempt with `groups_attempts.iOrder` = `from.order` and
//                `groups_attempts.ID` = `from.id` (`from.id` is required when `from.order` is present)
//   in: query
//   type: integer
//   format: int32
// - name: from.id
//   description: Start the page from the attempt next to the attempt with `groups_attempts.iOrder` = `from.order` and
//                `groups_attempts.ID` = `from.id` (`from.order` is required when `from.id` is present)
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
//         type: object
//         properties:
//           id:
//             type: string
//             format: int64
//           order:
//             type: integer
//             format: int32
//           score:
//             type: number
//             format: float
//           validated:
//             type: boolean
//           start_date:
//             description: Nullable
//             type: string
//             format: date-time
//           user_creator:
//             description: Nullable
//             type: object
//             required: [login, first_name, last_name]
//             properties:
//               login:
//                 type: string
//               first_name:
//                 description: Nullable
//                 type: string
//               last_name:
//                 description: Nullable
//                 type: string
//         required: [id, order, score, validated, start_date, user_creator]
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
		Joins("LEFT JOIN users AS creators ON creators.ID = groups_attempts.idUserCreator").
		Select(`
			groups_attempts.ID, groups_attempts.iOrder, groups_attempts.iScore, groups_attempts.bValidated,
			groups_attempts.sStartDate, creators.sLogin AS userCreator__sLogin,
			creators.sFirstName AS userCreator__sFirstName, creators.sLastName AS userCreator__sLastName`)
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query, map[string]*service.FieldSortingParams{
		"order": {ColumnName: "groups_attempts.iOrder", FieldType: "int64"},
		"id":    {ColumnName: "groups_attempts.ID", FieldType: "int64"},
	}, "order")
	if apiError != service.NoError {
		return apiError
	}
	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}
