package aws

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
)

var (
	ErrSecretSourceAWSInvalidLocation = fmt.Errorf(
		"invalid AWS Secrets Manager secret location format",
	)

	ErrSecretSourceAWSInvalidNameSuffix = fmt.Errorf(
		"secret name must not end with a hyphen followed by six characters",
	)

	// secretNameRegexp matches secrets referred to via name.
	secretNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9/_+=.@-]{0,511}$`)

	// secretNameDisallowedSuffixRegexp matches the disallowed suffix for secret names.
	secretNameDisallowedSuffixRegexp = regexp.MustCompile(`-[a-zA-Z0-9]{6}$`)

	// secretARNRegexp matches secrets referred to via ARN.
	secretARNRegexp = regexp.MustCompile(
		`^arn:aws(?:-[a-z]+)*:secretsmanager:[a-z0-9-]+:\d{12}:secret:[a-zA-Z0-9][a-zA-Z0-9/_+=.@-]{0,504}-[a-zA-Z0-9]{6}$`,
	)
)

// SecretSourceAWS digs up secrets from AWS Secrets Manager.
//
// The URI scheme for this source is "aws".
//
//	aws://<SECRET_NAME>
//	aws:///<SECRET_ARN>
//
// AWS Secrets Manager supports storing secrets either as a String, or as Binary (i.e. array of bytes).
// In the case of the latter, the API returns the Base64-encoded version of the bytes. Spelunk respects
// that and leaves it to you to either consume the secret in base64 form or decode it using the `?b64d` modifier.
//
// NOTE: When referring to a secret by ARN, it is important to use the prefix `aws:///` to ensure
// we don't confuse the internal Spelunk parser, given the "peculiar" format of AWS ARNs containing the `:` character.
//
// See https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_CreateSecret.html for supported name format.
//
// This types.SecretSource is a plug-in to spelunker.Spelunker and must be enabled explicitly.
type SecretSourceAWS struct {
	client *secretsmanager.Client
}

// WithAWS enables the SecretSourceAWS.
func WithAWS(client *secretsmanager.Client) spelunk.SpelunkerOption {
	source := &SecretSourceAWS{
		client: client,
	}
	return spelunk.WithSource(source)
}

var _ types.SecretSource = (*SecretSourceAWS)(nil)

func (s *SecretSourceAWS) Type() string {
	return "aws"
}

func (s *SecretSourceAWS) DigUp(ctx context.Context, coord types.SecretCoord) (string, error) {
	// Strip trailing slash if present (often happens when the URI contains query parameters e.g. /?jp=$.password)
	location := coord.Location
	if len(location) > 0 && location[len(location)-1] == '/' {
		location = location[:len(location)-1]
	}

	// Trim leading slash that might be present if the user used aws:///<ARN>,
	// and so the location was considered a path by the underlying URL parser.
	secretID := strings.TrimPrefix(location, "/")

	// Enforce 2 possible regexp
	switch {
	case secretARNRegexp.MatchString(secretID):
		// Valid ARN, nothing more to check
	case secretNameRegexp.MatchString(secretID):
		if secretNameDisallowedSuffixRegexp.MatchString(secretID) {
			return "", fmt.Errorf("%w: %q", ErrSecretSourceAWSInvalidNameSuffix, coord.Location)
		}
	default:
		return "", fmt.Errorf(
			"%w: expected <SECRET_NAME> or <SECRET_ARN>, got %q",
			ErrSecretSourceAWSInvalidLocation,
			coord.Location,
		)
	}

	// Retrieve secret
	res, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		// Differentiate between not found and other errors if possible
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			return "", fmt.Errorf("%w (%q): %w", types.ErrSecretNotFound, coord.Location, err)
		}
		return "", fmt.Errorf("%w (%q): %w", types.ErrCouldNotFetchSecret, coord.Location, err)
	}

	// Extract and return secret, or error if missing
	if res.SecretString != nil {
		// Secret is a string
		return *res.SecretString, nil
	}
	if res.SecretBinary != nil {
		// Secret is a binary, and AWS Secret Manager returns it as Base64-encoded array of bytes.
		// We return it as a string but leave it encoded: the user will decide how to handle it.
		return string(res.SecretBinary), nil
	}
	return "", fmt.Errorf(
		"%w (%q): secret contains no data",
		types.ErrSecretNotFound,
		coord.Location,
	)
}
