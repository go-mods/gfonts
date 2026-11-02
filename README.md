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

An error is returned if the URL does not start with `https://fonts.googleapis.com/css?family=`,
if a request fails or if a response status is not `200 OK`, for example for an unknown family.
`GetEmbeddedFont` also fails if the stylesheet references no font file, or a file outside
`https://fonts.gstatic.com/`.

