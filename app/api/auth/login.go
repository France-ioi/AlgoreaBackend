package auth

import (
	"errors"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /auth/login auth authLogin
// ---
// summary: Redirect with appropriate security parameters to the login module so that the user can log in
// description: Generates a state value (a random 32-char [0-9a-z] string) and a random cookie (a random 32-char [0-9a-z] string).
//              Stores the pair with an expiration time (2 hours) into the DB and sets the cookie value into the login_csrf cookie
//              while the state value goes to the login module as a 'state' URL parameter.
//
//   * No “Authorization” header should be present
// responses:
//   "302":
//     description: "Found. Success redirect to the login module's /oauth/authorize handler with needed parameters"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) login(w http.ResponseWriter, r *http.Request) service.APIError {
	if len(r.Header["Authorization"]) != 0 {
		return service.ErrInvalidRequest(errors.New("'Authorization' header should not be present"))
	}

	cookie, state, err := auth.CreateLoginState(srv.Store.LoginStates(), &srv.Config.Server)
	service.MustNotBeError(err)
	if strings.HasPrefix(srv.Config.Auth.CallbackURL, "https") {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)

	redirectURL := getOAuthConfig(&srv.Config.Auth).AuthCodeURL(state, oauth2.SetAuthURLParam("approval_prompt", "auto"))
	http.Redirect(w, r, redirectURL, http.StatusFound)
	return service.NoError
}
