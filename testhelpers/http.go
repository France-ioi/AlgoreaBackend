// +build !prod

package testhelpers

import (
	"errors"
	"net/http"
	"strings"
)

// ValidateJSONContentType validates the content-type header of the response is json
// If not, return an error
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
