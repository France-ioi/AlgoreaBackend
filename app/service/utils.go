package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// QueryParamToInt64Slice extracts from the query parameter of the request a list of integer separated by commas (',')
func QueryParamToInt64Slice(req *http.Request, paramName string) ([]int64, error) {

	idsStr := strings.Split(req.URL.Query().Get(paramName), ",")
	var ids []int64
	for _, idStr := range idsStr {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse one of the integer given as query arg (value: '%s', param: '%s')", idStr, paramName)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
