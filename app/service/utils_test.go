package service

import (
	"net/http"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

func TestQueryParamToInt64Slice(t *testing.T) {
	testCases := []struct {
		desc           string
		queryString    string
		expectedList   []int64
		expectedErrMsg string
	}{
		{
			desc:           "no param",
			queryString:    "",
			expectedList:   nil,
			expectedErrMsg: "",
		},
		{
			desc:           "wrong param name",
			queryString:    "id=1,2",
			expectedList:   nil,
			expectedErrMsg: "",
		},
		{
			desc:           "single value",
			queryString:    "ids=3",
			expectedList:   []int64{3},
			expectedErrMsg: "",
		},
		{
			desc:           "multiple value",
			queryString:    "ids=4,5",
			expectedList:   []int64{4, 5},
			expectedErrMsg: "",
		},
		{
			desc:           "not a int64 (string)",
			queryString:    "ids=6,7,etc",
			expectedList:   nil,
			expectedErrMsg: "unable to parse one of the integer given as query arg (value: 'etc', param: 'ids')",
		},
		{
			desc:           "not a int64 (empty val)",
			queryString:    "ids=8,,9",
			expectedList:   nil,
			expectedErrMsg: "unable to parse one of the integer given as query arg (value: '', param: 'ids')",
		},
		{
			desc:           "too big for int64",
			queryString:    "ids=123456789012345678901234567890",
			expectedList:   nil,
			expectedErrMsg: "unable to parse one of the integer given as query arg (value: '123456789012345678901234567890', param: 'ids')",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			assert := assertlib.New(t)

			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := QueryParamToInt64Slice(req, "ids")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(err, testCase.expectedErrMsg)
			} else {
				assert.NoError(err)
			}
			assert.Equal(testCase.expectedList, list)
		})
	}

}
