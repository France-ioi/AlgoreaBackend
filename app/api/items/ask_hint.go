package items

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
	"gopkg.in/jose.v1/crypto"

	"github.com/France-ioi/AlgoreaBackend/app/database"
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
	if apiError = checkAskHintUsers(user, &requestData); apiError != service.NoError {
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
		var hintsRequestedParsed []payloads.Anything
		hintsRequestedParsed, err = queryAndParsePreviouslyRequestedHints(&requestData, store, user, r)
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
			"sHintsRequested":            hintsRequestedNew,
			"nbHintsCached":              len(hintsRequestedParsed),
			"nbTasksWithHelp":            1,
			"sAncestorsComputationState": "todo",
			"sLastActivityDate":          gorm.Expr("NOW()"),
			"sLastHintDate":              gorm.Expr("NOW()"),
		}
		// Update groups_attempts with the hint request
		if requestData.TaskToken.Converted.AttemptID != nil {
			service.MustNotBeError(store.GroupAttempts().ByID(*requestData.TaskToken.Converted.AttemptID).
				UpdateColumn(columnsToUpdate).Error())
		}
		// Update users_items with the hint request
		query := store.UserItems().Where("idUser = ?", user.UserID).
			Where("idItem = ?", requestData.TaskToken.Converted.LocalItemID)
		if requestData.TaskToken.Converted.AttemptID != nil {
			query = query.Where("idAttemptActive = ?", *requestData.TaskToken.Converted.AttemptID)
		}
		service.MustNotBeError(query.UpdateColumn(columnsToUpdate).Error())
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

func queryAndParsePreviouslyRequestedHints(
	requestData *AskHintRequest, store *database.DataStore, user *database.User, r *http.Request) ([]payloads.Anything, error) {
	fieldsForLogging := map[string]interface{}{
		"idUser": user.UserID,
		"idItem": requestData.TaskToken.Converted.LocalItemID,
	}
	var query *database.DB
	if requestData.TaskToken.Converted.AttemptID != nil {
		query = store.GroupAttempts().ByID(*requestData.TaskToken.Converted.AttemptID)
		fieldsForLogging["idAttempt"] = *requestData.TaskToken.Converted.AttemptID
	} else {
		query = store.UserItems().Where("idUser = ?", user.UserID).
			Where("idItem = ?", requestData.TaskToken.Converted.LocalItemID)
	}
	var hintsRequested *string
	err := query.PluckFirst("sHintsRequested", &hintsRequested).Error()
	var hintsRequestedParsed []payloads.Anything
	if err == nil && hintsRequested != nil {
		hintsErr := json.Unmarshal(*(*[]byte)(unsafe.Pointer(hintsRequested)), &hintsRequestedParsed) //nolint:gosec
		if hintsErr != nil {
			fieldsForLoggingMarshaled, _ := json.Marshal(fieldsForLogging)
			logging.GetLogEntry(r).Warnf("Unable to parse sHintsRequested (%s): %s", fieldsForLoggingMarshaled, hintsErr.Error())
		}
	}
	return hintsRequestedParsed, err
}

func addHintToListIfNeeded(hintsList []payloads.Anything, hintToAdd payloads.Anything) []payloads.Anything {
	var hintFound bool
	for _, hint := range hintsList {
		if bytes.Equal(hint, hintToAdd) {
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
	TaskToken          *string           `json:"task_token"`
	HintRequestedToken payloads.Anything `json:"hint_requested"`
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
	if wrapper.HintRequestedToken == nil {
		return errors.New("missing hint_requested")
	}
	return requestData.unmarshalHintToken(&wrapper)
}

func (requestData *AskHintRequest) unmarshalHintToken(wrapper *askHintRequestWrapper) error {
	var platformInfo struct {
		UsesTokens bool   `gorm:"column:bUsesTokens"`
		PublicKey  string `gorm:"column:sPublicKey"`
	}
	var err error
	if err = requestData.store.Platforms().Select("bUsesTokens, sPublicKey").
		Joins("JOIN items ON items.idPlatform = platforms.ID").
		Where("items.ID = ?", requestData.TaskToken.Converted.LocalItemID).
		Scan(&platformInfo).Error(); gorm.IsRecordNotFoundError(err) {
		return fmt.Errorf("cannot find the platform for item %s", requestData.TaskToken.LocalItemID)
	}
	service.MustNotBeError(err)

	if platformInfo.UsesTokens {
		parsedPublicKey, err := crypto.ParseRSAPublicKeyFromPEM([]byte(platformInfo.PublicKey))
		if err != nil {
			logging.Warnf("cannot parse platform's public key for item with ID = %d: %s",
				requestData.TaskToken.Converted.LocalItemID, err.Error())
			return errors.New("invalid hint_requested: wrong platform's key")
		}
		requestData.HintToken = &token.Hint{PublicKey: parsedPublicKey}
		if err = requestData.HintToken.UnmarshalJSON([]byte(wrapper.HintRequestedToken)); err != nil {
			return fmt.Errorf("invalid hint_requested: %s", err.Error())
		}
	} else {
		hintToken := payloads.HintToken{}
		if err := hintToken.UnmarshalJSON(wrapper.HintRequestedToken); err != nil {
			return fmt.Errorf("invalid hint_requested: %s", err.Error())
		}
		requestData.HintToken = (*token.Hint)(&hintToken)
	}
	return nil
}

// Bind of AskHintRequest does nothing.
func (requestData *AskHintRequest) Bind(r *http.Request) error {
	return nil
}

func checkAskHintUsers(user *database.User, requestData *AskHintRequest) service.APIError {
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
	if user.UserID != requestData.HintToken.Converted.UserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token in hint_requested doesn't correspond to user session: got idUser=%d, expected %d",
			requestData.HintToken.Converted.UserID, user.UserID))
	}
	return service.NoError
}
