package answers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/types"
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

	attemptID := requestData.AttemptID.Value.(int64)
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
			"sState":  requestData.State.String.Value,
			"sAnswer": requestData.Answer.String.Value,
		}
		service.MustNotBeError(userAnswerStore.ByID(currentAnswerID).UpdateColumn(columnsToUpdate).Error())

		service.MustNotBeError(store.UserItems().Where("idUser = ?", user.UserID).
			Where("idItem = ?", itemID).
			Where("idAttemptActive = ?", requestData.AttemptID.Value).
			UpdateColumn(columnsToUpdate).Error())

		return nil
	})
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(rw, httpReq, service.UpdateSuccess(nil)))
	return service.NoError
}

type updateCurrentRequest struct {
	AttemptID types.RequiredInt64  `json:"attempt_id,string"`
	Answer    types.RequiredString `json:"answer"`
	State     types.RequiredString `json:"state"`
}

// Bind checks that all the needed request parameters are present
func (requestData *updateCurrentRequest) Bind(r *http.Request) error {
	return types.Validate([]string{"attempt_id", "answer", "state"},
		&requestData.AttemptID, &requestData.Answer, &requestData.State)
}
