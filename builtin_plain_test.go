package spelunk_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk"
)

func TestSecretSourcePlain_Type(t *testing.T) {
	s := &spelunk.SecretSourcePlain{}
	if got := s.Type(); got != "plain" {
		t.Errorf("SecretSourcePlain.Type() = %v, want %v", got, "plain")
	}
}

func TestSecretSourcePlain_DigUp(t *testing.T) {
	tests := []struct {
		name     string
		coordStr string
		want     string
		wantErr  bool
	}{
		{
			name:     "simple value",
			coordStr: "plain://my-secret",
			want:     "my-secret",
			wantErr:  false,
		},
		{
			name:     "value with path",
			coordStr: "plain://my/nested/secret",
			want:     "my/nested/secret",
			wantErr:  false,
		},
		{
			name:     "value with special chars",
			coordStr: "plain://user:pass@host",
			want:     "user:pass@host",
			wantErr:  false,
		},
		{
			name:     "empty value",
			coordStr: "plain://",
			want:     "",
			wantErr:  false, // It's valid to have an empty secret
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var coord *spelunk.SecretCoord
			var err error

			s := &spelunk.SecretSourcePlain{}

			// Attempt to parse coordinate
			coord, err = spelunk.NewSecretCoord(tt.coordStr)
			
			// Handle expected parsing failure for empty location
			if err != nil {
				if tt.coordStr == "plain://" && tt.want == "" && !tt.wantErr {
					coord = &spelunk.SecretCoord{Type: "plain", Location: ""}
				} else {
					t.Fatalf("failed to parse coord %q: %v", tt.coordStr, err)
				}
			}

			got, err := s.DigUp(context.Background(), *coord)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretSourcePlain.DigUp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SecretSourcePlain.DigUp() = %v, want %v", got, tt.want)
			}
		})
	}
}
