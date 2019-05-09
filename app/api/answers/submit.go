package answers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

func (srv *Service) submit(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	requestData := SubmitRequest{}

	var err error
	if err = render.Bind(httpReq, &requestData); err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	if err = user.Load(); err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if user.UserID != requestData.converted.UserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token doesn't correspond to user session: got idUser=%d, expected %d",
			requestData.converted.UserID, user.UserID))
	}

	var userAnswerID int64
	var hintsInfo struct {
		HintsRequested *string `gorm:"column:sHintsRequested"`
		HintsCached    int32   `gorm:"column:nbHintsCached"`
	}
	apiError := service.NoError

	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		var hasAccess bool
		var reason error
		hasAccess, reason, err = store.Items().CheckSubmissionRights(requestData.converted.ItemID, user)
		service.MustNotBeError(err)

		if !hasAccess {
			apiError = service.ErrForbidden(reason)
			return nil // commit! (CheckSubmissionRights() changes the DB sometimes)
		}

		userItemStore := store.UserItems()
		userItemStore.CreateIfMissing(user.UserID, requestData.converted.ItemID)

		userAnswerID, err = store.UserAnswers().SubmitNewAnswer(
			user.UserID, requestData.converted.ItemID, requestData.converted.AttemptID, *requestData.Answer)
		service.MustNotBeError(err)

		scope := userItemStore.Where("idUser = ? AND idItem = ?", user.UserID, requestData.TaskToken.LocalItemID)
		service.MustNotBeError(scope.WithWriteLock().Select("sHintsRequested, nbHintsCached").Scan(&hintsInfo).Error())
		service.MustNotBeError(scope.UpdateColumn(map[string]interface{}{
			"nbSubmissionsAttempts": gorm.Expr("nbSubmissionsAttempts + 1"),
		}).Error())
		return nil // commit
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(rw, httpReq, service.CreationSuccess(map[string]interface{}{
		"answer_token": &token.Answer{
			Answer:         *requestData.Answer,
			UserID:         requestData.TaskToken.UserID,
			ItemID:         requestData.TaskToken.ItemID,
			ItemURL:        requestData.TaskToken.ItemURL,
			LocalItemID:    requestData.TaskToken.LocalItemID,
			UserAnswerID:   strconv.FormatInt(userAnswerID, 10),
			RandomSeed:     requestData.TaskToken.RandomSeed,
			HintsRequested: hintsInfo.HintsRequested,
			HintsGiven:     strconv.FormatInt(int64(hintsInfo.HintsCached), 10),
			AttemptID:      requestData.TaskToken.AttemptID,
		},
	})))
	return service.NoError
}

// SubmitRequest represents a JSON request body format needed by answers.submit()
type SubmitRequest struct {
	TaskToken *token.Task `json:"task_token"`
	Answer    *string     `json:"answer"`

	converted struct {
		UserID    int64
		ItemID    int64
		AttemptID *int64
	}
}

// Bind checks that all the needed request parameters (task_token & answer) are present and
// all the needed values are valid.
func (requestData *SubmitRequest) Bind(r *http.Request) error {
	if requestData.TaskToken == nil {
		return errors.New("missing task_token")
	}

	if requestData.Answer == nil {
		return errors.New("missing answer")
	}

	var err error
	requestData.converted.UserID, err = strconv.ParseInt(requestData.TaskToken.UserID, 10, 64)
	if err != nil {
		return errors.New("wrong idUser in the token")
	}

	requestData.converted.ItemID, err = strconv.ParseInt(requestData.TaskToken.LocalItemID, 10, 64)
	if err != nil {
		return errors.New("wrong idItemLocal in the token")
	}

	if requestData.TaskToken.AttemptID != nil {
		var attemptIDValue int64
		attemptIDValue, err = strconv.ParseInt(*requestData.TaskToken.AttemptID, 10, 64)
		if err != nil {
			return errors.New("wrong idAttempt in the token")
		}
		requestData.converted.AttemptID = &attemptIDValue
	}
	return nil
}

var (
	_ render.Binder = (*SubmitRequest)(nil)
)
