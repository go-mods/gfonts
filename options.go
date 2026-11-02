package gfonts

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// options holds the description of the requested stylesheet.
type options struct {
	family  string
	weights []int
	italics []int
	subsets []string
	display string
}

// Option describes the stylesheet requested from GetFontWith or GetEmbeddedFontWith.
type Option func(*options)

// WithFamily sets the font family, for example "Open Sans". It is required. A call
// covers a single family: if the option is given several times, the last one applies.
func WithFamily(family string) Option {
	return func(o *options) { o.family = family }
}

// WithWeight adds regular styles, by weight: 400 for normal, 700 for bold. Without
// WithWeight or WithItalic, Google Fonts returns the normal weight only.
func WithWeight(weights ...int) Option {
	return func(o *options) { o.weights = append(o.weights, weights...) }
}

// WithItalic adds italic styles, by weight.
func WithItalic(weights ...int) Option {
	return func(o *options) { o.italics = append(o.italics, weights...) }
}

// WithSubset adds character subsets, for example "latin" or "cyrillic".
func WithSubset(subsets ...string) Option {
	return func(o *options) { o.subsets = append(o.subsets, subsets...) }
}

// WithDisplay sets the font-display value of the stylesheet, for example "swap".
func WithDisplay(display string) Option {
	return func(o *options) { o.display = display }
}

// GetFontWith is GetFont with a stylesheet described by options rather than by its URL.
// An error is returned if WithFamily is missing, or for the same reasons as GetFont.
func GetFontWith(opts ...Option) (string, error) {
	stylesheet, err := buildURL(opts)
	if err != nil {
		return "", err
	}

	return GetFont(stylesheet)
}

// GetEmbeddedFontWith is GetEmbeddedFont with a stylesheet described by options rather
// than by its URL. An error is returned if WithFamily is missing, or for the same
// reasons as GetEmbeddedFont.
func GetEmbeddedFontWith(opts ...Option) (string, error) {
	stylesheet, err := buildURL(opts)
	if err != nil {
		return "", err
	}

	return GetEmbeddedFont(stylesheet)
}

// buildURL returns the Google Fonts URL of the stylesheet described by the options
func buildURL(opts []Option) (string, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	if o.family == "" {
		return "", errors.New("font family is required, use WithFamily")
	}

	query := url.Values{"family": {o.familyParameter()}}
	if len(o.subsets) > 0 {
		query.Set("subset", strings.Join(o.subsets, ","))
	}
	if o.display != "" {
		query.Set("display", o.display)
	}

	stylesheet := url.URL{
		Scheme:   "https",
		Host:     stylesheetHost,
		Path:     stylesheetPath,
		RawQuery: query.Encode(),
	}
	return stylesheet.String(), nil
}

// familyParameter returns the family and its styles, for example "Roboto:400,700,400i"
func (o options) familyParameter() string {
	styles := make([]string, 0, len(o.weights)+len(o.italics))
	for _, weight := range o.weights {
		styles = append(styles, strconv.Itoa(weight))
	}
	for _, weight := range o.italics {
		styles = append(styles, strconv.Itoa(weight)+"i")
	}

	if len(styles) == 0 {
		return o.family
	}
	return o.family + ":" + strings.Join(styles, ",")
}
