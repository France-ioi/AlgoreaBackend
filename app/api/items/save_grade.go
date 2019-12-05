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

	var validated, hasUnlockedItems, ok bool
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		var hasAccess bool
		var reason error
		hasAccess, reason, err = store.Items().CheckSubmissionRights(requestData.TaskToken.Converted.LocalItemID, user)
		service.MustNotBeError(err)

		if !hasAccess {
			apiError = service.ErrForbidden(reason)
			return nil // commit! (CheckSubmissionRights() changes the DB sometimes)
		}

		validated, hasUnlockedItems, ok = saveGradingResultsIntoDB(store, user, &requestData)
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
		"task_token":         newTaskToken,
		"validated":          validated,
		"has_unlocked_items": hasUnlockedItems,
	})))
	return service.NoError
}

func saveGradingResultsIntoDB(store *database.DataStore, user *database.User,
	requestData *saveGradeRequestParsed) (validated, hasUnlockedItems, ok bool) {
	const todo = "todo"
	score := requestData.ScoreToken.Converted.Score

	gotFullScore := score == 100
	validated = gotFullScore // currently a validated task is only a task with a full score (score == 100)
	if !saveNewScoreIntoUserAnswer(store, user, requestData, score, validated) {
		return validated, hasUnlockedItems, false
	}

	// Build query to update groups_attempts
	// The score is set towards the end, so that the IF condition on
	// best_answer_at is computed before score is updated
	columnsToUpdate := []string{
		"tasks_tried",
		"latest_activity_at",
		"best_answer_at",
		"latest_answer_at",
		"score",
	}
	values := []interface{}{
		1,
		database.Now(),
		gorm.Expr("IF(? > score, ?, best_answer_at)", score, database.Now()),
		database.Now(),
		gorm.Expr("GREATEST(?, score)", score),
	}
	if validated {
		// Item was validated
		columnsToUpdate = append(columnsToUpdate,
			"ancestors_computation_state", "validated_at",
		)
		values = append(values,
			todo, gorm.Expr("IFNULL(validated_at, ?)", database.Now()))
	}
	if shouldUnlockItems(store, requestData.TaskToken.Converted.LocalItemID, score, gotFullScore) {
		hasUnlockedItems = true
		if !validated {
			// If validated, as the ancestor's recomputation will happen anyway
			columnsToUpdate = append(columnsToUpdate, "ancestors_computation_state")
			values = append(values, "todo")
		}
	}
	if score > 0 {
		// Always propagate attempts if the score was non-zero
		columnsToUpdate = append(columnsToUpdate, "ancestors_computation_state")
		values = append(values, todo)
	}

	updateExpr := "SET " + strings.Join(columnsToUpdate, " = ?, ") + " = ?"
	values = append(values, requestData.TaskToken.Converted.AttemptID)
	service.MustNotBeError(
		store.DB.Exec("UPDATE groups_attempts "+updateExpr+" WHERE id = ?", values...).Error()) // nolint:gosec
	service.MustNotBeError(store.GroupAttempts().ComputeAllGroupAttempts())
	return validated, hasUnlockedItems, true
}

func saveNewScoreIntoUserAnswer(store *database.DataStore, user *database.User,
	requestData *saveGradeRequestParsed, score float64, validated bool) bool {
	userAnswerID := requestData.ScoreToken.Converted.UserAnswerID
	userAnswerScope := store.UserAnswers().ByID(userAnswerID).
		Where("user_id = ?", user.GroupID).
		Where("item_id = ?", requestData.TaskToken.Converted.LocalItemID)

	updateResult := userAnswerScope.Where("score = ? OR score IS NULL", score).
		UpdateColumn(map[string]interface{}{
			"graded_at": database.Now(),
			"validated": validated,
			"score":     score,
		})
	service.MustNotBeError(updateResult.Error())

	if updateResult.RowsAffected() == 0 {
		var oldScore *float64
		err := userAnswerScope.PluckFirst("score", &oldScore).Error()
		if gorm.IsRecordNotFoundError(err) {
			return false
		}
		service.MustNotBeError(err)
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

	return true
}

func shouldUnlockItems(store *database.DataStore, itemID int64, score float64, gotFullScore bool) bool {
	if gotFullScore {
		return true
	}
	found, err := store.ItemUnlockingRules().
		Where("unlocking_item_id = ?", itemID).
		Where("score <= ?", score).HasRows()
	service.MustNotBeError(err)
	return found
}

type saveGradeRequestParsed struct {
	TaskToken   *token.Task
	ScoreToken  *token.Score
	AnswerToken *token.Answer

	store     *database.DataStore
	publicKey *rsa.PublicKey
}

type saveGradeRequest struct {
	TaskToken    *string            `json:"task_token"`
	ScoreToken   formdata.Anything  `json:"score_token"`
	Score        *float64           `json:"score"`
	AnswerToken  *formdata.Anything `json:"answer_token"`
	UserAnswerID *string            `json:"user_answer_id"`
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
func (requestData *saveGradeRequestParsed) Bind(r *http.Request) error {
	return nil
}
