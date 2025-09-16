//go:build !prod && !unit

package testhelpers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/google/go-cmp/cmp"
	"github.com/pmezard/go-difflib/difflib"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/encrypt"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

func (ctx *TestContext) ItShouldBeAJSONArrayWithEntries(count int) error { //nolint
	var objmap []map[string]*json.RawMessage

	if err := json.Unmarshal([]byte(ctx.lastResponseBody), &objmap); err != nil {
		return fmt.Errorf("unable to decode the response as JSON: %w\nData:%v", err, ctx.lastResponseBody)
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
		return nil, fmt.Errorf("getJSONPathOnResponse: Unmarshal response: %w", err)
	}

	return jsonpath.Get(jsonPath, JSONResponse)
}

// TheResponseAtShouldBeTheValue checks that the response at a JSONPath is a certain value.
func (ctx *TestContext) TheResponseAtShouldBeTheValue(jsonPath, value string) error {
	jsonPathRes, err := ctx.getJSONPathOnResponse(jsonPath)
	if err != nil {
		// The JSONPath is not defined.
		if value == undefinedValue {
			return nil
		}

		return fmt.Errorf("TheResponseAtShouldBeTheValue: JSONPath %v doesn't match value %v: %w", jsonPath, value, err)
	}

	value = ctx.replaceReferencesWithIDs(value)
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
	var expected interface{} = value

	switch jsonPathResultTyped := jsonPathRes.(type) {
	case bool:
		expected, _ = strconv.ParseBool(value)
	case string:
	case float64:
		expected, _ = strconv.ParseFloat(value, 64)
	case []interface{}:
		// When the result is an empty array, matches if we're looking for "[]".
		if len(jsonPathResultTyped) == 0 && value == "[]" {
			return true
		}
	case interface{}:
	}

	if jsonPathRes == nil && jsonPathValueConsideredNil(value) {
		return true
	}

	return jsonPathRes == expected
}

func jsonPathValueConsideredNil(value string) bool {
	return value == nullValue
}

// TheResponseAtShouldBe checks that the response at a JSONPath matches multiple values.
func (ctx *TestContext) TheResponseAtShouldBe(jsonPath string, wants *godog.Table) error {
	jsonPathRes, err := ctx.getJSONPathOnResponse(jsonPath)
	if err != nil {
		return err
	}

	switch typedJSONRes := jsonPathRes.(type) {
	case []interface{}:
		wantsHasHeader := len(wants.Rows[0].Cells) > 1

		wantLength := len(wants.Rows)
		if wantsHasHeader {
			wantLength--
		}

		// The result is an array (eg. "element": [...])
		if len(typedJSONRes) != wantLength {
			expectedJSONRows := make([]interface{}, len(typedJSONRes))
			for index, row := range typedJSONRes {
				if strValue, ok := row.(string); ok {
					expectedJSONRows[index] = ctx.readableValue(strValue)
				} else {
					expectedJSONRows[index] = row
				}
			}
			return fmt.Errorf(
				"TheResponseAtShouldBe: The JsonPath result length should be %v but is %v for %v",
				wantLength,
				len(typedJSONRes),
				expectedJSONRows,
			)
		}

		if wantsHasHeader {
			return ctx.wantRowsMatchesJSONPathResultArr(wants, typedJSONRes)
		}

		return ctx.wantValuesMatchesJSONPathResultArr(wants, typedJSONRes)
	default:
		if typedJSONRes == nil {
			return fmt.Errorf("TheResponseAtShouldBe: The JsonPath result at the path %v is %v", jsonPath, typedJSONRes)
		}
	}

	panic(fmt.Sprintf("TheResponseAtShouldBe: Result found at JSON Path %v should be an array but is: %v", jsonPath, jsonPathRes))
}

// TheResponseAtInJSONShouldBe checks that the response in JSON at a JSONPath matches.
func (ctx *TestContext) TheResponseAtInJSONShouldBe(jsonPath string, wants *godog.DocString) error {
	jsonPathRes, err := ctx.getJSONPathOnResponse(jsonPath)
	if err != nil {
		return err
	}

	actual, err := json.MarshalIndent(&jsonPathRes, "", "\t")
	if err != nil {
		return err
	}

	preprocessedWants := ctx.preprocessString(wants.Content)

	expected, err := indentJSON(preprocessedWants)
	if err != nil {
		return err
	}

	return compareStrings(string(expected), string(actual))
}

