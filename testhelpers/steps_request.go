//go:build !prod

package testhelpers

import (
	"net/http/httptest"
	"strings"

	"github.com/cucumber/godog"
)

func (ctx *TestContext) TheRequestHeaderIs(name, value string) error { //nolint
	value, err := ctx.preprocessString(value)
	if err != nil {
		return err
	}

	if value == undefinedHeaderValue {
		return nil
	}

	if ctx.requestHeaders == nil {
		ctx.requestHeaders = make(map[string][]string)
	}

	ctx.requestHeaders[name] = append(ctx.requestHeaders[name], value)
	return nil
}

func (ctx *TestContext) ISendrequestToWithBody(method string, path string, body *godog.DocString) error { // nolint
	return ctx.iSendrequestGeneric(method, path, body.Content)
}

func (ctx *TestContext) ISendrequestTo(method string, path string) error { //nolint
	return ctx.iSendrequestGeneric(method, path, "")
}

func (ctx *TestContext) iSendrequestGeneric(method, path, reqBody string) error {
	// put all data into the database before we send the request
	err := ctx.populateDatabase()
	if err != nil {
		return err
	}

	// app server
	testServer := httptest.NewServer(ctx.application.HTTPHandler)
	defer testServer.Close()

	var headers map[string][]string
	if ctx.userID != 0 {
		headers = make(map[string][]string, len(ctx.requestHeaders)+1)
		for key := range ctx.requestHeaders {
			headers[key] = make([]string, len(ctx.requestHeaders[key]))
			copy(headers[key], ctx.requestHeaders[key])
		}
		headers["Authorization"] = []string{"Bearer " + testAccessToken}
	} else {
		headers = ctx.requestHeaders
	}

	reqBody, err = ctx.preprocessString(reqBody)
	if err != nil {
		return err
	}

	path, err = ctx.preprocessString(path)
	if err != nil {
		return err
	}

	// do request
	response, body, err := testRequest(testServer, method, path, headers, strings.NewReader(reqBody))
	defer func() { _ = response.Body.Close() }()
	if err != nil {
		return err
	}
	ctx.lastResponse = response
	ctx.lastResponseBody = body

	return nil
}
