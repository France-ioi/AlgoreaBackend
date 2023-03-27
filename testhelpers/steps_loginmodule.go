//go:build !prod

package testhelpers

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/cucumber/messages-go/v10"
	"github.com/spf13/viper"
	"github.com/thingful/httpmock"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/loginmodule"
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

func (ctx *TestContext) TheLoginModuleTokenEndpointForCodeAndCodeVerifierAndRedirectURIReturns(code, codeVerifier, redirectURI string, statusCode int, body *messages.PickleStepArgument_PickleDocString) error { // nolint
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedCode, err := ctx.preprocessString(code)
	if err != nil {
		return err
	}
	preprocessedCodeVerifier, err := ctx.preprocessString(codeVerifier)
	if err != nil {
		return err
	}
	preprocessedRedirectURI, err := ctx.preprocessString(redirectURI)
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
		"redirect_uri":  {preprocessedRedirectURI},
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

	clientSecret := ctx.appAuthConfig().GetString("clientSecret")
	encodedResponseBody := loginmodule.Encode([]byte(preprocessedBody), clientSecret)

	responder := httpmock.NewStringResponder(statusCode, encodedResponseBody)
	requestBody, err := loginmodule.EncodeBody(map[string]string{"user_id": preprocessedUserID},
		ctx.appAuthConfig().GetString("clientID"), clientSecret)
	if err != nil {
		return err
	}
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.appAuthConfig().GetString("loginModuleURL")+"/platform_api/accounts_manager/unlink_client",
		responder, httpmock.WithHeader(&http.Header{"Content-Type": []string{"application/json"}}),
		httpmock.WithBody(bytes.NewReader(requestBody))))
	return nil
}

func (ctx *TestContext) TheLoginModuleLTIResultSendEndpointForUserIDContentIDScoreReturns( // nolint
	userID, contentID, score string, statusCode int, body *messages.PickleStepArgument_PickleDocString) error {
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedUserID, err := ctx.preprocessString(userID)
	if err != nil {
		return err
	}
	preprocessedConentID, err := ctx.preprocessString(contentID)
	if err != nil {
		return err
	}
	preprocessedScore, err := ctx.preprocessString(score)
	if err != nil {
		return err
	}
	preprocessedBody, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}

	clientSecret := ctx.appAuthConfig().GetString("clientSecret")
	encodedResponseBody := loginmodule.Encode([]byte(preprocessedBody), clientSecret)

	responder := httpmock.NewStringResponder(statusCode, encodedResponseBody)
	requestBody, err := loginmodule.EncodeBody(
		map[string]string{
			"user_id":    preprocessedUserID,
			"content_id": preprocessedConentID,
			"score":      preprocessedScore,
		},
		ctx.appAuthConfig().GetString("clientID"), clientSecret)
	if err != nil {
		return err
	}
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.appAuthConfig().GetString("loginModuleURL")+"/platform_api/lti_result/send",
		responder, httpmock.WithHeader(&http.Header{"Content-Type": []string{"application/json"}}),
		httpmock.WithBody(bytes.NewReader(requestBody))))
	return nil
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
	clientSecret := ctx.appAuthConfig().GetString("clientSecret")
	encodedResponseBody := loginmodule.Encode([]byte(preprocessedBody), clientSecret)

	parsedParams, err := url.ParseQuery(preprocessedParams)
	if err != nil {
		return err
	}
	paramsMap := make(map[string]string, len(parsedParams))
	for key := range parsedParams {
		paramsMap[key] = parsedParams.Get(key)
	}
	requestBody, err := loginmodule.EncodeBody(paramsMap, ctx.appAuthConfig().GetString("clientID"), clientSecret)
	if err != nil {
		return err
	}

	responder := httpmock.NewStringResponder(statusCode, encodedResponseBody)
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.appAuthConfig().GetString("LoginModuleURL")+"/platform_api/accounts_manager/"+endpoint,
		responder, httpmock.WithHeader(&http.Header{"Content-Type": []string{"application/json"}}),
		httpmock.WithBody(bytes.NewReader(requestBody))))
	return nil
}
