package testhelpers

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/thingful/httpmock"
)

func (ctx *TestContext) TheLoginModuleTokenEndpointForCodeReturns(code string, statusCode int, body *gherkin.DocString) error { // nolint
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedCode, err := ctx.preprocessString(code)
	if err != nil {
		return err
	}
	preprocessedBody, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}
	responder := httpmock.NewStringResponder(statusCode, preprocessedBody)
	params := url.Values{
		"client_id":     {ctx.application.Config.Auth.ClientID},
		"client_secret": {ctx.application.Config.Auth.ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {preprocessedCode},
		"redirect_uri":  {ctx.application.Config.Auth.CallbackURL},
	}
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.application.Config.Auth.LoginModuleURL+"/oauth/token", responder,
		httpmock.WithBody(
			bytes.NewBufferString(params.Encode()))))
	return nil
}

func (ctx *TestContext) TheLoginModuleTokenEndpointForRefreshTokenReturns(refreshToken string, statusCode int, body *gherkin.DocString) error { // nolint
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedRefreshToken, err := ctx.preprocessString(refreshToken)
	if err != nil {
		return err
	}
	preprocessedBody, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}
	responder := httpmock.NewStringResponder(statusCode, preprocessedBody)
	params := url.Values{
		"client_id":     {ctx.application.Config.Auth.ClientID},
		"client_secret": {ctx.application.Config.Auth.ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {preprocessedRefreshToken},
	}
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.application.Config.Auth.LoginModuleURL+"/oauth/token", responder,
		httpmock.WithBody(
			bytes.NewBufferString(params.Encode()))))
	return nil
}

func (ctx *TestContext) TheLoginModuleAccountEndpointForTokenReturns(token string, statusCode int, body *gherkin.DocString) error { // nolint
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedToken, err := ctx.preprocessString(token)
	if err != nil {
		return err
	}
	preprocessedBody, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}
	responder := httpmock.NewStringResponder(statusCode, preprocessedBody)
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("GET",
		ctx.application.Config.Auth.LoginModuleURL+"/user_api/account", responder,
		httpmock.WithHeader(&http.Header{"Authorization": {"Bearer " + preprocessedToken}})))
	return nil
}

func (ctx *TestContext) TheLoginModuleUnlinkClientEndpointForUserIDReturns( // nolint
	userID string, statusCode int, body *gherkin.DocString) error {
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedUserID, err := ctx.preprocessString(userID)
	if err != nil {
		return err
	}
	preprocessedBody, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}
	const size = 16
	mod := len(preprocessedBody) % size
	if mod != 0 {
		padding := byte(size - mod)
		preprocessedBody += strings.Repeat(string(padding), int(padding))
	}

	data := []byte(preprocessedBody)
	cipher, err := aes.NewCipher([]byte(ctx.application.Config.Auth.ClientSecret)[0:16])
	if err != nil {
		return err
	}
	encrypted := make([]byte, len(data))
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cipher.Encrypt(encrypted[bs:be], data[bs:be])
	}

	bodyBase64 := base64.StdEncoding.EncodeToString(encrypted)

	responder := httpmock.NewStringResponder(statusCode, bodyBase64)
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.application.Config.Auth.LoginModuleURL+"/platform_api/accounts_manager/unlink_client?client_id="+
			url.QueryEscape(ctx.application.Config.Auth.ClientID)+"&user_id="+url.QueryEscape(preprocessedUserID), responder))
	return nil
}
