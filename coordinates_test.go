package spelunk_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/detro/spelunk"
)

func TestNewSecretCoord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
		wantLoc  string
		wantMods map[string]string
		errMatch error
	}{
		{
			name:     "valid simple coordinate",
			input:    "vault://secret/data/myapp/config",
			wantType: "vault",
			wantLoc:  "secret/data/myapp/config",
			wantMods: map[string]string{},
		},
		{
			name:     "valid coordinate with jsonpath modifier",
			input:    "k8s://mynamespace/mysecret/mycredentials?jsonpath=.kafka.password",
			wantType: "k8s",
			wantLoc:  "mynamespace/mysecret/mycredentials",
			wantMods: map[string]string{"jsonpath": ".kafka.password"},
		},
		{
			name:     "valid coordinate with simple location",
			input:    "env://JUVE_MERDA",
			wantType: "env",
			wantLoc:  "JUVE_MERDA",
			wantMods: map[string]string{},
		},
		{
			name:     "valid coordinate with userinfo in the URI",
			input:    "env://JUVE@MERDA",
			wantType: "env",
			wantLoc:  "JUVE@MERDA",
			wantMods: map[string]string{},
		},
		{
			name:     "valid coordinate with userinfo in the URI",
			input:    "env://JUVE:MERDA@TORINO:1897",
			wantType: "env",
			wantLoc:  "JUVE:MERDA@TORINO:1897",
			wantMods: map[string]string{},
		},
		{
			name:     "valid coordinate with encoded characters in path",
			input:    "file:///etc/secrets/my%20secret.json",
			wantType: "file",
			wantLoc:  "/etc/secrets/my secret.json",
			wantMods: map[string]string{},
		},
		{
			name:     "valid coordinate with local file in in path",
			input:    "file://local/file.json",
			wantType: "file",
			wantLoc:  "local/file.json",
			wantMods: map[string]string{},
		},
		{
			name:     "valid coordinate with relative file in in path",
			input:    "file://./local/file.json",
			wantType: "file",
			wantLoc:  "./local/file.json",
			wantMods: map[string]string{},
		},
		{
			name:     "valid coordinate with encoded characters in modifiers",
			input:    "dummy://loc?jsonpath=%24.phoneNumbers%5B0%5D.type",
			wantType: "dummy",
			wantLoc:  "loc",
			wantMods: map[string]string{"jsonpath": "$.phoneNumbers[0].type"},
		},
		{
			name:     "invalid empty string",
			input:    "",
			errMatch: spelunk.ErrSecretCoordHaveNoType,
		},
		{
			name:     "invalid no scheme",
			input:    "just/a/path",
			errMatch: spelunk.ErrSecretCoordHaveNoType,
		},
		{
			name:     "invalid no location",
			input:    "scheme://",
			errMatch: spelunk.ErrSecretCoordHaveNoLocation,
		},
		{
			name:     "invalid modifier with percent causing unescape error",
			input:    "scheme://loc?key=100%25", // decodes to "100%", then fails 2nd unescape
			errMatch: spelunk.ErrSecretCoordFailedParsingModifiers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := spelunk.NewSecretCoord(tt.input)
			if (err != nil) != (tt.errMatch != nil) {
				t.Errorf("NewSecretCoord() error = %v, wantErr %v", err, tt.errMatch != nil)
				return
			}
			if tt.errMatch != nil {
				if !errors.Is(err, tt.errMatch) {
					t.Errorf("NewSecretCoord() error = %v, want error to contain %v", err, tt.errMatch)
				}
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("NewSecretCoord() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Location != tt.wantLoc {
				t.Errorf("NewSecretCoord() Location = %v, want %v", got.Location, tt.wantLoc)
			}
			if len(got.Modifiers) != len(tt.wantMods) {
				t.Errorf("NewSecretCoord() Modifiers length = %v, want %v", len(got.Modifiers), len(tt.wantMods))
			}
			for k, v := range tt.wantMods {
				if gotVal, ok := got.Modifiers[k]; !ok || gotVal != v {
					t.Errorf("NewSecretCoord() Modifier[%q] = %v, want %v", k, gotVal, v)
				}
			}
		})
	}
}

func TestNewSecretCoord_FromJson(t *testing.T) {
	type Config struct {
		Secret spelunk.SecretCoord `json:"secret"`
	}

	tests := []struct {
		name      string
		jsonInput string
		wantType  string
		wantLoc   string
		wantMods  map[string]string
		wantErr   bool
	}{
		{
			name:      "valid json",
			jsonInput: `{"secret": "vault://secret/data/myapp/config"}`,
			wantType:  "vault",
			wantLoc:   "secret/data/myapp/config",
			wantMods:  map[string]string{},
			wantErr:   false,
		},
		{
			name:      "valid json with modifiers",
			jsonInput: `{"secret": "k8s://ns/secret?key=value"}`,
			wantType:  "k8s",
			wantLoc:   "ns/secret",
			wantMods:  map[string]string{"key": "value"},
			wantErr:   false,
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
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cfg.Secret.Type != tt.wantType {
					t.Errorf("Secret.Type = %v, want %v", cfg.Secret.Type, tt.wantType)
				}
				if cfg.Secret.Location != tt.wantLoc {
					t.Errorf("Secret.Location = %v, want %v", cfg.Secret.Location, tt.wantLoc)
				}
				for k, v := range tt.wantMods {
					if gotVal, ok := cfg.Secret.Modifiers[k]; !ok || gotVal != v {
						t.Errorf("Secret.Modifiers[%q] = %v, want %v", k, gotVal, v)
					}
				}
			}
		})
	}
}
