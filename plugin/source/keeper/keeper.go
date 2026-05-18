package keeper

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
	ksm "github.com/keeper-security/secrets-manager-go/core"
)

var locationRegex = regexp.MustCompile(
	`^(?P<recordUID>[A-Za-z0-9-_]{22})(?:/(?P<field>.*))?$`,
)

// SecretSourceKeeper digs up secrets from Keeper Secrets Manager.
// The URI scheme for this source is "kp".
//
//	kp://RECORD_UID/FIELD
//	kp://RECORD_UID/
//	kp://RECORD_UID
//
// When `/FIELD` is appended, Spelunk extracts the specific field from the Keeper Record.
// Supported fields: "title", "password", "notes", or custom field types/labels.
// Otherwise, if it ends with `/`, it returns the whole Record representation as JSON.
//
// This types.SecretSource is a plug-in to spelunker.Spelunker and must be enabled explicitly.
type SecretSourceKeeper struct {
	client *ksm.SecretsManager
}

// WithKeeper enables the SecretSourceKeeper.
func WithKeeper(client *ksm.SecretsManager) spelunk.SpelunkerOption {
	return spelunk.WithSource(&SecretSourceKeeper{
		client: client,
	})
}

var _ types.SecretSource = (*SecretSourceKeeper)(nil)

func (s *SecretSourceKeeper) Type() string {
	return "kp"
}

func (s *SecretSourceKeeper) DigUp(
	_ context.Context,
	coord types.SecretCoord,
) (string, error) {
	matches := locationRegex.FindStringSubmatch(coord.Location)
	if matches == nil {
		return "", fmt.Errorf(
			"%w: expected a valid 22-character base64url RECORD UID optionally followed by /FIELD, got %q",
			types.ErrInvalidLocation,
			coord.Location,
		)
	}

	recordUID := matches[locationRegex.SubexpIndex("recordUID")]
	field := matches[locationRegex.SubexpIndex("field")]

	records, err := s.client.GetSecrets([]string{recordUID})
	if err != nil {
		return "", fmt.Errorf("%w: %w", types.ErrCouldNotFetchSecret, err)
	}

	if len(records) == 0 {
		return "", fmt.Errorf("%w (%q)", types.ErrSecretNotFound, coord.Location)
	}

	foundRecord := records[0]

	// No field requested: return the whole record as JSON
	if field == "" {
		if foundRecord.RawJson != "" {
			return foundRecord.RawJson, nil
		}
		return foundRecord.ToString(), nil
	}

	// Try standard fields first
	switch strings.ToLower(field) {
	case "title":
		return foundRecord.Title(), nil
	case "password":
		return foundRecord.Password(), nil
	case "notes":
		return foundRecord.Notes(), nil
	}

	// Try fetching by label
	if val := foundRecord.GetFieldValueByLabel(field); val != "" {
		return val, nil
	}
	if val := foundRecord.GetCustomFieldValueByLabel(field); val != "" {
		return val, nil
	}

	return "", fmt.Errorf(
		"%w: unknown or empty field %q in %q",
		types.ErrSecretKeyNotFound,
		field,
		coord.Location,
	)
}
