package answers

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) get(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	userAnswerID, err := service.ResolveURLQueryPathInt64Field(httpReq, "answer_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	var result []map[string]interface{}
	err = srv.Store.UserAnswers().Visible(user).
		Where("users_answers.ID = ?", userAnswerID).
		Select(`users_answers.ID, users_answers.idUser, users_answers.idItem, users_answers.idAttempt,
			users_answers.sType, users_answers.sState, users_answers.sAnswer,
			users_answers.sSubmissionDate, users_answers.iScore, users_answers.bValidated,
			users_answers.sGradingDate, users_answers.idUserGrader`).
		ScanIntoSliceOfMaps(&result).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)
	if len(result) == 0 {
		return service.InsufficientAccessRightsError
	}
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)[0]

	render.Respond(rw, httpReq, convertedResult)
	return service.NoError
}
