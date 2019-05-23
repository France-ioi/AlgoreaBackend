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
	"github.com/France-ioi/AlgoreaBackend/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

func (srv *Service) saveGrade(w http.ResponseWriter, r *http.Request) service.APIError {
	requestData := SaveGradeRequest{store: srv.Store, publicKey: srv.TokenConfig.PublicKey}

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

	var validated, keyObtained bool
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		var hasAccess bool
		var reason error
		hasAccess, reason, err = store.Items().CheckSubmissionRights(requestData.TaskToken.Converted.LocalItemID, user)
		service.MustNotBeError(err)

		if !hasAccess {
			apiError = service.ErrForbidden(reason)
			return nil // commit! (CheckSubmissionRights() changes the DB sometimes)
		}

		validated, keyObtained = saveGradingResultsIntoDB(store, user, &requestData)
		return nil
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	if validated && requestData.TaskToken.Converted.AccessSolutions != nil && !(*requestData.TaskToken.Converted.AccessSolutions) {
		requestData.TaskToken.AccessSolutions = formdata.AnythingFromString(`"1"`)
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
	requestData *SaveGradeRequest) (validated, keyObtained bool) {
	const todo = "todo"
	score := requestData.ScoreToken.Converted.Score
	userAnswerID := requestData.ScoreToken.Converted.UserAnswerID

	// TODO: handle validation in a proper way (what did he mean??)
	if score > 99 {
		validated = true
	}
	service.MustNotBeError(store.UserAnswers().ByID(userAnswerID).
		Where("idUser = ?", user.UserID).
		Where("idItem = ?", requestData.TaskToken.Converted.LocalItemID).
		UpdateColumn(map[string]interface{}{
			"sGradingDate": gorm.Expr("NOW()"),
			"bValidated":   validated,
			"iScore":       score,
		}).Error())
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
			"sAncestorsComputationState", "bValidated", "bKeyObtained", "sValidationDate",
		)
		values = append(values,
			todo, 1, 1, gorm.Expr("IFNULL(sValidationDate, NOW())"))
		keyObtained = true
	} else {
		// Item wasn't validated, check if we unlocked something
		var unlockedInfo struct {
			UnlockedItemID string  `gorm:"column:idItemUnlocked"`
			ScoreMinUnlock float64 `gorm:"column:iScoreMinUnlock"`
		}
		service.MustNotBeError(store.Items().ByID(requestData.TaskToken.Converted.LocalItemID).Select("idItemUnlocked, iScoreMinUnlock").
			Take(&unlockedInfo).Error())
		if unlockedInfo.UnlockedItemID != "" && unlockedInfo.ScoreMinUnlock < score {
			keyObtained = true
			// Update sAncestorsComputationState only if we hadn't obtained the key before
			columnsToUpdate = append(columnsToUpdate,
				"sAncestorsComputationState", "bKeyObtained",
			)
			values = append(values, gorm.Expr("IF(bKeyObtained = 0, 'todo', sAncestorsComputationState)"), 1)
		}
	}
	if score > 0 && requestData.TaskToken.Converted.AttemptID != nil {
		// Always propagate attempts if the score was non-zero
		columnsToUpdate = append(columnsToUpdate, "sAncestorsComputationState")
		values = append(values, todo)
	}

	updateExpr := "SET " + strings.Join(columnsToUpdate, " = ?, ") + " = ?"
	userItemsValues := make([]interface{}, len(values)+2)
	copy(userItemsValues, values)
	userItemsValues[len(userItemsValues)-2] = user.UserID
	userItemsValues[len(userItemsValues)-1] = requestData.TaskToken.Converted.LocalItemID
	service.MustNotBeError(
		store.DB.Exec("UPDATE users_items "+updateExpr+" WHERE idUser = ? AND idItem = ?", userItemsValues...).Error()) // nolint:gosec
	if requestData.TaskToken.Converted.AttemptID != nil {
		values = append(values, *requestData.TaskToken.Converted.AttemptID)
		service.MustNotBeError(
			store.DB.Exec("UPDATE groups_attempts "+updateExpr+" WHERE ID = ?", values...).Error()) // nolint:gosec
	}
	service.MustNotBeError(store.GroupAttempts().After())
	return validated, keyObtained
}

// SaveGradeRequest represents a JSON request body format needed by items.saveGrade()
type SaveGradeRequest struct {
	TaskToken   *token.Task
	ScoreToken  *token.Score
	AnswerToken *token.Answer

	store     *database.DataStore
	publicKey *rsa.PublicKey
}

type saveGradeRequestWrapper struct {
	TaskToken    *string            `json:"task_token"`
	ScoreToken   formdata.Anything  `json:"score_token"`
	Score        *float64           `json:"score"`
	AnswerToken  *formdata.Anything `json:"answer_token"`
	UserAnswerID *string            `json:"user_answer_id"`
}

// UnmarshalJSON unmarshals the items/saveGrade request data from JSON
func (requestData *SaveGradeRequest) UnmarshalJSON(raw []byte) error {
	var wrapper saveGradeRequestWrapper
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

func (requestData *SaveGradeRequest) unmarshalScoreToken(wrapper *saveGradeRequestWrapper) error {
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

func (requestData *SaveGradeRequest) unmarshalAnswerToken(wrapper *saveGradeRequestWrapper) error {
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

func (requestData *SaveGradeRequest) reconstructScoreTokenData(wrapper *saveGradeRequestWrapper) error {
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

// Bind of SaveGradeRequest does nothing.
func (requestData *SaveGradeRequest) Bind(r *http.Request) error {
	return nil
}

func checkSaveGradeTokenParams(user *database.User, requestData *SaveGradeRequest) service.APIError {
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
