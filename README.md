# gfonts

[![Go Reference](https://pkg.go.dev/badge/github.com/go-mods/gfonts.svg)](https://pkg.go.dev/github.com/go-mods/gfonts)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-mods/gfonts)](https://goreportcard.com/report/github.com/go-mods/gfonts)
[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

gfonts is a Go library that gets a font family from the Google Fonts API and returns its
stylesheet encoded as a base64 string. The stylesheet can reference the font files hosted by
Google, or embed them so that displaying the fonts requires no request to Google.
It has no dependency outside the standard library.

<details>
<summary>Table of contents</summary>

- [Installation](#installation)
- [Usage](#usage)
- [Functions](#functions)
- [Options](#options)
- [Errors](#errors)
- [Development](#development)
- [License](#license)

</details>

## Installation

Prerequisite: Go 1.20 or later.

```shell
go get github.com/go-mods/gfonts
```

## Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/go-mods/gfonts"
)

func main() {
    css, err := gfonts.GetEmbeddedFontWith(
        gfonts.WithFamily("Open Sans"),
        gfonts.WithWeight(400, 700),
        gfonts.WithItalic(400),
        gfonts.WithDisplay("swap"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Use the stylesheet in a page
    fmt.Printf("<link rel=\"stylesheet\" href=\"data:text/css;base64,%s\">\n", css)
}
```

A Google Fonts URL can be given instead of options:

```go
css, err := gfonts.GetFont("https://fonts.googleapis.com/css?family=Roboto:400,700")
```

## Functions

Every function returns a stylesheet encoded as a base64 string:

| Function              | Input   | Result                                  | Displaying the fonts            |
|-----------------------|---------|-----------------------------------------|---------------------------------|
| `GetFont`             | URL     | `@font-face` rules only                 | Downloads the files from Google |
| `GetFontWith`         | Options | `@font-face` rules only                 | Downloads the files from Google |
| `GetEmbeddedFont`     | URL     | `@font-face` rules and their font files | Requires no request to Google   |
| `GetEmbeddedFontWith` | Options | `@font-face` rules and their font files | Requires no request to Google   |

`GetEmbeddedFont` and `GetEmbeddedFontWith` download every referenced font file, one at a time,
so their result is much larger and they take longer.

A URL must be `https://fonts.googleapis.com/css` with a `family` parameter. The `css2` endpoint
is not supported.

## Options

| Option                     | Description                                       | Example                     |
|----------------------------|---------------------------------------------------|-----------------------------|
| `WithFamily(family)`       | Font family, required                             | `WithFamily("Open Sans")`   |
| `WithWeight(weights...)`   | Regular styles, by weight                         | `WithWeight(400, 700)`      |
| `WithItalic(weights...)`   | Italic styles, by weight                          | `WithItalic(400)`           |
| `WithSubset(subsets...)`   | Character subsets                                 | `WithSubset("cyrillic")`    |
| `WithDisplay(display)`     | `font-display` value of the stylesheet            | `WithDisplay("swap")`       |

A call covers a single family: if `WithFamily` is given several times, the last one applies.
Without `WithWeight` or `WithItalic`, Google Fonts returns the normal weight only.

## Errors

An error is returned if:

- the URL is invalid, or `WithFamily` is missing;
- a request fails;
- a response status is not `200 OK`, for example for an unknown family or weight.

`GetEmbeddedFont` and `GetEmbeddedFontWith` also fail if the stylesheet references no font file,
if a file is not on `https://fonts.gstatic.com/`, or if it is not a font.

Requests time out after 30 seconds.

## Development

Development requires [Task](https://taskfile.dev/installation/). Install the pinned tools once:

```shell
task go-tools:install
```

Then:

```shell
task go:default   # clean, generate, format, vet, tidy, lint and test
task go:test      # tests
task go:lint      # golangci-lint, gosec included
task go:fmt       # format the code and check EditorConfig
```

`TestGetFont` and `TestGetEmbeddedFont` call the Google Fonts API, so the tests require network access.

## License

MIT, see [LICENSE](LICENSE).

