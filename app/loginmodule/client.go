package loginmodule

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
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
		logging.Warnf("Can't retrieve user's profile (status code = %d, response = %q)", response.StatusCode, body)
		return nil, fmt.Errorf("can't retrieve user's profile (status code = %d)", response.StatusCode)
	}
	var decoded map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	err = decoder.Decode(&decoded)
	if err != nil {
		logging.Warnf("Can't parse user's profile (response = %q, error = %q)", body, err)
		return nil, errors.New("can't parse user's profile")
	}

	profile, err = convertUserProfile(decoded)
	if err != nil {
		logging.Warnf("User's profile is invalid (response = %q, error = %q)", body, err)
		return nil, errors.New("user's profile is invalid")
	}
	return profile, nil
}

// UnlinkClient discards our client authorization for the login module user
func (client *Client) UnlinkClient(ctx context.Context, clientID, clientKey string, userLoginID int64) (err error) {
	defer recoverPanics(&err)

	request, err := http.NewRequest("POST", client.url+"/platform_api/accounts_manager/unlink_client"+
		"?client_id="+url.QueryEscape(clientID)+"&user_id="+strconv.FormatInt(userLoginID, 10),
		nil)
	mustNotBeError(err)
	request = request.WithContext(ctx)
	response, err := http.DefaultClient.Do(request)
	mustNotBeError(err)
	body, err := ioutil.ReadAll(io.LimitReader(response.Body, 1<<20)) // 1Mb
	_ = response.Body.Close()
	mustNotBeError(err)
	if response.StatusCode != http.StatusOK {
		logging.Warnf("Can't unlink the user (status code = %d, response = %q)", response.StatusCode, body)
		return fmt.Errorf("can't unlink the user")
	}

	decodedBody := make([]byte, base64.StdEncoding.DecodedLen(len(body)))
	n, err := base64.StdEncoding.Decode(decodedBody, body)
	decodedBody = decodedBody[0:n]
	if err != nil {
		logging.Warnf("Can't decode response from the login module (status code = %d, response = %q): %s", response.StatusCode, body, err)
		return fmt.Errorf("can't unlink the user")
	}
	decryptedBody := decryptAes128Ecb(decodedBody, []byte(clientKey)[:16]) // note that only the first 16 bytes are used
	var decodedResponse struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	err = json.Unmarshal(decryptedBody, &decodedResponse)
	if err != nil {
		logging.Warnf("Can't parse response from the login module (decrypted response = %q, encrypted response = %q): %s",
			decryptedBody, decodedBody, err)
		return fmt.Errorf("can't unlink the user")
	}

	if !decodedResponse.Success {
		logging.Warnf("Can't unlink the user. The login module returned an error: %s", decodedResponse.Error)
		return fmt.Errorf("can't unlink the user")

	}
	return nil
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

func decryptAes128Ecb(data, key []byte) []byte {
	cipher, err := aes.NewCipher(key)
	mustNotBeError(err)
	decrypted := make([]byte, len(data))
	size := 16

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cipher.Decrypt(decrypted[bs:be], data[bs:be])
	}

	// see https://secure.php.net/manual/de/function.openssl-encrypt.php#109598
	if len(decrypted) > 0 {
		paddingCharacter := decrypted[len(decrypted)-1]
		stripPadding := true
		for i := 1; i <= int(paddingCharacter); i++ {
			if decrypted[len(decrypted)-i] != paddingCharacter {
				stripPadding = false
				break
			}
		}
		if stripPadding {
			decrypted = decrypted[0 : len(decrypted)-int(paddingCharacter)]
		}
	}

	return decrypted
}
