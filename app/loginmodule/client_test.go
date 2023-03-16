package loginmodule

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thingful/httpmock"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/loggingtest"
)

func TestNewClient(t *testing.T) {
	assert.Equal(t, &Client{url: "someurl"}, NewClient("someurl"))
}

func Test_recoverPanics_RecoversError(t *testing.T) {
	expectedError := errors.New("some error")
	err := func() (err error) {
		defer recoverPanics(&err)
		panic(expectedError)
	}()
	assert.Equal(t, expectedError, err)
}

func Test_recoverPanics_PanicsOnRuntimeError(t *testing.T) {
	panicValue := func() (panicValue interface{}) {
		defer func() {
			if p := recover(); p != nil {
				panicValue = p
			}
		}()

		_ = func() (err error) {
			defer recoverPanics(&err)
			var a []int
			a[0]++ // nolint:govet // runtime error
			return nil
		}()

		return nil
	}()

	assert.Implements(t, (*runtime.Error)(nil), panicValue)
	assert.Equal(t, "runtime error: index out of range [0] with length 0", panicValue.(error).Error())
}

func TestClient_GetUserProfile(t *testing.T) {
	tests := []struct {
		name            string
		responseCode    int
		response        string
		expectedProfile map[string]interface{}
		expectedErr     error
		expectedLog     string
	}{
		{
			name:         "all fields are set",
			responseCode: 200,
			response: `
				{
					"id":100000001, "login":"jane","login_updated_at":"2019-07-16 01:56:25","login_fixed":0,
					"login_revalidate_required":0,"login_change_required":0,"language":"en","first_name":"Jane",
					"last_name":"Doe","real_name_visible":false,"timezone":"Europe\/London","country_code":"GB",
					"address":null,"city":null,"zipcode":null,"primary_phone":null,"secondary_phone":null,
					"role":"student","school_grade":null,"student_id":"456789012","ministry_of_education":null,
					"ministry_of_education_fr":false,"birthday":"2001-08-03","presentation":"I'm Jane Doe",
					"website":"http:\/\/jane.freepages.com","ip":"192.168.11.1","picture":"http:\/\/127.0.0.1:8000\/images\/user.png",
					"gender":"f","graduation_year":2021,"graduation_grade_expire_at":"2020-07-01 00:00:00",
					"graduation_grade":0,"created_at":"2019-07-16 01:56:25","last_login":"2019-07-22 14:47:18",
					"logout_config":null,"last_password_recovery_at":null,"merge_group_id":null,
					"origin_instance_id":null,"creator_client_id":null,"nationality":"GB",
					"primary_email":"janedoe@gmail.com","secondary_email":"jane.doe@gmail.com",
					"primary_email_verified":1,"secondary_email_verified":null,"has_picture":true,
					"badges": [
						{
							"id": 110504,
							"url": "https:\/\/badges.example.com\/examples\/one",
							"code": "examplebadge001",
							"do_not_possess": false,
							"data": {"category": "", "round": null},
							"manager": false,
							"badge_info": {
								"name": "Example #1",
								"group_path": [
									{"url": "https:\/\/badges.example.com\/", "name": "Example badges", "manager": true},
									{"url": "https:\/\/badges.example.com\/parents", "name": "Example badges with multiple parents", "manager": false}
								]
							},
							"last_update": "2022-07-18T16:07:12+0000"
						}
        ],"client_id":1,"verification":[],"subscription_news":true
				}`,
			expectedProfile: map[string]interface{}{
				"login_id": int64(100000001), "sex": "Female", "land_line_number": nil, "city": nil, "default_language": "en",
				"free_text": "I'm Jane Doe", "graduation_year": int64(2021), "country_code": "gb", "email": "janedoe@gmail.com",
				"student_id": "456789012", "cell_phone_number": nil, "web_site": "http://jane.freepages.com", "grade": int64(0),
				"last_name": "Doe", "birth_date": "2001-08-03", "first_name": "Jane", "zipcode": nil, "address": nil,
				"login": "jane", "email_verified": true, "time_zone": "Europe/London",
				"notify_news": true, "photo_autoload": true, "public_first_name": false, "public_last_name": false,
				"badges": []database.Badge{
					{
						URL:     "https://badges.example.com/examples/one",
						Manager: false,
						BadgeInfo: database.BadgeInfo{
							Name: "Example #1",
							GroupPath: []database.BadgeGroupPathElement{
								{Name: "Example badges", URL: "https://badges.example.com/", Manager: true},
								{Name: "Example badges with multiple parents", URL: "https://badges.example.com/parents", Manager: false},
							},
						},
					},
				},
			},
		},
		{
			name:         "null fields",
			responseCode: 200,
			response: `
				{
					"id":100000001, "login":"jane","login_updated_at":null,"login_fixed":0,
					"login_revalidate_required":0,"login_change_required":0,"language":null,"first_name":null,
					"last_name":null,"real_name_visible":false,"timezone":null,"country_code":null,
					"address":null,"city":null,"zipcode":null,"primary_phone":null,"secondary_phone":null,
					"role":null,"school_grade":null,"student_id":null,"ministry_of_education":null,
					"ministry_of_education_fr":false,"birthday":null,"presentation":null,
					"website":null,"ip":null,"picture":null,
					"gender":null,"graduation_year":null,"graduation_grade_expire_at":null,
					"graduation_grade":null,"created_at":null,"last_login":null,
					"logout_config":null,"last_password_recovery_at":null,"merge_group_id":null,
					"origin_instance_id":null,"creator_client_id":null,"nationality":null,
					"primary_email":null,"secondary_email":null,
					"primary_email_verified":null,"secondary_email_verified":null,"has_picture":false,
					"badges":null,"client_id":null,"verification":null,"subscription_news":null
				}`,
			expectedProfile: map[string]interface{}{
				"graduation_year": int64(0), "address": nil, "sex": nil, "web_site": nil, "last_name": nil,
				"student_id": nil, "cell_phone_number": nil, "country_code": "", "default_language": nil,
				"email_verified": false, "birth_date": nil, "grade": nil, "city": nil, "first_name": nil,
				"login_id": int64(100000001), "email": nil, "login": "jane", "zipcode": nil, "land_line_number": nil,
				"free_text": nil, "time_zone": nil, "notify_news": false, "photo_autoload": false,
				"public_first_name": false, "public_last_name": false, "badges": []database.Badge(nil),
			},
		},
		{
			name:         "wrong response code",
			responseCode: 500,
			response:     "Unknown error",
			expectedErr:  errors.New("can't retrieve user's profile (status code = 500)"),
		},
		{
			name:         "invalid response",
			responseCode: 200,
			response:     "{",
			expectedErr:  errors.New("can't parse user's profile"),
			expectedLog:  `level=warning msg="Can't parse user's profile (response = \"{\", error = \"unexpected EOF\")"`,
		},
		{
			name:         "invalid profile",
			responseCode: 200,
			response:     "{}",
			expectedErr:  errors.New("user's profile is invalid"),
			expectedLog: `level=warning msg="User's profile is invalid (response = \"{}\", ` +
				`error = \"no id in user's profile\")"`,
		},
		{
			name:         "invalid badges",
			responseCode: 200,
			response:     `{"id":100000001,"login":"jane","badges":1234}`,
			expectedErr:  errors.New("user's profile is invalid"),
			expectedLog: `level=warning msg="User's profile is invalid ` +
				`(response = \"{\\\"id\\\":100000001,\\\"login\\\":\\\"jane\\\",\\\"badges\\\":1234}\", ` +
				`error = \"invalid badges data\")"`,
		},
	}

	const moduleURL = "http://login.url.com"
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{url: moduleURL}
			httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
			defer httpmock.DeactivateAndReset()
			responder := httpmock.NewStringResponder(tt.responseCode, tt.response)
			httpmock.RegisterStubRequests(httpmock.NewStubRequest("GET",
				moduleURL+"/user_api/account", responder,
				httpmock.WithHeader(&http.Header{"Authorization": {"Bearer accesstoken"}})))

			hook, restoreLogFunc := logging.MockSharedLoggerHook()
			defer restoreLogFunc()

			gotProfile, err := client.GetUserProfile(context.Background(), "accesstoken")

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedProfile, gotProfile)
			if tt.expectedLog != "" {
				assert.Contains(t, (&loggingtest.Hook{Hook: hook}).GetAllStructuredLogs(), tt.expectedLog)
			}
			assert.NoError(t, httpmock.AllStubsCalled())
		})
	}
}

