package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupRecentActivityResponseRow
type groupRecentActivityResponseRow struct {
	// `answers.id`
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	CreatedAt *database.Time `json:"created_at"`
	// Nullable
	// required: true
	Score *float32 `json:"score"`
	// required: true
	User struct {
		// required: true
		Login string `json:"login"`
		// Nullable
		// required: true
		FirstName *string `json:"first_name"`
		// Nullable
		// required: true
		LastName *string `json:"last_name"`
	} `json:"user" gorm:"embedded;embedded_prefix:user__"`
	// required: true
	Item struct {
		// required: true
		ID int64 `json:"id,string"`
		// required: true
		// enum: Chapter,Task,Course
		Type string `json:"type"`
		// required: true
		String struct {
			// Nullable
			// required: true
			Title *string `json:"title"`
		} `json:"string" gorm:"embedded;embedded_prefix:string__"`
	} `json:"item" gorm:"embedded;embedded_prefix:item__"`
}

// swagger:operation GET /groups/{group_id}/recent_activity groups groupRecentActivity
// ---
// summary: Get recent activity of a group
// description: >
//   Returns rows from `answers` with `type` = "Submission" and additional info on users and items.
//
//
//   If possible, items titles are shown in the authenticated user's default language.
//   Otherwise, the item's default language is used.
//
//
//   All rows of the result are related to users who are descendants of the `group_id` and items that are
//   descendants of `item_id` and visible to the authenticated user (at least 'info' access).
//
//
//   If the `validated` parameter is given, only `answers` with `score` = 100 (if `validated` = 1)
//   or with `score` != 100 (otherwise) are returned.
//
//
//   The authenticated user should be a manager of `group_id`, otherwise the 'forbidden' error is returned.
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
//   default: [-created_at,id]
//   type: array
//   items:
//     type: string
//     enum: [created_at,-created_at,id,-id]
// - name: from.created_at
//   description: Start the page from the row next to the row with `answers.created_at` = `from.created_at`
//                (`from.id` is required when `from.created_at` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the row next to the row with `answers.id`=`from.id`
//                (`from.created_at` is required when from.id is present)
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
//         "$ref": "#/definitions/groupRecentActivityResponseRow"
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

	if apiError := checkThatUserCanManageTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	itemDescendants := srv.Store.ItemAncestors().DescendantsOf(itemID).Select("child_item_id")
	query := srv.Store.Answers().WithUsers().WithItems().
		Joins("LEFT JOIN gradings ON gradings.answer_id = answers.id").
		Select(`
			answers.id as id, answers.created_at, gradings.score,
			items.id AS item__id, items.type AS item__type,
			users.login AS user__login,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.first_name, NULL) AS user__first_name,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.last_name, NULL) AS user__last_name,
			IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS item__string__title`,
			user.GroupID, user.GroupID).
		WithPersonalInfoViewApprovals(user).
		JoinsUserAndDefaultItemStrings(user).
		Where("attempts.item_id IN ?", itemDescendants.SubQuery()).
		Where("answers.type='Submission'").
		WhereItemsAreVisible(user.GroupID).
		WhereUsersAreDescendantsOfGroup(groupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query = srv.filterByValidated(r, query)

	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"created_at": {ColumnName: "answers.created_at", FieldType: "time"},
			"id":         {ColumnName: "answers.id", FieldType: "int64"}},
		"-created_at,id", []string{"id"}, false)
	if apiError != service.NoError {
		return apiError
	}

	var result []groupRecentActivityResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}

func (srv *Service) filterByValidated(r *http.Request, query *database.DB) *database.DB {
	validated, err := service.ResolveURLQueryGetBoolField(r, "validated")
	if err == nil {
		condition := "gradings.score "
		if !validated {
			condition += "IS NULL OR gradings.score !"
		}
		condition += "= 100"
		query = query.Where(condition)
	}
	return query
}
