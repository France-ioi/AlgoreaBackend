package answers

import (
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"net/http"
	"strconv"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getAnswers(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	userID, err := resolveURLQueryGetInt64Field(httpReq, "user_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	itemID, err := resolveURLQueryGetInt64Field(httpReq, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)

	if userID != user.UserID {
		count := 0
		givenUserSelfGroup := srv.Store.Users().ByID(userID).Select("idGroupSelf").SubQuery()
		if err := srv.Store.GroupAncestors().OwnedByUser(user).
			Where("idGroupChild=?", givenUserSelfGroup).
			Count(&count).Error();
			err != nil {
				return service.ErrUnexpected(err)
			}
		if count == 0 {
			return service.ErrForbidden(errors.New("insufficient access rights"))
		}
	}

	accessDetailsMap, err := srv.Store.Items().GetAccessDetailsMapForIDs(user, []int64{itemID})
	if err != nil {
		return service.ErrUnexpected(err)
	}

	accessDetails, ok := accessDetailsMap[itemID]
	if !ok || (!accessDetails.FullAccess && !accessDetails.PartialAccess && !accessDetails.GrayedAccess) {
		return service.ErrNotFound(errors.New("insufficient access rights on the given item id"))
	}

	if !accessDetails.FullAccess && !accessDetails.PartialAccess {
		return service.ErrForbidden(errors.New("insufficient access rights on the given item id"))
	}

	type dbData struct{
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

	var result []dbData

	if err := srv.Store.UserAnswers().WithUsers().
		Select(`users_answers.ID, users_answers.sName, users_answers.sType, users_answers.sLangProg,
            users_answers.sSubmissionDate, users_answers.iScore, users_answers.bValidated,
            users.sLogin, users.sFirstName, users.sLastName`).
		Where("idItem=? AND idUser=?", itemID, userID).Order("sSubmissionDate DESC").Scan(&result).Error();
	  err != nil {
		return service.ErrUnexpected(err)
	}

	type responseAnswerUser struct{
		Login          string  `json:"login"`
		FirstName      *string `json:"first_name,omitempty"`
		LastName       *string `json:"last_name,omitempty"`
	}

	type responseAnswer struct{
		ID             int64    `json:"id"`
		Name           *string  `json:"name,omitempty"`
		Type           string   `json:"type"`
		LangProg       *string  `json:"lang_prog,omitempty"`
		SubmissionDate string   `json:"submission_date"`
		Score          *float32 `json:"score,omitempty"`
		Validated      *bool    `json:"validated,omitempty"`

		User           responseAnswerUser `json:"user"`
	}

	type response struct {
		Answers        []responseAnswer `json:"answers"`
	}

	responseData := response{Answers:make([]responseAnswer, 0, len(result))}
	for _, row := range result {
		fmt.Printf("%#v", row)
		responseData.Answers = append(responseData.Answers, responseAnswer{
			ID:             row.ID,
			Name:           row.Name,
			Type:           row.Type,
			LangProg:       row.LangProg,
			SubmissionDate: row.SubmissionDate,
			Score:          row.Score,
			Validated:      row.Validated,
			User:           responseAnswerUser{
				Login: row.UserLogin,
				FirstName: row.UserFirstName,
				LastName: row.UserLastName,
			},
		})
	}
	fmt.Printf("%#v", responseData)
	render.Respond(rw, httpReq, responseData)
	return service.NoError
}

func resolveURLQueryGetInt64Field(httpReq *http.Request, name string) (int64, error) {
	strValue := httpReq.URL.Query().Get(name)
	int64Value, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("missing %s", name)
	}
	return int64Value, nil
}
