package gfonts

import "testing"

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
