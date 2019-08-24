package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id}/recent_activity groups users groupRecentActivity
// ---
// summary: Get recent activity of a group
// description: >
//   Returns rows from `users_answers` with `sType` = "Submission" and additional info on users and items.
//
//
//   If possible, items titles are shown in the authenticated user's default language.
//   Otherwise, the item's default language is used.
//
//
//   All rows of the result are related to users who are descendants of the `group_id` and items that are
//   descendants of `item_id` and visible to the authenticated user (at least grayed access).
//
//
//   If the `validated` parameter is true, only validated `users_answers` (with `bValidated`=1) are returned.
//
//
//   The authenticated user should be an owner of `group_id`, otherwise the 'forbidden' error is returned.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: item_id
//   in: query
//   type: integer
//   required: true
// - name: validated
//   in: query
//   type: boolean
//   default: false
// - name: sort
//   in: query
//   default: [-submission_date,id]
//   type: array
//   items:
//     type: string
//     enum: [submission_date,-submission_date,id,-id]
// - name: from.submission_date
//   description: Start the page from the row next to the row with `users_answers.sSubmissionDate` = `from.submission_date`
//                (`from.id` is required when `from.submission_date` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the row next to the row with `users_answers.ID`=`from.id`
//                (`from.submission_date` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display the first N rows
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. The array of users answers
//     schema:
//       type: array
//       items:
//         type: object
//         required: [id, submission_date, score, validated, user, item]
//         properties:
//           id:
//             description: "`users_answers.ID`"
//             type: string
//             format: int64
//           submission_date:
//             type: string
//             format: date-time
//           score:
//             description: Nullable
//             type: number
//             format: float
//           validated:
//             description: Nullable
//             type: boolean
//           user:
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
//           item:
//             type: object
//             required: [id, type, string]
//             properties:
//               id:
//                 type: string
//                 format: int64
//               type:
//                 type: string
//                 enum: [Root, Category, Chapter, Task, Course]
//               string:
//                 type: object
//                 required: [title]
//                 properties:
//                   title:
//                     description: Nullable
//                     type: string
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getRecentActivity(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryGetInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	itemDescendants := srv.Store.ItemAncestors().DescendantsOf(itemID).Select("idItemChild")
	query := srv.Store.UserAnswers().WithUsers().WithItems().
		Select(
			`users_answers.ID as ID, users_answers.sSubmissionDate, users_answers.bValidated, users_answers.iScore,
       items.ID AS Item__ID, items.sType AS Item__sType,
		   users.sLogin AS User__sLogin, users.sFirstName AS User__sFirstName, users.sLastName AS User__sLastName,
			 IF(user_strings.idLanguage IS NULL, default_strings.sTitle, user_strings.sTitle) AS Item__String__sTitle`).
		JoinsUserAndDefaultItemStrings(user).
		Where("users_answers.idItem IN ?", itemDescendants.SubQuery()).
		Where("users_answers.sType='Submission'").
		WhereItemsAreVisible(user).
		WhereUsersAreDescendantsOfGroup(groupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query = srv.filterByValidated(r, query)

	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"submission_date": {ColumnName: "users_answers.sSubmissionDate", FieldType: "time"},
			"id":              {ColumnName: "users_answers.ID", FieldType: "int64"}},
		"-submission_date")
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	if err := query.ScanIntoSliceOfMaps(&result).Error(); err != nil {
		return service.ErrUnexpected(err)
	}
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}

func (srv *Service) filterByValidated(r *http.Request, query *database.DB) *database.DB {
	validated, err := service.ResolveURLQueryGetBoolField(r, "validated")
	if err == nil {
		query = query.Where("users_answers.bValidated = ?", validated)
	}
	return query
}
