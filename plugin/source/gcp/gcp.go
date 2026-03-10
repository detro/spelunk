package gcp

import (
	"context"
	"fmt"
	"regexp"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrSecretSourceGCPInvalidLocation = fmt.Errorf(
		"invalid GCP Secret Manager secret location format",
	)

	// fullSecretVersionNameRegexp matches secret version names that include a specific version indicator.
	fullSecretVersionNameRegexp = regexp.MustCompile(
		`^projects/(?:[a-z][-a-z0-9]{4,28}[a-z0-9]|\d{5,20})/secrets/([a-zA-Z0-9_-]{1,255})/versions/(?:\d+|latest)$`,
	)

	// latestSecretVersionShortNameRegexp matches secret version names that refer to the latest version of a secret.
	latestSecretVersionShortNameRegexp = regexp.MustCompile(
		`^projects/(?:[a-z][-a-z0-9]{4,28}[a-z0-9]|\d{5,20})/secrets/([a-zA-Z0-9_-]{1,255})$`,
	)
)

// SecretSourceGCP digs up secrets from Google Cloud Secret Manager.
//
// The URI scheme for this source is "gcp".
//
//	gcp://projects/<PROJECT_ID_OR_NUM>/secrets/<SECRET_NAME>
//	gcp://projects/<PROJECT_ID_OR_NUM>/secrets/<SECRET_NAME>/versions/<VERSION>
//
// If the version is omitted, a "/versions/latest" suffix is appended.
//
// Expected format of `<PROJECT_ID_OR_NUM>` is documented at: https://google.aip.dev/cloud/2510.
// Expected format of `<SECRET_NAME>` is documented at: https://cloud.google.com/security/products/secret-manager.
//
// This types.SecretSource is a plug-in to spelunker.Spelunker and must be enabled explicitly.
type SecretSourceGCP struct {
	client *secretmanager.Client
}

// WithGCP enables the SecretSourceGCP.
func WithGCP(client *secretmanager.Client) spelunk.SpelunkerOption {
	return spelunk.WithSource(&SecretSourceGCP{
		client: client,
	})
}

var _ types.SecretSource = (*SecretSourceGCP)(nil)

func (s *SecretSourceGCP) Type() string {
	return "gcp"
}

func (s *SecretSourceGCP) DigUp(ctx context.Context, coord types.SecretCoord) (string, error) {
	// Strip trailing slash if present (often happens when the URI contains query parameters e.g. `/?jp=$.password`)
	location := coord.Location
	if len(location) > 0 && location[len(location)-1] == '/' {
		location = location[:len(location)-1]
	}

	// Enforce one of 2 possible regexp to validate the location format
	var secretVersionName string
	switch {
	case fullSecretVersionNameRegexp.MatchString(location):
		secretVersionName = location
	case latestSecretVersionShortNameRegexp.MatchString(location):
		secretVersionName = fmt.Sprintf("%s/versions/latest", location)
	default:
		return "", fmt.Errorf(
			"%w: expected 'projects/<PROJECT_ID_OR_NUM>/secrets/<SECRET_NAME>[/versions/<VERSION>]', got %q",
			ErrSecretSourceGCPInvalidLocation,
			coord.Location,
		)
	}

	// Retrieve secret
	res, err := s.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretVersionName,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return "", fmt.Errorf("%w (%q): %w", types.ErrSecretNotFound, coord.Location, err)
		}
		return "", fmt.Errorf("%w (%q): %w", types.ErrCouldNotFetchSecret, coord.Location, err)
	}

	// Extract and return payload, or error fi missing
	if res.Payload == nil || res.Payload.Data == nil {
		return "", fmt.Errorf(
			"%w (%q): secret contains no data",
			types.ErrSecretNotFound,
			coord.Location,
		)
	}
	return string(res.Payload.Data), nil
}
