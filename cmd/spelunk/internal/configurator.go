package internal

import (
	"context"

	"github.com/detro/spelunk/v2"
)

// SecretSourceConfigurator is an object capable of handling the configuration of a specific types.SecretSource.
type SecretSourceConfigurator interface {
	// Type returns the identifier of the types.SecretSource plugin.
	Type() string

	// CredentialsDetected returns true if the necessary credentials for the given plugin are detected.
	CredentialsDetected() bool

	// SpelunkerOption returns a configuration option for Spelunker.
	// If credentials/flags for the plugin are present, it initializes the SDK client
	// and returns the actual With<Source>(client) option.
	//
	// If requirements are NOT met (no credentials detected), it returns a nil option.
	//
	// In all other cases, a descriptive error of what went wrong.
	SpelunkerOption(ctx context.Context) (spelunk.SpelunkerOption, error)

	// CredentialsValid validates credentials, via a lightweight non-mutating operation against the underlying secret provider.
	// Returns an error if something goes wrong, including if the credentials were not even detected in the first
	// place (i.e. CredentialsDetected already returned false).
	CredentialsValid(ctx context.Context) error
}
