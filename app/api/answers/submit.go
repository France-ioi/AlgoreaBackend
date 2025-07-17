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

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/doc"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

// SubmitRequest represents a JSON request body format needed by answers.submit()
// swagger:ignore
type SubmitRequest struct {
	TaskToken *token.Token[payloads.TaskToken] `json:"task_token"`
	Answer    *string                          `json:"answer"`

	PublicKey *rsa.PublicKey
}

// swagger:model
type submitRequestWrapper struct {
	// required:true
	TaskToken *string `json:"task_token"`
	// required:true
	Answer *string `json:"answer"`
}

// Created. Success response with answer_token
// swagger:model answerSubmitResponse
type answerSubmitResponse struct { //nolint:unused
	// description
	// swagger:allOf
	doc.CreatedResponse
	// required:true
	Data struct {
		AnswerToken string `json:"answer_token"`
	} `json:"data"`
}

// swagger:operation POST /answers answers itemGetAnswerToken
//
//	---
//	summary: Generate an answer token
//	description: >
//		Generate and return an answer token from user's answer and task token.
//		It is used to bind an answer with task parameters so that the TaskGrader can check if they have not been altered.
//
//
//		This service doesn't require authentication. The user is identified by the task token.
//
//
//		* The task token's user should have submission rights on `task_token.idItemLocal`.
//
//		* The attempt should allow submission (`attempts.allows_submissions_until` should be a time in the future).
//
//		If any of the preconditions fails, the 'forbidden' error is returned.
//	parameters:
//		- name: answer information
//			in: body
//			required: true
//			schema:
//				"$ref": "#/definitions/submitRequestWrapper"
//	responses:
//		"201":
//			description: "Created. Success response with answer_token"
//			in: body
//			schema:
//				"$ref": "#/definitions/answerSubmitResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) submit(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	requestData := SubmitRequest{PublicKey: srv.TokenConfig.PublicKey}

	var err error
	if err = render.Bind(httpRequest, &requestData); err != nil {
		return service.ErrInvalidRequest(err)
	}

	var answerID int64
	var hintsInfo *database.HintsInfo

	logging.LogEntrySetField(httpRequest, "user_id", requestData.TaskToken.Payload.Converted.UserID)

	err = srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
		var hasAccess bool
		var reason error
		hasAccess, reason, err = store.Items().
			CheckSubmissionRights(requestData.TaskToken.Payload.Converted.ParticipantID, requestData.TaskToken.Payload.Converted.LocalItemID)
		service.MustNotBeError(err)

		if !hasAccess {
			return service.ErrForbidden(reason) // rollback
		}

		hintsInfo, err = store.Results().GetHintsInfoForActiveAttempt(
			requestData.TaskToken.Payload.Converted.ParticipantID,
			requestData.TaskToken.Payload.Converted.AttemptID,
			requestData.TaskToken.Payload.Converted.LocalItemID)

		if gorm.IsRecordNotFoundError(err) {
			return service.ErrForbidden(errors.New("no active attempt found")) // rollback
		}
		service.MustNotBeError(err)

		answerID, err = store.Answers().SubmitNewAnswer(
			requestData.TaskToken.Payload.Converted.UserID,
			requestData.TaskToken.Payload.Converted.ParticipantID,
			requestData.TaskToken.Payload.Converted.AttemptID,
			requestData.TaskToken.Payload.Converted.LocalItemID,
			*requestData.Answer)
		service.MustNotBeError(err)

		resultStore := store.Results()
		service.MustNotBeError(resultStore.
			ByID(requestData.TaskToken.Payload.Converted.ParticipantID, requestData.TaskToken.Payload.Converted.AttemptID,
				requestData.TaskToken.Payload.Converted.LocalItemID).
			UpdateColumn(map[string]interface{}{
				"submissions":          gorm.Expr("submissions + 1"),
				"latest_submission_at": database.Now(),
				"latest_activity_at":   database.Now(),
			}).Error())
		service.MustNotBeError(resultStore.MarkAsToBePropagated(
			requestData.TaskToken.Payload.Converted.ParticipantID, requestData.TaskToken.Payload.Converted.AttemptID,
			requestData.TaskToken.Payload.Converted.LocalItemID, true))
		return nil
	})

	service.MustNotBeError(err)

	answerToken, err := (&token.Token[payloads.AnswerToken]{Payload: payloads.AnswerToken{
		Answer:          *requestData.Answer,
		UserID:          requestData.TaskToken.Payload.UserID,
		ItemID:          requestData.TaskToken.Payload.ItemID,
		ItemURL:         requestData.TaskToken.Payload.ItemURL,
		LocalItemID:     requestData.TaskToken.Payload.LocalItemID,
		UserAnswerID:    strconv.FormatInt(answerID, 10),
		RandomSeed:      requestData.TaskToken.Payload.RandomSeed,
		HintsRequested:  hintsInfo.HintsRequested,
		HintsGivenCount: strconv.FormatInt(int64(hintsInfo.HintsCached), 10),
		AttemptID:       requestData.TaskToken.Payload.AttemptID,
		PlatformName:    srv.TokenConfig.PlatformName,
	}}).Sign(srv.TokenConfig.PrivateKey)
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.CreationSuccess(map[string]interface{}{
		"answer_token": answerToken,
	})))
	return nil
}

// UnmarshalJSON loads SubmitRequest from JSON passing a public key into TaskToken.
func (requestData *SubmitRequest) UnmarshalJSON(raw []byte) error {
	var wrapper submitRequestWrapper
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return err
	}
	if wrapper.TaskToken != nil {
		requestData.TaskToken = &token.Token[payloads.TaskToken]{PublicKey: requestData.PublicKey}
		if err := requestData.TaskToken.UnmarshalString(*wrapper.TaskToken); err != nil {
			return fmt.Errorf("invalid task_token: %s", err.Error())
		}
	}
	requestData.Answer = wrapper.Answer
	return nil
}

// Bind checks that all the needed request parameters (task_token & answer) are present and
// all the needed values are valid.
func (requestData *SubmitRequest) Bind(_ *http.Request) error {
	if requestData.TaskToken == nil {
		return errors.New("missing task_token")
	}

	if requestData.Answer == nil {
		return errors.New("missing answer")
	}

	return nil
}

var _ render.Binder = (*SubmitRequest)(nil)
