package spelunk

import (
	"encoding"
	"fmt"
	"net/url"
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
	Modifiers map[string]string
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

	coord := &SecretCoord{
		Type:      u.Scheme,
		Location:  fmt.Sprintf("%s%s", u.Host, u.Path),
		Modifiers: make(map[string]string),
	}
	if len(coord.Type) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrSecretCoordHaveNoType, secretCoordURI)
	}
	if len(coord.Location) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrSecretCoordHaveNoLocation, secretCoordURI)
	}

	// Aggregate and URL-unescape modifiers
	for qk, qv := range u.Query() {
		// Ignore empty-value modifiers
		if len(qv) == 0 {
			continue
		}

		coord.Modifiers[qk], err = url.QueryUnescape(qv[0])
		if err != nil {
			return nil, fmt.Errorf("%w (%q): %w", ErrSecretCoordFailedParsingModifiers, qv[0], err)
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
