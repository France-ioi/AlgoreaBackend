package answers

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

// swagger:operation GET /items/{item_id}/answers answers answersList
//
//	---
//	summary: List answers
//	description: >
//		Return answers (i.e., saved answers, current answer and submissions)
//		for a given item and user, or from a given attempt.
//
//		* One of `author_id` or `attempt_id` is required.
//
//		* The user should have at least 'content' access to the item.
//
//		* If `author_id` is given, the authenticated user should be the input `author_id`
//			or a manager of a group containing the input `author_id`.
//
//		* If `attempt_id` is given, the authenticated user should be a member
//		or a manager of the group attached to the attempt.
//
//
//		Users' `first_name` and `last_name` are only shown for the authenticated user or if the user
//		approved access to their personal info for some group managed by the authenticated user.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: author_id
//			in: query
//			type: integer
//			format: int64
//		- name: attempt_id
//			in: query
//			type: integer
//			format: int64
//		- name: sort
//			in: query
//			default: [-created_at,id]
//			type: array
//			items:
//				type: string
//				enum: [created_at,-created_at,id,-id]
//		- name: from.id
//			description: Start the page from the answer next to the answer with `answers.id`=`{from.id}`
//			in: query
//			type: integer
//			format: int64
//		- name: limit
//			description: Display the first N answers
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. Success response with an array of answers
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/answersResponseAnswer"
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
func (srv *Service) listAnswers(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	itemID, itemIDError := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if itemIDError != nil {
		return service.ErrInvalidRequest(itemIDError)
	}

	found, err := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("view", "content").
		Where("item_id = ?", itemID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.ErrAPIInsufficientAccessRights
	}

	dataQuery := store.Answers().WithUsers().WithResults().
		Joins("LEFT JOIN gradings ON gradings.answer_id = answers.id").
		Select(`
			answers.id, answers.type, answers.created_at, gradings.score,
			users.login,
			users.group_id = ? OR personal_info_view_approvals.approved AS show_personal_info,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.first_name, NULL) AS first_name,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.last_name, NULL) AS last_name`,
			user.GroupID, user.GroupID, user.GroupID).
		Where("answers.item_id = ?", itemID).
		WithPersonalInfoViewApprovals(user)

	authorIDIsSet := len(httpRequest.URL.Query()["author_id"]) > 0

	if !authorIDIsSet { // attempt_id
		attemptIDIsSet := len(httpRequest.URL.Query()["attempt_id"]) > 0
		if !attemptIDIsSet {
			return service.ErrInvalidRequest(errors.New("either author_id or attempt_id must be present"))
		}

		attemptID, attemptIDError := service.ResolveURLQueryGetInt64Field(httpRequest, "attempt_id")
		if attemptIDError != nil {
			return service.ErrInvalidRequest(attemptIDError)
		}

		service.MustNotBeError(srv.checkAccessRightsForGetAnswersByAttemptID(store, attemptID, user))

		dataQuery = dataQuery.Where("answers.attempt_id = ?", attemptID)
	} else { // author_id
		authorID, authorIDError := service.ResolveURLQueryGetInt64Field(httpRequest, "author_id")
		if authorIDError != nil {
			return service.ErrInvalidRequest(authorIDError)
		}

		service.MustNotBeError(srv.checkAccessRightsForGetAnswersByAuthorID(store, authorID, user))

		dataQuery = dataQuery.Where("author_id = ?", authorID)
	}

	dataQuery, err = service.ApplySortingAndPaging(
		httpRequest, dataQuery,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"created_at": {ColumnName: "answers.created_at"},
				"id":         {ColumnName: "answers.id"},
			},
			DefaultRules: "-created_at,id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	service.MustNotBeError(err)

	dataQuery = service.NewQueryLimiter().Apply(httpRequest, dataQuery)

	var result []rawAnswersData
	service.MustNotBeError(dataQuery.Scan(&result).Error())

	responseData := srv.convertDBDataToResponse(result)

	render.Respond(responseWriter, httpRequest, responseData)
	return nil
}

// swagger:ignore
type rawAnswersData struct {
	ID               int64
	Type             string
	CreatedAt        database.Time
	Score            *float32
	UserLogin        string  `sql:"column:login"`
	UserFirstName    *string `sql:"column:first_name"`
	UserLastName     *string `sql:"column:last_name"`
	ShowPersonalInfo bool
}

type answersResponseAnswerUser struct {
	// required: true
	Login string `json:"login"`
	*structures.UserPersonalInfo
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
	// required: true
	Score *float32 `json:"score"`

	// required: true
	User answersResponseAnswerUser `json:"user"`
}

func (srv *Service) convertDBDataToResponse(rawData []rawAnswersData) (response *[]answersResponseAnswer) {
	responseData := make([]answersResponseAnswer, 0, len(rawData))
	for _, row := range rawData {
		responseDataRow := answersResponseAnswer{
			ID:        row.ID,
			Type:      row.Type,
			CreatedAt: row.CreatedAt,
			Score:     row.Score,
			User: answersResponseAnswerUser{
				Login: row.UserLogin,
			},
		}
		if row.ShowPersonalInfo {
			responseDataRow.User.UserPersonalInfo = &structures.UserPersonalInfo{
				FirstName: row.UserFirstName,
				LastName:  row.UserLastName,
			}
		}
		responseData = append(responseData, responseDataRow)
	}
	return &responseData
}

func (srv *Service) checkAccessRightsForGetAnswersByAttemptID(
	store *database.DataStore, attemptID int64, user *database.User,
) error {
	var count int64
	groupsManagedByUser := store.GroupAncestors().ManagedByUser(user).Select("groups_ancestors.child_group_id")
	groupsWhereUserIsMember := store.GroupGroups().WhereUserIsMember(user).Select("parent_group_id")

	service.MustNotBeError(store.Attempts().ByID(attemptID).
		Where("(attempts.participant_id IN ?) OR (attempts.participant_id IN ?) OR attempts.participant_id = ?",
			groupsManagedByUser.SubQuery(),
			groupsWhereUserIsMember.SubQuery(),
			user.GroupID).
		Count(&count).Error())
	if count == 0 {
		return service.ErrAPIInsufficientAccessRights
	}
	return nil
}

func (srv *Service) checkAccessRightsForGetAnswersByAuthorID(
	store *database.DataStore, authorID int64, user *database.User,
) error {
	if authorID != user.GroupID {
		found, err := store.GroupAncestors().ManagedByUser(user).
			Where("groups_ancestors.child_group_id=?", authorID).HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrAPIInsufficientAccessRights
		}
	}

	return nil
}
