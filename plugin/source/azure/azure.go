package azure

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
)

var (
	ErrSecretSourceAzureInvalidLocation = fmt.Errorf(
		"invalid Azure Key Vault secret location format",
	)

	// secretVersionNameRegexp matches secret locations that include a specific version.
	// Format: `<SECRET_NAME>/<VERSION>`.
	//
	// Each part follows these rules:
	// - `SECRET_NAME`: Length 1-127, Alphanumerics and hyphens.
	// - `VERSION`: 32 character hex string.
	secretVersionNameRegexp = regexp.MustCompile(
		`^([a-zA-Z0-9-]{1,127})/([a-fA-F0-9]{32})$`,
	)

	// latestSecretVersionShortNameRegexp matches secret locations that refer to the latest version of a secret.
	// Format: `<SECRET_NAME>`.
	latestSecretVersionShortNameRegexp = regexp.MustCompile(
		`^([a-zA-Z0-9-]{1,127})$`,
	)
)

// SecretSourceAzure digs up secrets from Azure Key Vault.
//
// The URI scheme for this source is "az".
//
//	az://<SECRET_NAME>
//	az://<SECRET_NAME>/<VERSION>
//
// NOTE: Since the Azure Key Vault client (`azsecrets.Client`) is explicitly bound to a specific
// vault URL when instantiated, the dug-up secret is assumed to be present in the vault the client is bound to.
// Spelunk expects just the secret name and optional version.
//
// Expected format of `<SECRET_NAME>` is documented at: https://learn.microsoft.com/en-us/azure/key-vault/general/about-keys-secrets-certificates#vault-name-and-object-name
//
// This types.SecretSource is a plug-in to spelunker.Spelunker and must be enabled explicitly.
type SecretSourceAzure struct {
	client *azsecrets.Client
}

// WithAzure enables the SecretSourceAzure.
func WithAzure(client *azsecrets.Client) spelunk.SpelunkerOption {
	return spelunk.WithSource(&SecretSourceAzure{
		client: client,
	})
}

var _ types.SecretSource = (*SecretSourceAzure)(nil)

func (s *SecretSourceAzure) Type() string {
	return "az"
}

func (s *SecretSourceAzure) DigUp(ctx context.Context, coord types.SecretCoord) (string, error) {
	// Strip trailing slash if present (often happens when the URI contains query parameters e.g. /?jp=$.password)
	location := coord.Location
	if len(location) > 0 && location[len(location)-1] == '/' {
		location = location[:len(location)-1]
	}

	// Trim leading slash that might be present if the user used az:///<SECRET_NAME>
	location = strings.TrimPrefix(location, "/")

	// Enforce one of 2 possible regexp to validate the location format and extract parts
	var secretName, version string

	switch {
	case secretVersionNameRegexp.MatchString(location):
		matches := secretVersionNameRegexp.FindStringSubmatch(location)
		secretName = matches[1]
		version = matches[2]
	case latestSecretVersionShortNameRegexp.MatchString(location):
		matches := latestSecretVersionShortNameRegexp.FindStringSubmatch(location)
		secretName = matches[1]
		version = "" // API gets latest when version is empty
	default:
		return "", fmt.Errorf(
			"%w: expected <SECRET_NAME>[/<VERSION>], got %q",
			ErrSecretSourceAzureInvalidLocation,
			coord.Location,
		)
	}

	res, err := s.client.GetSecret(ctx, secretName, version, nil)
	if err != nil {
		// Use azcore.ResponseError to accurately detect 404
		if respErr, errMatched := errors.AsType[*azcore.ResponseError](
			err,
		); errMatched &&
			respErr.StatusCode == 404 {
			return "", fmt.Errorf("%w (%q): %w", types.ErrSecretNotFound, coord.Location, err)
		}

		// Fallback to string matching for other cases
		if strings.Contains(err.Error(), "SecretNotFound") ||
			strings.Contains(err.Error(), "NotFoundException") {
			return "", fmt.Errorf("%w (%q): %w", types.ErrSecretNotFound, coord.Location, err)
		}

		return "", fmt.Errorf("%w (%q): %w", types.ErrCouldNotFetchSecret, coord.Location, err)
	}

	if res.Value == nil {
		return "", fmt.Errorf(
			"%w (%q): secret contains no data",
			types.ErrSecretNotFound,
			coord.Location,
		)
	}

	return *res.Value, nil
}
