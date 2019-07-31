package loginmodule

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

// A Client is the login module client
type Client struct {
	url string
}

// NewClient creates a new login module client
func NewClient(loginModuleURL string) *Client {
	return &Client{url: loginModuleURL}
}

// GetUserProfile returns a user profile for given access token
func (client *Client) GetUserProfile(ctx context.Context, accessToken string) (profile map[string]interface{}, err error) {
	defer recoverPanics(&err)

	request, err := http.NewRequest("GET", client.url+"/user_api/account", nil)
	mustNotBeError(err)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request = request.WithContext(ctx)
	response, err := http.DefaultClient.Do(request)
	mustNotBeError(err)
	body, err := ioutil.ReadAll(io.LimitReader(response.Body, 1<<20)) // 1Mb
	_ = response.Body.Close()
	mustNotBeError(err)
	if response.StatusCode != http.StatusOK {
		logging.Warnf("Can't retrieve user's profile (status code = %d, response = %q, accessToken = %q)",
			response.StatusCode, body, accessToken)
		return nil, fmt.Errorf("can't retrieve user's profile (status code = %d)", response.StatusCode)
	}
	var decoded map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	err = decoder.Decode(&decoded)
	if err != nil {
		logging.Warnf("Can't parse user's profile (response = %q, error = %s, accessToken = %q)",
			body, err, accessToken)
		return nil, errors.New("can't parse user's profile")
	}

	profile, err = convertUserProfile(decoded)
	if err != nil {
		logging.Warnf("User's profile is invalid (response = %q, error = %s, accessToken = %q)",
			body, err, accessToken)
		return nil, errors.New("user's profile is invalid")
	}
	return profile, nil
}

func convertUserProfile(source map[string]interface{}) (map[string]interface{}, error) {
	dest := make(map[string]interface{}, len(source)+2)
	mapping := map[string]string{
		"loginID":          "id", // unsigned int
		"sLogin":           "login",
		"sEmail":           "primary_email",
		"sFirstName":       "first_name",
		"sLastName":        "last_name",
		"sStudentId":       "student_id",
		"sCountryCode":     "country_code",
		"sBirthDate":       "birthday",
		"iGraduationYear":  "graduation_year",  // int
		"iGrade":           "graduation_grade", // int
		"sAddress":         "address",
		"sZipcode":         "zipcode",
		"sCity":            "city",
		"sLandLineNumber":  "primary_phone",
		"sCellPhoneNumber": "secondary_phone",
		"sDefaultLanguage": "language",
		"sFreeText":        "presentation",
		"sWebSite":         "website",
		"bEmailVerified":   "primary_email_verified",
	}
	for destKey, sourceKey := range mapping {
		dest[destKey] = source[sourceKey]
		if number, ok := dest[destKey].(json.Number); ok {
			dest[destKey], _ = number.Int64()
		}
	}
	dest["sSex"] = nil
	switch source["gender"] {
	case "m":
		dest["sSex"] = "Male"
	case "f":
		dest["sSex"] = "Female"
	}
	dest["bEmailVerified"] = (dest["bEmailVerified"] == true) || (dest["bEmailVerified"] == int64(1))
	if countryCode, ok := dest["sCountryCode"].(string); ok {
		dest["sCountryCode"] = strings.ToLower(countryCode)
	} else {
		dest["sCountryCode"] = ""
	}

	if dest["loginID"] == nil {
		return nil, errors.New("no id in user's profile")
	}

	if _, ok := dest["sLogin"].(string); !ok {
		return nil, errors.New("no login in user's profile")
	}

	if dest["iGraduationYear"] == nil {
		dest["iGraduationYear"] = int64(0)
	}

	return dest, nil
}

func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}

func recoverPanics(returnErr *error) { // nolint:gocritic
	if p := recover(); p != nil {
		switch e := p.(type) {
		case runtime.Error:
			panic(e)
		default:
			*returnErr = p.(error)
		}
	}
}
