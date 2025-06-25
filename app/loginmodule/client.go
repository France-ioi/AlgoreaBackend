// Package loginmodule provides functions related to user login.
package loginmodule

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

const oneMegabyte = 1 << 20

// A Client is the login module client.
type Client struct {
	url string
}

// NewClient creates a new login module client.
func NewClient(loginModuleURL string) *Client {
	return &Client{url: loginModuleURL}
}

// GetUserProfile returns a user profile for given access token.
func (client *Client) GetUserProfile(ctx context.Context, accessToken string) (profile map[string]interface{}, err error) {
	defer recoverPanics(&err)

	request, err := http.NewRequest("GET", client.url+"/user_api/account", http.NoBody)
	mustNotBeError(err)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request = request.WithContext(ctx)
	response, err := http.DefaultClient.Do(request)
	mustNotBeError(err)
	body, err := io.ReadAll(io.LimitReader(response.Body, oneMegabyte))
	_ = response.Body.Close()
	mustNotBeError(err)
	if response.StatusCode != http.StatusOK {
		logging.SharedLogger.WithContext(ctx).
			Warnf("Can't retrieve user's profile (status code = %d, response = %q)", response.StatusCode, body)
		return nil, fmt.Errorf("can't retrieve user's profile (status code = %d)", response.StatusCode)
	}
	var decoded map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	err = decoder.Decode(&decoded)
	if err != nil {
		logging.SharedLogger.WithContext(ctx).
			Warnf("Can't parse user's profile (response = %q, error = %q)", body, err)
		return nil, errors.New("can't parse user's profile")
	}

	profile, err = convertUserProfile(decoded)
	if err != nil {
		logging.SharedLogger.WithContext(ctx).
			Warnf("User's profile is invalid (response = %q, error = %q)", body, err)
		return nil, errors.New("user's profile is invalid")
	}
	return profile, nil
}

// CreateUsersParams represents parameters for Client.CreateUsers().
type CreateUsersParams struct {
	Prefix         string
	Amount         int
	PostfixLength  int
	PasswordLength int
	LoginFixed     *bool
	Language       *string
}

// CreateUsersResponseDataRow represents an element of the array returned by Client.CreateUsers() (id, login, password).
type CreateUsersResponseDataRow struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// CreateUsers creates a batch of users in the login module.
func (client *Client) CreateUsers(ctx context.Context, clientID, clientKey string,
	params *CreateUsersParams,
) (bool, []CreateUsersResponseDataRow, error) {
	urlParams := map[string]string{
		"prefix":          params.Prefix,
		"amount":          strconv.Itoa(params.Amount),
		"postfix_length":  strconv.Itoa(params.PostfixLength),
		"password_length": strconv.Itoa(params.PasswordLength),
	}
	if params.LoginFixed != nil {
		loginFixed := "0"
		if *params.LoginFixed {
			loginFixed = "1"
		}
		urlParams["login_fixed"] = loginFixed
	}
	if params.Language != nil {
		urlParams["language"] = *params.Language
	}
	response, err := client.requestAccountsManagerAndDecode(ctx, "/platform_api/accounts_manager/create",
		urlParams, clientID, clientKey)
	if err != nil {
		return false, nil, fmt.Errorf("can't create users: %s", err.Error())
	}

	var resultRows []CreateUsersResponseDataRow

	if response.Success {
		mustNotBeError(json.Unmarshal(response.Data, &resultRows))
	}

	return response.Success, resultRows, nil
}

// DeleteUsers deletes users specified by the given login prefix from the login module.
func (client *Client) DeleteUsers(ctx context.Context, clientID, clientKey, loginPrefix string) (bool, error) {
	params := map[string]string{
		"prefix": loginPrefix,
	}
	response, err := client.requestAccountsManagerAndDecode(ctx, "/platform_api/accounts_manager/delete",
		params, clientID, clientKey)
	if err != nil {
		return false, fmt.Errorf("can't delete users: %s", err.Error())
	}
	return response.Success, nil
}

