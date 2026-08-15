package cli

import (
	"context"
	"testing"

	"github.com/detro/spelunk/cmd/spelunk/internal/configurator"
	"github.com/stretchr/testify/require"
)

func TestPluginConfigs_Types(t *testing.T) {
	var cfg Configurators
	all := cfg.All()
	require.Len(t, all, 8)

	expectedTypes := []string{
		"aws", "az", "gcp", "vault", "k8s", "op", "bw", "kp",
	}

	for i, p := range all {
		require.Equal(t, expectedTypes[i], p.Type())
	}
}

func TestPluginConfigs_NoCredentials_ReturnsNil(t *testing.T) {
	ctx := context.Background()
	var cfg Configurators

	opts, err := cfg.SpelunkerOptions(ctx)
	require.NoError(t, err)
	require.Empty(t, opts)
}

func TestPluginConfigs_NoCredentials_CredentialsValid_ReturnsError(t *testing.T) {
	ctx := context.Background()
	var cfg Configurators

	for _, p := range cfg.All() {
		require.False(t, p.CredentialsDetected())
		err := p.CredentialsValid(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, configurator.ErrCredentialsNotDetected)
	}
}

func TestConfig_VerifyAll_NoCredentials_NoError(t *testing.T) {
	ctx := context.Background()
	var cfg Configurators

	err := cfg.VerifyAll(ctx)
	require.NoError(t, err)
}
