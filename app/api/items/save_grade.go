package items

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// swagger:operation POST /items/save-grade items saveGrade
//
//	---
//	summary: Save the grade
//	description: >
//
//		This service doesn't require authentication. The user is identified by the answer token or by the score token.
//
//
//		Saves the grade returned by a grading app into the `gradings` table and updates the attempt results in the DB.
//		When the `score` is big enough, the service unlocks locked dependent items (if any).
//
//
//		Restrictions:
//
//	 	* `score_token`/`answer_token` should belong to the current user, otherwise the "bad request"
//	 		response is returned;
//		* the answer should exist and should have not been graded, otherwise the "forbidden" response is returned.
//	parameters:
//		- in: body
//			name: data
//			required: true
//			schema:
//				type: object
//				properties:
//					score_token:
//						description: A score token generated by the grader (required for platforms supporting tokens)
//						type: string
//					answer_token:
//						description: An answer token generated by AlgoreaBackend (required for platforms not supporting tokens)
//						type: string
//					score:
//						description: A score returned by the grader (required for platforms not supporting tokens)
//						type: number
//	responses:
//		"201":
//			description: "Created. Success response."
//			schema:
//					type: object
//					required: [success, message, data]
//					properties:
//						success:
//							description: "true"
//							type: boolean
//							enum: [true]
//						message:
//							description: created
//							type: string
//							enum: [created]
//						data:
//							type: object
//							required: [validated]
//							properties:
//								validated:
//									description: Whether the full score was obtained on this grading
//									type: boolean
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) saveGrade(w http.ResponseWriter, r *http.Request) service.APIError {
	store := srv.GetStore(r)
	requestData := saveGradeRequestParsed{store: store, publicKey: srv.TokenConfig.PublicKey}

	var err error
	if err = render.Bind(r, &requestData); err != nil {
		return service.ErrInvalidRequest(err)
	}

	var validated, ok bool
	err = store.InTransaction(func(store *database.DataStore) error {
		validated, ok = saveGradingResultsIntoDB(store, &requestData)
		return nil
	})
	service.MustNotBeError(err)

	if !ok {
		return service.ErrForbidden(errors.New("the answer has been already graded or is not found"))
	}

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(map[string]interface{}{
		"validated": validated,
	})))
	return service.NoError
}

func saveGradingResultsIntoDB(store *database.DataStore, requestData *saveGradeRequestParsed) (validated, ok bool) {
	score := requestData.ScoreToken.Converted.Score

	gotFullScore := score == 100
	validated = gotFullScore // currently a validated task is only a task with a full score (score == 100)
	if !saveNewScoreIntoGradings(store, requestData, score) {
		return validated, false
	}

	// Build query to update results
	columnsToUpdate := []string{
		"tasks_tried",
		"score_obtained_at",
		"score_computed",
	}
	newScoreExpression := gorm.Expr(`
			LEAST(GREATEST(
				CASE score_edit_rule
					WHEN 'set' THEN score_edit_value
					WHEN 'diff' THEN ? + score_edit_value
					ELSE ?
				END, score_computed, 0), 100)`, score, score)
	values := []interface{}{
		requestData.ScoreToken.Converted.UserAnswerID, // for join
		1, // tasks_tried
		// for score_computed we compare patched scores
		gorm.Expr(`
			CASE
			  -- New best score or no time saved yet
				-- Note that when the score = 0, score_obtained_at is the time of the first submission
				WHEN score_obtained_at IS NULL OR score_computed < ? THEN answers.created_at
				-- We may get the result of an earlier submission after one with the same score
				WHEN score_computed = ? THEN LEAST(score_obtained_at, answers.created_at)
				-- New score if lower than the best score
				ELSE score_obtained_at
			END`, newScoreExpression, newScoreExpression), // score_obtained_at
		newScoreExpression, // score_computed
	}
	if validated {
		// Item was validated
		columnsToUpdate = append(columnsToUpdate, "validated_at")
		values = append(values, gorm.Expr("LEAST(IFNULL(validated_at, answers.created_at), answers.created_at)"))
	}

	updateExpr := "SET " + strings.Join(columnsToUpdate, " = ?, ") + " = ?"
	values = append(values,
		requestData.Converted.ParticipantID,
		requestData.Converted.AttemptID,
		requestData.Converted.LocalItemID,
	)
	service.MustNotBeError(
		store.DB.Exec("UPDATE results JOIN answers ON answers.id = ? "+ // nolint:gosec
			updateExpr+" WHERE results.participant_id = ? AND results.attempt_id = ? AND results.item_id = ?", values...).
			Error()) // nolint:gosec
	resultStore := store.Results()
	service.MustNotBeError(resultStore.MarkAsToBePropagated(
		requestData.Converted.ParticipantID, requestData.Converted.AttemptID,
		requestData.Converted.LocalItemID, true))
	return validated, true
}

