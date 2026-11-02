package gfonts

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

// roundTripFunc answers the requests of httpClient without using the network.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// stubResponse makes httpClient answer every request with the given status and body.
func stubResponse(t *testing.T, status int, body string) {
	t.Helper()

	original := httpClient
	t.Cleanup(func() { httpClient = original })

	httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(status)
		_, _ = rec.WriteString(body)
		return rec.Result(), nil
	})}
}

func TestGetFont(t *testing.T) {
	url := "https://fonts.googleapis.com/css?family=Roboto:300,400,500,700"
	font, err := GetFont(url)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if font == "" {
		t.Errorf("Font is empty")
	}
}

func TestGetFontEncodesStylesheet(t *testing.T) {
	const css = "@font-face { font-family: 'Roboto'; }"
	stubResponse(t, http.StatusOK, css)

	got, err := GetFont("https://fonts.googleapis.com/css?family=Roboto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := base64.StdEncoding.EncodeToString([]byte(css))
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGetFontRejectsErrorStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{name: "unknown family", status: http.StatusBadRequest},
		{name: "server failure", status: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubResponse(t, tt.status, "<!DOCTYPE html>")

			got, err := GetFont("https://fonts.googleapis.com/css?family=Unknown")
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if got != "" {
				t.Errorf("expected an empty result, got %q", got)
			}
		})
	}
}

func TestGetFontRejectsInvalidURL(t *testing.T) {
	if _, err := GetFont("https://example.com/css?family=Roboto"); err == nil {
		t.Fatal("expected an error, got nil")
	}
}
