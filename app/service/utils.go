package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"unicode"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

// QueryParamToInt64Slice extracts from the query parameter of the request a list of integer separated by commas (',')
// returns `nil` for no IDs
func QueryParamToInt64Slice(req *http.Request, paramName string) ([]int64, error) {
	var ids []int64
	paramValue := req.URL.Query().Get(paramName)
	if paramValue == "" {
		return ids, nil
	}
	idsStr := strings.Split(paramValue, ",")
	for _, idStr := range idsStr {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse one of the integer given as query arg (value: '%s', param: '%s')", idStr, paramName)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ResolveURLQueryGetInt64Field extracts a get-parameter of type int64 from the query
func ResolveURLQueryGetInt64Field(httpReq *http.Request, name string) (int64, error) {
	strValue := httpReq.URL.Query().Get(name)
	int64Value, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("missing %s", name)
	}
	return int64Value, nil
}

// ResolveURLQueryGetStringField extracts a get-parameter of type string from the query, fails if the value is empty
func ResolveURLQueryGetStringField(httpReq *http.Request, name string) (string, error) {
	strValue := httpReq.URL.Query().Get(name)
	if strValue == "" {
		return "", fmt.Errorf("missing %s", name)
	}
	return strValue, nil
}

// ResolveURLQueryGetBoolField extracts a get-parameter of type bool (0 or 1) from the query, fails if the value is empty
func ResolveURLQueryGetBoolField(httpReq *http.Request, name string) (bool, error) {
	strValue := httpReq.URL.Query().Get(name)
	if strValue == "" {
		return false, fmt.Errorf("missing %s", name)
	}
	return strValue == "1", nil
}

// ResolveURLQueryPathInt64Field extracts a path element of type int64 from the query
func ResolveURLQueryPathInt64Field(httpReq *http.Request, name string) (int64, error) {
	strValue := chi.URLParam(httpReq, name)
	int64Value, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("missing %s", name)
	}
	return int64Value, nil
}

// ConvertSliceOfMapsFromDBToJSON given a slice of maps that represents DB result data,
// converts it to a slice of maps for rendering JSON so that:
// 1) all maps keys with "__" are considered as paths in JSON (converts "User__ID":... to "user":{"id": ...})
// 2) all maps keys are converted to snake case
// 3) prefixes are stripped, values are converted to needed types accordingly
// 4) fields with nil values are skipped
func ConvertSliceOfMapsFromDBToJSON(dbMaps []map[string]interface{}) []map[string]interface{} {
	convertedResult := make([]map[string]interface{}, len(dbMaps))
	for index := range dbMaps {
		convertedResult[index] = ConvertMapFromDBToJSON(dbMaps[index])
	}
	return convertedResult
}

// ConvertMapFromDBToJSON given a map that represents DB result data,
// converts it a map for rendering JSON so that:
// 1) all map keys with "__" are considered as paths in JSON (converts "User__ID":... to "user":{"id": ...})
// 2) all map keys are converted to snake case
// 3) prefixes are stripped, values are converted to needed types accordingly
// 4) fields with nil values are skipped
func ConvertMapFromDBToJSON(dbMap map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range dbMap {
		currentMap := result

		subKeys := strings.Split(key, "__")
		for subKeyIndex, subKey := range subKeys {
			if subKeyIndex == len(subKeys)-1 {
				setConvertedValueToJSONMap(subKey, value, currentMap)
			} else {
				subKey = toSnakeCase(subKey)
				shouldCreateSubMap := true
				if subMap, hasSubMap := currentMap[subKey]; hasSubMap {
					if subMap, ok := subMap.(map[string]interface{}); ok {
						currentMap = subMap
						shouldCreateSubMap = false
					}
				}
				if shouldCreateSubMap {
					currentMap[subKey] = map[string]interface{}{}
					currentMap = currentMap[subKey].(map[string]interface{})
				}
			}
		}
	}
	return result
}

func setConvertedValueToJSONMap(valueName string, value interface{}, result map[string]interface{}) {
	if value == nil {
		return
	}

	if valueName == "ID" {
		result["id"] = value.(int64)
		return
	}

	if valueName[:2] == "id" {
		valueName = toSnakeCase(valueName[2:]) + "_id"
		result[valueName] = value.(int64)
		return
	}

	switch valueName[0] {
	case 'b':
		value = value == int64(1)
		fallthrough
	case 's', 'i':
		valueName = valueName[1:]
	}
	result[toSnakeCase(valueName)] = value
}

// toSnakeCase convert the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func toSnakeCase(in string) string {
	runes := []rune(in)

	var out []rune
	for i := 0; i < len(runes); i++ {
		if i > 0 && (unicode.IsUpper(runes[i]) || unicode.IsNumber(runes[i])) &&
			((i+1 < len(runes) && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

// GetResponseForTheRouteWithMockedDBAndUser executes a route for unit tests
// auth.UserIDFromContext is stubbed to return the given userID.
// The test should provide functions that prepare the router and the sql mock
func GetResponseForTheRouteWithMockedDBAndUser(
	method string, path string, requestBody string, userID int64,
	setMockExpectationsFunc func(sqlmock.Sqlmock),
	setRouterFunc func(router *chi.Mux, baseService *Base)) (*http.Response, sqlmock.Sqlmock, string, error) {

	logs := setupLogsCaptureForTests()
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }() // nolint: gosec

	setMockExpectationsFunc(mock)

	base := Base{Store: database.NewDataStore(db), Config: nil}
	router := chi.NewRouter()
	setRouterFunc(router, &base)

	monkey.Patch(auth.UserIDFromContext, func(context context.Context) int64 {
		return userID
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	request, err := http.NewRequest(method, ts.URL+path, strings.NewReader(requestBody))
	var response *http.Response
	if err == nil {
		response, err = http.DefaultClient.Do(request)
	}
	return response, mock, logs.String(), err
}

func setupLogsCaptureForTests() *bytes.Buffer { // nolint: deadcode
	logs := &bytes.Buffer{}
	monkey.Patch(logging.GetLogEntry, func(r *http.Request) logrus.FieldLogger {
		logger := logrus.New()
		logger.SetOutput(logs)
		return logger
	})
	return logs
}
