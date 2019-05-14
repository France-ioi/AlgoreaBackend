package testhelpers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/pmezard/go-difflib/difflib"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (ctx *TestContext) ItShouldBeAJSONArrayWithEntries(count int) error { // nolint
	var objmap []map[string]*json.RawMessage

	if err := json.Unmarshal([]byte(ctx.lastResponseBody), &objmap); err != nil {
		return fmt.Errorf("unable to decode the response as JSON: %s\nData:%v", err, ctx.lastResponseBody)
	}

	if count != len(objmap) {
		return fmt.Errorf("the result does not have the expected length. Expected: %d, received: %d", count, len(objmap))
	}

	return nil
}

func (ctx *TestContext) TheResponseCodeShouldBe(code int) error { // nolint
	if code != ctx.lastResponse.StatusCode {
		return fmt.Errorf("expected http response code: %d, actual is: %d. \n Data: %s", code, ctx.lastResponse.StatusCode, ctx.lastResponseBody)
	}
	return nil
}

func (ctx *TestContext) TheResponseBodyShouldBeJSON(body *gherkin.DocString) (err error) { // nolint
	return ctx.TheResponseDecodedBodyShouldBeJSON("", body)
}

func (ctx *TestContext) TheResponseDecodedBodyShouldBeJSON(responseType string, body *gherkin.DocString) (err error) { // nolint
	// verify the content type
	if err = ValidateJSONContentType(ctx.lastResponse); err != nil {
		return
	}

	expectedBody, err := ctx.preprocessJSONBody(body.Content)
	if err != nil {
		return err
	}

	// re-encode expected response
	var exp interface{}
	err = json.Unmarshal([]byte(expectedBody), &exp)
	if err != nil {
		return err
	}
	var expected, actual []byte
	if expected, err = json.MarshalIndent(&exp, "", "\t"); err != nil {
		return err
	}

	var act interface{}
	if responseType == "" {
		var value interface{}
		act = &value
	} else {
		act, err = getZeroStructPtr(responseType)
		if err != nil {
			return err
		}
		reflect.ValueOf(act).Elem().FieldByName("PublicKey").Set(reflect.ValueOf(ctx.application.TokenConfig.PublicKey))
	}

	// re-encode actual response too
	if err = json.Unmarshal([]byte(ctx.lastResponseBody), act); err != nil {
		return fmt.Errorf("unable to decode the response as JSON: %s -- Data: %v", err, ctx.lastResponseBody)
	}

	if responseType != "" {
		act = payloads.ConvertIntoMap(act)
	}
	if actual, err = json.MarshalIndent(act, "", "\t"); err != nil {
		return
	}

	return compareStrings(string(expected), string(actual))
}

func compareStrings(expected, actual string) error {
	if expected != actual {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{ // nolint: gosec
			A:        difflib.SplitLines(expected),
			B:        difflib.SplitLines(actual),
			FromFile: "Expected",
			FromDate: "",
			ToFile:   "Actual",
			ToDate:   "",
			Context:  1,
		})

		return fmt.Errorf(
			"expected JSON does not match actual.\n     Diff:\n%s",
			diff,
		)
	}
	return nil
}

func (ctx *TestContext) TheResponseHeaderShouldBe(headerName string, headerValue string) (err error) { // nolint
	if ctx.lastResponse.Header.Get(headerName) != headerValue {
		return fmt.Errorf("headers %s different from expected. Expected: %s, got: %s",
			headerName, headerValue, ctx.lastResponse.Header.Get(headerName))
	}
	return nil
}

func (ctx *TestContext) TheResponseErrorMessageShouldContain(s string) (err error) { // nolint

	errorResp := service.ErrorResponse{}
	// decode response
	if err = json.Unmarshal([]byte(ctx.lastResponseBody), &errorResp); err != nil {
		return fmt.Errorf("unable to decode the response as JSON: %s -- Data: %v", err, ctx.lastResponseBody)
	}
	if !strings.Contains(errorResp.ErrorText, s) {
		return fmt.Errorf("cannot find expected `%s` in error text: `%s`", s, errorResp.ErrorText)
	}

	return nil
}

func (ctx *TestContext) TheResponseShouldBe(kind string) error { // nolint
	var expectedCode int
	switch kind {
	case "updated":
		expectedCode = 200
	case "created":
		expectedCode = 201
	default:
		return fmt.Errorf("unknown response kind: %q", kind)
	}
	if err := ctx.TheResponseCodeShouldBe(expectedCode); err != nil {
		return err
	}
	if err := ctx.TheResponseBodyShouldBeJSON(&gherkin.DocString{
		Content: `
		{
			"message": "` + kind + `",
			"success": true
		}`}); err != nil {
		return err
	}
	return nil
}
