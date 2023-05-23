//go:build !prod

package testhelpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/cucumber/messages-go/v10"
	"github.com/google/go-cmp/cmp"
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
		return fmt.Errorf("the result does not have the expected length. Expected: %d, received: %d on %v",
			count, len(objmap), ctx.lastResponseBody)
	}

	return nil
}

func (ctx *TestContext) getJSONPathOnResponse(jsonPath string) (interface{}, error) {
	var JSONResponse interface{}
	err := json.Unmarshal([]byte(ctx.lastResponseBody), &JSONResponse)
	if err != nil {
		return nil, fmt.Errorf("getJSONPathOnResponse: Unmarshal response: %v", err)
	}

	jsonPathRes, err := jsonpath.Get(jsonPath, JSONResponse)
	if err != nil {
		return nil, fmt.Errorf("getJSONPathOnResponse: Cannot get JsonPath: %v", err)
	}

	return jsonPathRes, nil
}

// TheResponseAtShouldBeTheValue checks that the response at a JSONPath is a certain value.
func (ctx *TestContext) TheResponseAtShouldBeTheValue(jsonPath, value string) error {
	jsonPathRes, err := ctx.getJSONPathOnResponse(jsonPath)
	if err != nil {
		// When an empty value is provided, not finding the jsonPath because it doesn't exist is a success.
		if value == "" {
			return nil
		}

		return err
	}

	value = ctx.replaceReferencesByIDs(value)
	if jsonPathResultMatchesValue(jsonPathRes, value) {
		return nil
	}

	return fmt.Errorf(
		"TheResponseAtShouldBeTheValue: JSONPath %v doesn't match value %v: %v != %v",
		jsonPath,
		ctx.lastResponseBody,
		jsonPathRes,
		value,
	)
}

func jsonPathResultMatchesValue(jsonPathRes interface{}, value string) bool {
	switch jsonPathResultTyped := jsonPathRes.(type) {
	case string:
	case float64:
		valueFloat, _ := strconv.ParseFloat(value, 64)
		if valueFloat == jsonPathResultTyped {
			return true
		}
	case []interface{}:
		// When the result is an empty array, matches if we're looking for an empty value.
		if len(jsonPathResultTyped) == 0 && value == "" {
			return true
		}
	case interface{}:
	}

	return jsonPathRes == value
}

// TheResponseAtShouldBe checks that the response at a JSONPath matches multiple values.
func (ctx *TestContext) TheResponseAtShouldBe(jsonPath string, wants *messages.PickleStepArgument_PickleTable) error {
	jsonPathRes, err := ctx.getJSONPathOnResponse(jsonPath)
	if err != nil {
		return err
	}

	wantsHasHeader := len(wants.Rows[0].Cells) > 1

	wantLength := len(wants.Rows)
	if wantsHasHeader {
		wantLength--
	}

	jsonPathResArr := jsonPathRes.([]interface{})
	if len(jsonPathResArr) != wantLength {
		return fmt.Errorf(
			"TheResponseAtShouldBe: The JsonPath result length should be %v but is %v for %v",
			wantLength,
			len(jsonPathResArr),
			jsonPathResArr,
		)
	}

	if wantsHasHeader {
		return ctx.wantRowsMatchesJSONPathResult(wants, jsonPathResArr)
	}

	return ctx.wantValuesMatchesJSONPathResult(wants, jsonPathResArr)
}

func (ctx *TestContext) wantRowsMatchesJSONPathResult(
	wants *messages.PickleStepArgument_PickleTable,
	jsonPathResArr []interface{},
) error {
	// The jsonPathResult and want rows are put in a slice of maps with the fields that are wanted.
	// Those slices are then sorted, so they can be easily compared element by element.

	headerCells := wants.Rows[0].Cells

	sortedResults := make([]map[string]string, len(wants.Rows)-1)
	sortedWants := make([]map[string]string, len(wants.Rows)-1)

	for i := 1; i < len(wants.Rows); i++ {
		curWant := make(map[string]string)
		curResult := make(map[string]string)

		wantRow := wants.Rows[i]
		for j := 0; j < len(headerCells); j++ {
			curHeader := headerCells[j].Value

			curWant[curHeader] = ctx.replaceReferencesByIDs(wantRow.Cells[j].Value)
			sortedWants[i-1] = curWant

			// The header is a JSONPath (e.g. "title", "strings.title").
			curJSONPathResult, err := jsonpath.Get(curHeader, jsonPathResArr[i-1])
			if err != nil {
				return fmt.Errorf("wantRowsMatchesJSONPathResult: Couldn't retrieve JSONPath %v at %v", curHeader, jsonPathResArr[i-1])
			}

			curResult[curHeader] = stringifyJSONPathResultValue(curJSONPathResult)
			sortedResults[i-1] = curResult
		}
	}

	sortedWants = sortSliceForEasyComparison(sortedWants, headerCells)
	sortedResults = sortSliceForEasyComparison(sortedResults, headerCells)

	if !cmp.Equal(sortedResults, sortedWants) {
		return fmt.Errorf("wantRowsMatchesJSONPathResult: The values (sorted) are %v but should have been: %v", sortedResults, sortedWants)
	}

	return nil
}

