package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

// ResolveURLQueryGetInt64SliceField extracts from the query parameter of the request a list of integer separated by commas (',')
// returns `nil` for no IDs.
func ResolveURLQueryGetInt64SliceField(req *http.Request, paramName string) ([]int64, error) {
	if err := checkQueryGetFieldIsNotMissing(req, paramName); err != nil {
		return nil, err
	}

	paramValue := req.URL.Query().Get(paramName)
	if paramValue == "" {
		return []int64(nil), nil
	}
	idsStr := strings.Split(paramValue, ",")
	ids := make([]int64, 0, len(idsStr))
	for _, idStr := range idsStr {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse one of the integers given as query args (value: '%s', param: '%s')", idStr, paramName)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ResolveURLQueryGetInt64Field extracts a get-parameter of type int64 from the query.
func ResolveURLQueryGetInt64Field(httpReq *http.Request, name string) (int64, error) {
	if err := checkQueryGetFieldIsNotMissing(httpReq, name); err != nil {
		return 0, err
	}
	strValue := httpReq.URL.Query().Get(name)
	int64Value, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("wrong value for %s (should be int64)", name)
	}
	return int64Value, nil
}

// ResolveURLQueryGetStringField extracts a get-parameter of type string from the query, fails if the value is empty.
func ResolveURLQueryGetStringField(httpReq *http.Request, name string) (string, error) {
	if err := checkQueryGetFieldIsNotMissing(httpReq, name); err != nil {
		return "", err
	}
	return httpReq.URL.Query().Get(name), nil
}

// ResolveURLQueryGetStringSliceField extracts from the query parameter of the request a list of strings separated by commas (',')
// returns `nil` the parameter is missing.
func ResolveURLQueryGetStringSliceField(req *http.Request, paramName string) ([]string, error) {
	if err := checkQueryGetFieldIsNotMissing(req, paramName); err != nil {
		return nil, err
	}

	paramValue := req.URL.Query().Get(paramName)
	return strings.Split(paramValue, ","), nil
}

// ResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters extracts a list of values
// out from '<fieldName>_include'/'<fieldName>_exclude' request parameters:
//  1. If none of '<fieldName>_include'/'<fieldName>_exclude' is present, all the known values are returned.
//  2. If '<fieldName>_include' is present, then it becomes the result list.
//  3. If '<fieldName>_exclude' is present, then we exclude all its values from the result list.
//
// All values from both the request parameters are checked against the list of known values.
func ResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters(
	r *http.Request, fieldName string, knownValuesMap map[string]bool,
) ([]string, error) {
	var valuesMap map[string]bool
	valuesToInclude, err := ResolveURLQueryGetStringSliceField(r, fieldName+"_include")
	if err == nil {
		valuesMap = make(map[string]bool, len(valuesToInclude))
		for _, value := range valuesToInclude {
			if !knownValuesMap[value] {
				return nil, fmt.Errorf("wrong value in '%s_include': %q", fieldName, value)
			}
			valuesMap[value] = true
		}
	} else {
		valuesMap = make(map[string]bool, len(knownValuesMap))
		for value := range knownValuesMap {
			valuesMap[value] = true
		}
	}

	valuesToExclude, err := ResolveURLQueryGetStringSliceField(r, fieldName+"_exclude")
	if err == nil && len(valuesToExclude) != 0 {
		for _, valueToExclude := range valuesToExclude {
			if !knownValuesMap[valueToExclude] {
				return nil, fmt.Errorf("wrong value in '%s_exclude': %q", fieldName, valueToExclude)
			}
			delete(valuesMap, valueToExclude)
		}
	}
	valuesList := make([]string, 0, len(valuesMap))
	for value := range valuesMap {
		valuesList = append(valuesList, value)
	}
	return valuesList, nil
}

// ResolveURLQueryGetTimeField extracts a get-parameter of type time.Time (rfc3339) from the query.
func ResolveURLQueryGetTimeField(httpReq *http.Request, name string) (time.Time, error) {
	if err := checkQueryGetFieldIsNotMissing(httpReq, name); err != nil {
		return time.Time{}, err
	}
	result, err := time.Parse(time.RFC3339Nano, httpReq.URL.Query().Get(name))
	if err != nil {
		return time.Time{}, fmt.Errorf("wrong value for %s (should be time (rfc3339Nano))", name)
	}
	return result, nil
}

// ResolveURLQueryGetBoolField extracts a get-parameter of type bool (0 or 1) from the query, fails if the value is empty.
func ResolveURLQueryGetBoolField(httpReq *http.Request, name string) (bool, error) {
	err := checkQueryGetFieldIsNotMissing(httpReq, name)
	if err != nil {
		return false, err
	}
	strValue := httpReq.URL.Query().Get(name)
	if strValue == "0" {
		return false, nil
	}
	if strValue == "1" {
		return true, nil
	}
	return false, fmt.Errorf("wrong value for %s (should have a boolean value (0 or 1))", name)
}

// ResolveURLQueryGetBoolFieldWithDefault extracts a get-parameter of type bool (0 or 1) from the query.
// If it is not provided, `defaultValue` is returned.
func ResolveURLQueryGetBoolFieldWithDefault(httpReq *http.Request, name string, defaultValue bool) (bool, error) {
	if !URLQueryPathHasField(httpReq, name) {
		return defaultValue, nil
	}

	strValue := httpReq.URL.Query().Get(name)
	if strValue == "0" {
		return false, nil
	}
	if strValue == "1" {
		return true, nil
	}

	return false, fmt.Errorf("wrong value for %s (should have a boolean value (0 or 1))", name)
}

// ResolveURLQueryPathInt64Field extracts a path element of type int64 from the query.
func ResolveURLQueryPathInt64Field(httpReq *http.Request, name string) (int64, error) {
	strValue := chi.URLParam(httpReq, name)
	if strValue == "" {
		return 0, fmt.Errorf("missing %s", name)
	}
	int64Value, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("wrong value for %s (should be int64)", name)
	}
	return int64Value, nil
}

