package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestResolveURLQueryGetInt64SliceField(t *testing.T) {
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
			expectedErrMsg: "missing ids",
		},
		{
			desc:           "empty param",
			queryString:    "ids=",
			expectedList:   nil,
			expectedErrMsg: "",
		},
		{
			desc:           "wrong param name",
			queryString:    "id=1,2",
			expectedList:   nil,
			expectedErrMsg: "missing ids",
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
			expectedErrMsg: "unable to parse one of the integers given as query args (value: 'etc', param: 'ids')",
		},
		{
			desc:           "not a int64 (empty val)",
			queryString:    "ids=8,,9",
			expectedList:   nil,
			expectedErrMsg: "unable to parse one of the integers given as query args (value: '', param: 'ids')",
		},
		{
			desc:           "too big for int64",
			queryString:    "ids=123456789012345678901234567890",
			expectedList:   nil,
			expectedErrMsg: "unable to parse one of the integers given as query args (value: '123456789012345678901234567890', param: 'ids')",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.desc, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := ResolveURLQueryGetInt64SliceField(req, "ids")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, err, testCase.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.expectedList, list)
		})
	}
}

func TestResolveURLQueryPathInt64Field(t *testing.T) {
	testCases := []struct {
		desc           string
		routeString    string
		queryString    string
		expectedValue  int64
		expectedErrMsg string
	}{
		{
			desc:           "single value",
			routeString:    "/{id}",
			queryString:    "/3",
			expectedValue:  3,
			expectedErrMsg: "",
		},
		{
			desc:           "multiple value",
			routeString:    "/{id}",
			queryString:    "/4,5",
			expectedValue:  0,
			expectedErrMsg: "wrong value for id (should be int64)",
		},
		{
			desc:           "not an int64 (string)",
			routeString:    "/{id}",
			queryString:    "/word",
			expectedValue:  0,
			expectedErrMsg: "wrong value for id (should be int64)",
		},
		{
			desc:           "not a int64 (empty val)",
			routeString:    "/{id}/",
			queryString:    "//",
			expectedValue:  0,
			expectedErrMsg: "missing id",
		},
		{
			desc:           "too big for int64",
			routeString:    "/{id}",
			queryString:    "/123456789012345678901234567890",
			expectedValue:  0,
			expectedErrMsg: "wrong value for id (should be int64)",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.desc, func(t *testing.T) {
			r := chi.NewRouter()
			called := false
			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				value, err := ResolveURLQueryPathInt64Field(r, "id")
				if testCase.expectedErrMsg != "" {
					assert.EqualError(t, err, testCase.expectedErrMsg)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, testCase.expectedValue, value)
			}
			r.Get(testCase.routeString, handler)

			ts := httptest.NewServer(r)
			request, _ := http.NewRequest("GET", ts.URL+testCase.queryString, nil)
			response, err := http.DefaultClient.Do(request)
			if err == nil {
				_ = response.Body.Close()
			}
			ts.Close()
			assert.True(t, called, "The handler was not called")
		})
	}
}

func TestResolveURLQueryPathInt64SliceField(t *testing.T) {
	testCases := []struct {
		desc           string
		queryString    string
		expectedList   []int64
		expectedErrMsg string
	}{
		{
			desc:         "no param",
			queryString:  "/something",
			expectedList: nil,
		},
		{
			desc:         "empty param",
			queryString:  "///something",
			expectedList: nil,
		},
		{
			desc:         "single value",
			queryString:  "/3/something",
			expectedList: []int64{3},
		},
		{
			desc:         "multiple values",
			queryString:  "/4/5/something",
			expectedList: []int64{4, 5},
		},
		{
			desc:         "multiple value with slashes",
			queryString:  "////4/5////something",
			expectedList: []int64{4, 5},
		},
		{
			desc:           "not an int64 (string)",
			queryString:    "/6/7/etc/something",
			expectedList:   nil,
			expectedErrMsg: "unable to parse one of the integers given as query args (value: 'etc', param: 'ids')",
		},
		{
			desc:           "not an int64 (empty val)",
			queryString:    "/8//9/something",
			expectedList:   nil,
			expectedErrMsg: "unable to parse one of the integers given as query args (value: '', param: 'ids')",
		},
		{
			desc:           "too big for int64",
			queryString:    "/123456789012345678901234567890/something",
			expectedList:   nil,
			expectedErrMsg: "unable to parse one of the integers given as query args (value: '123456789012345678901234567890', param: 'ids')",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.desc, func(t *testing.T) {
			called := false
			r := chi.NewRouter()
			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				value, err := ResolveURLQueryPathInt64SliceField(r, "ids")
				if testCase.expectedErrMsg != "" {
					assert.EqualError(t, err, testCase.expectedErrMsg)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, testCase.expectedList, value)
			}
			r.Get(`/{ids:.*}something`, handler)
			ts := httptest.NewServer(r)
			request, _ := http.NewRequest("GET", ts.URL+testCase.queryString, nil)
			response, err := http.DefaultClient.Do(request)
			if err == nil {
				_ = response.Body.Close()
			}
			ts.Close()
			assert.True(t, called, "The handler was not called")
		})
	}
}

