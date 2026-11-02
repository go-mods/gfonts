# gfonts

[![GoDoc](https://godoc.org/github.com/go-mods/gfonts?status.svg)](https://godoc.org/github.com/go-mods/gfonts)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-mods/gfonts)](https://goreportcard.com/report/github.com/go-mods/gfonts)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/go-mods/gfonts/blob/master/LICENSE)

Get a font family from the Google Fonts API, encoded as a base64 string.

## Install

```shell
go get github.com/go-mods/gfonts
```

## Usage

```go
css, err := gfonts.GetFont("https://fonts.googleapis.com/css?family=Roboto:400,700")
if err != nil {
    log.Fatal(err)
}

// Embed the stylesheet in a page
fmt.Printf("<link rel=\"stylesheet\" href=\"data:text/css;base64,%s\">\n", css)
```

Two functions are available, both returning a stylesheet encoded as a base64 string:

| Function          | Result                                  | Displaying the fonts            |
|-------------------|-----------------------------------------|---------------------------------|
| `GetFont`         | `@font-face` rules only                 | Downloads the files from Google |
| `GetEmbeddedFont` | `@font-face` rules and their font files | Requires no request to Google   |

`GetEmbeddedFont` downloads every referenced font file, so its result is much larger.

`GetFontWith` and `GetEmbeddedFontWith` are their variants taking options instead of a URL:

```go
css, err := gfonts.GetFontWith(
    gfonts.WithFamily("Open Sans"),
    gfonts.WithWeight(400, 700),
    gfonts.WithItalic(400),
    gfonts.WithSubset("latin"),
    gfonts.WithDisplay("swap"),
)
```

`WithFamily` is required, and a call covers a single family.

An error is returned if the URL is not `https://fonts.googleapis.com/css` with a `family` parameter,
if a request fails or if a response status is not `200 OK`, for example for an unknown family.
`GetEmbeddedFont` and `GetEmbeddedFontWith` also fail if the stylesheet references no font file,
or a file outside `https://fonts.gstatic.com/`.

