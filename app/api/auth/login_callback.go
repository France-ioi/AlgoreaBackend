package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /auth/login-callback users auth userLoginCallback
// ---
// summary: Callback to which the user is redirected after authentication with the login module
// description: Creates or updates the authenticated user in the DB using the data returned by the login module,
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

	userProfile, apiErr := srv.retrieveUserProfile(r, token)
	if apiErr != service.NoError {
		return apiErr
	}

	domainConfig := domain.ConfigFromContext(r.Context())

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		userID := createOrUpdateUser(store.Users(), userProfile, domainConfig)
		service.MustNotBeError(store.Sessions().InsertMap(map[string]interface{}{
			"sAccessToken":    token.AccessToken,
			"sExpirationDate": token.Expiry.UTC(),
			"idUser":          userID,
			"sIssuer":         "login-module",
			"sIssuedAtDate":   database.Now(),
		}))

		service.MustNotBeError(store.Exec(
			"INSERT INTO refresh_tokens (idUser, sRefreshToken) VALUES (?, ?) ON DUPLICATE KEY UPDATE sRefreshToken = ?",
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

func (srv *Service) retrieveUserProfile(r *http.Request, token *oauth2.Token) (map[string]interface{}, service.APIError) {
	request, err := http.NewRequest("GET", srv.Config.Auth.LoginModuleURL+"/user_api/account", nil)
	service.MustNotBeError(err)
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	request = request.WithContext(r.Context())
	response, err := http.DefaultClient.Do(request)
	service.MustNotBeError(err)
	body, err := ioutil.ReadAll(io.LimitReader(response.Body, 1<<20)) // 1Mb
	_ = response.Body.Close()
	service.MustNotBeError(err)
	if response.StatusCode != http.StatusOK {
		logging.Warnf("Can't retrieve user's profile (status code = %d, response = %q, accessToken = %q)",
			response.StatusCode, body, token.AccessToken)
		return nil, service.ErrUnexpected(fmt.Errorf("can't retrieve user's profile (status code = %d)", response.StatusCode))
	}
	var decoded map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	err = decoder.Decode(&decoded)
	if err != nil {
		logging.Warnf("Can't parse user's profile (response = %q, error = %s, accessToken = %q)",
			body, err, token.AccessToken)
		return nil, service.ErrUnexpected(errors.New("can't parse user's profile"))
	}

	converted, err := convertUserProfile(decoded, strings.SplitN(r.RemoteAddr, ":", 2)[0])
	if err != nil {
		logging.Warnf("User's profile is invalid (response = %q, error = %s, accessToken = %q)",
			body, err, token.AccessToken)
		return nil, service.ErrUnexpected(errors.New("user's profile is invalid"))
	}
	return converted, service.NoError
}

func convertUserProfile(source map[string]interface{}, remoteAddr string) (map[string]interface{}, error) {
	dest := make(map[string]interface{}, len(source)+2)
	mapping := map[string]string{
		"loginID":          "id", // unsigned int
		"sLogin":           "login",
		"sEmail":           "primary_email",
		"sFirstName":       "first_name",
		"sLastName":        "last_name",
		"sStudentId":       "student_id",
		"sCountryCode":     "country_code",
		"sBirthDate":       "birthday",
		"iGraduationYear":  "graduation_year",  // int
		"iGrade":           "graduation_grade", // int
		"sAddress":         "address",
		"sZipcode":         "zipcode",
		"sCity":            "city",
		"sLandLineNumber":  "primary_phone",
		"sCellPhoneNumber": "secondary_phone",
		"sDefaultLanguage": "language",
		"sFreeText":        "presentation",
		"sWebSite":         "website",
		"bEmailVerified":   "primary_email_verified",
	}
	for destKey, sourceKey := range mapping {
		dest[destKey] = source[sourceKey]
		if number, ok := dest[destKey].(json.Number); ok {
			dest[destKey], _ = number.Int64()
		}
	}
	dest["sSex"] = nil
	switch source["gender"] {
	case "m":
		dest["sSex"] = "Male"
	case "f":
		dest["sSex"] = "Female"
	}
	dest["bEmailVerified"] = (dest["bEmailVerified"] == true) || (dest["bEmailVerified"] == int64(1))
	if countryCode, ok := dest["sCountryCode"].(string); ok {
		dest["sCountryCode"] = strings.ToLower(countryCode)
	} else {
		dest["sCountryCode"] = ""
	}
	dest["sLastIP"] = remoteAddr

	if dest["loginID"] == nil {
		return dest, errors.New("no id in user's profile")
	}
	fmt.Printf("LoginID type: %t", dest["loginID"])

	if _, ok := dest["sLogin"].(string); !ok {
		return dest, errors.New("no login in user's profile")
	}

	if dest["iGraduationYear"] == nil {
		dest["iGraduationYear"] = 0
	}

	return dest, nil
}

func createOrUpdateUser(s *database.UserStore, userData map[string]interface{}, domainConfig *domain.Configuration) int64 {
	var userID int64
	err := s.WithWriteLock().
		Where("loginID = ?", userData["loginID"]).PluckFirst("ID", &userID).Error()

	userData["sLastLoginDate"] = database.Now()
	userData["sLastActivityDate"] = database.Now()

	if gorm.IsRecordNotFoundError(err) {
		ownedGroupID, selfGroupID := createGroupsFromLogin(s.Groups(), userData["sLogin"].(string), domainConfig)
		userData["tempUser"] = 0
		userData["sRegistrationDate"] = database.Now()
		userData["idGroupSelf"] = selfGroupID
		userData["idGroupOwned"] = ownedGroupID

		service.MustNotBeError(s.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
			userID = s.NewID()
			userData["ID"] = userID
			return s.Users().InsertMap(userData)
		}))
		return userID
	}

	service.MustNotBeError(err)
	service.MustNotBeError(s.ByID(userID).UpdateColumn(userData).Error())
	return userID
}

func createGroupsFromLogin(store *database.GroupStore, login string, domainConfig *domain.Configuration) (ownedGroupID, selfGroupID int64) {
	service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryIDStore *database.DataStore) error {
		selfGroupID = retryIDStore.NewID()
		return retryIDStore.Groups().InsertMap(map[string]interface{}{
			"ID":           selfGroupID,
			"sName":        login,
			"sType":        "UserSelf",
			"sDescription": login,
			"sDateCreated": database.Now(),
			"bOpened":      false,
			"bSendEmails":  false,
		})
	}))
	service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryIDStore *database.DataStore) error {
		ownedGroupID = retryIDStore.NewID()
		adminGroupName := login + "-admin"
		return retryIDStore.Groups().InsertMap(map[string]interface{}{
			"ID":           ownedGroupID,
			"sName":        adminGroupName,
			"sType":        "UserAdmin",
			"sDescription": adminGroupName,
			"sDateCreated": database.Now(),
			"bOpened":      false,
			"bSendEmails":  false,
		})
	}))

	service.MustNotBeError(store.GroupGroups().CreateRelationsWithoutChecking([]database.ParentChild{
		{ParentID: domainConfig.RootSelfGroupID, ChildID: selfGroupID},
		{ParentID: domainConfig.RootAdminGroupID, ChildID: ownedGroupID},
	}))

	return ownedGroupID, selfGroupID
}
