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

// CreateLoginState creates a new cookie/state pair for the login process and stores it into the DB.
// Returns the generated cookie/state pair.
func CreateLoginState(s *database.LoginStateStore, conf *config.Server) (*http.Cookie, string, error) {
	var state string
	state, err := GenerateKey()
	if err != nil {
		return nil, "", err
	}
	var cookie string
	err = s.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
		cookie, err = GenerateKey()
		if err != nil {
			return err
		}
		return retryStore.LoginStates().InsertMap(map[string]interface{}{
			"sCookie":         cookie,
			"sState":          state,
			"sExpirationDate": gorm.Expr("? + INTERVAL ? SECOND", database.Now(), loginStateLifetimeInSeconds),
		})
	})
	if err != nil {
		return nil, "", err
	}
	return &http.Cookie{
		Name:    loginCsrfCookieName,
		Value:   cookie,
		Expires: time.Now().Add(time.Duration(loginStateLifetimeInSeconds) * time.Second),
		MaxAge:  int(loginStateLifetimeInSeconds),
		Domain:  conf.Domain, Path: conf.RootPath,
		HttpOnly: true,
	}, state, nil
}
