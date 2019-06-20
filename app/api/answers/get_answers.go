package answers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /answers answers users attempts items itemAnswersView
// ---
// summary: View answers history
// description: Display the history of answers (submissions, saved and current)
//   for a given item and user, or from a given attempt.
//
//   * One of (user_id, item_id) pair or attempt_id is required.
//
//   * The user should have at least partial access to the item.
//
//   * If item_id and user_id are given, the authenticated user should be either the input user
//   or an owner of a group containing the selfGroup of the input user.
//
//   * If attempt_id is given, the authenticated user should be a member of the group
//   or an owner of the group attached to the attempt.
// parameters:
// - name: user_id
//   in: query
//   type: integer
// - name: item_id
//   in: query
//   type: integer
// - name: attempt_id
//   in: query
//   type: integer
// responses:
//   "200":
//     "$ref": "#/responses/itemAnswersViewResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getAnswers(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	user := srv.GetUser(httpReq)

	dataQuery := srv.Store.UserAnswers().WithUsers().
		Select(`users_answers.ID, users_answers.sName, users_answers.sType, users_answers.sLangProg,
            users_answers.sSubmissionDate, users_answers.iScore, users_answers.bValidated,
            users.sLogin, users.sFirstName, users.sLastName`).
		Order("sSubmissionDate DESC")

	userID, userIDError := service.ResolveURLQueryGetInt64Field(httpReq, "user_id")
	itemID, itemIDError := service.ResolveURLQueryGetInt64Field(httpReq, "item_id")

	if userIDError != nil || itemIDError != nil { // attempt_id
		attemptID, attemptIDError := service.ResolveURLQueryGetInt64Field(httpReq, "attempt_id")
		if attemptIDError != nil {
			return service.ErrInvalidRequest(fmt.Errorf("either user_id & item_id or attempt_id must be present"))
		}

		if result := srv.checkAccessRightsForGetAnswersByAttemptID(attemptID, user); result != service.NoError {
			return result
		}

		// we should create an index on `users_answers`.`idAttempt` for this query
		dataQuery = dataQuery.Where("idAttempt = ?", attemptID)
	} else { // user_id + item_id
		if result := srv.checkAccessRightsForGetAnswersByUserIDAndItemID(userID, itemID, user); result != service.NoError {
			return result
		}

		dataQuery = dataQuery.Where("idItem = ? AND idUser = ?", itemID, userID)
	}

	var result []rawAnswersData
	if err := dataQuery.Scan(&result).Error(); err != nil {
		return service.ErrUnexpected(err)
	}

	responseData := srv.convertDBDataToResponse(result)

	render.Respond(rw, httpReq, responseData)
	return service.NoError
}

type rawAnswersData struct {
	ID             int64    `sql:"column:ID"`
	Name           *string  `sql:"column:sName"`
	Type           string   `sql:"column:sType"`
	LangProg       *string  `sql:"column:sLangProg"`
	SubmissionDate string   `sql:"column:sSubmissionDate"`
	Score          *float32 `sql:"column:iScore"`
	Validated      *bool    `sql:"column:bValidated"`
	UserLogin      string   `sql:"column:sLogin"`
	UserFirstName  *string  `sql:"column:sFirstName"`
	UserLastName   *string  `sql:"column:sLastName"`
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

type answersResponseAnswer struct {
	// required: true
	ID int64 `json:"id,string"`
	// Nullable
	// required: true
	Name *string `json:"name"`
	// required: true
	// enum: Submission,Saved,Current
	Type string `json:"type"`
	// Nullable
	// required: true
	LangProg *string `json:"lang_prog"`
	// required: true
	SubmissionDate string `json:"submission_date"`
	// Nullable
	// required: true
	Score *float32 `json:"score"`
	// Nullable
	// required: true
	Validated *bool `json:"validated"`

	// required: true
	User answersResponseAnswerUser `json:"user"`
}

// OK. Success response of the itemAnswersView service
// swagger:response itemAnswersViewResponse
type itemAnswersViewResponse struct { // nolint:unused,deadcode
	// description: The returned answers
	// in:body
	Answers []answersResponseAnswer
}

func (srv *Service) convertDBDataToResponse(rawData []rawAnswersData) (response *[]answersResponseAnswer) {
	responseData := make([]answersResponseAnswer, 0, len(rawData))
	for _, row := range rawData {
		responseData = append(responseData, answersResponseAnswer{
			ID:             row.ID,
			Name:           row.Name,
			Type:           row.Type,
			LangProg:       row.LangProg,
			SubmissionDate: row.SubmissionDate,
			Score:          row.Score,
			Validated:      row.Validated,
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
	itemsUserCanAccess := srv.Store.Items().AccessRights(user).
		Having("fullAccess>0 OR partialAccess>0")
	if itemsUserCanAccess.Error() == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(itemsUserCanAccess.Error())

	groupsOwnedByUser := srv.Store.GroupAncestors().OwnedByUser(user).Select("idGroupChild")
	service.MustNotBeError(groupsOwnedByUser.Error())

	groupsWhereUserIsMember := srv.Store.GroupGroups().WhereUserIsMember(user).Select("idGroupParent")
	service.MustNotBeError(groupsWhereUserIsMember.Error())

	userSelfGroupID, _ := user.SelfGroupID()
	service.MustNotBeError(srv.Store.GroupAttempts().ByID(attemptID).
		Joins("JOIN ? rights ON rights.idItem = groups_attempts.idItem", itemsUserCanAccess.SubQuery()).
		Where("(groups_attempts.idGroup IN ?) OR (groups_attempts.idGroup IN ?) OR groups_attempts.idGroup = ?",
			groupsOwnedByUser.SubQuery(),
			groupsWhereUserIsMember.SubQuery(),
			userSelfGroupID).
		Count(&count).Error())
	if count == 0 {
		return service.InsufficientAccessRightsError
	}
	return service.NoError
}

func (srv *Service) checkAccessRightsForGetAnswersByUserIDAndItemID(userID, itemID int64, user *database.User) service.APIError {
	if userID != user.UserID {
		count := 0
		givenUserSelfGroup := srv.Store.Users().ByID(userID).Select("idGroupSelf")
		service.MustNotBeError(givenUserSelfGroup.Error())
		err := srv.Store.GroupAncestors().OwnedByUser(user).
			Where("idGroupChild=?", givenUserSelfGroup.SubQuery()).
			Count(&count).Error()
		if err == database.ErrUserNotFound {
			return service.InsufficientAccessRightsError
		}
		service.MustNotBeError(err)
		if count == 0 {
			return service.InsufficientAccessRightsError
		}
	}

	accessDetails, err := srv.Store.Items().GetAccessDetailsForIDs(user, []int64{itemID})
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if len(accessDetails) == 0 || accessDetails[0].IsForbidden() {
		return service.ErrNotFound(errors.New("insufficient access rights on the given item id"))
	}

	if accessDetails[0].IsGrayed() {
		return service.ErrForbidden(errors.New("insufficient access rights on the given item id"))
	}

	return service.NoError
}
