package types_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
)

func TestNewSecretCoord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
		wantLoc  string
		wantMods [][2]string
		errMatch error
	}{
		{
			name:     "valid simple coordinate",
			input:    "vault://secret/data/myapp/config",
			wantType: "vault",
			wantLoc:  "secret/data/myapp/config",
			wantMods: [][2]string{},
		},
		{
			name:     "valid coordinate with jsonpath modifier",
			input:    "k8s://mynamespace/mysecret/mycredentials?jsonpath=.kafka.password",
			wantType: "k8s",
			wantLoc:  "mynamespace/mysecret/mycredentials",
			wantMods: [][2]string{
				{"jsonpath", ".kafka.password"},
			},
		},
		{
			name:     "valid coordinate with simple location",
			input:    "env://JUVE_MERDA",
			wantType: "env",
			wantLoc:  "JUVE_MERDA",
			wantMods: [][2]string{},
		},
		{
			name:     "valid coordinate with userinfo in the URI",
			input:    "env://JUVE@MERDA",
			wantType: "env",
			wantLoc:  "JUVE@MERDA",
			wantMods: [][2]string{},
		},
		{
			name:     "valid coordinate with userinfo in the URI",
			input:    "env://JUVE:MERDA@TORINO:1897",
			wantType: "env",
			wantLoc:  "JUVE:MERDA@TORINO:1897",
			wantMods: [][2]string{},
		},
		{
			name:     "valid coordinate with encoded characters in path",
			input:    "file:///etc/secrets/my%20secret.json",
			wantType: "file",
			wantLoc:  "/etc/secrets/my secret.json",
			wantMods: [][2]string{},
		},
		{
			name:     "valid coordinate with local file in in path",
			input:    "file://local/file.json",
			wantType: "file",
			wantLoc:  "local/file.json",
			wantMods: [][2]string{},
		},
		{
			name:     "valid coordinate with relative file in in path",
			input:    "file://./local/file.json",
			wantType: "file",
			wantLoc:  "./local/file.json",
			wantMods: [][2]string{},
		},
		{
			name:     "valid coordinate with encoded characters in modifiers",
			input:    "dummy://loc?jsonpath=%24.phoneNumbers%5B0%5D.type",
			wantType: "dummy",
			wantLoc:  "loc",
			wantMods: [][2]string{
				{"jsonpath", "$.phoneNumbers[0].type"},
			},
		},
		{
			name:     "valid coordinate with multiple ordered modifiers",
			input:    "k8s://ns/name/key?first=1&second=2&first=3",
			wantType: "k8s",
			wantLoc:  "ns/name/key",
			wantMods: [][2]string{
				{"first", "1"},
				{"second", "2"},
				{"first", "3"},
			},
		},
		{
			name:     "valid coordinate with modifier without value",
			input:    "k8s://ns/name/key?m1&m2=v&m3=&m4=v",
			wantType: "k8s",
			wantLoc:  "ns/name/key",
			wantMods: [][2]string{
				{"m1", ""},
				{"m2", "v"},
				{"m3", ""},
				{"m4", "v"},
			},
		},
		{
			name:     "invalid empty string",
			input:    "",
			errMatch: types.ErrSecretCoordHaveNoType,
		},
		{
			name:     "invalid no scheme",
			input:    "just/a/path",
			errMatch: types.ErrSecretCoordHaveNoType,
		},
		{
			name:     "invalid no location",
			input:    "scheme://",
			errMatch: types.ErrSecretCoordHaveNoLocation,
		},
		{
			name:     "invalid modifier with percent causing unescape error",
			input:    "scheme://loc?key=100%",
			errMatch: types.ErrSecretCoordFailedParsingModifiers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.NewSecretCoord(tt.input)
			if tt.errMatch != nil {
				require.Error(t, err)
				require.True(
					t,
					errors.Is(err, tt.errMatch),
					"expected error %v, got %v",
					tt.errMatch,
					err,
				)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantType, got.Type)
			require.Equal(t, tt.wantLoc, got.Location)

			if len(tt.wantMods) == 0 {
				require.Empty(t, got.Modifiers)
			} else {
				require.Equal(t, tt.wantMods, got.Modifiers)
			}
		})
	}
}

func TestNewSecretCoord_FromJson(t *testing.T) {
	type Config struct {
		Secret types.SecretCoord `json:"secret"`
	}

	tests := []struct {
		name      string
		jsonInput string
		wantType  string
		wantLoc   string
		wantMods  [][2]string
		wantErr   bool
	}{
		{
			name:      "valid json",
			jsonInput: `{"secret": "vault://secret/data/myapp/config"}`,
			wantType:  "vault",
			wantLoc:   "secret/data/myapp/config",
			wantMods:  [][2]string{},
			wantErr:   false,
		},
		{
			name:      "valid json with modifiers",
			jsonInput: `{"secret": "k8s://ns/secret?key=value"}`,
			wantType:  "k8s",
			wantLoc:   "ns/secret",
			wantMods: [][2]string{
				{"key", "value"},
			},
			wantErr: false,
		},
		{
			name:      "valid json with multiple modifiers",
			jsonInput: `{"secret": "k8s://ns/secret?k1=v1&k2=v2"}`,
			wantType:  "k8s",
			wantLoc:   "ns/secret",
			wantMods: [][2]string{
				{"k1", "v1"},
				{"k2", "v2"},
			},
			wantErr: false,
		},
		{
			name:      "valid json with multiple modifiers that repeat",
			jsonInput: `{"secret": "k8s://ns/secret?k1=v1&k2=v2&k1=v3"}`,
			wantType:  "k8s",
			wantLoc:   "ns/secret",
			wantMods: [][2]string{
				{"k1", "v1"},
				{"k2", "v2"},
				{"k1", "v3"},
			},
			wantErr: false,
		},
		{
			name:      "invalid json format",
			jsonInput: `{"secret": 123}`,
			wantErr:   true,
		},
		{
			name:      "invalid secret format inside json",
			jsonInput: `{"secret": "invalid-scheme"}`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			err := json.Unmarshal([]byte(tt.jsonInput), &cfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.wantType, cfg.Secret.Type)
			require.Equal(t, tt.wantLoc, cfg.Secret.Location)

			if len(tt.wantMods) == 0 {
				require.Empty(t, cfg.Secret.Modifiers)
			} else {
				require.Equal(t, tt.wantMods, cfg.Secret.Modifiers)
			}
		})
	}
}
