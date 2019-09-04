package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-chi/chi"
)

// ResolveURLQueryGetInt64SliceField extracts from the query parameter of the request a list of integer separated by commas (',')
// returns `nil` for no IDs
func ResolveURLQueryGetInt64SliceField(req *http.Request, paramName string) ([]int64, error) {
	if err := checkQueryGetFieldIsNotMissing(req, paramName); err != nil {
		return nil, err
	}

	var ids []int64
	paramValue := req.URL.Query().Get(paramName)
	if paramValue == "" {
		return ids, nil
	}
	idsStr := strings.Split(paramValue, ",")
	for _, idStr := range idsStr {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse one of the integers given as query args (value: '%s', param: '%s')", idStr, paramName)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ResolveURLQueryGetInt64Field extracts a get-parameter of type int64 from the query
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

// ResolveURLQueryGetStringField extracts a get-parameter of type string from the query, fails if the value is empty
func ResolveURLQueryGetStringField(httpReq *http.Request, name string) (string, error) {
	if err := checkQueryGetFieldIsNotMissing(httpReq, name); err != nil {
		return "", err
	}
	return httpReq.URL.Query().Get(name), nil
}

// ResolveURLQueryGetStringSliceField extracts from the query parameter of the request a list of strings separated by commas (',')
// returns `nil` the parameter is missing
func ResolveURLQueryGetStringSliceField(req *http.Request, paramName string) ([]string, error) {
	if err := checkQueryGetFieldIsNotMissing(req, paramName); err != nil {
		return nil, err
	}

	paramValue := req.URL.Query().Get(paramName)
	return strings.Split(paramValue, ","), nil
}

// ResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters extracts a list of values
// out from '<fieldName>_include'/'<fieldName>_exclude' request parameters:
//   1. If none of '<fieldName>_include'/'<fieldName>_exclude' is present, all the known values are returned.
//   2. If '<fieldName>_include' is present, then it becomes the result list.
//   3. If '<fieldName>_exclude' is present, then we exclude all its values from the result list.
//
// All values from both the request parameters are checked against the list of known values.
func ResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters(
	r *http.Request, fieldName string, knownValuesMap map[string]bool) ([]string, error) {
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

// ResolveURLQueryGetTimeField extracts a get-parameter of type time.Time (rfc3339) from the query
func ResolveURLQueryGetTimeField(httpReq *http.Request, name string) (time.Time, error) {
	if err := checkQueryGetFieldIsNotMissing(httpReq, name); err != nil {
		return time.Time{}, err
	}
	result, err := time.Parse(time.RFC3339, httpReq.URL.Query().Get(name))
	if err != nil {
		return time.Time{}, fmt.Errorf("wrong value for %s (should be time (rfc3339))", name)
	}
	return result, nil
}

// ResolveURLQueryGetBoolField extracts a get-parameter of type bool (0 or 1) from the query, fails if the value is empty
func ResolveURLQueryGetBoolField(httpReq *http.Request, name string) (bool, error) {
	if len(httpReq.URL.Query()[name]) == 0 {
		return false, fmt.Errorf("missing %s", name)
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

// ResolveURLQueryPathInt64Field extracts a path element of type int64 from the query
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

func checkQueryGetFieldIsNotMissing(httpReq *http.Request, name string) error {
	if len(httpReq.URL.Query()[name]) == 0 {
		return fmt.Errorf("missing %s", name)
	}
	return nil
}

// ResolveURLQueryPathInt64SliceField extracts a list of integers separated by commas (',') from the query path of the request
func ResolveURLQueryPathInt64SliceField(req *http.Request, paramName string) ([]int64, error) {
	paramValue := chi.URLParam(req, paramName)
	paramValue = strings.Trim(paramValue, "/")
	var ids []int64
	if paramValue == "" {
		return ids, nil
	}
	idsStr := strings.Split(paramValue, "/")
	for _, idStr := range idsStr {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse one of the integers given as query args (value: '%s', param: '%s')", idStr, paramName)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ConvertSliceOfMapsFromDBToJSON given a slice of maps that represents DB result data,
// converts it to a slice of maps for rendering JSON so that:
// 1) all maps keys with "__" are considered as paths in JSON (converts "User__ID":... to "user":{"id": ...})
// 2) all maps keys are converted to snake case
// 3) prefixes are stripped, values are converted to needed types accordingly
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

	replaceEmptySubMapsWithNils(result)
	return result
}

func replaceEmptySubMapsWithNils(mapToProcess map[string]interface{}) bool {
	for key := range mapToProcess {
		if subMap, ok := mapToProcess[key].(map[string]interface{}); ok {
			if replaceEmptySubMapsWithNils(subMap) {
				mapToProcess[key] = nil
			}
		}
	}
	for key := range mapToProcess {
		if mapToProcess[key] != nil {
			return false
		}
	}
	return true
}

func setConvertedValueToJSONMap(valueName string, value interface{}, result map[string]interface{}) {
	snakeCaseName := toSnakeCase(valueName)
	underscoreIndex := strings.IndexByte(snakeCaseName, '_')
	prefix := ""
	if underscoreIndex > 0 {
		prefix = snakeCaseName[:underscoreIndex]
	}

	switch prefix {
	case "id":
		snakeCaseName = snakeCaseName[3:] + "_id"
	case "nb":
		value = int32(value.(int64))
	case "b":
		value = value == int64(1)
	case "i":
		value = parseINumber(value)
	}
	if map[string]bool{"nb": true, "b": true, "i": true, "s": true}[prefix] {
		snakeCaseName = snakeCaseName[underscoreIndex+1:]
	}

	if valueInt64, ok := value.(int64); ok {
		value = strconv.FormatInt(valueInt64, 10)
	}

	result[snakeCaseName] = value
}

func parseINumber(value interface{}) interface{} {
	switch typedValue := value.(type) {
	case int64:
		value = int32(typedValue)
	case string:
		if parsed, err := strconv.ParseInt(typedValue, 10, 32); err == nil {
			value = int32(parsed)
		} else if parsed, err := strconv.ParseFloat(typedValue, 64); err == nil {
			value = float32(parsed)
		}
	}
	return value
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
