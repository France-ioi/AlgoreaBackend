package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
