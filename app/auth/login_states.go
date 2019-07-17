package auth

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

const loginStateLifetimeInSeconds = int32(2 * time.Hour / time.Second) // 2 hours (7200 seconds)
const loginCsrfCookieName = "login_csrf"

// SetNewLoginStateCookie creates a new cookie/state pair for the login process and sets the cookie.
// Returns the generated state value.
func SetNewLoginStateCookie(s *database.LoginStateStore, conf *config.Server, w http.ResponseWriter) (string, error) {
	var state string
	state, err := GenerateRandomString()
	if err != nil {
		return "", err
	}
	var cookie string
	err = s.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
		cookie, err = GenerateRandomString()
		if err != nil {
			return err
		}
		return retryStore.LoginStates().InsertMap(map[string]interface{}{
			"sCookie":         cookie,
			"sState":          state,
			"sExpirationDate": gorm.Expr("NOW() + INTERVAL ? SECOND", loginStateLifetimeInSeconds),
		})
	})
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:    loginCsrfCookieName,
		Value:   cookie,
		Expires: time.Now().Add(time.Duration(loginStateLifetimeInSeconds) * time.Second),
		MaxAge:  int(loginStateLifetimeInSeconds),
		Domain:  conf.Domain, Path: conf.RootPath,
		HttpOnly: true,
	})
	return state, nil
}
