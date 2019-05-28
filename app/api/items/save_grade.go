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
	if err = user.Load(); err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	apiError := service.NoError
	if apiError = checkSaveGradeTokenParams(user, &requestData); apiError != service.NoError {
		return apiError
	}

	var validated, keyObtained, ok bool
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		var hasAccess bool
		var reason error
		hasAccess, reason, err = store.Items().CheckSubmissionRights(requestData.TaskToken.Converted.LocalItemID, user)
		service.MustNotBeError(err)

		if !hasAccess {
			apiError = service.ErrForbidden(reason)
			return nil // commit! (CheckSubmissionRights() changes the DB sometimes)
		}

		validated, keyObtained, ok = saveGradingResultsIntoDB(store, user, &requestData)
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
		"task_token":   newTaskToken,
		"validated":    validated,
		"key_obtained": keyObtained,
	})))
	return service.NoError
}

func saveGradingResultsIntoDB(store *database.DataStore, user *database.User,
	requestData *saveGradeRequestParsed) (validated, keyObtained, ok bool) {
	const todo = "todo"
	score := requestData.ScoreToken.Converted.Score

	gotFullScore := score == 100
	validated = gotFullScore // currently a validated task is only a task with a full score (score == 100)
	if !saveNewScoreIntoUserAnswer(store, user, requestData, score, validated) {
		return validated, keyObtained, false
	}

	// Build query to update users_items
	// The iScore is set towards the end, so that the IF condition on
	// sBestAnswerDate is computed before iScore is updated
	columnsToUpdate := []string{
		"nbTasksTried",
		"sLastActivityDate",
		"sBestAnswerDate",
		"sLastAnswerDate",
		"iScore",
	}
	values := []interface{}{
		1,
		gorm.Expr("NOW()"),
		gorm.Expr("IF(? > iScore, NOW(), sBestAnswerDate)", score),
		gorm.Expr("NOW()"),
		gorm.Expr("GREATEST(?, iScore)", score),
	}
	if validated {
		// Item was validated
		columnsToUpdate = append(columnsToUpdate,
			"sAncestorsComputationState", "bValidated", "sValidationDate",
		)
		values = append(values,
			todo, 1, gorm.Expr("IFNULL(sValidationDate, NOW())"))
	}
	if shouldUnlockItems(store, requestData.TaskToken.Converted.LocalItemID, score, gotFullScore) {
		keyObtained = true
		if !validated {
			// If validated, as the ancestor's recomputation will happen anyway
			// Update sAncestorsComputationState only if we hadn't obtained the key before
			columnsToUpdate = append(columnsToUpdate, "sAncestorsComputationState")
			values = append(values, gorm.Expr("IF(bKeyObtained = 0, 'todo', sAncestorsComputationState)"))
		}
		columnsToUpdate = append(columnsToUpdate, "bKeyObtained")
		values = append(values, 1)
	}
	if score > 0 && requestData.TaskToken.Converted.AttemptID != nil {
		// Always propagate attempts if the score was non-zero
		columnsToUpdate = append(columnsToUpdate, "sAncestorsComputationState")
		values = append(values, todo)
	}

	updateExpr := "SET " + strings.Join(columnsToUpdate, " = ?, ") + " = ?"
	userItemsValues := make([]interface{}, 0, len(values)+2)
	userItemsValues = append(userItemsValues, values...)
	userItemsValues = append(userItemsValues, user.UserID, requestData.TaskToken.Converted.LocalItemID)
	service.MustNotBeError(
		store.DB.Exec("UPDATE users_items "+updateExpr+" WHERE idUser = ? AND idItem = ?", userItemsValues...).Error()) // nolint:gosec
	if requestData.TaskToken.Converted.AttemptID != nil {
		values = append(values, *requestData.TaskToken.Converted.AttemptID)
		service.MustNotBeError(
			store.DB.Exec("UPDATE groups_attempts "+updateExpr+" WHERE ID = ?", values...).Error()) // nolint:gosec
	}
	service.MustNotBeError(store.GroupAttempts().After())
	return validated, keyObtained, true
}

func saveNewScoreIntoUserAnswer(store *database.DataStore, user *database.User,
	requestData *saveGradeRequestParsed, score float64, validated bool) bool {
	userAnswerID := requestData.ScoreToken.Converted.UserAnswerID
	userAnswerScope := store.UserAnswers().ByID(userAnswerID).
		Where("idUser = ?", user.UserID).
		Where("idItem = ?", requestData.TaskToken.Converted.LocalItemID)

	updateResult := userAnswerScope.Where("iScore = ? OR iScore IS NULL", score).
		UpdateColumn(map[string]interface{}{
			"sGradingDate": gorm.Expr("NOW()"),
			"bValidated":   validated,
			"iScore":       score,
		})
	service.MustNotBeError(updateResult.Error())

	if updateResult.RowsAffected() == 0 {
		var oldScore *float64
		err := userAnswerScope.PluckFirst("iScore", &oldScore).Error()
		if gorm.IsRecordNotFoundError(err) {
			return false
		}
		service.MustNotBeError(err)
		if oldScore != nil {
			if *oldScore != score {
				fieldsForLoggingMarshaled, _ := json.Marshal(map[string]interface{}{
					"idAttempt":    requestData.TaskToken.Converted.AttemptID,
					"idItem":       requestData.TaskToken.Converted.LocalItemID,
					"idUser":       user.UserID,
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
	var unlockedInfo struct {
		UnlockedItemID string  `gorm:"column:idItemUnlocked"`
		ScoreMinUnlock float64 `gorm:"column:iScoreMinUnlock"`
	}
	service.MustNotBeError(store.Items().ByID(itemID).Select("idItemUnlocked, iScoreMinUnlock").
		Take(&unlockedInfo).Error())
	return unlockedInfo.UnlockedItemID != "" && unlockedInfo.ScoreMinUnlock <= score
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
	if (requestData.AnswerToken.AttemptID == nil) != (requestData.TaskToken.AttemptID == nil) ||
		(requestData.AnswerToken.AttemptID != nil &&
			*requestData.AnswerToken.AttemptID != *requestData.TaskToken.AttemptID) {
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
		ItemURL: requestData.TaskToken.ItemURL,
	}
	return nil
}

// Bind of saveGradeRequestParsed does nothing.
func (requestData *saveGradeRequestParsed) Bind(r *http.Request) error {
	return nil
}

func checkSaveGradeTokenParams(user *database.User, requestData *saveGradeRequestParsed) service.APIError {
	if user.UserID != requestData.TaskToken.Converted.UserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token in task_token doesn't correspond to user session: got idUser=%d, expected %d",
			requestData.TaskToken.Converted.UserID, user.UserID))
	}
	if user.UserID != requestData.ScoreToken.Converted.UserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token in score_token doesn't correspond to user session: got idUser=%d, expected %d",
			requestData.ScoreToken.Converted.UserID, user.UserID))
	}
	if requestData.TaskToken.ItemURL != requestData.ScoreToken.ItemURL {
		return service.ErrInvalidRequest(errors.New("wrong itemUrl in score_token"))
	}
	return service.NoError
}