func TestResolveURLQueryGetStringField(t *testing.T) {
	testCases := []struct {
		desc           string
		queryString    string
		expectedValue  string
		expectedErrMsg string
	}{
		{
			desc:           "no param",
			queryString:    "",
			expectedValue:  "",
			expectedErrMsg: "missing name",
		},
		{
			desc:           "wrong param name",
			queryString:    "name1=value",
			expectedValue:  "",
			expectedErrMsg: "missing name",
		},
		{
			desc:           "value given",
			queryString:    "name=value",
			expectedValue:  "value",
			expectedErrMsg: "",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.desc, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := ResolveURLQueryGetStringField(req, "name")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, err, testCase.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.expectedValue, list)
		})
	}
}

func TestResolveURLQueryGetBoolField(t *testing.T) {
	testCases := []struct {
		desc           string
		queryString    string
		expectedValue  bool
		expectedErrMsg string
	}{
		{
			desc:           "no param",
			queryString:    "",
			expectedValue:  false,
			expectedErrMsg: "missing flag",
		},
		{
			desc:           "wrong param name",
			queryString:    "flag1=1",
			expectedValue:  false,
			expectedErrMsg: "missing flag",
		},
		{
			desc:           "true value given",
			queryString:    "flag=1",
			expectedValue:  true,
			expectedErrMsg: "",
		},
		{
			desc:           "false value given",
			queryString:    "flag=0",
			expectedValue:  false,
			expectedErrMsg: "",
		},
		{
			desc:           "wrong value given",
			queryString:    "flag=2",
			expectedValue:  false,
			expectedErrMsg: "wrong value for flag (should have a boolean value (0 or 1))",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.desc, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := ResolveURLQueryGetBoolField(req, "flag")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, err, testCase.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.expectedValue, list)
		})
	}
}

func TestResolveURLQueryGetTimeField(t *testing.T) {
	testCases := []struct {
		desc           string
		queryString    string
		expectedValue  time.Time
		expectedErrMsg string
	}{
		{
			desc:           "no param",
			queryString:    "",
			expectedErrMsg: "missing time",
		},
		{
			desc:        "correct value given",
			queryString: "time=" + url.QueryEscape("2006-01-02T15:04:05+07:00"),
			expectedValue: time.Date(2006, 1, 2, 15, 4, 5, 0,
				time.FixedZone("+0700", 7*3600)),
		},
		{
			desc:           "wrong value given",
			queryString:    "time=2006-01-02",
			expectedErrMsg: "wrong value for time (should be time (rfc3339))",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.desc, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			dateTime, err := ResolveURLQueryGetTimeField(req, "time")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, err, testCase.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, testCase.expectedValue.Equal(dateTime))
		})
	}
}