func sortSliceForEasyComparison(
	slice []map[string]string,
	headerCells []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell,
) []map[string]string {
	sort.Slice(slice, func(i, j int) bool {
		for curHeader := 0; curHeader < len(headerCells); curHeader++ {
			header := headerCells[curHeader].Value

			curComparison := strings.Compare(slice[i][header], slice[j][header])
			if curComparison < 0 {
				return true
			} else if curComparison > 0 {
				return false
			}
		}

		return true
	})

	return slice
}

func (ctx *TestContext) wantValuesMatchesJSONPathResult(
	wants *messages.PickleStepArgument_PickleTable,
	jsonPathResArr []interface{},
) error {
	// Sort the jsonPath and wants values so that we can check them sequentially.
	// This also makes sure that the values are checked one to one, and are not just present in the "wants".

	sortedResults := make([]string, len(wants.Rows))
	sortedWants := make([]string, len(wants.Rows))
	for i := 0; i < len(wants.Rows); i++ {
		sortedResults[i] = stringifyJSONPathResultValue(jsonPathResArr[i])
		sortedWants[i] = ctx.replaceReferencesByIDs(wants.Rows[i].Cells[0].Value)
	}

	sort.Strings(sortedResults)
	sort.Strings(sortedWants)

	if !cmp.Equal(sortedResults, sortedWants) {
		return fmt.Errorf("wantValuesMatchesJSONPathResult: The values (sorted) are %v but should have been: %v", sortedResults, sortedWants)
	}

	return nil
}

func stringifyJSONPathResultValue(value interface{}) string {
	switch typedValue := value.(type) {
	case bool:
		// Convert boolean results to strings because the values we check are coming from Gherkin as strings.
		return strconv.FormatBool(typedValue)
	default:
		return typedValue.(string)
	}
}

// TheResponseShouldNotBeDefinedAt checks that the provided jsonPath doesn't exist.
func (ctx *TestContext) TheResponseShouldNotBeDefinedAt(jsonPath string) error {
	var JSONResponse interface{}
	err := json.Unmarshal([]byte(ctx.lastResponseBody), &JSONResponse)
	if err != nil {
		return fmt.Errorf("TheResponseShouldNotBeDefinedAt: Unmarshal response: %v", err)
	}

	jsonPathRes, err := jsonpath.Get(jsonPath, JSONResponse)
	if err != nil {
		//nolint:nilerr // We want jsonpath.Get to return an error.
		return nil
	}

	return fmt.Errorf("TheResponseShouldNotBeDefinedAt: JsonPath: %v is defined with value %v", jsonPath, jsonPathRes)
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

// TheResponseBodyShouldBe checks that the response is the same as the one provided.
func (ctx *TestContext) TheResponseBodyShouldBe(body *messages.PickleStepArgument_PickleDocString) (err error) {
	expectedBody, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}
	return compareStrings(expectedBody, ctx.lastResponseBody)
}

func compareStrings(expected, actual string) error {
	if expected != actual {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
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

// TheResponseHeaderShouldBe checks that the response header matches the provided value.
func (ctx *TestContext) TheResponseHeaderShouldBe(headerName, headerValue string) (err error) {
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

// TheResponseHeadersShouldBe checks that the response header matches the multiline provided value.
func (ctx *TestContext) TheResponseHeadersShouldBe(
	headerName string,
	headersValue *messages.PickleStepArgument_PickleDocString,
) (err error) {
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

// TheResponseErrorMessageShouldContain checks that the response contains the provided string.
func (ctx *TestContext) TheResponseErrorMessageShouldContain(s string) (err error) {
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

// TheResponseShouldBe checks that the response status of the response is of the given kind.
func (ctx *TestContext) TheResponseShouldBe(kind string) error {
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
