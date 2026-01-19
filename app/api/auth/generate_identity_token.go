package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

const identityTokenLifetime = 2 * time.Hour

// swagger:operation POST /auth/identity-token auth generateIdentityToken
//
//	---
//	summary: Generate an identity token
//	description: >
//		Generates a signed JWS token that allows external services to verify the user's identity.
//		The token contains the user's ID and an expiration time (2 hours from generation).
//
//	responses:
//		"201":
//			"$ref": "#/responses/identityTokenResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) generateIdentityToken(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)

	expiresIn := int32(identityTokenLifetime / time.Second)
	expirationTime := time.Now().Add(identityTokenLifetime)

	identityToken, err := (&token.Token[payloads.IdentityToken]{Payload: payloads.IdentityToken{
		UserID:     strconv.FormatInt(user.GroupID, 10),
		IsTempUser: user.IsTempUser,
		Exp:        expirationTime.Unix(),
	}}).Sign(srv.TokenConfig.PrivateKey)
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.CreationSuccess(map[string]interface{}{
		"identity_token": identityToken,
		"expires_in":     expiresIn,
	})))

	return nil
}
