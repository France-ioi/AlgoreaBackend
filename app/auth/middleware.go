package auth

import (
	"context"
	"encoding/json"
	"net/http"
)

type ctxKey int

const (
	ctxUserID ctxKey = iota
)

type authResponse struct {
	UserID int64  `json:"userID"`
	Error  string `json:"error"`
}

// UserIDMiddleware is a middleware retrieving user ID from the request content
// Created by giving the reverse proxy used for getting the auth info
func UserIDMiddleware(serviceURL string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authCookieName := "PHPSESSID"
			var authCookie *http.Cookie
			var err error

			if authCookie, err = r.Cookie(authCookieName); err != nil {
				http.Error(w, "Unable to get the expected auth cookie from the request", http.StatusUnauthorized)
				return
			}

			// create a new url from the raw RequestURI sent by the client
			cookieParam := "?sessionid=" + authCookie.Value
			authRequest, _ := http.NewRequest("GET", serviceURL+cookieParam, nil)

			httpClient := http.Client{}
			var resp *http.Response
			resp, err = httpClient.Do(authRequest)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}

			auth := &authResponse{}
			if err = json.NewDecoder(resp.Body).Decode(auth); err != nil {
				http.Error(w, "Unable to parse response for auth server: "+err.Error(), http.StatusBadGateway)
				return
			}
			if auth.Error != "" {
				http.Error(w, "Unable to validate the session ID: "+auth.Error, http.StatusUnauthorized)
				return
			}
			if auth.UserID <= 0 {
				http.Error(w, "Invalid response from auth server. No error by userID:"+string(auth.UserID), http.StatusBadGateway)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, auth.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserID returs the user id stored in the request context
func UserID(r *http.Request) int64 {
	return r.Context().Value(ctxUserID).(int64)
}