// TheResponseAtShouldBeTheBase64OfAnAES256GCMEncryptedJSONObjectContaining checks that the response at a JSONPath is
// an AES256GCM encrypted JSON object.
func (ctx *TestContext) TheResponseAtShouldBeTheBase64OfAnAES256GCMEncryptedJSONObjectContaining(
	jsonPath string,
	expectedJSONParam *godog.DocString,
) error {
	hexCipher, err := ctx.getJSONPathOnResponse(jsonPath)
	if err != nil {
		return err
	}

	//nolint:forcetypeassert // panic if hexCipher is not a string
	cipherText, err := hex.DecodeString(hexCipher.(string))
	if err != nil {
		return err
	}

	key := []byte(app.AuthConfig(ctx.application.Config).GetString("clientSecret")[0:32])
	plainJSON, err := encrypt.DecryptAES256GCM(key, cipherText)
	if err != nil {
		return err
	}

	expectedJSON := strings.ReplaceAll(expectedJSONParam.Content, " ", "")
	expectedJSON = strings.ReplaceAll(expectedJSON, "\n", "")
	expectedJSON = ctx.preprocessString(expectedJSON)

	return compareStrings(expectedJSON, string(plainJSON))
}

// indentJSON indents the JSON string.
// Works by re-encoding the JSON string with indentation.
func indentJSON(preprocessedWants string) ([]byte, error) {
	var exp interface{}
	err := json.Unmarshal([]byte(preprocessedWants), &exp)
	if err != nil {
		return nil, err
	}

	var expected []byte
	if expected, err = json.MarshalIndent(&exp, "", "\t"); err != nil {
		return nil, err
	}

	return expected, nil
}

