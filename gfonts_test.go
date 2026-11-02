package gfonts

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	stylesheetURL = "https://fonts.googleapis.com/css?family=Roboto"
	fontFile      = "https://fonts.gstatic.com/s/roboto/regular.ttf"

	// ttfContent starts with the signature of a TrueType font
	ttfContent = "\x00\x01\x00\x00font"
)

// roundTripFunc answers the requests of httpClient without using the network.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// stubResponses makes httpClient answer each URL with the given body, any other URL
// with a 404, and returns the number of requests received per URL.
func stubResponses(t *testing.T, bodies map[string]string) map[string]int {
	t.Helper()

	original := httpClient
	t.Cleanup(func() { httpClient = original })

	requests := map[string]int{}
	httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		url := r.URL.String()
		requests[url]++

		rec := httptest.NewRecorder()
		body, found := bodies[url]
		if !found {
			rec.WriteHeader(http.StatusNotFound)
			return rec.Result(), nil
		}
		_, _ = rec.WriteString(body)
		return rec.Result(), nil
	})}

	return requests
}

// stubStatus makes httpClient answer every request with the given status.
func stubStatus(t *testing.T, status int) {
	t.Helper()

	original := httpClient
	t.Cleanup(func() { httpClient = original })

	httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(status)
		_, _ = rec.WriteString("<!DOCTYPE html>")
		return rec.Result(), nil
	})}
}

// decode returns the content of a base64 string.
func decode(t *testing.T, encoded string) string {
	t.Helper()

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("invalid base64: %v", err)
	}
	return string(decoded)
}

// getters lists the exported functions sharing the same stylesheet handling.
var getters = []struct {
	name string
	get  func(string) (string, error)
}{
	{name: "GetFont", get: GetFont},
	{name: "GetEmbeddedFont", get: GetEmbeddedFont},
}

func TestGetFont(t *testing.T) {
	url := "https://fonts.googleapis.com/css?family=Roboto:300,400,500,700"
	font, err := GetFont(url)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if font == "" {
		t.Errorf("Font is empty")
	}
}

func TestGetEmbeddedFont(t *testing.T) {
	encoded, err := GetEmbeddedFont("https://fonts.googleapis.com/css?family=Roboto:400")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	css := decode(t, encoded)
	if !strings.Contains(css, "url(data:font/") {
		t.Errorf("no embedded font file in the result")
	}
	if strings.Contains(css, "url(https:") {
		t.Errorf("the result still references a remote file")
	}
}

func TestGetFontEncodesStylesheet(t *testing.T) {
	const css = "@font-face { font-family: 'Roboto'; src: url(" + fontFile + "); }"
	requests := stubResponses(t, map[string]string{stylesheetURL: css})

	got, err := GetFont(stylesheetURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decoded := decode(t, got); decoded != css {
		t.Errorf("got %q, want %q", decoded, css)
	}
	if requests[fontFile] != 0 {
		t.Errorf("the font file must not be downloaded")
	}
}

func TestGetEmbeddedFontEmbedsFontFiles(t *testing.T) {
	const otherFile = "https://fonts.gstatic.com/s/roboto/bold.ttf"
	css := "@font-face { font-weight: 400; src: url(" + fontFile + ") format('truetype'); }\n" +
		"@font-face { font-weight: 500; src: url(" + fontFile + ") format('truetype'); }\n" +
		"@font-face { font-weight: 700; src: url(" + otherFile + ") format('truetype'); }\n"
	requests := stubResponses(t, map[string]string{
		stylesheetURL: css,
		fontFile:      ttfContent,
		otherFile:     ttfContent + "-bold",
	})

	got, err := GetEmbeddedFont(stylesheetURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	regular := "url(data:font/ttf;base64," + base64.StdEncoding.EncodeToString([]byte(ttfContent)) + ")"
	bold := "url(data:font/ttf;base64," + base64.StdEncoding.EncodeToString([]byte(ttfContent+"-bold")) + ")"
	want := "@font-face { font-weight: 400; src: " + regular + " format('truetype'); }\n" +
		"@font-face { font-weight: 500; src: " + regular + " format('truetype'); }\n" +
		"@font-face { font-weight: 700; src: " + bold + " format('truetype'); }\n"
	if decoded := decode(t, got); decoded != want {
		t.Errorf("got %q, want %q", decoded, want)
	}
	if requests[fontFile] != 1 {
		t.Errorf("shared font file downloaded %d times, want once", requests[fontFile])
	}
}

func TestGetEmbeddedFontRejectsInvalidFontFiles(t *testing.T) {
	const otherHost = "https://example.com/roboto.ttf"
	tests := []struct {
		name   string
		bodies map[string]string
	}{
		{
			name:   "no font file",
			bodies: map[string]string{stylesheetURL: "@font-face { font-family: 'Roboto'; }"},
		},
		{
			name: "unexpected host",
			bodies: map[string]string{
				stylesheetURL: "@font-face { src: url(" + otherHost + "); }",
				otherHost:     ttfContent,
			},
		},
		{
			name:   "unavailable file",
			bodies: map[string]string{stylesheetURL: "@font-face { src: url(" + fontFile + "); }"},
		},
		{
			name: "not a font",
			bodies: map[string]string{
				stylesheetURL: "@font-face { src: url(" + fontFile + "); }",
				fontFile:      "<!DOCTYPE html>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := stubResponses(t, tt.bodies)

			got, err := GetEmbeddedFont(stylesheetURL)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if got != "" {
				t.Errorf("expected an empty result, got %q", got)
			}
			if requests[otherHost] != 0 {
				t.Errorf("a file was downloaded from an unexpected host")
			}
		})
	}
}

func TestGettersRejectErrorStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{name: "unknown family", status: http.StatusBadRequest},
		{name: "server failure", status: http.StatusInternalServerError},
	}

	for _, getter := range getters {
		for _, tt := range tests {
			t.Run(getter.name+"/"+tt.name, func(t *testing.T) {
				stubStatus(t, tt.status)

				got, err := getter.get(stylesheetURL)
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				if got != "" {
					t.Errorf("expected an empty result, got %q", got)
				}
			})
		}
	}
}

func TestGettersRejectInvalidURL(t *testing.T) {
	for _, getter := range getters {
		t.Run(getter.name, func(t *testing.T) {
			if _, err := getter.get("https://example.com/css?family=Roboto"); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}
