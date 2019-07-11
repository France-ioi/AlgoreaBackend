package testhelpers

import (
	"net/http/httptest"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

func (ctx *TestContext) RequestHeadersAre(data *gherkin.DataTable) error { // nolint
	for _, row := range data.Rows {
		headerName := row.Cells[0].Value
		if ctx.requestHeaders == nil {
			ctx.requestHeaders = make(map[string][]string, len(data.Rows))
		}
		ctx.requestHeaders[headerName] = append(ctx.requestHeaders[headerName], row.Cells[1].Value)
	}
	return nil
}

func (ctx *TestContext) ISendrequestToWithBody(method string, path string, body *gherkin.DocString) error { // nolint
	return ctx.iSendrequestGeneric(method, path, body.Content)
}

func (ctx *TestContext) ISendrequestTo(method string, path string) error { // nolint
	return ctx.iSendrequestGeneric(method, path, "")
}

func (ctx *TestContext) iSendrequestGeneric(method, path, reqBody string) error {
	// app server
	testServer := httptest.NewServer(ctx.application.HTTPHandler)
	defer testServer.Close()

	// auth proxy server
	authProxyServer := ctx.setupAuthProxyServer()
	defer authProxyServer.Close()

	reqBody, err := ctx.preprocessString(reqBody)
	if err != nil {
		return err
	}

	// do request
	response, body, err := testRequest(testServer, method, path, ctx.requestHeaders, strings.NewReader(reqBody))
	if err != nil {
		return err
	}
	ctx.lastResponse = response
	ctx.lastResponseBody = body

	return nil
}