// URLQueryPathHasField checks whether a field is present in the query.
func URLQueryPathHasField(httpReq *http.Request, name string) bool {
	return len(httpReq.URL.Query()[name]) > 0
}

func checkQueryGetFieldIsNotMissing(httpReq *http.Request, name string) error {
	if !URLQueryPathHasField(httpReq, name) {
		return fmt.Errorf("missing %s", name)
	}

	return nil
}

// ResolveURLQueryPathInt64SliceField extracts a list of integers separated by commas (',') from the query path of the request.
func ResolveURLQueryPathInt64SliceField(req *http.Request, paramName string) ([]int64, error) {
	paramValue := chi.URLParam(req, paramName)
	paramValue = strings.Trim(paramValue, "/")
	if paramValue == "" {
		return []int64(nil), nil
	}
	idsStr := strings.Split(paramValue, "/")
	ids := make([]int64, 0, len(idsStr))
	for _, idStr := range idsStr {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse one of the integers given as query args (value: '%s', param: '%s')", idStr, paramName)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ResolveURLQueryPathInt64SliceFieldWithLimit extracts a list of integers separated by commas (',')
// from the query path of the request applying the given limit.
func ResolveURLQueryPathInt64SliceFieldWithLimit(r *http.Request, paramName string, limit int) ([]int64, error) {
	ids, err := ResolveURLQueryPathInt64SliceField(r, paramName)
	if err != nil {
		return nil, err
	}
	if len(ids) > limit {
		return nil, fmt.Errorf("no more than %d %s expected", limit, paramName)
	}
	return ids, nil
}

// ResolveJSONBodyIntoMap reads the request body and parses it as JSON into a map.
// As it reads out the body, it can only be called once.
func ResolveJSONBodyIntoMap(r *http.Request) (map[string]interface{}, error) {
	var rawRequestData map[string]interface{}
	defer func() { _, _ = io.Copy(io.Discard, r.Body) }()
	err := json.NewDecoder(r.Body).Decode(&rawRequestData)
	if err != nil {
		return nil, ErrInvalidRequest(fmt.Errorf("invalid input JSON: %w", err))
	}
	return rawRequestData, nil
}
