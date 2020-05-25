// +build !prod

package testhelpers

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"

	"github.com/cucumber/messages-go/v10"
	"github.com/spf13/viper"
	"github.com/thingful/httpmock"

	"github.com/France-ioi/AlgoreaBackend/app"
)

func (ctx *TestContext) appAuthConfig() *viper.Viper {
	return app.AuthConfig(ctx.application.Config)
}

func (ctx *TestContext) TheLoginModuleTokenEndpointForCodeReturns(code string, statusCode int, body *messages.PickleStepArgument_PickleDocString) error { // nolint
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
		"client_id":     {ctx.appAuthConfig().GetString("ClientID")},
		"client_secret": {ctx.appAuthConfig().GetString("ClientSecret")},
		"grant_type":    {"authorization_code"},
		"code":          {preprocessedCode},
	}
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.appAuthConfig().GetString("LoginModuleURL")+"/oauth/token", responder,
		httpmock.WithBody(
			bytes.NewBufferString(params.Encode()))))
	return nil
}

func (ctx *TestContext) TheLoginModuleTokenEndpointForCodeAndCodeVerifierReturns(code, codeVerifier string, statusCode int, body *messages.PickleStepArgument_PickleDocString) error { // nolint
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedCode, err := ctx.preprocessString(code)
	if err != nil {
		return err
	}
	preprocessedCodeVerifier, err := ctx.preprocessString(codeVerifier)
	if err != nil {
		return err
	}
	preprocessedBody, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}
	responder := httpmock.NewStringResponder(statusCode, preprocessedBody)
	authConfig := app.AuthConfig(ctx.application.Config)
	params := url.Values{
		"client_id":     {authConfig.GetString("ClientID")},
		"client_secret": {authConfig.GetString("ClientSecret")},
		"grant_type":    {"authorization_code"},
		"code":          {preprocessedCode},
		"code_verifier": {preprocessedCodeVerifier},
	}
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		authConfig.GetString("LoginModuleURL")+"/oauth/token", responder,
		httpmock.WithBody(
			bytes.NewBufferString(params.Encode()))))
	return nil
}

func (ctx *TestContext) TheLoginModuleTokenEndpointForRefreshTokenReturns(refreshToken string, statusCode int, body *messages.PickleStepArgument_PickleDocString) error { // nolint
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
		"client_id":     {ctx.appAuthConfig().GetString("ClientID")},
		"client_secret": {ctx.appAuthConfig().GetString("ClientSecret")},
		"grant_type":    {"refresh_token"},
		"refresh_token": {preprocessedRefreshToken},
	}
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.appAuthConfig().GetString("LoginModuleURL")+"/oauth/token", responder,
		httpmock.WithBody(
			bytes.NewBufferString(params.Encode()))))
	return nil
}

func (ctx *TestContext) TheLoginModuleAccountEndpointForTokenReturns(token string, statusCode int, body *messages.PickleStepArgument_PickleDocString) error { // nolint
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
		ctx.appAuthConfig().GetString("LoginModuleURL")+"/user_api/account", responder,
		httpmock.WithHeader(&http.Header{"Authorization": {"Bearer " + preprocessedToken}})))
	return nil
}

func (ctx *TestContext) TheLoginModuleUnlinkClientEndpointForUserIDReturns( // nolint
	userID string, statusCode int, body *messages.PickleStepArgument_PickleDocString) error {
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedUserID, err := ctx.preprocessString(userID)
	if err != nil {
		return err
	}
	preprocessedBody, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}
	bodyBase64, err := ctx.encodeLoginModuleResponse(preprocessedBody)
	if err != nil {
		return err
	}

	responder := httpmock.NewStringResponder(statusCode, bodyBase64)
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.appAuthConfig().GetString("LoginModuleURL")+"/platform_api/accounts_manager/unlink_client?client_id="+
			url.QueryEscape(ctx.appAuthConfig().GetString("ClientID"))+"&user_id="+url.QueryEscape(preprocessedUserID), responder))
	return nil
}

func (ctx *TestContext) encodeLoginModuleResponse(preprocessedBody string) (string, error) {
	const size = 16
	mod := len(preprocessedBody) % size
	if mod != 0 {
		padding := byte(size - mod)
		preprocessedBody += strings.Repeat(string(padding), int(padding))
	}

	data := []byte(preprocessedBody)
	cipher, err := aes.NewCipher([]byte(ctx.appAuthConfig().GetString("ClientSecret"))[0:16])
	if err != nil {
		return "", err
	}
	encrypted := make([]byte, len(data))
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cipher.Encrypt(encrypted[bs:be], data[bs:be])
	}

	bodyBase64 := base64.StdEncoding.EncodeToString(encrypted)
	return bodyBase64, nil
}

func (ctx *TestContext) TheLoginModuleCreateEndpointWithParamsReturns( // nolint
	params string, statusCode int, body *messages.PickleStepArgument_PickleDocString) error {
	return ctx.theLoginModuleAccountsManagerEndpointWithParamsReturns("create", params, statusCode, body)
}

func (ctx *TestContext) TheLoginModuleDeleteEndpointWithParamsReturns( // nolint
	params string, statusCode int, body *messages.PickleStepArgument_PickleDocString) error {
	return ctx.theLoginModuleAccountsManagerEndpointWithParamsReturns("delete", params, statusCode, body)
}

func (ctx *TestContext) theLoginModuleAccountsManagerEndpointWithParamsReturns( // nolint
	endpoint, params string, statusCode int, body *messages.PickleStepArgument_PickleDocString) error {
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedParams, err := ctx.preprocessString(params)
	if err != nil {
		return err
	}
	preprocessedBody, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}
	bodyBase64, err := ctx.encodeLoginModuleResponse(preprocessedBody)
	if err != nil {
		return err
	}

	urlValues, err := url.ParseQuery("client_id=" + url.QueryEscape(ctx.appAuthConfig().GetString("ClientID")) + "&" + preprocessedParams)
	if err != nil {
		return err
	}
	responder := httpmock.NewStringResponder(statusCode, bodyBase64)
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.appAuthConfig().GetString("LoginModuleURL")+"/platform_api/accounts_manager/"+endpoint+"?"+urlValues.Encode(), responder))
	return nil
}
