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
	"github.com/France-ioi/AlgoreaBackend/app/doc"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// swagger:operation POST /answers items itemGetAnswerToken
// ---
// summary: Generate an answer token
// description: Generate and return an answer token from user s answer and task token.
//   It is used to bind an answer with task parameters so that the TaskGrader can check if they have not been altered.
//
//   * task_token.idUser should be the current user
//
//   * The user should have submission rights on `task_token.idItemLocal`
// parameters:
// - name: answer information
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/submitRequestWrapper"
// responses:
//   "201":
//     description: "Created. Success response with answer_token"
//     in: body
//     schema:
//       "$ref": "#/definitions/answerSubmitResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) submit(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	requestData := SubmitRequest{PublicKey: srv.TokenConfig.PublicKey}

	var err error
	if err = render.Bind(httpReq, &requestData); err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)

	if user.GroupID != requestData.TaskToken.Converted.UserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token doesn't correspond to user session: got idUser=%d, expected %d",
			requestData.TaskToken.Converted.UserID, user.GroupID))
	}

	var userAnswerID int64
	var hintsInfo struct {
		HintsRequested *string
		HintsCached    int32
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

		userAnswerID, err = store.UserAnswers().SubmitNewAnswer(
			user.GroupID, requestData.TaskToken.Converted.AttemptID, *requestData.Answer)
		service.MustNotBeError(err)

		groupAttemptsScope := store.GroupAttempts().ByID(requestData.TaskToken.Converted.AttemptID)
		service.MustNotBeError(
			groupAttemptsScope.WithWriteLock().Select("hints_requested, hints_cached").Scan(&hintsInfo).Error())

		return groupAttemptsScope.UpdateColumn(map[string]interface{}{
			"submissions":        gorm.Expr("submissions + 1"),
			"latest_activity_at": database.Now(),
		}).Error()
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
// swagger:ignore
type SubmitRequest struct {
	TaskToken *token.Task `json:"task_token"`
	Answer    *string     `json:"answer"`

	PublicKey *rsa.PublicKey
}

// swagger:model
type submitRequestWrapper struct {
	// required:true
	TaskToken *string `json:"task_token"`
	// required:true
	Answer *string `json:"answer"`
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

// Created. Success response with answer_token
// swagger:model answerSubmitResponse
type answerSubmitResponse struct { // nolint:unused,deadcode
	// description
	// swagger:allOf
	doc.CreatedResponse
	// required:true
	Data struct {
		AnswerToken string `json:"answer_token"`
	} `json:"data"`
}