func Test_convertUserProfile(t *testing.T) {
	tests := []struct {
		name          string
		source        map[string]interface{}
		expected      map[string]interface{}
		expectedError error
	}{
		{
			name: "all fields are set",
			source: map[string]interface{}{
				"id": int64(100000001), "login": "jane", "login_updated_at": "2019-07-16 01:56:25", "login_fixed": int64(0),
				"login_revalidate_required": int64(0), "login_change_required": int64(0), "language": "en", "first_name": "Jane",
				"last_name": "Doe", "real_name_visible": true, "timezone": "Europe/London", "country_code": "GB",
				"address": nil, "city": nil, "zipcode": nil, "primary_phone": nil, "secondary_phone": nil,
				"role": "student", "school_grade": nil, "student_id": "456789012", "ministry_of_education": nil,
				"ministry_of_education_fr": false, "birthday": "2001-08-03", "presentation": "I'm Jane Doe",
				"website": "http://jane.freepages.com", "ip": "192.168.11.1", "picture": "http://127.0.0.1:8000/images/user.png",
				"gender": "f", "graduation_year": int64(2021), "graduation_grade_expire_at": "2020-07-01 00:00:00",
				"graduation_grade": int64(-1), "created_at": "2019-07-16 01:56:25", "last_login": "2019-07-22 14:47:18",
				"logout_config": nil, "last_password_recovery_at": nil, "merge_group_id": nil,
				"origin_instance_id": nil, "creator_client_id": nil, "nationality": "GB",
				"primary_email": "janedoe@gmail.com", "secondary_email": "jane.doe@gmail.com",
				"primary_email_verified": int64(1), "secondary_email_verified": nil, "has_picture": true,
				"badges": []interface{}{
					map[string]interface{}{
						"id":             110501,
						"url":            "https://badges.castor-informatique.fr/qualification_demi_finale/2020",
						"code":           "t9556",
						"do_not_possess": false,
						"data":           map[string]interface{}{"category": "", "round": nil},
						"manager":        false,
						"badge_info": map[string]interface{}{
							"name": "Concours Castor 2020",
							"group_path": []interface{}{
								map[string]interface{}{"url": "https://badges.castor-informatique.fr/", "name": "Concours Castor", "manager": false},
							},
						},
						"last_update": "2022-07-11T12:54:37+0000",
					},
				},
				"client_id": int64(1), "verification": []interface{}(nil),
			},
			expected: map[string]interface{}{
				"free_text": "I'm Jane Doe", "email": "janedoe@gmail.com", "grade": int64(-1),
				"badges": []database.Badge{
					{
						URL:     "https://badges.castor-informatique.fr/qualification_demi_finale/2020",
						Manager: false,
						BadgeInfo: database.BadgeInfo{
							Name: "Concours Castor 2020",
							GroupPath: []database.BadgeGroupPathElement{
								{URL: "https://badges.castor-informatique.fr/", Name: "Concours Castor", Manager: false},
							},
						},
					},
				},
				"web_site": "http://jane.freepages.com", "email_verified": true, "land_line_number": nil, "last_name": "Doe",
				"zipcode": nil, "sex": "Female", "login_id": int64(100000001), "country_code": "gb", "first_name": "Jane",
				"cell_phone_number": nil, "login": "jane", "address": nil, "birth_date": "2001-08-03", "graduation_year": int64(2021),
				"default_language": "en", "city": nil, "student_id": "456789012", "time_zone": "Europe/London",
				"notify_news": false, "public_first_name": true, "public_last_name": true, "photo_autoload": true,
			},
		},
		{
			name: "null fields",
			source: map[string]interface{}{
				"id": int64(100000001), "login": "jane", "login_updated_at": nil, "login_fixed": int64(0),
				"badges":                    []interface{}(nil),
				"login_revalidate_required": int64(0), "login_change_required": int64(0), "language": nil, "first_name": nil,
				"last_name": nil, "real_name_visible": false, "timezone": nil, "country_code": nil,
				"address": nil, "city": nil, "zipcode": nil, "primary_phone": nil, "secondary_phone": nil,
				"role": nil, "school_grade": nil, "student_id": nil, "ministry_of_education": nil,
				"ministry_of_education_fr": false, "birthday": nil, "presentation": nil,
				"website": nil, "ip": nil, "picture": nil,
				"gender": nil, "graduation_year": nil, "graduation_grade_expire_at": nil,
				"graduation_grade": nil, "created_at": nil, "last_login": nil,
				"logout_config": nil, "last_password_recovery_at": nil, "merge_group_id": nil,
				"origin_instance_id": nil, "creator_client_id": nil, "nationality": nil,
				"primary_email": nil, "secondary_email": nil,
				"primary_email_verified": nil, "secondary_email_verified": nil, "has_picture": false,
				"client_id": nil, "verification": nil, "public_first_name": false, "public_last_name": false,
				"subscription_news": nil,
			},
			expected: map[string]interface{}{
				"land_line_number": nil, "login_id": int64(100000001), "login": "jane", "free_text": nil, "sex": nil,
				"badges":     []database.Badge(nil),
				"student_id": nil, "email_verified": false, "cell_phone_number": nil, "grade": nil, "address": nil,
				"zipcode": nil, "birth_date": nil, "email": nil, "graduation_year": int64(0), "city": nil,
				"default_language": nil, "web_site": nil, "last_name": nil, "first_name": nil, "country_code": "",
				"time_zone": nil, "notify_news": false, "photo_autoload": false,
				"public_first_name": false, "public_last_name": false,
			},
		},
		{
			name:   "gender: male",
			source: map[string]interface{}{"id": int64(100000001), "login": "john", "gender": "m"},
			expected: map[string]interface{}{
				"land_line_number": nil, "login_id": int64(100000001), "login": "john",
				"badges": []database.Badge(nil), "free_text": nil, "sex": "Male",
				"student_id": nil, "email_verified": false, "cell_phone_number": nil, "grade": nil, "address": nil,
				"zipcode": nil, "birth_date": nil, "email": nil, "graduation_year": int64(0), "city": nil,
				"default_language": nil, "web_site": nil, "last_name": nil, "first_name": nil, "country_code": "",
				"time_zone": nil, "notify_news": false, "photo_autoload": false,
				"public_first_name": false, "public_last_name": false,
			},
		},
		{
			name:   "primary email verified: true",
			source: map[string]interface{}{"id": int64(100000001), "login": "john", "primary_email_verified": true},
			expected: map[string]interface{}{
				"land_line_number": nil, "login_id": int64(100000001),
				"badges": []database.Badge(nil), "login": "john", "free_text": nil, "sex": nil,
				"student_id": nil, "email_verified": true, "cell_phone_number": nil, "grade": nil, "address": nil,
				"zipcode": nil, "birth_date": nil, "email": nil, "graduation_year": int64(0), "city": nil,
				"default_language": nil, "web_site": nil, "last_name": nil, "first_name": nil, "country_code": "",
				"time_zone": nil, "notify_news": false, "photo_autoload": false,
				"public_first_name": false, "public_last_name": false,
			},
		},
		{
			name:   "country code",
			source: map[string]interface{}{"id": int64(100000001), "login": "john", "country_code": "US"},
			expected: map[string]interface{}{
				"land_line_number": nil, "login_id": int64(100000001),
				"badges": []database.Badge(nil), "login": "john", "free_text": nil, "sex": nil,
				"student_id": nil, "email_verified": false, "cell_phone_number": nil, "grade": nil, "address": nil,
				"zipcode": nil, "birth_date": nil, "email": nil, "graduation_year": int64(0), "city": nil,
				"default_language": nil, "web_site": nil, "last_name": nil, "first_name": nil, "country_code": "us",
				"time_zone": nil, "notify_news": false, "photo_autoload": false,
				"public_first_name": false, "public_last_name": false,
			},
		},
		{
			name:          "no id",
			source:        map[string]interface{}{"login": "john", "country_code": "US"},
			expectedError: errors.New("no id in user's profile"),
		},
		{
			name:          "no login",
			source:        map[string]interface{}{"id": int64(1234), "country_code": "US"},
			expectedError: errors.New("no login in user's profile"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertUserProfile(tt.source)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func encodeAccountsManagerResponse(response, clientSecret string) string {
	const size = 16
	mod := len(response) % size
	if mod != 0 {
		padding := byte(size - mod)
		response += strings.Repeat(string(padding), int(padding))
	}

	data := []byte(response)
	cipher, err := aes.NewCipher([]byte(clientSecret)[0:16])
	if err != nil {
		panic(err)
	}
	encrypted := make([]byte, len(data))
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cipher.Encrypt(encrypted[bs:be], data[bs:be])
	}

	return base64.StdEncoding.EncodeToString(encrypted)
}

func TestClient_AccountsManagerEndpoints(t *testing.T) {
	for _, testSuite := range []struct {
		endpoint     string
		errorMessage string
		params       string
		action       func(*Client) (bool, error)
	}{
		{
			endpoint:     "accounts_manager/unlink_client",
			errorMessage: "can't unlink the user",
			params:       "user_id=123456",
			action: func(client *Client) (bool, error) {
				return client.UnlinkClient(context.Background(), "clientID", "clientKeyclientKey", 123456)
			}},
		{
			endpoint:     "accounts_manager/delete",
			errorMessage: "can't delete users",
			params:       "prefix=prefix_subprefix_",
			action: func(client *Client) (bool, error) {
				return client.DeleteUsers(context.Background(), "clientID", "clientKeyclientKey", "prefix_subprefix_")
			},
		},
		{
			endpoint:     "lti_result/send",
			errorMessage: "can't publish score",
			params:       "user_id=1234&content_id=5678&score=99.9",
			action: func(client *Client) (bool, error) {
				return client.SendLTIResult(context.Background(), "clientID", "clientKeyclientKey", 1234, 5678, 99.9)
			},
		},
	} {
		testSuite := testSuite
		t.Run(testSuite.endpoint, func(t *testing.T) {
			tests := []struct {
				name           string
				responseCode   int
				response       string
				expectedResult bool
				expectedErr    error
				expectedLog    string
			}{
				{
					name:           "success",
					responseCode:   200,
					response:       encodeAccountsManagerResponse(`{"success":true}`, "clientKeyclientKey"),
					expectedResult: true,
				},
				{
					name:         "wrong status code",
					responseCode: 500,
					response:     "Unexpected error",
					expectedErr:  fmt.Errorf(testSuite.errorMessage+": %s", "bad response code"),
					expectedLog: `level=warning msg="Login module returned a bad status code for /platform_api/` +
						testSuite.endpoint + ` (status code = 500, response = \"Unexpected error\")"`,
				},
				{
					name:         "corrupted base64",
					responseCode: 200,
					response:     "Some text",
					expectedErr:  fmt.Errorf(testSuite.errorMessage+": %s", "illegal base64 data at input byte 4"),
					expectedLog: `level=warning msg="Can't decode response from the login module for /platform_api/` +
						testSuite.endpoint + ` (status code = 200, response = \"Some text\"): illegal base64 data at input byte 4"`,
				},
				{
					name:         "can't unmarshal",
					responseCode: 200,
					response:     encodeAccountsManagerResponse(`{"success":true}`, "anotherClientKey"),
					expectedErr:  fmt.Errorf(testSuite.errorMessage+": %s", "invalid character 'Ý' in literal true (expecting 'r')"),
					expectedLog: `level=warning msg="Can't parse response from the login module for /platform_api/` +
						testSuite.endpoint +
						` (decrypted response = \"t\\xdd\\t\\xc0\\x02\\xe9M.{0\\xa5\\xba\\xff\\xcb@|\", ` +
						`encrypted response = \"K\\f_Bd\\xa5et\\xa5̡\\xfa蠐x\"): invalid character 'Ý' in literal true (expecting 'r')"`,
				},
				{
					name:         "'success' is false",
					responseCode: 200,
					response:     encodeAccountsManagerResponse(`{"error":"unknown error"}`, "clientKeyclientKey"),
					expectedLog: `level=warning msg="The login module returned an error for /platform_api/` +
						testSuite.endpoint + `: unknown error"`,
				},
			}
			const moduleURL = "http://login.url.com"
			for _, tt := range tests {
				tt := tt
				t.Run(tt.name, func(t *testing.T) {
					client := &Client{url: moduleURL}
					httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
					defer httpmock.DeactivateAndReset()
					responder := httpmock.NewStringResponder(tt.responseCode, tt.response)

					parsedParams, err := url.ParseQuery(testSuite.params)
					assert.NoError(t, err)
					paramsMap := make(map[string]string, len(parsedParams))
					for key := range parsedParams {
						paramsMap[key] = parsedParams.Get(key)
					}
					requestBody, err := EncodeBody(paramsMap, "clientID", "clientKeyclientKey")
					assert.NoError(t, err)

					httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
						moduleURL+"/platform_api/"+testSuite.endpoint, responder,
						httpmock.WithHeader(&http.Header{"Content-Type": []string{"application/json"}}),
						httpmock.WithBody(bytes.NewReader(requestBody))))

					hook, restoreLogFunc := logging.MockSharedLoggerHook()
					defer restoreLogFunc()

					result, err := testSuite.action(client)

					assert.Equal(t, tt.expectedResult, result)
					assert.Equal(t, tt.expectedErr, err)
					if tt.expectedLog != "" {
						assert.Contains(t, (&loggingtest.Hook{Hook: hook}).GetAllStructuredLogs(), tt.expectedLog)
					}
					assert.NoError(t, httpmock.AllStubsCalled())
				})
			}
		})
	}
}

func TestClient_CreateUsers(t *testing.T) {
	tests := []struct {
		name              string
		params            *CreateUsersParams
		loginModuleParams string
		responseCode      int
		response          string
		expectedResult    bool
		expectedData      []CreateUsersResponseDataRow
		expectedErr       error
		expectedLog       string
	}{
		{
			name:           "success",
			responseCode:   200,
			response:       encodeAccountsManagerResponse(`{"success":true, "data":[]}`, "clientKeyclientKey"),
			expectedResult: true,
			expectedData:   []CreateUsersResponseDataRow{},
		},
		{
			name: "success with all the parameters set",
			params: &CreateUsersParams{
				Prefix:         "pref",
				Amount:         10,
				PostfixLength:  3,
				PasswordLength: 4,
				LoginFixed:     func(b bool) *bool { return &b }(false),
				Language:       func(s string) *string { return &s }("fr"),
			},
			loginModuleParams: "amount=10&language=fr&login_fixed=0&password_length=4&postfix_length=3&prefix=pref",
			responseCode:      200,
			response: encodeAccountsManagerResponse(`{
				"success":true, "data":[
					{"id":12345678901234, "login":"pref_abcd", "password": "efgh"},
					{"id":12345678901235, "login":"pref_bcde", "password": "jklm"}
				]}`, "clientKeyclientKey"),
			expectedResult: true,
			expectedData: []CreateUsersResponseDataRow{
				{ID: 12345678901234, Login: "pref_abcd", Password: "efgh"},
				{ID: 12345678901235, Login: "pref_bcde", Password: "jklm"},
			},
		},
		{
			name: "success with all the parameters set and login_fixed=true",
			params: &CreateUsersParams{
				Prefix:         "pref",
				Amount:         10,
				PostfixLength:  3,
				PasswordLength: 4,
				LoginFixed:     func(b bool) *bool { return &b }(true),
				Language:       func(s string) *string { return &s }("en"),
			},
			loginModuleParams: "amount=10&language=en&login_fixed=1&password_length=4&postfix_length=3&prefix=pref",
			responseCode:      200,
			response:          encodeAccountsManagerResponse(`{"success":true, "data":[]}`, "clientKeyclientKey"),
			expectedResult:    true,
			expectedData:      []CreateUsersResponseDataRow{},
		},
		{
			name:         "wrong status code",
			responseCode: 500,
			response:     "Unexpected error",
			expectedErr:  fmt.Errorf("can't create users: %s", "bad response code"),
			expectedLog: `level=warning msg="Login module returned a bad status code for /platform_api/accounts_manager/create ` +
				`(status code = 500, response = \"Unexpected error\")"`,
		},
		{
			name:         "corrupted base64",
			responseCode: 200,
			response:     "Some text",
			expectedErr:  fmt.Errorf("can't create users: %s", "illegal base64 data at input byte 4"),
			expectedLog: `level=warning msg="Can't decode response from the login module for /platform_api/accounts_manager/create ` +
				`(status code = 200, response = \"Some text\"): illegal base64 data at input byte 4"`,
		},
		{
			name:         "can't unmarshal",
			responseCode: 200,
			response:     encodeAccountsManagerResponse(`{"success":true}`, "anotherClientKey"),
			expectedErr:  fmt.Errorf("can't create users: %s", "invalid character 'Ý' in literal true (expecting 'r')"),
			expectedLog: `level=warning msg="Can't parse response from the login module for /platform_api/accounts_manager/create ` +
				`(decrypted response = \"t\\xdd\\t\\xc0\\x02\\xe9M.{0\\xa5\\xba\\xff\\xcb@|\", ` +
				`encrypted response = \"K\\f_Bd\\xa5et\\xa5̡\\xfa蠐x\"): invalid character 'Ý' in literal true (expecting 'r')"`,
		},
		{
			name:         "'success' is false",
			responseCode: 200,
			response:     encodeAccountsManagerResponse(`{"error":"unknown error"}`, "clientKeyclientKey"),
			expectedLog:  `level=warning msg="The login module returned an error for /platform_api/accounts_manager/create: unknown error"`,
		},
	}

	const moduleURL = "http://login.url.com"
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{url: moduleURL}
			httpmock.Activate(httpmock.WithAllowedHosts("127.0.0.1"))
			defer httpmock.DeactivateAndReset()
			responder := httpmock.NewStringResponder(tt.responseCode, tt.response)
			if tt.loginModuleParams == "" {
				tt.loginModuleParams = "amount=0&password_length=0&postfix_length=0&prefix="
			}

			parsedParams, err := url.ParseQuery(tt.loginModuleParams)
			assert.NoError(t, err)
			paramsMap := make(map[string]string, len(parsedParams))
			for key := range parsedParams {
				paramsMap[key] = parsedParams.Get(key)
			}
			requestBody, err := EncodeBody(paramsMap, "clientID", "clientKeyclientKey")
			assert.NoError(t, err)

			httpmock.RegisterStubRequests(httpmock.NewStubRequest("POST",
				moduleURL+"/platform_api/accounts_manager/create", responder,
				httpmock.WithHeader(&http.Header{"Content-Type": []string{"application/json"}}),
				httpmock.WithBody(bytes.NewReader(requestBody))))

			hook, restoreLogFunc := logging.MockSharedLoggerHook()
			defer restoreLogFunc()

			if tt.params == nil {
				tt.params = &CreateUsersParams{}
			}
			result, data, err := client.CreateUsers(context.Background(), "clientID", "clientKeyclientKey", tt.params)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedData, data)
			if tt.expectedLog != "" {
				assert.Contains(t, (&loggingtest.Hook{Hook: hook}).GetAllStructuredLogs(), tt.expectedLog)
			}
			assert.NoError(t, httpmock.AllStubsCalled())
		})
	}
}

