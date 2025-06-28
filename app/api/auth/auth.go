// Package auth provides API services related to authentication.
package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"golang.org/x/oauth2"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// Service is the mount point for services related to `auth`.
type Service struct {
	*service.Base
}

// SetRoutes defines the routes for this package in a route group.
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Post("/auth/temp-user", service.AppHandler(srv.createTempUser).ServeHTTP)

	router.With(middleware.AllowContentType("", "application/json", "application/x-www-form-urlencoded")).
		Post("/auth/token", service.AppHandler(srv.createAccessToken).ServeHTTP)
	router.With(auth.UserMiddleware(srv.Base)).
		Post("/auth/logout", service.AppHandler(srv.logout).ServeHTTP)
}

func validateAndGetExpiresInFromOAuth2Token(token *oauth2.Token) (expiresIn int32, err error) {
	if !token.Valid() {
		return 0, &service.APIError{
			HTTPStatusCode: http.StatusUnauthorized,
			EmbeddedError:  errors.New("got an invalid OAuth2 token"),
		}
	}

	expiresIn64 := int64(time.Until(token.Expiry).Round(time.Second) / time.Second)
	//nolint:gosec // G115: The oauth2 package guarantees that the "expires_in" value is always <= 2^31-1.
	// Also, the "expires_in" value is always greater than 0 in valid tokens.
	return int32(expiresIn64), nil
}
