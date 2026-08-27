package cli

import (
	"errors"
	"testing"

	"github.com/detro/spelunk/cmd/spelunk/internal/configurator"
	"github.com/stretchr/testify/require"
)

func TestPluginConfigs_Types(t *testing.T) {
	var cfg Configurators
	all := cfg.All()
	require.Len(t, all, 7)

	expectedTypes := []string{
		"aws", "az", "gcp", "vault", "k8s", "op", "kp",
	}

	for i, p := range all {
		require.Equal(t, expectedTypes[i], p.Type())
	}
}

func TestPluginConfigs_SpelunkerOptions_HostAware(t *testing.T) {
	ctx := t.Context()
	var cfg Configurators

	var detectedCount int
	for _, p := range cfg.All() {
		if p.CredentialsDetected() {
			detectedCount++
		}
	}

	opts, err := cfg.SpelunkerOptions(ctx)
	require.NoError(t, err)
	require.Len(t, opts, detectedCount)

	for _, opt := range opts {
		require.NotNil(t, opt)
	}
}

func TestPluginConfigs_CredentialsValid_HostAware(t *testing.T) {
	ctx := t.Context()
	var cfg Configurators

	for _, p := range cfg.All() {
		t.Run(p.Type(), func(t *testing.T) {
			err := p.CredentialsValid(ctx)
			if !p.CredentialsDetected() {
				require.Error(t, err)
				require.ErrorIs(t, err, configurator.ErrCredentialsNotDetected)
			} else {
				// If detected on host machine, validation can succeed or fail at provider level,
				// but must NOT fail with ErrCredentialsNotDetected
				require.False(
					t,
					errors.Is(err, configurator.ErrCredentialsNotDetected),
					"plugin %s detected credentials but returned ErrCredentialsNotDetected",
					p.Type(),
				)
			}
		})
	}
}

func TestConfig_VerifyAll_HostAware(t *testing.T) {
	ctx := t.Context()
	var cfg Configurators

	var expectedErrors []error
	for _, p := range cfg.All() {
		if p.CredentialsDetected() {
			if err := p.CredentialsValid(ctx); err != nil {
				expectedErrors = append(expectedErrors, err)
			}
		}
	}

	err := cfg.VerifyAll(ctx)
	if len(expectedErrors) == 0 {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}
}
