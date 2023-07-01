package gfonts

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
)

// GetFont Get the font from the Google Fonts API
// the encode it as a base64 string
// The result is a string that can be used as a CSS font-family
func GetFont(url string) (string, error) {
	// Verify the URL
	if !verifyURL(url) {
		return "", errors.New("invalid URL, must start with https://fonts.googleapis.com/css?family=")
	}

	// Get the font from the Google Fonts API
	font, err := fetch(url)
	if err != nil {
		return "", err
	}
	if font == nil {
		return "", errors.New("font not found")
	}

	// Encode the font as a base64 string
	font64, err := encodeFont(font)
	if err != nil {
		return "", err
	}

	// Return the font as a string
	return font64, nil
}

// verifyURL Verify the URL is a valid Google Fonts URL
// the URL must start with https://fonts.googleapis.com/css?family=
func verifyURL(url string) bool {
	return strings.HasPrefix(url, "https://fonts.googleapis.com/css?family=")
}

// fetch Get the font from the Google Fonts API
func fetch(url string) ([]byte, error) {
	// Get the font from the Google Fonts API
	resp, err := http.Get(url) // #nosec:G107
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Return the font
	return body, nil
}

// encodeFont Encode the font as a base64 string
func encodeFont(font []byte) (string, error) {
	// Encode the font as a base64 string
	font64 := base64.StdEncoding.EncodeToString(font)

	// Return the font as a string
	return font64, nil
}
