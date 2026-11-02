// Package gfonts retrieves the CSS of a font family from the Google Fonts API and
// returns it encoded as a base64 string.
package gfonts

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GetFont gets the CSS stylesheet at the given Google Fonts URL and returns it encoded
// as a base64 string, for example to embed it in a data:text/css;base64 URI.
//
// The result holds the @font-face rules, not the font files: those rules still
// reference the font files hosted by Google.
//
// The URL must start with https://fonts.googleapis.com/css?family=. An error is
// returned if the URL is invalid, if the request fails or if the response status is
// not 200 OK, which is the case for an unknown font family.
func GetFont(url string) (string, error) {
	// Verify the URL
	if !verifyURL(url) {
		return "", errors.New("invalid URL, must start with https://fonts.googleapis.com/css?family=")
	}

	// Get the stylesheet from the Google Fonts API
	css, err := fetch(url)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(css), nil
}

// verifyURL Verify the URL is a valid Google Fonts URL
// the URL must start with https://fonts.googleapis.com/css?family=
func verifyURL(url string) bool {
	return strings.HasPrefix(url, "https://fonts.googleapis.com/css?family=")
}

// httpClient is used to call the Google Fonts API; the timeout prevents a slow or
// unresponsive server from blocking the caller indefinitely.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// fetch gets the body of a successful response from the Google Fonts API
func fetch(url string) ([]byte, error) {
	resp, err := httpClient.Get(url) // #nosec:G107
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// An error page must not be mistaken for a stylesheet
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
