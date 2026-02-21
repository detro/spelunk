package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
	"github.com/hashicorp/vault/api"
)

var ErrSecretSourceVaultInvalidLocation = fmt.Errorf(
	"invalid Vault secret location format",
)

// SecretSourceVault digs up secrets from HashiCorp Vault KV Secrets Engine.
// It supports both KV engine versions 1 and 2 (https://developer.hashicorp.com/vault/docs/secrets/kv),
// and transparently handles the differences in the response format between the 2 engines.
//
// The URI scheme for this source is "vault".
//
//	vault://<ENGINE_MOUNT>/<PATH/TO/SECRET>/KEY
//	vault://<ENGINE_MOUNT>/<PATH/TO/SECRET>/
//
// When `/KEY` is appended, Spelunk extracts the specific value in the secret's data key-value map.
// Otherwise, if it ends with `/`, it returns the whole secret's data key-value map as JSON.
//
// Note that secrets stored in a KV version 2 requires `/data/` to be places between
// `<ENGINE_MOUNT>` and `<PATH/TO/SECRET>`:
//
//	vault://<ENGINE_MOUNT>/data/<PATH/TO/SECRET>/KEY
//	vault://<ENGINE_MOUNT>/data/<PATH/TO/SECRET>/
//
// This types.SecretSource is a plug-in to spelunker.Spelunker and must be enabled explicitly.
type SecretSourceVault struct {
	vaultClient *api.Client
}

// WithVault enables the SecretSourceVault.
func WithVault(vaultClient *api.Client) spelunk.SpelunkerOption {
	source := &SecretSourceVault{
		vaultClient,
	}
	return spelunk.WithSource(source)
}

var _ types.SecretSource = (*SecretSourceVault)(nil)

func (s *SecretSourceVault) Type() string {
	return "vault"
}

func (s *SecretSourceVault) DigUp(
	ctx context.Context,
	coord types.SecretCoord,
) (string, error) {
	parts := strings.Split(coord.Location, "/")

	if len(parts) < 3 {
		return "", fmt.Errorf(
			"%w: expected <MOUNT>/<PATH/TO/SECRET>/<KEY> or <MOUNT>/<PATH/TO/SECRET>/, got %q",
			ErrSecretSourceVaultInvalidLocation,
			coord.Location,
		)
	}

	path := strings.Join(parts[:len(parts)-1], "/")
	key := parts[len(parts)-1]

	// Retrieve
	secret, err := s.vaultClient.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", types.ErrCouldNotFetchSecret, coord.Location, err)
	}
	if secret == nil {
		return "", fmt.Errorf("%w (%q)", types.ErrSecretNotFound, coord.Location)
	}
	if secret.Data == nil {
		return "", fmt.Errorf(
			"%w (%q): secret contains no data",
			types.ErrSecretNotFound,
			coord.Location,
		)
	}

	// Vault KV v2 wraps data in a "data" field
	var data map[string]interface{}
	if v2Data, ok := secret.Data["data"].(map[string]interface{}); ok {
		data = v2Data
	} else {
		// KV v1 or other logical paths
		data = secret.Data
	}

	// No key requested: return the whole `data` map
	if len(key) == 0 {
		dataJsonBytes, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(dataJsonBytes), nil
	}

	// Return specific key
	if val, found := data[key]; found {
		return fmt.Sprintf("%v", val), nil
	}

	return "", fmt.Errorf("%w (%q)", types.ErrSecretKeyNotFound, coord.Location)
}