func saveNewScoreIntoGradings(store *database.DataStore, requestData *saveGradeRequestParsed, score float64) bool {
	answerID := requestData.ScoreToken.Converted.UserAnswerID
	gradingStore := store.Gradings()

	insertError := gradingStore.InsertMap(map[string]interface{}{
		"answer_id": answerID, "score": score, "graded_at": database.Now(),
	})

	// ERROR 1452 (23000): Cannot add or update a child row: a foreign key constraint fails (the answer has been removed)
	if insertError != nil && database.IsForeignConstraintError(insertError) {
		return false
	}

	// ERROR 1062 (23000): Duplicate entry (already graded)
	if insertError != nil && database.IsDuplicateEntryError(insertError) {
		var oldScore *float64
		service.MustNotBeError(gradingStore.
			Where("answer_id = ?", answerID).PluckFirst("score", &oldScore).Error())
		if oldScore != nil {
			if *oldScore != score {
				fieldsForLoggingMarshaled, _ := json.Marshal(map[string]interface{}{
					"idAttempt":    requestData.ScoreToken.AttemptID,
					"idItem":       requestData.ScoreToken.LocalItemID,
					"idUser":       requestData.ScoreToken.UserID,
					"idUserAnswer": requestData.ScoreToken.UserAnswerID,
					"newScore":     score,
					"oldScore":     *oldScore,
				})
				logging.Warnf("A user tries to replay a score token with a different score value (%s)", fieldsForLoggingMarshaled)
			}
			return false
		}
	}
	service.MustNotBeError(insertError)

	return true
}

type saveGradeRequestParsed struct {
	ScoreToken  *token.Score
	AnswerToken *token.Answer
	Converted   struct {
		UserID        int64
		ParticipantID int64
		AttemptID     int64
		LocalItemID   int64
		ItemURL       string
	}

	store     *database.DataStore
	publicKey *rsa.PublicKey
}

type saveGradeRequest struct {
	ScoreToken  formdata.Anything  `json:"score_token"`
	Score       *float64           `json:"score"`
	AnswerToken *formdata.Anything `json:"answer_token"`
}

// UnmarshalJSON unmarshals the items/saveGrade request data from JSON.
func (requestData *saveGradeRequestParsed) UnmarshalJSON(raw []byte) error {
	var wrapper saveGradeRequest
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return err
	}

	return requestData.unmarshalScoreToken(&wrapper)
}

func (requestData *saveGradeRequestParsed) unmarshalScoreToken(wrapper *saveGradeRequest) error {
	hasScoreToken := wrapper.ScoreToken.Bytes() != nil
	hasPlatformKey := false
	if hasScoreToken {
		// We need the `idItemLocal` to get the platform's public key, and verify the signature of the token.
		// So we need to extract it before we can unmarshal (which also verifies the signature) the token.
		localItemIDRaw, err := token.GetUnsafeFromToken(wrapper.ScoreToken.Bytes(), "idItemLocal")
		if err != nil {
			return errors.Join(errors.New("invalid score_token"), err)
		}

		localItemID, err := strconv.ParseInt(localItemIDRaw.(string), 10, 64)
		service.MustNotBeError(err)

		hasPlatformKey, err = token.UnmarshalDependingOnItemPlatform(
			requestData.store,
			localItemID,
			&requestData.ScoreToken,
			wrapper.ScoreToken.Bytes(),
			"score_token",
		)
		if err != nil {
			return err
		}
	}

	if !hasScoreToken || !hasPlatformKey {
		err := requestData.reconstructScoreTokenData(wrapper)
		if err != nil {
			return err
		}
	}
	return nil
}

