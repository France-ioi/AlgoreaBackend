package answers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /answers answers itemAnswersView
// ---
// summary: List answers
// description: Return answers (i.e., history of submissions and current answer)
//   for a given item and user, or from a given attempt.
//
//   * One of (`author_id`, `item_id`) pair or `attempt_id` is required.
//
//   * The user should have at least 'content' access to the item.
//
//   * If `item_id` and `author_id` are given, the authenticated user should have `group_id` equal to the input `author_id`
//   or be an owner of a group containing the input `author_id`.
//
//   * If `attempt_id` is given, the authenticated user should be a member of the group
//   or an owner of the group attached to the attempt.
// parameters:
// - name: author_id
//   in: query
//   type: integer
// - name: item_id
//   in: query
//   type: integer
// - name: attempt_id
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [-created_at,id]
//   type: array
//   items:
//     type: string
//     enum: [created_at,-created_at,id,-id]
// - name: from.created_at
//   description: Start the page from the answer next to the answer with `created_at` = `from.created_at`
//                and `answers.id` = `from.id`
//                (`from.id` is required when `from.created_at` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the answer next to the answer with `created_at`=`from.created_at`
//                and `answers.id`=`from.id`
//                (`from.created_at` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display the first N answers
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with an array of answers
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/answersResponseAnswer"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "404":
//     "$ref": "#/responses/notFoundResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getAnswers(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	user := srv.GetUser(httpReq)

	dataQuery := srv.Store.Answers().WithUsers().WithGroupAttempts().
		Joins("LEFT JOIN gradings ON gradings.answer_id = answers.id").
		Select(`answers.id, answers.type, answers.created_at, gradings.score,
		        users.login, users.first_name, users.last_name`)

	authorID, authorIDError := service.ResolveURLQueryGetInt64Field(httpReq, "author_id")
	itemID, itemIDError := service.ResolveURLQueryGetInt64Field(httpReq, "item_id")

	if authorIDError != nil || itemIDError != nil { // attempt_id
		attemptID, attemptIDError := service.ResolveURLQueryGetInt64Field(httpReq, "attempt_id")
		if attemptIDError != nil {
			return service.ErrInvalidRequest(fmt.Errorf("either author_id & item_id or attempt_id must be present"))
		}

		if result := srv.checkAccessRightsForGetAnswersByAttemptID(attemptID, user); result != service.NoError {
			return result
		}

		dataQuery = dataQuery.Where("attempt_id = ?", attemptID)
	} else { // author_id + item_id
		if result := srv.checkAccessRightsForGetAnswersByAuthorIDAndItemID(authorID, itemID, user); result != service.NoError {
			return result
		}

		dataQuery = dataQuery.Where("item_id = ? AND author_id = ?", itemID, authorID)
	}

	dataQuery, apiError := service.ApplySortingAndPaging(httpReq, dataQuery, map[string]*service.FieldSortingParams{
		"created_at": {ColumnName: "answers.created_at", FieldType: "time"},
		"id":         {ColumnName: "answers.id", FieldType: "int64"},
	}, "-created_at,id", "id", false)
	if apiError != service.NoError {
		return apiError
	}
	dataQuery = service.NewQueryLimiter().Apply(httpReq, dataQuery)

	var result []rawAnswersData
	service.MustNotBeError(dataQuery.Scan(&result).Error())

	responseData := srv.convertDBDataToResponse(result)

	render.Respond(rw, httpReq, responseData)
	return service.NoError
}

// swagger:ignore
type rawAnswersData struct {
	ID            int64
	Type          string
	CreatedAt     database.Time
	Score         *float32
	UserLogin     string  `sql:"column:login"`
	UserFirstName *string `sql:"column:first_name"`
	UserLastName  *string `sql:"column:last_name"`
}

type answersResponseAnswerUser struct {
	// required: true
	Login string `json:"login"`
	// Nullable
	// required: true
	FirstName *string `json:"first_name"`
	// Nullable
	// required: true
	LastName *string `json:"last_name"`
}

// swagger:model
type answersResponseAnswer struct {
	// `answers.id`
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	// enum: Submission,Saved,Current
	Type string `json:"type"`
	// required: true
	CreatedAt database.Time `json:"created_at"`
	// Nullable
	// required: true
	Score *float32 `json:"score"`

	// required: true
	User answersResponseAnswerUser `json:"user"`
}

func (srv *Service) convertDBDataToResponse(rawData []rawAnswersData) (response *[]answersResponseAnswer) {
	responseData := make([]answersResponseAnswer, 0, len(rawData))
	for _, row := range rawData {
		responseData = append(responseData, answersResponseAnswer{
			ID:        row.ID,
			Type:      row.Type,
			CreatedAt: row.CreatedAt,
			Score:     row.Score,
			User: answersResponseAnswerUser{
				Login:     row.UserLogin,
				FirstName: row.UserFirstName,
				LastName:  row.UserLastName,
			},
		})
	}
	return &responseData
}

func (srv *Service) checkAccessRightsForGetAnswersByAttemptID(attemptID int64, user *database.User) service.APIError {
	var count int64
	itemsUserCanAccess := srv.Store.Permissions().WithViewPermissionForUser(user, "content")

	groupsManagedByUser := srv.Store.GroupAncestors().ManagedByUser(user).Select("groups_ancestors.child_group_id")
	groupsWhereUserIsMember := srv.Store.GroupGroups().WhereUserIsMember(user).Select("parent_group_id")

	service.MustNotBeError(srv.Store.GroupAttempts().ByID(attemptID).
		Joins("JOIN ? rights ON rights.item_id = groups_attempts.item_id", itemsUserCanAccess.SubQuery()).
		Where("(groups_attempts.group_id IN ?) OR (groups_attempts.group_id IN ?) OR groups_attempts.group_id = ?",
			groupsManagedByUser.SubQuery(),
			groupsWhereUserIsMember.SubQuery(),
			user.GroupID).
		Count(&count).Error())
	if count == 0 {
		return service.InsufficientAccessRightsError
	}
	return service.NoError
}

func (srv *Service) checkAccessRightsForGetAnswersByAuthorIDAndItemID(authorID, itemID int64, user *database.User) service.APIError {
	if authorID != user.GroupID {
		count := 0
		err := srv.Store.GroupAncestors().ManagedByUser(user).
			Where("groups_ancestors.child_group_id=?", authorID).
			Count(&count).Error()
		service.MustNotBeError(err)
		if count == 0 {
			return service.InsufficientAccessRightsError
		}
	}

	accessDetails, err := srv.Store.Items().GetAccessDetailsForIDs(user, []int64{itemID})
	service.MustNotBeError(err)

	if len(accessDetails) == 0 || accessDetails[0].IsForbidden() {
		return service.ErrNotFound(errors.New("insufficient access rights on the given item id"))
	}

	if accessDetails[0].IsInfo() {
		return service.ErrForbidden(errors.New("insufficient access rights on the given item id"))
	}

	return service.NoError
}
