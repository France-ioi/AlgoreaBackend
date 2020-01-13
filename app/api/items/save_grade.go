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
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

func (srv *Service) saveGrade(w http.ResponseWriter, r *http.Request) service.APIError {
	requestData := saveGradeRequestParsed{store: srv.Store, publicKey: srv.TokenConfig.PublicKey}

	var err error
	if err = render.Bind(r, &requestData); err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	apiError := service.NoError
	if apiError = checkHintOrScoreTokenRequiredFields(user, requestData.TaskToken, "score_token",
		requestData.ScoreToken.Converted.UserID, requestData.ScoreToken.LocalItemID,
		requestData.ScoreToken.ItemURL, requestData.ScoreToken.AttemptID); apiError != service.NoError {
		return apiError
	}

	var validated, ok bool
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		var hasAccess bool
		var reason error
		hasAccess, reason, err = store.Items().CheckSubmissionRights(requestData.TaskToken.Converted.LocalItemID, user)
		service.MustNotBeError(err)

		if !hasAccess {
			apiError = service.ErrForbidden(reason)
			return nil // commit! (CheckSubmissionRights() changes the DB sometimes)
		}

		validated, ok = saveGradingResultsIntoDB(store, user, &requestData)
		return nil
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	if !ok {
		return service.ErrForbidden(errors.New("the answer has been already graded or is not found"))
	}

	if validated && requestData.TaskToken.AccessSolutions != nil && !(*requestData.TaskToken.AccessSolutions) {
		requestData.TaskToken.AccessSolutions = ptrBool(true)
	}
	requestData.TaskToken.PlatformName = srv.TokenConfig.PlatformName
	newTaskToken, err := requestData.TaskToken.Sign(srv.TokenConfig.PrivateKey)
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(map[string]interface{}{
		"task_token": newTaskToken,
		"validated":  validated,
	})))
	return service.NoError
}

func saveGradingResultsIntoDB(store *database.DataStore, user *database.User,
	requestData *saveGradeRequestParsed) (validated, ok bool) {
	score := requestData.ScoreToken.Converted.Score

	gotFullScore := score == 100
	validated = gotFullScore // currently a validated task is only a task with a full score (score == 100)
	if !saveNewScoreIntoGradings(store, user, requestData, score) {
		return validated, false
	}

	// Build query to update groups_attempts
	columnsToUpdate := []string{
		"tasks_tried",
		"latest_answer_at",
		"latest_answer_at",
		"score_obtained_at",
		"score_computed",
		"result_propagation_state",
	}
	newScoreExpression := gorm.Expr(`
			LEAST(GREATEST(
				CASE score_edit_rule
					WHEN 'set' THEN score_edit_value
					WHEN 'diff' THEN ? + score_edit_value
					ELSE ?
				END, score_computed, 0), 100)`, score, score)
	values := []interface{}{
		1,
		// store the old value into a temporary variable
		gorm.Expr("(@old_latest_answer_at:=latest_answer_at)"),
		// use latest_answer_at to store the answer's submission time
		gorm.Expr("(SELECT created_at FROM answers WHERE id = ? FOR UPDATE)", requestData.ScoreToken.Converted.UserAnswerID),
		// for score_computed we compare patched scores
		gorm.Expr(`
			CASE
			  -- New best score or no time saved yet
				-- Note that when the score = 0, score_obtained_at is the time of the first submission
				WHEN score_obtained_at IS NULL OR score_computed < ? THEN latest_answer_at
				-- We may get the result of an earlier submission after one with the same score
				WHEN score_computed = ? THEN LEAST(score_obtained_at, latest_answer_at)
				-- New score if lower than the best score
				ELSE score_obtained_at
			END`, newScoreExpression, newScoreExpression),
		newScoreExpression,
		"changed",
	}
	if validated {
		// Item was validated
		columnsToUpdate = append(columnsToUpdate, "validated_at")
		values = append(values, gorm.Expr("LEAST(IFNULL(validated_at, latest_answer_at), latest_answer_at)"))
	}
	columnsToUpdate = append(columnsToUpdate, "latest_answer_at")
	values = append(values, gorm.Expr("GREATEST(latest_answer_at, IFNULL(@old_latest_answer_at, latest_answer_at))"))

	updateExpr := "SET " + strings.Join(columnsToUpdate, " = ?, ") + " = ?"
	values = append(values, requestData.TaskToken.Converted.AttemptID)
	service.MustNotBeError(
		store.DB.Exec("UPDATE groups_attempts "+updateExpr+" WHERE id = ?", values...).Error()) // nolint:gosec
	service.MustNotBeError(store.GroupAttempts().ComputeAllGroupAttempts())
	return validated, true
}