func (requestData *saveGradeRequestParsed) unmarshalAnswerToken(wrapper *saveGradeRequest) error {
	if wrapper.AnswerToken == nil {
		return errors.New("missing answer_token")
	}
	requestData.AnswerToken = &token.Answer{PublicKey: requestData.publicKey}
	if err := requestData.AnswerToken.UnmarshalJSON(wrapper.AnswerToken.Bytes()); err != nil {
		return fmt.Errorf("invalid answer_token: %s", err.Error())
	}

	return nil
}

func (requestData *saveGradeRequestParsed) reconstructScoreTokenData(wrapper *saveGradeRequest) error {
	if err := requestData.unmarshalAnswerToken(wrapper); err != nil {
		return err
	}
	if wrapper.Score == nil {
		return errors.New("missing score")
	}
	userAnswerID, err := strconv.ParseInt(requestData.AnswerToken.UserAnswerID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid idUserAnswer in answer_token: %s", err.Error())
	}

	userID, err := strconv.ParseInt(requestData.AnswerToken.UserID, 10, 64)
	service.MustNotBeError(err)

	requestData.ScoreToken = &token.Score{
		Converted: payloads.ScoreTokenConverted{
			Score:        *wrapper.Score,
			UserID:       userID,
			UserAnswerID: userAnswerID,
		},
		ItemURL:     requestData.AnswerToken.ItemURL,
		AttemptID:   requestData.AnswerToken.AttemptID,
		LocalItemID: requestData.AnswerToken.LocalItemID,
	}
	return nil
}

// Bind of saveGradeRequestParsed does nothing.
func (requestData *saveGradeRequestParsed) Bind(*http.Request) error {
	requestData.bindUserID()
	requestData.bindParticipantIDAndAttemptID()
	requestData.bindLocalItemID()
	requestData.bindItemURL()

	return nil
}

// Binds the user ID from the request, coming from either the score token or the answer token.
func (requestData *saveGradeRequestParsed) bindUserID() {
	userID := requestData.ScoreToken.Converted.UserID

	if userID == 0 {
		var err error
		userID, err = strconv.ParseInt(requestData.AnswerToken.UserID, 10, 64)
		service.MustNotBeError(err)
	}

	requestData.Converted.UserID = userID
}

// Binds the participant ID and the attempt ID from the request, coming from either the score token or the answer token.
func (requestData *saveGradeRequestParsed) bindParticipantIDAndAttemptID() {
	var participantID, attemptID int64
	_, err := fmt.Sscanf(requestData.ScoreToken.AttemptID, "%d/%d", &participantID, &attemptID)
	if err != nil {
		_, err = fmt.Sscanf(requestData.AnswerToken.AttemptID, "%d/%d", &participantID, &attemptID)
		service.MustNotBeError(err)
	}

	requestData.Converted.ParticipantID = participantID
	requestData.Converted.AttemptID = attemptID
}

// Bind the local item ID from the request, coming from either the score token or the answer token.
func (requestData *saveGradeRequestParsed) bindLocalItemID() {
	localItemID, err := strconv.ParseInt(requestData.ScoreToken.LocalItemID, 10, 64)
	if localItemID == 0 || err != nil {
		localItemID, err = strconv.ParseInt(requestData.AnswerToken.LocalItemID, 10, 64)
		service.MustNotBeError(err)
	}

	requestData.Converted.LocalItemID = localItemID
}

// Bind the item URL from the request, coming from either the score token or the answer token.
func (requestData *saveGradeRequestParsed) bindItemURL() {
	if requestData.ScoreToken.ItemURL != "" {
		requestData.Converted.ItemURL = requestData.ScoreToken.ItemURL
	} else {
		requestData.Converted.ItemURL = requestData.AnswerToken.ItemURL
	}
}
