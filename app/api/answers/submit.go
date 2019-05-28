package answers

import (
	"crypto/rsa"
	"encoding/json"
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
	requestData := SubmitRequest{PublicKey: srv.TokenConfig.PublicKey}

	var err error
	if err = render.Bind(httpReq, &requestData); err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	if err = user.Load(); err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if user.UserID != requestData.TaskToken.Converted.UserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token doesn't correspond to user session: got idUser=%d, expected %d",
			requestData.TaskToken.Converted.UserID, user.UserID))
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
		hasAccess, reason, err = store.Items().CheckSubmissionRights(requestData.TaskToken.Converted.LocalItemID, user)
		service.MustNotBeError(err)

		if !hasAccess {
			apiError = service.ErrForbidden(reason)
			return nil // commit! (CheckSubmissionRights() changes the DB sometimes)
		}

		userItemStore := store.UserItems()
		err = userItemStore.CreateIfMissing(user.UserID, requestData.TaskToken.Converted.LocalItemID)
		service.MustNotBeError(err)

		userAnswerID, err = store.UserAnswers().SubmitNewAnswer(
			user.UserID, requestData.TaskToken.Converted.LocalItemID, requestData.TaskToken.Converted.AttemptID, *requestData.Answer)
		service.MustNotBeError(err)

		scope := userItemStore.Where("idUser = ? AND idItem = ?", user.UserID, requestData.TaskToken.LocalItemID)
		service.MustNotBeError(scope.WithWriteLock().Select("sHintsRequested, nbHintsCached").Scan(&hintsInfo).Error())
		columnsToUpdate := map[string]interface{}{
			"nbSubmissionsAttempts": gorm.Expr("nbSubmissionsAttempts + 1"),
			"sLastActivityDate":     gorm.Expr("NOW()"),
		}
		service.MustNotBeError(scope.UpdateColumn(columnsToUpdate).Error())
		service.MustNotBeError(store.GroupAttempts().ByID(requestData.TaskToken.Converted.AttemptID).
			UpdateColumn(columnsToUpdate).Error())
		return nil // commit
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	answerToken, err := (&token.Answer{
		Answer:          *requestData.Answer,
		UserID:          requestData.TaskToken.UserID,
		ItemID:          requestData.TaskToken.ItemID,
		ItemURL:         requestData.TaskToken.ItemURL,
		LocalItemID:     requestData.TaskToken.LocalItemID,
		UserAnswerID:    strconv.FormatInt(userAnswerID, 10),
		RandomSeed:      requestData.TaskToken.RandomSeed,
		HintsRequested:  hintsInfo.HintsRequested,
		HintsGivenCount: strconv.FormatInt(int64(hintsInfo.HintsCached), 10),
		AttemptID:       requestData.TaskToken.AttemptID,
		PlatformName:    srv.TokenConfig.PlatformName,
	}).Sign(srv.TokenConfig.PrivateKey)
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(rw, httpReq, service.CreationSuccess(map[string]interface{}{
		"answer_token": answerToken,
	})))
	return service.NoError
}

// SubmitRequest represents a JSON request body format needed by answers.submit()
type SubmitRequest struct {
	TaskToken *token.Task `json:"task_token"`
	Answer    *string     `json:"answer"`

	PublicKey *rsa.PublicKey
}

type submitRequestWrapper struct {
	TaskToken *string `json:"task_token"`
	Answer    *string `json:"answer"`
}

// UnmarshalJSON loads SubmitRequest from JSON passing a public key into TaskToken
func (requestData *SubmitRequest) UnmarshalJSON(raw []byte) error {
	var wrapper submitRequestWrapper
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return err
	}
	if wrapper.TaskToken != nil {
		requestData.TaskToken = &token.Task{PublicKey: requestData.PublicKey}
		if err := requestData.TaskToken.UnmarshalString(*wrapper.TaskToken); err != nil {
			return fmt.Errorf("invalid task_token: %s", err.Error())
		}
	}
	requestData.Answer = wrapper.Answer
	return nil
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

	return nil
}

var (
	_ render.Binder = (*SubmitRequest)(nil)
)