func Test_mustNotBeError(t *testing.T) {
	mustNotBeError(nil)
	expectedError := errors.New("some error")
	assert.PanicsWithValue(t, expectedError, func() {
		mustNotBeError(expectedError)
	})
}

func TestEncodeBody(t *testing.T) {
	params := map[string]string{
		"p1": "v1",
		"p2": "v2",
	}
	const clientID = "1234"
	const clientKey = "abcdefghijklmnop"
	encoded, err := EncodeBody(params, clientID, clientKey)
	assert.NoError(t, err)
	var unmarshaled map[string]string
	err = json.Unmarshal(encoded, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, clientID, unmarshaled["client_id"])
	assert.Equal(t, 2, len(unmarshaled))
	assert.Contains(t, unmarshaled, "data")
	decodedData := make([]byte, base64.StdEncoding.DecodedLen(len(unmarshaled["data"])))
	n, err := base64.StdEncoding.Decode(decodedData, []byte(unmarshaled["data"]))
	assert.NoError(t, err)
	decodedData = decodedData[0:n]
	decryptedData := decryptAes128Ecb(decodedData, []byte(clientKey)[:16])
	decoder := json.NewDecoder(bytes.NewReader(decryptedData))
	decoder.UseNumber()
	var parsedData map[string]string
	err = decoder.Decode(&parsedData)
	assert.NoError(t, err)
	assert.Equal(t, params, parsedData)
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectedLen int
	}{
		{name: "empty", data: []byte(""), expectedLen: 24},
		{name: "15 bytes", data: []byte("123456789012345"), expectedLen: 24},
		{name: "16 bytes", data: []byte("1234567890123456"), expectedLen: 44},
		{name: "17 bytes", data: []byte("12345678901234567"), expectedLen: 44},
	}

	const clientKey = "1234567890123456"
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(tt.data, clientKey)
			got1 := Encode(tt.data, clientKey+"7")
			got2 := Encode(tt.data, "0"+clientKey)
			assert.Equal(t, got, got1)
			assert.NotEqual(t, got, got2)
			assert.NotEqual(t, tt.data, got)
			assert.Len(t, got, tt.expectedLen)
			decodedData := make([]byte, base64.StdEncoding.DecodedLen(len(got)))
			n, err := base64.StdEncoding.Decode(decodedData, []byte(got))
			assert.NoError(t, err)
			decodedData = decodedData[0:n]
			decryptedData := decryptAes128Ecb(decodedData, []byte(clientKey)[:16])
			assert.Equal(t, tt.data, decryptedData)
		})
	}
}
