package gfonts

import (
	"strings"
	"testing"
)

func TestBuildURL(t *testing.T) {
	const prefix = "https://fonts.googleapis.com/css?"

	tests := []struct {
		name string
		opts []Option
		want string
	}{
		{
			name: "family only",
			opts: []Option{WithFamily("Roboto")},
			want: prefix + "family=Roboto",
		},
		{
			name: "family with a space",
			opts: []Option{WithFamily("Open Sans")},
			want: prefix + "family=Open+Sans",
		},
		{
			name: "weights",
			opts: []Option{WithFamily("Roboto"), WithWeight(300, 400), WithWeight(700)},
			want: prefix + "family=Roboto%3A300%2C400%2C700",
		},
		{
			name: "weights and italics",
			opts: []Option{WithFamily("Roboto"), WithItalic(400), WithWeight(700)},
			want: prefix + "family=Roboto%3A700%2C400i",
		},
		{
			name: "subsets and display",
			opts: []Option{WithFamily("Roboto"), WithSubset("latin", "cyrillic"), WithDisplay("swap")},
			want: prefix + "display=swap&family=Roboto&subset=latin%2Ccyrillic",
		},
		{
			name: "last family applies",
			opts: []Option{WithFamily("Roboto"), WithFamily("Lato")},
			want: prefix + "family=Lato",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildURL(tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			if !verifyURL(got) {
				t.Errorf("%q is rejected by verifyURL", got)
			}
		})
	}
}

func TestGetFontWith(t *testing.T) {
	const url = "https://fonts.googleapis.com/css?family=Roboto%3A400"
	const css = "@font-face { font-family: 'Roboto'; src: url(" + fontFile + "); }"
	stubResponses(t, map[string]string{url: css})

	got, err := GetFontWith(WithFamily("Roboto"), WithWeight(400))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded := decode(t, got); decoded != css {
		t.Errorf("got %q, want %q", decoded, css)
	}
}

func TestGetEmbeddedFontWith(t *testing.T) {
	const url = "https://fonts.googleapis.com/css?family=Roboto%3A400"
	stubResponses(t, map[string]string{
		url:      "@font-face { src: url(" + fontFile + "); }",
		fontFile: ttfContent,
	})

	got, err := GetEmbeddedFontWith(WithFamily("Roboto"), WithWeight(400))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded := decode(t, got); !strings.Contains(decoded, "url(data:font/ttf;base64,") {
		t.Errorf("no embedded font file in %q", decoded)
	}
}

func TestOptionGettersRequireFamily(t *testing.T) {
	tests := []struct {
		name string
		get  func(...Option) (string, error)
	}{
		{name: "GetFontWith", get: GetFontWith},
		{name: "GetEmbeddedFontWith", get: GetEmbeddedFontWith},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.get(WithWeight(400)); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}
