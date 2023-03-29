//go:build !prod

package testhelpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/cucumber/messages-go/v10"
	"github.com/pmezard/go-difflib/difflib"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (ctx *TestContext) ItShouldBeAJSONArrayWithEntries(count int) error { //nolint
	var objmap []map[string]*json.RawMessage

	if err := json.Unmarshal([]byte(ctx.lastResponseBody), &objmap); err != nil {
		return fmt.Errorf("unable to decode the response as JSON: %s\nData:%v", err, ctx.lastResponseBody)
	}

	if count != len(objmap) {
		return fmt.Errorf("the result does not have the expected length. Expected: %d, received: %d", count, len(objmap))
	}

	return nil
}

func (ctx *TestContext) TheResponseCodeShouldBe(code int) error { //nolint
	if code != ctx.lastResponse.StatusCode {
		return fmt.Errorf("expected http response code: %d, actual is: %d. \n Data: %s", code, ctx.lastResponse.StatusCode, ctx.lastResponseBody)
	}
	return nil
}

func (ctx *TestContext) TheResponseBodyShouldBeJSON(body *messages.PickleStepArgument_PickleDocString) (err error) { // nolint
	return ctx.TheResponseDecodedBodyShouldBeJSON("", body)
}

func (ctx *TestContext) TheResponseDecodedBodyShouldBeJSON(responseType string, body *messages.PickleStepArgument_PickleDocString) (err error) { // nolint
	// verify the content type
	if err = ValidateJSONContentType(ctx.lastResponse); err != nil {
		return
	}

	expectedBody, err := ctx.preprocessString(body.Content)
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
		config, _ := app.TokenConfig(ctx.application.Config)
		reflect.ValueOf(act).Elem().FieldByName("PublicKey").Set(reflect.ValueOf(config.PublicKey))
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

func (ctx *TestContext) TheResponseBodyShouldBe(body *messages.PickleStepArgument_PickleDocString) (err error) { // nolint
	expectedBody, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}
	return compareStrings(expectedBody, ctx.lastResponseBody)
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
			"expected string does not match actual.\n     Diff:\n%s",
			diff,
		)
	}
	return nil
}

const nullHeaderValue = "[NULL]"

func (ctx *TestContext) TheResponseHeaderShouldBe(headerName string, headerValue string) (err error) { //nolint
	headerValue, err = ctx.preprocessString(headerValue)
	if err != nil {
		return err
	}
	headerName = http.CanonicalHeaderKey(headerName)
	if headerValue != nullHeaderValue {
		if len(ctx.lastResponse.Header[headerName]) == 0 {
			return fmt.Errorf("no such header '%s' in the response", headerName)
		}
		realValue := strings.Join(ctx.lastResponse.Header[headerName], "\n")
		if realValue != headerValue {
			return fmt.Errorf("headers %s different from expected.\nExpected:\n%s\ngot:\n%s",
				headerName, headerValue, realValue)
		}
	} else if len(ctx.lastResponse.Header[headerName]) != 0 {
		return fmt.Errorf("there should not be a '%s' header, but at least one is found", headerName)
	}
	return nil
}

func (ctx *TestContext) TheResponseHeadersShouldBe(headerName string, headersValue *messages.PickleStepArgument_PickleDocString) (err error) { // nolint
	headerValue, err := ctx.preprocessString(headersValue.Content)
	if err != nil {
		return err
	}
	lines := strings.Split(headerValue, "\n")
	trimmed := make([]string, 0, len(lines))
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
		if lines[i] != "" {
			trimmed = append(trimmed, lines[i])
		}
	}
	headerValue = strings.Join(trimmed, "\n")
	return ctx.TheResponseHeaderShouldBe(headerName, headerValue)
}

func (ctx *TestContext) TheResponseErrorMessageShouldContain(s string) (err error) { //nolint
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

func (ctx *TestContext) TheResponseShouldBe(kind string) error { //nolint
	var expectedCode int
	switch kind {
	case "updated", "deleted":
		expectedCode = 200
	case "created":
		expectedCode = 201
	default:
		return fmt.Errorf("unknown response kind: %q", kind)
	}
	if err := ctx.TheResponseCodeShouldBe(expectedCode); err != nil {
		return err
	}
	if err := ctx.TheResponseBodyShouldBeJSON(&messages.PickleStepArgument_PickleDocString{
		Content: `
		{
			"message": "` + kind + `",
			"success": true
		}`,
	}); err != nil {
		return err
	}
	return nil
}