// UnlinkClient discards our client authorization for the login module user.
func (client *Client) UnlinkClient(ctx context.Context, clientID, clientKey string, userLoginID int64) (bool, error) {
	response, err := client.requestAccountsManagerAndDecode(ctx, "/platform_api/accounts_manager/unlink_client",
		map[string]string{"user_id": strconv.FormatInt(userLoginID, 10)}, clientID, clientKey)
	if err != nil {
		return false, fmt.Errorf("can't unlink the user: %s", err.Error())
	}
	return response.Success, nil
}

// SendLTIResult sends item score to LTI.
func (client *Client) SendLTIResult(
	ctx context.Context, clientID, clientKey string, userLoginID, itemID int64, score float32,
) (bool, error) {
	response, err := client.requestAccountsManagerAndDecode(ctx, "/platform_api/lti_result/send",
		map[string]string{
			"user_id":    strconv.FormatInt(userLoginID, 10),
			"content_id": strconv.FormatInt(itemID, 10),
			"score":      strconv.FormatFloat(float64(score), 'f', -1, 32),
		}, clientID, clientKey)
	if err != nil {
		return false, fmt.Errorf("can't publish score: %s", err.Error())
	}
	return response.Success, nil
}

type accountManagerResponse struct {
	Success bool            `json:"success"`
	Error   string          `json:"error"`
	Data    json.RawMessage `json:"data"`
}

func (client *Client) requestAccountsManagerAndDecode(ctx context.Context, urlPath string, requestParams map[string]string,
	clientID, clientKey string,
) (decodedResponse *accountManagerResponse, err error) {
	defer recoverPanics(&err)

	apiURL, err := url.Parse(client.url + urlPath)
	mustNotBeError(err)
	values := apiURL.Query()
	apiURL.RawQuery = values.Encode()

	params, err := EncodeBody(requestParams, clientID, clientKey)

	request, err := http.NewRequest("POST", apiURL.String(), bytes.NewBuffer(params))
	mustNotBeError(err)
	request = request.WithContext(ctx)
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	mustNotBeError(err)
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, oneMegabyte))
	_ = response.Body.Close()
	mustNotBeError(err)
	if response.StatusCode != http.StatusOK {
		logging.SharedLogger.WithContext(ctx).
			Warnf("Login module returned a bad status code for %s (status code = %d, response = %q)",
				urlPath, response.StatusCode, responseBody)
		panic(errors.New("bad response code"))
	}

	decodedBody := make([]byte, base64.StdEncoding.DecodedLen(len(responseBody)))
	n, err := base64.StdEncoding.Decode(decodedBody, responseBody)
	decodedBody = decodedBody[0:n]
	if err != nil {
		logging.SharedLogger.WithContext(ctx).
			Warnf("Can't decode response from the login module for %s (status code = %d, response = %q): %s",
				urlPath, response.StatusCode, responseBody, err)
		panic(err)
	}
	decryptedBody := decryptAes128Ecb(decodedBody, []byte(clientKey)[:16]) // note that only the first 16 bytes are used
	decoder := json.NewDecoder(bytes.NewReader(decryptedBody))
	decoder.UseNumber()
	err = decoder.Decode(&decodedResponse)
	if err != nil {
		logging.SharedLogger.WithContext(ctx).
			Warnf("Can't parse response from the login module for %s (decrypted response = %q, encrypted response = %q): %s",
				urlPath, decryptedBody, decodedBody, err)
		panic(err)
	}
	if !decodedResponse.Success {
		logging.SharedLogger.WithContext(ctx).
			Warnf("The login module returned an error for %s: %s", urlPath, decodedResponse.Error)
	}
	return decodedResponse, nil
}

// EncodeBody forms a request body with the given parameters for the login module: `{"client_id": ..., "data": _encoded_}`.
func EncodeBody(requestParams map[string]string, clientID, clientKey string) (result []byte, err error) {
	defer recoverPanics(&err)
	paramsJSON, _ := json.Marshal(requestParams)
	encodedParams := Encode(paramsJSON, clientKey)
	params, _ := json.Marshal(map[string]string{"client_id": clientID, "data": encodedParams})
	return params, err
}

// Encode encodes the given bytes array using the given key for the login module (AES128ECB + BASE64).
func Encode(data []byte, clientKey string) string {
	encrypted := encryptAes128Ecb(data, []byte(clientKey)[:16])
	return base64.StdEncoding.EncodeToString(encrypted)
}

