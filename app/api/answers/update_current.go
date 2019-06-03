package answers

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) updateCurrent(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	var requestData updateCurrentRequest

	var err error
	if err = render.Bind(httpReq, &requestData); err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	if err = user.Load(); err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	foundItemID, itemID, err := srv.Store.GroupAttempts().GetAttemptItemIDIfUserHasAccess(*requestData.AttemptID, user)
	service.MustNotBeError(err)
	if !foundItemID {
		return service.InsufficientAccessRightsError
	}

	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		userAnswerStore := store.UserAnswers()
		var currentAnswerID int64
		currentAnswerID, err = userAnswerStore.GetOrCreateCurrentAnswer(user.UserID, itemID, requestData.AttemptID)
		service.MustNotBeError(err)

		columnsToUpdate := map[string]interface{}{
			"sState":  *requestData.State,
			"sAnswer": *requestData.Answer,
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

type updateCurrentRequest struct {
	AttemptID *int64  `json:"attempt_id,string"`
	Answer    *string `json:"answer"`
	State     *string `json:"state"`
}

// Bind checks that all the needed request parameters are present
func (requestData *updateCurrentRequest) Bind(r *http.Request) error {
	if requestData.AttemptID == nil {
		return errors.New("missing attempt_id")
	}

	if requestData.Answer == nil {
		return errors.New("missing answer")
	}

	if requestData.State == nil {
		return errors.New("missing state")
	}

	return nil
}
