//go:build !prod

package testhelpers

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/cucumber/godog"
	"github.com/spf13/viper"
	"github.com/thingful/httpmock"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loginmodule"
)

func (ctx *TestContext) appAuthConfig() *viper.Viper {
	return app.AuthConfig(ctx.application.Config)
}

// TheLoginModuleTokenEndpointForCodeReturns mocks the return of the login module /oauth/token,
// called with a provided code.
func (ctx *TestContext) TheLoginModuleTokenEndpointForCodeReturns(
	code string,
	statusCode int,
	body *godog.DocString,
) error {
	preprocessedCode := ctx.preprocessString(code)
	preprocessedBody := ctx.preprocessString(body.Content)
	params := url.Values{
		"grant_type": {"authorization_code"},
		"code":       {preprocessedCode},
	}
	ctx.stubOAuth2RequestToLoginModule(params, statusCode, preprocessedBody)
	return nil
}

func (ctx *TestContext) stubOAuth2RequestToLoginModule(requestParams url.Values, statusCode int, body string) {
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	responder := httpmock.NewStringResponder(statusCode, body)
	params := make(url.Values, len(requestParams))
	for key, values := range requestParams {
		params[key] = values
	}
	params.Set("client_id", ctx.appAuthConfig().GetString("ClientID"))
	params.Set("client_secret", ctx.appAuthConfig().GetString("ClientSecret"))
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
		ctx.appAuthConfig().GetString("LoginModuleURL")+"/oauth/token", responder,
		httpmock.WithBody(
			bytes.NewBufferString(params.Encode()))))
}

// TheLoginModuleTokenEndpointForCodeAndCodeVerifierReturns mocks the return of the login module /oauth/token,
// called with the provided code and code_verifier.
func (ctx *TestContext) TheLoginModuleTokenEndpointForCodeAndCodeVerifierReturns(
	code, codeVerifier string,
	statusCode int,
	body *godog.DocString,
) error {
	preprocessedCode := ctx.preprocessString(code)
	preprocessedCodeVerifier := ctx.preprocessString(codeVerifier)
	preprocessedBody := ctx.preprocessString(body.Content)
	params := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {preprocessedCode},
		"code_verifier": {preprocessedCodeVerifier},
	}
	ctx.stubOAuth2RequestToLoginModule(params, statusCode, preprocessedBody)
	return nil
}

// TheLoginModuleTokenEndpointForCodeAndCodeVerifierAndRedirectURIReturns mocks the return of the login module /oauth/token,
// called with the provided code, code_verifier, and redirect_uri.
func (ctx *TestContext) TheLoginModuleTokenEndpointForCodeAndCodeVerifierAndRedirectURIReturns(
	code, codeVerifier, redirectURI string,
	statusCode int,
	body *godog.DocString,
) error {
	preprocessedCode := ctx.preprocessString(code)
	preprocessedCodeVerifier := ctx.preprocessString(codeVerifier)
	preprocessedRedirectURI := ctx.preprocessString(redirectURI)
	preprocessedBody := ctx.preprocessString(body.Content)
	params := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {preprocessedCode},
		"code_verifier": {preprocessedCodeVerifier},
		"redirect_uri":  {preprocessedRedirectURI},
	}
	ctx.stubOAuth2RequestToLoginModule(params, statusCode, preprocessedBody)
	return nil
}

// TheLoginModuleTokenEndpointForRefreshTokenReturns mocks the return of the login module /oauth/token,
// called with the provided refresh_token.
func (ctx *TestContext) TheLoginModuleTokenEndpointForRefreshTokenReturns(
	refreshToken string,
	statusCode int,
	body *godog.DocString,
) error {
	preprocessedRefreshToken := ctx.preprocessString(refreshToken)
	preprocessedBody := ctx.preprocessString(body.Content)
	params := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {preprocessedRefreshToken},
	}
	ctx.stubOAuth2RequestToLoginModule(params, statusCode, preprocessedBody)
	return nil
}

// TheLoginModuleAccountEndpointForTokenReturns mocks the return of the login module /user_api/account with
// the provided authorization token.
func (ctx *TestContext) TheLoginModuleAccountEndpointForTokenReturns(
	authToken string,
	statusCode int,
	body *godog.DocString,
) error {
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedToken := ctx.preprocessString(authToken)
	preprocessedBody := ctx.preprocessString(body.Content)
	responder := httpmock.NewStringResponder(statusCode, preprocessedBody)
	httpmock.RegisterStubRequests(httpmock.NewStubRequest("GET",
		ctx.appAuthConfig().GetString("LoginModuleURL")+"/user_api/account", responder,
		httpmock.WithHeader(&http.Header{"Authorization": {"Bearer " + preprocessedToken}})))
	return nil
}

// TheLoginModuleUnlinkClientEndpointForUserIDReturns mocks the return of the login module /platform_api/accounts_manager/unlink_client
// with the provided user_id.
func (ctx *TestContext) TheLoginModuleUnlinkClientEndpointForUserIDReturns(
	userID string, statusCode int, body *godog.DocString,
) error {
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedUserID := ctx.preprocessString(userID)
	preprocessedBody := ctx.preprocessString(body.Content)

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

// TheLoginModuleLTIResultSendEndpointForUserIDContentIDScoreReturns mocks the return of the login module
// /platform_api/lti_result/send with the provided user_id, content_id, and score.
func (ctx *TestContext) TheLoginModuleLTIResultSendEndpointForUserIDContentIDScoreReturns(
	userID, contentID, score string, statusCode int, body *godog.DocString,
) error {
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedUserID := ctx.preprocessString(userID)
	preprocessedContentID := ctx.preprocessString(contentID)
	preprocessedScore := ctx.preprocessString(score)
	preprocessedBody := ctx.preprocessString(body.Content)

	clientSecret := ctx.appAuthConfig().GetString("clientSecret")
	encodedResponseBody := loginmodule.Encode([]byte(preprocessedBody), clientSecret)

	responder := httpmock.NewStringResponder(statusCode, encodedResponseBody)
	requestBody, err := loginmodule.EncodeBody(
		map[string]string{
			"user_id":    preprocessedUserID,
			"content_id": preprocessedContentID,
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

// TheLoginModuleCreateEndpointWithParamsReturns mocks the return of the login module
// /platform_api/accounts_manager/create with the provided parameters, return status, and body.
func (ctx *TestContext) TheLoginModuleCreateEndpointWithParamsReturns(
	params string, statusCode int, body *godog.DocString,
) error {
	return ctx.theLoginModuleAccountsManagerEndpointWithParamsReturns("create", params, statusCode, body)
}

// TheLoginModuleDeleteEndpointWithParamsReturns mocks the return of the login module
// /platform_api/accounts_manager/delete with the provided parameters, return status, and body.
func (ctx *TestContext) TheLoginModuleDeleteEndpointWithParamsReturns(
	params string, statusCode int, body *godog.DocString,
) error {
	return ctx.theLoginModuleAccountsManagerEndpointWithParamsReturns("delete", params, statusCode, body)
}

func (ctx *TestContext) theLoginModuleAccountsManagerEndpointWithParamsReturns(
	endpoint, params string, statusCode int, body *godog.DocString,
) error {
	httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
	preprocessedParams := ctx.preprocessString(params)
	preprocessedBody := ctx.preprocessString(body.Content)
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