func saveNewScoreIntoGradings(store *database.DataStore, user *database.User,
	requestData *saveGradeRequestParsed, score float64) bool {
	answerID := requestData.ScoreToken.Converted.UserAnswerID
	gradingStore := store.Gradings()

	insertError := gradingStore.InsertMap(map[string]interface{}{
		"answer_id": answerID, "score": score, "graded_at": database.Now(),
	})

	// ERROR 1452 (23000): Cannot add or update a child row: a foreign key constraint fails (the answer has been removed)
	if e, ok := insertError.(*mysql.MySQLError); ok && e.Number == 1452 {
		return false
	}

	// ERROR 1062 (23000): Duplicate entry (already graded)
	if e, ok := insertError.(*mysql.MySQLError); ok && e.Number == 1062 {
		var oldScore *float64
		service.MustNotBeError(gradingStore.
			Where("answer_id = ?", answerID).PluckFirst("score", &oldScore).Error())
		if oldScore != nil {
			if *oldScore != score {
				fieldsForLoggingMarshaled, _ := json.Marshal(map[string]interface{}{
					"idAttempt":    requestData.TaskToken.Converted.AttemptID,
					"idItem":       requestData.TaskToken.Converted.LocalItemID,
					"idUser":       user.GroupID,
					"idUserAnswer": requestData.ScoreToken.Converted.UserAnswerID,
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
	TaskToken   *token.Task
	ScoreToken  *token.Score
	AnswerToken *token.Answer

	store     *database.DataStore
	publicKey *rsa.PublicKey
}

type saveGradeRequest struct {
	TaskToken   *string            `json:"task_token"`
	ScoreToken  formdata.Anything  `json:"score_token"`
	Score       *float64           `json:"score"`
	AnswerToken *formdata.Anything `json:"answer_token"`
	AnswerID    *string            `json:"answer_id"`
}

// UnmarshalJSON unmarshals the items/saveGrade request data from JSON
func (requestData *saveGradeRequestParsed) UnmarshalJSON(raw []byte) error {
	var wrapper saveGradeRequest
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return err
	}
	if wrapper.TaskToken == nil {
		return errors.New("missing task_token")
	}
	requestData.TaskToken = &token.Task{PublicKey: requestData.publicKey}
	if err := requestData.TaskToken.UnmarshalString(*wrapper.TaskToken); err != nil {
		return fmt.Errorf("invalid task_token: %s", err.Error())
	}
	return requestData.unmarshalScoreToken(&wrapper)
}

func (requestData *saveGradeRequestParsed) unmarshalScoreToken(wrapper *saveGradeRequest) error {
	err := token.UnmarshalDependingOnItemPlatform(requestData.store, requestData.TaskToken.Converted.LocalItemID,
		&requestData.ScoreToken, wrapper.ScoreToken.Bytes(), "score_token")
	if err != nil && !token.IsUnexpectedError(err) {
		return err
	}
	service.MustNotBeError(err)
	if requestData.ScoreToken == nil {
		err = requestData.reconstructScoreTokenData(wrapper)
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
	if requestData.AnswerToken.UserID != requestData.TaskToken.UserID {
		return errors.New("wrong idUser in answer_token")
	}
	if requestData.AnswerToken.LocalItemID != requestData.TaskToken.LocalItemID {
		return errors.New("wrong idItemLocal in answer_token")
	}
	if requestData.AnswerToken.ItemURL != requestData.TaskToken.ItemURL {
		return errors.New("wrong itemUrl in answer_token")
	}
	if requestData.AnswerToken.AttemptID != requestData.TaskToken.AttemptID {
		return errors.New("wrong idAttempt in answer_token")
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
	requestData.ScoreToken = &token.Score{
		Converted: payloads.ScoreTokenConverted{
			Score:        *wrapper.Score,
			UserID:       requestData.TaskToken.Converted.UserID,
			UserAnswerID: userAnswerID,
		},
		ItemURL:     requestData.TaskToken.ItemURL,
		AttemptID:   requestData.AnswerToken.AttemptID,
		LocalItemID: requestData.AnswerToken.LocalItemID,
	}
	return nil
}

// Bind of saveGradeRequestParsed does nothing.
func (requestData *saveGradeRequestParsed) Bind(*http.Request) error {
	return nil
}
