// Package gfonts retrieves font families from the Google Fonts API, encoded as base64
// strings: either the CSS alone (GetFont), or the CSS with its font files embedded
// (GetEmbeddedFont).
package gfonts

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// stylesheetHost and stylesheetPath locate the stylesheets of the Google Fonts API.
	stylesheetHost = "fonts.googleapis.com"
	stylesheetPath = "/css"

	// fontFileHost is the only host font files are downloaded from: the URLs come from
	// the content of a response, they must not make the caller reach another server.
	fontFileHost = "fonts.gstatic.com"
)

// fontFileURL matches the references to a file in a stylesheet.
var fontFileURL = regexp.MustCompile(`url\(([^)]*)\)`)

// GetFont gets the CSS stylesheet at the given Google Fonts URL and returns it encoded
// as a base64 string, for example to embed it in a data:text/css;base64 URI.
//
// The result holds the @font-face rules, not the font files: those rules still
// reference the font files hosted by Google. Use GetEmbeddedFont to embed them.
//
// The URL must be https://fonts.googleapis.com/css with a family parameter. An error is
// returned if the URL is invalid, if the request fails or if the response status is
// not 200 OK, which is the case for an unknown font family.
func GetFont(url string) (string, error) {
	css, err := getStylesheet(url)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(css), nil
}

// GetEmbeddedFont gets the stylesheet at the given Google Fonts URL, embeds the font
// files it references, and returns the result encoded as a base64 string, for example
// to use it in a data:text/css;base64 URI.
//
// The result is self-contained: displaying the fonts requires no request to Google. It
// is much larger than the result of GetFont and every font file is downloaded, one at
// a time, so the call can take several times the timeout of a single request.
//
// The URL must be https://fonts.googleapis.com/css with a family parameter. An error is
// returned if the URL is invalid, if a request fails, if a response status is not
// 200 OK, or if the stylesheet references no font file or an unexpected one.
func GetEmbeddedFont(url string) (string, error) {
	css, err := getStylesheet(url)
	if err != nil {
		return "", err
	}

	embedded, err := embedFontFiles(string(css))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString([]byte(embedded)), nil
}

// getStylesheet verifies the URL and gets the stylesheet from the Google Fonts API
func getStylesheet(url string) ([]byte, error) {
	if !verifyURL(url) {
		return nil, errors.New("invalid URL, must be https://" + stylesheetHost + stylesheetPath + " with a family parameter")
	}

	return fetch(url)
}

// embedFontFiles replaces the font file URLs of a stylesheet with data URIs
func embedFontFiles(css string) (string, error) {
	references := fontFileURL.FindAllStringSubmatch(css, -1)
	if len(references) == 0 {
		return "", errors.New("no font file found in the stylesheet")
	}

	for _, reference := range references {
		// A file shared by several rules has been replaced everywhere already
		if !strings.Contains(css, reference[0]) {
			continue
		}

		dataURI, err := fontFileDataURI(reference[1])
		if err != nil {
			return "", fmt.Errorf("font file %s: %w", reference[1], err)
		}
		css = strings.ReplaceAll(css, reference[0], "url("+dataURI+")")
	}

	return css, nil
}

// fontFileDataURI downloads a font file and returns it as a data URI
func fontFileDataURI(rawURL string) (string, error) {
	if u, err := url.Parse(rawURL); err != nil || u.Scheme != "https" || u.Host != fontFileHost {
		return "", errors.New("unexpected location, must be on https://" + fontFileHost)
	}

	font, err := fetch(rawURL)
	if err != nil {
		return "", err
	}

	mediaType := http.DetectContentType(font)
	if !strings.HasPrefix(mediaType, "font/") {
		return "", fmt.Errorf("unexpected content type: %s", mediaType)
	}

	return "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(font), nil
}

// verifyURL reports whether the URL is a Google Fonts stylesheet URL naming a family
func verifyURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return u.Scheme == "https" &&
		u.Host == stylesheetHost &&
		u.Path == stylesheetPath &&
		u.Query().Get("family") != ""
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
