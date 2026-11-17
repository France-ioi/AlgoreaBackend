package auth

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /auth/contestant auth contestantAccountCreate
//
//	---
//	summary: Create a contestant account and sign that contestant in
//	description: >
//		Asks the login module for one contestant account and a bridge code, then exchanges that code
//		the same way `POST /auth/token` exchanges a coordinator's authorization code, and sets the session cookie.
//
//		No participation code is required. No session is required: a temporary session cookie, or no session, is accepted.
//		The client secret and any password stay on the server. The response data is the contestant's login and `expires_in`.
//		When `{use_cookie}` is not 1, the access token is returned in the data instead of a cookie, as for `POST /auth/token`.
//
//		The cookie's `SameSite` attribute is always `Strict`.
//		If `{cookie_secure}` is not set explicitly to `0`, the cookie is also issued with the `Secure` attribute.
//	security: []
//	parameters:
//		- name: use_cookie
//			in: query
//			description: If 1, set a cookie instead of returning the access token in the data
//			type: integer
//			enum: [0,1]
//			default: 0
//		- name: cookie_secure
//			in: query
//			description: >
//				Whether to set the cookie with the `Secure` attribute.
//				Defaults to 1; set to 0 only when TLS is unavailable (e.g. local development over http://localhost).
//			type: integer
//			enum: [0,1]
//			default: 1
//	responses:
//		"201":
//			description: Created. The contestant is signed in.
//			schema:
//				type: object
//				required: [success, message, data]
//				properties:
//					success:
//						type: boolean
//						enum: [true]
//					message:
//						type: string
//						enum: [created]
//					data:
//						type: object
//						required: [login, expires_in]
//						properties:
//							login:
//								type: string
//							expires_in:
//								type: integer
//								format: int32
//							access_token:
//								type: string
//								description: Only if the cookie is not enabled
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) createContestant(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	cookieAttributes, err := srv.resolveSessionCookieAttributesFromRequest(httpRequest)
	service.MustNotBeError(err)

	bridge, err := loginmodule.NewClient(srv.AuthConfig.GetString("loginModuleURL")).
		CreateContestant(httpRequest.Context(), srv.AuthConfig.GetString("clientID"), srv.AuthConfig.GetString("clientSecret"))
	if err != nil {
		return service.ErrUnexpected(err)
	}

	accessToken, expiresIn, login := srv.createSessionFromAuthorizationCode(
		httpRequest, bridge.Code, nil, &bridge.RedirectURI)
	srv.respondWithNewAccessToken(
		responseWriter, httpRequest, service.CreationSuccess[map[string]any], accessToken, expiresIn, cookieAttributes,
		map[string]any{"login": login})
	return nil
}
