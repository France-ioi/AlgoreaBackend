package service

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
			_, _ = http.DefaultClient.Do(request)
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
					"user": map[string]interface{}{"id": "1"},
					"item": map[string]interface{}{"string": map[string]interface{}{"title": "Chapter 1", "id": "2"}},
				},
			},
		},
		{
			"converts to snake case",
			[]map[string]interface{}{{
				"TheGreatestUser": "root", "MyID": int64(1), "ID": "2",
			}}, // gorm returns numbers as int64
			[]map[string]interface{}{{
				"the_greatest_user": "root", "my_id": "1", "id": "2",
			}},
		},
		{
			"handles prefixes",
			[]map[string]interface{}{{
				"ID":          int64(123),
				"idUser":      int64(1),
				"bTrueFlag":   int64(1),
				"bFalseFlag":  1,
				"bFalseFlag2": int64(2),
				"bFalseFlag3": int64(0),
				"sString":     "value",
			}}, // gorm returns numbers as int64
			[]map[string]interface{}{{
				"id":           "123",
				"user_id":      "1",
				"true_flag":    true,
				"false_flag":   false,
				"false_flag_2": false,
				"false_flag_3": false,
				"string":       "value",
			}},
		},
		{
			"keeps nil fields",
			[]map[string]interface{}{{"TheGreatestUser": nil, "otherField": 1}},
			[]map[string]interface{}{{"the_greatest_user": nil, "other_field": 1}},
		},
		{
			"replaces empty sub-maps with nils",
			[]map[string]interface{}{{"TheGreatestUser": nil, "EmptySubMap__Field1": nil, "EmptySubMap__Field2": nil}},
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
				"int_64":           "123",
				"int_32":           int32(1234),
				"corrections_read": int32(12345),
				"grade":            int32(-1),
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