func TestResolveURLQueryGetStringSliceField(t *testing.T) {
	testCases := []struct {
		desc           string
		queryString    string
		expectedList   []string
		expectedErrMsg string
	}{
		{
			desc:           "no param",
			queryString:    "",
			expectedList:   nil,
			expectedErrMsg: "missing values",
		},
		{
			desc:           "empty param",
			queryString:    "values=",
			expectedList:   []string{""},
			expectedErrMsg: "",
		},
		{
			desc:           "wrong param name",
			queryString:    "value=1,2",
			expectedList:   nil,
			expectedErrMsg: "missing values",
		},
		{
			desc:           "single value",
			queryString:    "values=3",
			expectedList:   []string{"3"},
			expectedErrMsg: "",
		},
		{
			desc:           "multiple values",
			queryString:    "values=4,abc",
			expectedList:   []string{"4", "abc"},
			expectedErrMsg: "",
		},
		{
			desc:           "empty val",
			queryString:    "values=abc,,def",
			expectedList:   []string{"abc", "", "def"},
			expectedErrMsg: "",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.desc, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := ResolveURLQueryGetStringSliceField(req, "values")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, err, testCase.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.expectedList, list)
		})
	}
}

func TestResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters(t *testing.T) {
	testCases := []struct {
		desc           string
		queryString    string
		expectedList   []string
		expectedErrMsg string
	}{
		{
			desc:           "no params",
			queryString:    "",
			expectedList:   []string{"apple", "orange", "pear"},
			expectedErrMsg: "",
		},
		{
			desc:           "empty include param",
			queryString:    "fruits_include=",
			expectedList:   nil,
			expectedErrMsg: `wrong value in 'fruits_include': ""`,
		},
		{
			desc:           "empty exclude param",
			queryString:    "fruits_exclude=",
			expectedList:   nil,
			expectedErrMsg: `wrong value in 'fruits_exclude': ""`,
		},
		{
			desc:           "wrong value in include param",
			queryString:    "fruits_include=cat",
			expectedList:   nil,
			expectedErrMsg: `wrong value in 'fruits_include': "cat"`,
		},
		{
			desc:           "wrong value in exclude param",
			queryString:    "fruits_exclude=dog",
			expectedList:   nil,
			expectedErrMsg: `wrong value in 'fruits_exclude': "dog"`,
		},
		{
			desc:           "include param is given",
			queryString:    "fruits_include=apple,orange",
			expectedList:   []string{"apple", "orange"},
			expectedErrMsg: "",
		},
		{
			desc:           "exclude param is given",
			queryString:    "fruits_exclude=orange,pear",
			expectedList:   []string{"apple"},
			expectedErrMsg: "",
		},
		{
			desc:           "both params are given",
			queryString:    "fruits_include=apple,orange&fruits_exclude=apple",
			expectedList:   []string{"orange"},
			expectedErrMsg: "",
		},
		{
			desc:           "exclude an absent value",
			queryString:    "fruits_include=apple,orange&fruits_exclude=pear",
			expectedList:   []string{"apple", "orange"},
			expectedErrMsg: "",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.desc, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := ResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters(req, "fruits",
				map[string]bool{"apple": true, "orange": true, "pear": true})
			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, err, testCase.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
			sort.Strings(list)
			sort.Strings(testCase.expectedList)
			assert.Equal(t, testCase.expectedList, list)
		})
	}
}

func TestResolveURLQueryPathInt64SliceFieldWithLimit(t *testing.T) {
	expectedRequest := &http.Request{}
	expectedParamName := "param_name"
	expectedError := errors.New("some error")
	tests := []struct {
		name       string
		limit      int
		mockResult []int64
		mockError  error
		want       []int64
		wantErr    error
	}{
		{
			name:       "normal",
			limit:      10,
			mockResult: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			want:       []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:       "limit",
			limit:      2,
			mockResult: []int64{1, 2, 3},
			want:       nil,
			wantErr:    fmt.Errorf("no more than %d %s expected", 2, expectedParamName),
		},
		{
			name:       "error",
			limit:      2,
			mockResult: []int64{1, 2, 3},
			mockError:  expectedError,
			wantErr:    expectedError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			patch := monkey.Patch(ResolveURLQueryPathInt64SliceField, func(r *http.Request, paramName string) ([]int64, error) {
				assert.Equal(t, expectedRequest, r)
				assert.Equal(t, expectedParamName, paramName)
				return tt.mockResult, tt.mockError
			})
			defer patch.Unpatch()

			got, err := ResolveURLQueryPathInt64SliceFieldWithLimit(expectedRequest, expectedParamName, tt.limit)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