func (ctx *TestContext) wantRowsMatchesJSONPathResultArr(
	wants *godog.Table,
	jsonPathResArr []interface{},
) error {
	// The jsonPathResult and want rows are put in a slice of maps with the fields that are wanted.
	// Those slices are then sorted, so they can be easily compared element by element.

	headerCells := wants.Rows[0].Cells

	sortedResults := make([]map[string]string, len(wants.Rows)-1)
	sortedWants := make([]map[string]string, len(wants.Rows)-1)

	for rowIndex := 1; rowIndex < len(wants.Rows); rowIndex++ {
		curWant := make(map[string]string)
		curResult := make(map[string]string)

		wantRow := wants.Rows[rowIndex]
		for cellIndex := 0; cellIndex < len(headerCells); cellIndex++ {
			curHeader := headerCells[cellIndex].Value

			curWant[curHeader] = ctx.replaceReferencesWithIDs(wantRow.Cells[cellIndex].Value)
			sortedWants[rowIndex-1] = curWant

			// The header is a JSONPath (e.g. "title", "strings.title").
			curJSONPathResult, err := jsonpath.Get(curHeader, jsonPathResArr[rowIndex-1])
			if err != nil {
				return fmt.Errorf("wantRowsMatchesJSONPathResult: Couldn't retrieve JSONPath %v at %v", curHeader, jsonPathResArr[rowIndex-1])
			}

			curResult[curHeader] = stringifyJSONPathResultValue(curJSONPathResult)
			sortedResults[rowIndex-1] = curResult
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
	headerCells []*messages.PickleTableCell,
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

func (ctx *TestContext) wantValuesMatchesJSONPathResultArr(
	wants *godog.Table,
	jsonPathResArr []interface{},
) error {
	// Sort the jsonPath and wants values so that we can check them sequentially.
	// This also makes sure that the values are checked one to one, and are not just present in the "wants".

	sortedResults := make([]string, len(wants.Rows))
	sortedWants := make([]string, len(wants.Rows))
	for i := 0; i < len(wants.Rows); i++ {
		sortedResults[i] = stringifyJSONPathResultValue(jsonPathResArr[i])
		sortedWants[i] = ctx.replaceReferencesWithIDs(wants.Rows[i].Cells[0].Value)
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
		// The value is nil when the JSONPath is not defined.
		if value == nil {
			return undefinedValue
		}

		//nolint:forcetypeassert // panic if the value is neither string nor bool
		return typedValue.(string)
	}
}

func (ctx *TestContext) TheResponseCodeShouldBe(code int) error { //nolint
	if code != ctx.lastResponse.StatusCode {
		return fmt.Errorf("expected http response code: %d, actual is: %d. \n Data: %s", code, ctx.lastResponse.StatusCode, ctx.lastResponseBody)
	}
	return nil
}

// TheResponseBodyShouldBeJSON checks that the response body is a valid JSON and matches the given JSON.
func (ctx *TestContext) TheResponseBodyShouldBeJSON(body *godog.DocString) (err error) {
	return ctx.TheResponseDecodedBodyShouldBeJSON("", body)
}

// TheResponseDecodedBodyShouldBeJSON checks that the response body, after being decoded, is a valid JSON and matches the given JSON.
func (ctx *TestContext) TheResponseDecodedBodyShouldBeJSON(responseType string, body *godog.DocString) error {
	// verify the content type
	if err := ValidateJSONContentType(ctx.lastResponse); err != nil {
		return err
	}

	expectedBody := ctx.preprocessString(body.Content)

	expected, err := indentJSON(expectedBody)
	if err != nil {
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
		return fmt.Errorf("unable to decode the response as JSON: %w -- Data: %v", err, ctx.lastResponseBody)
	}

	if responseType != "" {
		act = payloads.ConvertIntoMap(act)
	}
	actual, err := json.MarshalIndent(act, "", "\t")
	if err != nil {
		return err
	}

	return compareStrings(string(expected), string(actual))
}

// TheResponseBodyShouldBe checks that the response is the same as the one provided.
func (ctx *TestContext) TheResponseBodyShouldBe(body *godog.DocString) (err error) {
	expectedBody := ctx.preprocessString(body.Content)
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

const (
	undefinedHeaderValue = "[Header not defined]"
	nullValue            = "<null>"
	undefinedValue       = "<undefined>"
)

// TheResponseHeaderShouldBe checks that the response header matches the provided value.
func (ctx *TestContext) TheResponseHeaderShouldBe(headerName, headerValue string) (err error) {
	if headerValue == undefinedHeaderValue {
		return ctx.TheResponseHeaderShouldNotBeSet(headerName)
	}

	headerValue = ctx.preprocessString(headerValue)
	headerName = http.CanonicalHeaderKey(headerName)

	if len(ctx.lastResponse.Header[headerName]) == 0 {
		return fmt.Errorf("no such header '%s' in the response", headerName)
	}
	realValue := strings.Join(ctx.lastResponse.Header[headerName], "\n")
	if realValue != headerValue {
		return fmt.Errorf("headers %s different from expected.\nExpected:\n%s\ngot:\n%s",
			headerName, headerValue, realValue)
	}

	return nil
}

// TheResponseHeaderShouldNotBeSet checks that the response header is not set.
func (ctx *TestContext) TheResponseHeaderShouldNotBeSet(headerName string) (err error) {
	headerName = http.CanonicalHeaderKey(headerName)

	if len(ctx.lastResponse.Header[headerName]) != 0 {
		return fmt.Errorf("there should not be a '%s' header, but at least one is found", headerName)
	}

	return nil
}

// TheResponseHeadersShouldBe checks that the response header matches the multiline provided value.
func (ctx *TestContext) TheResponseHeadersShouldBe(
	headerName string,
	headersValue *godog.DocString,
) (err error) {
	headerValue := ctx.preprocessString(headersValue.Content)
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
func (ctx *TestContext) TheResponseErrorMessageShouldContain(needle string) (err error) {
	errorResp := service.ErrorResponse[interface{}]{}
	// decode response
	if err = json.Unmarshal([]byte(ctx.lastResponseBody), &errorResp); err != nil {
		return fmt.Errorf("unable to decode the response as JSON: %w -- Data: %v", err, ctx.lastResponseBody)
	}
	if !strings.Contains(errorResp.ErrorText, needle) {
		return fmt.Errorf("cannot find expected `%s` in error text: `%s`", needle, errorResp.ErrorText)
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
	return ctx.TheResponseBodyShouldBeJSON(&godog.DocString{
		Content: `
		{
			"message": "` + kind + `",
			"success": true
		}`,
	})
}
