//go:build !prod

package testhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// SendTestHTTPRequest sends an HTTP request to a given path on the test server.
// It returns the response and the response body as a string. The response body is read completely and closed.
// The timeout for requests is set to 5 seconds.
func SendTestHTTPRequest(ts *httptest.Server, method, path string, headers map[string][]string, body io.Reader) (
	response *http.Response, responseBody string, err error,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, body)
	if err != nil {
		return nil, "", err
	}

	// add headers
	for name, values := range headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	// execute the query
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	defer func() { /* #nosec */ _ = resp.Body.Close() }()

	return resp, string(respBody), nil
}

// ValidateJSONContentType validates the content-type header of the response is json
// If not, return an error.
func ValidateJSONContentType(resp *http.Response) error {
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return errors.New("missing Content-Type header")
	}
	mediaType := strings.Split(contentType, ";")[0]
	if mediaType != "application/json" {
		return errors.New("Unexpected Content-Type header. Expected 'application/json', got: " + mediaType)
	}
	return nil
}

// VerifyTestHTTPRequestWithToken makes an HTTP request to the given path on the test server with the given token
// and verifies that the response status code is as expected.
// Note, that the request is made with a 5-second timeout.
func VerifyTestHTTPRequestWithToken(t *testing.T, hookedAppServer *httptest.Server,
	token string, expectedStatusCode int,
	method string, path string, headers map[string][]string, body interface{},
) {
	t.Helper()

	headersWithToken := make(map[string][]string, len(headers)+1)
	for key := range headers {
		headersWithToken[key] = make([]string, len(headers[key]))
		copy(headersWithToken[key], headers[key])
	}
	headersWithToken["Authorization"] = []string{"Bearer " + token}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(mustMarshalJSON(body))
	}

	response, responseBody, err := SendTestHTTPRequest(
		hookedAppServer, method, path, headersWithToken, bodyReader)
	if err != nil {
		t.Errorf("cannot send %s request to %s with token '%s': %v", method, path, token, err)
		return
	}
	if response != nil {
		_ = response.Body.Close()
		if response.StatusCode != expectedStatusCode {
			t.Errorf("unexpected status code %d (expected %d) of %s request to %s with token '%s', status: %s, response body: %s",
				response.StatusCode, expectedStatusCode, method, path, token, response.Status, responseBody)
		}
	}
}

func mustMarshalJSON(v any) []byte {
	result, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return result
}
