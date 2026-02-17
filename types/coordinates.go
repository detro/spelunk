package types

import (
	"encoding"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrSecretCoordFailedParsing          = fmt.Errorf("failed to parse coordinates")
	ErrSecretCoordHaveNoType             = fmt.Errorf("coordinates have no type (URI scheme)")
	ErrSecretCoordHaveNoLocation         = fmt.Errorf("coordinates point to no location (URI authority+path)")
	ErrSecretCoordFailedParsingModifiers = fmt.Errorf("failed to parse modifiers")
)

// SecretCoord are the coordinates to a secret.
// Coordinates have a Type (to determine the source of the secret),
// a Location (to determine how to get to it) and optional Modifiers.
//
// This type implements encoding.TextUnmarshaler, so it
// can be decoded by any idiomatic Go codebase from plain text
// and even as part of json.Unmarshal calls.
type SecretCoord struct {
	Type      string
	Location  string
	Modifiers [][2]string
}

// NewSecretCoord creates a new SecretCoord from a URI-based coordinates string.
//
// Splunker will then dig-up the secret using the correct SecretSource, identified using the SecretCoord.Type (scheme).
// The specific SecretSource will then use the SecretCoord.Location (authority + path)
// and the SecretCoord.Modifiers (query) to finish the dig-up.
//
// Each SecretSource defines the URI format it supports.
func NewSecretCoord(secretCoordURI string) (*SecretCoord, error) {
	u, err := url.Parse(secretCoordURI)
	if err != nil {
		return nil, fmt.Errorf("%w (%q): %w", ErrSecretCoordFailedParsing, secretCoordURI, err)
	}

	loc := fmt.Sprintf("%s%s", u.Host, u.Path)
	if u.User != nil {
		loc = fmt.Sprintf("%s@%s", u.User.String(), loc)
	}

	coord := &SecretCoord{
		Type:      u.Scheme,
		Location:  loc,
		Modifiers: make([][2]string, 0),
	}
	if len(coord.Type) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrSecretCoordHaveNoType, secretCoordURI)
	}
	if len(coord.Location) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrSecretCoordHaveNoLocation, secretCoordURI)
	}

	// Aggregate and URL-unescape modifiers
	// We parse RawQuery manually to preserve order and allow for modifiers
	// to be applied more than once
	if len(u.RawQuery) > 0 {
		for _, pair := range strings.Split(u.RawQuery, "&") {
			if len(pair) == 0 {
				continue
			}

			var key, value string
			splitPair := strings.SplitN(pair, "=", 2)
			key, err = url.QueryUnescape(splitPair[0])
			if err != nil {
				return nil, fmt.Errorf("%w (pair: %q / key: %q): %w",
					ErrSecretCoordFailedParsingModifiers,
					pair, splitPair[0],
					err,
				)
			}

			if len(splitPair) > 1 {
				value, err = url.QueryUnescape(splitPair[1])
				if err != nil {
					return nil, fmt.Errorf("%w (pair: %q / val: %q): %w",
						ErrSecretCoordFailedParsingModifiers,
						pair, splitPair[1],
						err,
					)
				}
			}

			coord.Modifiers = append(coord.Modifiers, [2]string{key, value})
		}
	}

	return coord, nil
}

var _ encoding.TextUnmarshaler = (*SecretCoord)(nil)

func (sc *SecretCoord) UnmarshalText(text []byte) error {
	coord, err := NewSecretCoord(string(text))
	if err != nil {
		return err
	}
	*sc = *coord
	return nil
}