func convertUserProfile(source map[string]interface{}) (map[string]interface{}, error) {
	//nolint:gomnd // we are going to add two fields: public_first_name and public_last_name
	dest := make(map[string]interface{}, len(source)+2)
	/*
	 We ignore fields: birthday_year, client_id, created_at, creator_client_id,
	 school_grade, graduation_grade_expire_at, ip, last_password_recovery_at, last_login,
	 login_change_required, login_fixed, login_revalidate_required, login_updated_at,
	 logout_config, merge_group_id, ministry_of_education, ministry_of_education_fr,
	 nationality (capitalized country code), origin_instance_id, picture (URL),
	 role, secondary_email, secondary_email_verified, subscription_results (bool), verification.

	 "badges" are returned as []database.Badge (always set, but can be nil), all unnecessary inner properties are skipped.
	*/

	mapping := map[string]string{
		"login_id":          "id", // unsigned int
		"login":             "login",
		"email":             "primary_email",
		"first_name":        "first_name",
		"last_name":         "last_name",
		"student_id":        "student_id",
		"country_code":      "country_code",
		"birth_date":        "birthday",
		"graduation_year":   "graduation_year",  // int
		"grade":             "graduation_grade", // int
		"address":           "address",
		"zipcode":           "zipcode",
		"city":              "city",
		"land_line_number":  "primary_phone",
		"cell_phone_number": "secondary_phone",
		"default_language":  "language",
		"free_text":         "presentation",
		"web_site":          "website",
		"time_zone":         "timezone",
		"email_verified":    "primary_email_verified",
		"photo_autoload":    "has_picture",
		"notify_news":       "subscription_news",
	}
	for destKey, sourceKey := range mapping {
		dest[destKey] = source[sourceKey]
		if number, ok := dest[destKey].(json.Number); ok {
			dest[destKey], _ = number.Int64()
		}
	}

	convertUserGender(source, dest)
	normalizeUserProfileBooleanFields(dest)

	if countryCode, ok := dest["country_code"].(string); ok {
		dest["country_code"] = strings.ToLower(countryCode)
	} else {
		dest["country_code"] = ""
	}

	if dest["login_id"] == nil {
		return nil, errors.New("no id in user's profile")
	}

	if _, ok := dest["login"].(string); !ok {
		return nil, errors.New("no login in user's profile")
	}

	if dest["graduation_year"] == nil {
		dest["graduation_year"] = int64(0)
	}

	var realNameVisible bool
	if value, ok := source["real_name_visible"]; ok && value == true {
		realNameVisible = true
	}
	dest["public_first_name"] = realNameVisible
	dest["public_last_name"] = realNameVisible

	err := convertBadges(source, dest)
	if err != nil {
		return nil, fmt.Errorf("invalid badges data")
	}

	return dest, nil
}

func normalizeUserProfileBooleanFields(dest map[string]interface{}) {
	for _, fieldName := range [...]string{"email_verified", "notify_news", "photo_autoload"} {
		dest[fieldName] = (dest[fieldName] == true) || (dest[fieldName] == int64(1))
	}
}

func convertUserGender(source, dest map[string]interface{}) {
	dest["sex"] = nil
	switch source["gender"] {
	case "m":
		dest["sex"] = "Male"
	case "f":
		dest["sex"] = "Female"
	}
}

func convertBadges(source, dest map[string]interface{}) error {
	if badges, ok := source["badges"]; !ok || badges == nil {
		dest["badges"] = []database.Badge(nil)
		return nil
	}
	data := struct {
		Badges []database.Badge `json:"badges"`
	}{}
	form := formdata.NewFormData(&data)
	form.AllowUnknownFields()
	if err := form.ParseMapData(source); err != nil {
		return err
	}
	dest["badges"] = data.Badges
	return nil
}

func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}

func recoverPanics(
	returnErr *error, //nolint:gocritic // we need the pointer as we replace the error with a panic
) {
	if p := recover(); p != nil {
		switch e := p.(type) {
		case runtime.Error:
			panic(e)
		default:
			*returnErr = p.(error)
		}
	}
}
