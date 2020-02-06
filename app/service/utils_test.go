package service

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/go-chi/chi"
	assertlib "github.com/stretchr/testify/assert"
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
			assert := assertlib.New(t)

			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := ResolveURLQueryGetInt64SliceField(req, "ids")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(err, testCase.expectedErrMsg)
			} else {
				assert.NoError(err)
			}
			assert.Equal(testCase.expectedList, list)
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
			assert := assertlib.New(t)

			r := chi.NewRouter()
			called := false
			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				value, err := ResolveURLQueryPathInt64Field(r, "id")
				if testCase.expectedErrMsg != "" {
					assert.EqualError(err, testCase.expectedErrMsg)
				} else {
					assert.NoError(err)
				}
				assert.Equal(testCase.expectedValue, value)
			}
			r.Get(testCase.routeString, handler)

			ts := httptest.NewServer(r)
			request, _ := http.NewRequest("GET", ts.URL+testCase.queryString, nil)
			response, err := http.DefaultClient.Do(request)
			if err == nil {
				_ = response.Body.Close()
			}
			ts.Close()
			assert.True(called, "The handler was not called")
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
			assert := assertlib.New(t)

			called := false
			r := chi.NewRouter()
			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				value, err := ResolveURLQueryPathInt64SliceField(r, "ids")
				if testCase.expectedErrMsg != "" {
					assert.EqualError(err, testCase.expectedErrMsg)
				} else {
					assert.NoError(err)
				}
				assert.Equal(testCase.expectedList, value)
			}
			r.Get(`/{ids:.*}something`, handler)
			ts := httptest.NewServer(r)
			request, _ := http.NewRequest("GET", ts.URL+testCase.queryString, nil)
			response, err := http.DefaultClient.Do(request)
			if err == nil {
				_ = response.Body.Close()
			}
			ts.Close()
			assert.True(called, "The handler was not called")
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
			assert := assertlib.New(t)

			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := ResolveURLQueryGetStringField(req, "name")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(err, testCase.expectedErrMsg)
			} else {
				assert.NoError(err)
			}
			assert.Equal(testCase.expectedValue, list)
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
			assert := assertlib.New(t)

			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := ResolveURLQueryGetBoolField(req, "flag")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(err, testCase.expectedErrMsg)
			} else {
				assert.NoError(err)
			}
			assert.Equal(testCase.expectedValue, list)
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
			assert := assertlib.New(t)

			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			dateTime, err := ResolveURLQueryGetTimeField(req, "time")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(err, testCase.expectedErrMsg)
			} else {
				assert.NoError(err)
			}
			assert.True(testCase.expectedValue.Equal(dateTime))
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
			assert := assertlib.New(t)

			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := ResolveURLQueryGetStringSliceField(req, "values")
			if testCase.expectedErrMsg != "" {
				assert.EqualError(err, testCase.expectedErrMsg)
			} else {
				assert.NoError(err)
			}
			assert.Equal(testCase.expectedList, list)
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
			assert := assertlib.New(t)

			req, _ := http.NewRequest("GET", "/health-check?"+testCase.queryString, nil)
			list, err := ResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters(req, "fruits",
				map[string]bool{"apple": true, "orange": true, "pear": true})
			if testCase.expectedErrMsg != "" {
				assert.EqualError(err, testCase.expectedErrMsg)
			} else {
				assert.NoError(err)
			}
			sort.Strings(list)
			sort.Strings(testCase.expectedList)
			assert.Equal(testCase.expectedList, list)
		})
	}
}

func TestConvertSliceOfMapsFromDBToJSON(t *testing.T) {
	tests := []struct {
		name  string
		dbMap []map[string]interface{}
		want  []map[string]interface{}
	}{
		{
			"nested structures",
			[]map[string]interface{}{{
				"User__ID":            int64(1),
				"Item__String__Title": "Chapter 1",
				"Item__String__ID":    "2",
			}},
			[]map[string]interface{}{
				{
					"User": map[string]interface{}{"ID": "1"},
					"Item": map[string]interface{}{"String": map[string]interface{}{"Title": "Chapter 1", "ID": "2"}},
				},
			},
		},
		{
			"keeps nil fields",
			[]map[string]interface{}{{"TheGreatestUser": nil, "otherField": 1}},
			[]map[string]interface{}{{"TheGreatestUser": nil, "otherField": 1}},
		},
		{
			"replaces empty sub-maps with nils",
			[]map[string]interface{}{{"the_greatest_user": nil, "empty_sub_map__field1": nil, "empty_sub_map__field2": nil}},
			[]map[string]interface{}{{"the_greatest_user": nil, "empty_sub_map": nil}},
		},
		{
			"converts int64 into string",
			[]map[string]interface{}{{
				"int64":             int64(123),
				"int32":             int32(1234),
				"nbCorrectionsRead": int64(12345),
				"iGrade":            int64(-1),
			}}, // gorm returns numbers as int64
			[]map[string]interface{}{{
				"int64":             "123",
				"int32":             int32(1234),
				"nbCorrectionsRead": "12345",
				"iGrade":            "-1",
			}},
		},
		{
			"handles datetime",
			[]map[string]interface{}{{
				"my_date":   "2019-05-30 11:00:00",
				"null_date": nil,
			}},
			[]map[string]interface{}{{
				"my_date":   "2019-05-30T11:00:00Z",
				"null_date": nil,
			}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertSliceOfMapsFromDBToJSON(tt.dbMap)
			assertlib.Equal(t, tt.want, got)
		})
	}
}

func TestConvertSliceOfMapsFromDBToJSON_PanicsWhenDatetimeIsInvalid(t *testing.T) {
	assertlib.Panics(t, func() {
		ConvertSliceOfMapsFromDBToJSON([]map[string]interface{}{{"some_date": "1234:13:05 24:60:60"}})
	})
}
