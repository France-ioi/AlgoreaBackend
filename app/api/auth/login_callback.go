package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /auth/login-callback auth userLoginCallback
// ---
// summary: User login callback
// description: Callback to which the user is redirected after authentication with the login module.
//
//
//              Creates or updates the authenticated user in the DB using the data returned by the login module,
//              saves the access & refresh tokens in DB as well.
//
//   * No “Authorization” header should be present
//
//   * `login_csrf` cookie set by `/auth/login` should be present
// security: []
// parameters:
// - name: code
//   in: query
//   description: OAuth2 code
//   type: string
//   required: true
// - name: state
//   in: query
//   description: OAuth2 state
//   type: string
//   required: true
// responses:
//   "201":
//     description: "Created. Success response with the new access token"
//     in: body
//     schema:
//       "$ref": "#/definitions/userCreateTmpResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) loginCallback(w http.ResponseWriter, r *http.Request) service.APIError {
	if len(r.Header["Authorization"]) != 0 {
		return service.ErrInvalidRequest(errors.New("'Authorization' header should not be present"))
	}

	code, err := service.ResolveURLQueryGetStringField(r, "code")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	state, err := service.ResolveURLQueryGetStringField(r, "state")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	loginState, err := auth.LoadLoginState(srv.Store.LoginStates(), r, state)
	service.MustNotBeError(err)
	if !loginState.IsOK() {
		return service.ErrInvalidRequest(errors.New("wrong state"))
	}

	oauthConfig := getOAuthConfig(&srv.Config.Auth)
	token, err := oauthConfig.Exchange(r.Context(), code)
	service.MustNotBeError(err)

	userProfile, err := loginmodule.NewClient(srv.Config.Auth.LoginModuleURL).GetUserProfile(r.Context(), token.AccessToken)
	service.MustNotBeError(err)
	userProfile["last_ip"] = strings.SplitN(r.RemoteAddr, ":", 2)[0]

	domainConfig := domain.ConfigFromContext(r.Context())

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		userID := createOrUpdateUser(store.Users(), userProfile, domainConfig)
		service.MustNotBeError(store.Sessions().InsertNewOAuth(userID, token))

		service.MustNotBeError(store.Exec(
			"INSERT INTO refresh_tokens (user_id, refresh_token) VALUES (?, ?) ON DUPLICATE KEY UPDATE refresh_token = ?",
			userID, token.RefreshToken, token.RefreshToken).Error())

		expiredCookie, err := loginState.Delete(store.LoginStates(), &srv.Config.Server)
		service.MustNotBeError(err)
		if strings.HasPrefix(srv.Config.Auth.CallbackURL, "https") {
			expiredCookie.Secure = true
		}
		http.SetCookie(w, expiredCookie)

		return nil
	}))

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(map[string]interface{}{
		"access_token": token.AccessToken,
		"expires_in":   time.Until(token.Expiry).Round(time.Second) / time.Second,
	})))
	return service.NoError
}

func createOrUpdateUser(s *database.UserStore, userData map[string]interface{}, domainConfig *domain.Configuration) int64 {
	var groupID int64
	err := s.WithWriteLock().
		Where("login_id = ?", userData["login_id"]).PluckFirst("group_id", &groupID).Error()

	userData["latest_login_at"] = database.Now()
	userData["latest_activity_at"] = database.Now()

	if defaultLanguage, ok := userData["default_language"]; ok && defaultLanguage == nil {
		userData["default_language"] = database.Default()
	}

	if gorm.IsRecordNotFoundError(err) {
		selfGroupID := createGroupsFromLogin(s.Groups(), userData["login"].(string), domainConfig)
		userData["temp_user"] = 0
		userData["registered_at"] = database.Now()
		userData["group_id"] = selfGroupID

		service.MustNotBeError(s.Users().InsertMap(userData))
		return selfGroupID
	}

	found, err := s.GroupGroups().WithWriteLock().Where("parent_group_id = ?", domainConfig.RootSelfGroupID).
		Where("child_group_id = ?", groupID).HasRows()
	service.MustNotBeError(err)
	groupsToCreate := make([]database.ParentChild, 0, 2)
	if !found {
		groupsToCreate = append(groupsToCreate,
			database.ParentChild{ParentID: domainConfig.RootSelfGroupID, ChildID: groupID})
	}

	service.MustNotBeError(s.GroupGroups().CreateRelationsWithoutChecking(groupsToCreate))

	service.MustNotBeError(err)
	service.MustNotBeError(s.ByID(groupID).UpdateColumn(userData).Error())
	return groupID
}

func createGroupsFromLogin(store *database.GroupStore, login string, domainConfig *domain.Configuration) (selfGroupID int64) {
	service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryIDStore *database.DataStore) error {
		selfGroupID = retryIDStore.NewID()
		return retryIDStore.Groups().InsertMap(map[string]interface{}{
			"id":          selfGroupID,
			"name":        login,
			"type":        "User",
			"description": login,
			"created_at":  database.Now(),
			"opened":      false,
			"send_emails": false,
		})
	}))

	service.MustNotBeError(store.GroupGroups().CreateRelationsWithoutChecking([]database.ParentChild{
		{ParentID: domainConfig.RootSelfGroupID, ChildID: selfGroupID},
	}))

	return selfGroupID
}
