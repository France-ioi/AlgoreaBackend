package items

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

func (srv *Service) askHint(w http.ResponseWriter, r *http.Request) service.APIError {
	requestData := AskHintRequest{store: srv.Store, publicKey: srv.TokenConfig.PublicKey}

	var err error
	if err = render.Bind(r, &requestData); err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	apiError := service.NoError
	if apiError = checkAskHintUsersAndItemURL(user, &requestData); apiError != service.NoError {
		return apiError
	}

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

		// Get the previous hints requested JSON data
		var hintsRequestedParsed []formdata.Anything
		hintsRequestedParsed, err = queryAndParsePreviouslyRequestedHints(requestData.TaskToken, store, user, r)
		if err == gorm.ErrRecordNotFound {
			apiError = service.ErrNotFound(errors.New("can't find previously requested hints info"))
			return nil // commit
		}
		service.MustNotBeError(err)

		// Add the new requested hint to the list if it's not in the list yet
		hintsRequestedParsed = addHintToListIfNeeded(hintsRequestedParsed, requestData.HintToken.AskedHint)

		var hintsRequestedNew []byte
		hintsRequestedNew, err = json.Marshal(hintsRequestedParsed)
		service.MustNotBeError(err)
		hintsRequestedNewString := string(hintsRequestedNew)
		requestData.TaskToken.HintsRequested = &hintsRequestedNewString
		hintsGivenCountString := strconv.Itoa(len(hintsRequestedParsed))
		requestData.TaskToken.HintsGivenCount = &hintsGivenCountString

		columnsToUpdate := map[string]interface{}{
			"nbTasksWithHelp":            1,
			"sAncestorsComputationState": "todo",
			"sLastActivityDate":          gorm.Expr("NOW()"),
			"sLastHintDate":              gorm.Expr("NOW()"),
		}
		// Update users_items with the hint request
		service.MustNotBeError(store.UserItems().Where("idUser = ?", user.UserID).
			Where("idItem = ?", requestData.TaskToken.Converted.LocalItemID).
			Where("idAttemptActive = ?", requestData.TaskToken.Converted.AttemptID).
			UpdateColumn(columnsToUpdate).Error())

		// Update groups_attempts with the hint request
		columnsToUpdate["sHintsRequested"] = hintsRequestedNew
		columnsToUpdate["nbHintsCached"] = len(hintsRequestedParsed)
		service.MustNotBeError(store.GroupAttempts().ByID(requestData.TaskToken.Converted.AttemptID).
			UpdateColumn(columnsToUpdate).Error())

		service.MustNotBeError(store.GroupAttempts().After())

		return nil
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	requestData.TaskToken.PlatformName = srv.TokenConfig.PlatformName
	newTaskToken, err := requestData.TaskToken.Sign(srv.TokenConfig.PrivateKey)
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(map[string]interface{}{
		"task_token": newTaskToken,
	})))
	return service.NoError
}

func queryAndParsePreviouslyRequestedHints(taskToken *token.Task, store *database.DataStore,
	user *database.User, r *http.Request) ([]formdata.Anything, error) {
	var hintsRequested *string
	err := store.GroupAttempts().ByID(taskToken.Converted.AttemptID).PluckFirst("sHintsRequested", &hintsRequested).Error()
	var hintsRequestedParsed []formdata.Anything
	if err == nil && hintsRequested != nil {
		hintsErr := json.Unmarshal([]byte(*hintsRequested), &hintsRequestedParsed)
		if hintsErr != nil {
			hintsRequestedParsed = nil
			fieldsForLoggingMarshaled, _ := json.Marshal(map[string]interface{}{
				"idUser":    user.UserID,
				"idItem":    taskToken.Converted.LocalItemID,
				"idAttempt": taskToken.Converted.AttemptID,
			})
			logging.GetLogEntry(r).Warnf("Unable to parse sHintsRequested (%s) having value %q: %s", fieldsForLoggingMarshaled,
				*hintsRequested, hintsErr.Error())
		}
	}
	return hintsRequestedParsed, err
}

func addHintToListIfNeeded(hintsList []formdata.Anything, hintToAdd formdata.Anything) []formdata.Anything {
	var hintFound bool
	for _, hint := range hintsList {
		if bytes.Equal(hint.Bytes(), hintToAdd.Bytes()) {
			hintFound = true
			break
		}
	}
	if !hintFound {
		hintsList = append(hintsList, hintToAdd)
	}
	return hintsList
}

// AskHintRequest represents a JSON request body format needed by items.askHint()
type AskHintRequest struct {
	TaskToken *token.Task
	HintToken *token.Hint

	store     *database.DataStore
	publicKey *rsa.PublicKey
}

type askHintRequestWrapper struct {
	TaskToken          *string            `json:"task_token"`
	HintRequestedToken *formdata.Anything `json:"hint_requested"`
}

// UnmarshalJSON unmarshals the items/askHint request data from JSON
func (requestData *AskHintRequest) UnmarshalJSON(raw []byte) error {
	var wrapper askHintRequestWrapper
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
	return requestData.unmarshalHintToken(&wrapper)
}

func (requestData *AskHintRequest) unmarshalHintToken(wrapper *askHintRequestWrapper) error {
	if wrapper.HintRequestedToken == nil {
		return errors.New("missing hint_requested")
	}

	err := token.UnmarshalDependingOnItemPlatform(requestData.store, requestData.TaskToken.Converted.LocalItemID,
		&requestData.HintToken, wrapper.HintRequestedToken.Bytes(), "hint_requested")
	if err != nil && !token.IsUnexpectedError(err) {
		return err
	}
	service.MustNotBeError(err)

	if requestData.HintToken == nil {
		hintToken := payloads.HintToken{}
		if err := json.Unmarshal(wrapper.HintRequestedToken.Bytes(), &hintToken); err != nil {
			return fmt.Errorf("invalid hint_requested: %s", err.Error())
		}
		requestData.HintToken = (*token.Hint)(&hintToken)
	}
	return nil
}

// Bind of AskHintRequest checks that the asked hint is present
func (requestData *AskHintRequest) Bind(r *http.Request) error {
	if len(requestData.HintToken.AskedHint.Bytes()) == 0 || bytes.Equal([]byte("null"), requestData.HintToken.AskedHint.Bytes()) {
		return fmt.Errorf("asked hint should not be empty")
	}
	return nil
}

func checkAskHintUsersAndItemURL(user *database.User, requestData *AskHintRequest) service.APIError {
	var err error
	if err = user.Load(); err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if user.UserID != requestData.TaskToken.Converted.UserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token in task_token doesn't correspond to user session: got idUser=%d, expected %d",
			requestData.TaskToken.Converted.UserID, user.UserID))
	}
	if requestData.HintToken.Converted.UserID != nil && user.UserID != *requestData.HintToken.Converted.UserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token in hint_requested doesn't correspond to user session: got idUser=%d, expected %d",
			*requestData.HintToken.Converted.UserID, user.UserID))
	}
	if requestData.HintToken.ItemURL != nil && requestData.TaskToken.ItemURL != *requestData.HintToken.ItemURL {
		return service.ErrInvalidRequest(errors.New("wrong itemUrl in hint_requested token"))
	}
	if requestData.HintToken.AttemptID != requestData.TaskToken.AttemptID {
		return service.ErrInvalidRequest(errors.New("wrong idAttempt in hint_requested token"))
	}
	return service.NoError
}
