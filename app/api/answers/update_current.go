package answers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type updateCurrentRequest struct {
	AttemptID int64  `json:"attempt_id,string" validate:"required"`
	Answer    string `json:"answer" validate:"required"`
	State     string `json:"state" validate:"required"`
}

func (srv *Service) updateCurrent(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	var requestData updateCurrentRequest

	formData := formdata.NewFormData(&requestData)
	err := formData.ParseJSONRequestData(httpReq)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	if err = user.Load(); err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	attemptID := requestData.AttemptID
	found, itemID, err := srv.Store.GroupAttempts().GetAttemptItemIDIfUserHasAccess(attemptID, user)
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		userAnswerStore := store.UserAnswers()
		var currentAnswerID int64
		currentAnswerID, err = userAnswerStore.GetOrCreateCurrentAnswer(user.UserID, itemID, &attemptID)
		service.MustNotBeError(err)

		columnsToUpdate := map[string]interface{}{
			"sState":  requestData.State,
			"sAnswer": requestData.Answer,
		}
		service.MustNotBeError(userAnswerStore.ByID(currentAnswerID).UpdateColumn(columnsToUpdate).Error())

		service.MustNotBeError(store.UserItems().Where("idUser = ?", user.UserID).
			Where("idItem = ?", itemID).
			Where("idAttemptActive = ?", requestData.AttemptID).
			UpdateColumn(columnsToUpdate).Error())

		return nil
	})
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(rw, httpReq, service.UpdateSuccess(nil)))
	return service.NoError
}
