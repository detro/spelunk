package spelunk_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/detro/spelunk"
)

// mockSource implements spelunk.SecretSource for testing
type mockSource struct {
	typ string
	val string
	err error
}

func (m *mockSource) Type() string {
	return m.typ
}

func (m *mockSource) DigUp(_ context.Context, _ spelunk.SecretCoord) (string, error) {
	return m.val, m.err
}

func TestSpelunker_DigUp(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		opts      []spelunk.SpelunkerOption
		coordStr  string
		want      string
		wantErr   bool
		errTarget error
	}{
		{
			name: "success with single source",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&mockSource{typ: "test", val: "secret-value"}),
			},
			coordStr: "test://loc",
			want:     "secret-value",
			wantErr:  false,
		},
		{
			name: "success with multiple sources",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&mockSource{typ: "src1", val: "val1"}),
				spelunk.WithSource(&mockSource{typ: "src2", val: "val2"}),
			},
			coordStr: "src2://loc",
			want:     "val2",
			wantErr:  false,
		},
		{
			name: "trim value by default",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&mockSource{typ: "test", val: "  secret  \n"}),
			},
			coordStr: "test://loc",
			want:     "secret",
			wantErr:  false,
		},
		{
			name: "disable trim value",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&mockSource{typ: "test", val: "  secret  \n"}),
				spelunk.WithoutTrimValue(),
			},
			coordStr: "test://loc",
			want:     "  secret  \n",
			wantErr:  false,
		},
		{
			name:     "unsupported source type",
			opts:     []spelunk.SpelunkerOption{}, // No sources
			coordStr: "unknown://loc",
			want:     "",
			wantErr:  true,
			errTarget: spelunk.ErrUnsupportedSecretSourceType,
		},
		{
			name: "source returns error",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&mockSource{typ: "fail", err: errors.New("boom")}),
			},
			coordStr: "fail://loc",
			want:     "",
			wantErr:  true,
			errTarget: spelunk.ErrFailedToDigUpSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Helper to create valid coordinates for testing
			coord, err := spelunk.NewSecretCoord(tt.coordStr)
			if err != nil {
				t.Fatalf("failed to create coord from %q: %v", tt.coordStr, err)
			}

			// Capture panic if WithSource fails due to uninitialized map
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic during test: %v", r)
				}
			}()

			s := spelunk.NewSpelunker(tt.opts...)
			got, err := s.DigUp(ctx, coord)

			if (err != nil) != tt.wantErr {
				t.Errorf("Spelunker.DigUp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr {
				if tt.errTarget != nil && !errors.Is(err, tt.errTarget) {
					t.Errorf("Spelunker.DigUp() error = %v, want target %v", err, tt.errTarget)
				}
				// Also check if error string contains the wrapped error
				if tt.name == "source returns error" && !strings.Contains(err.Error(), "boom") {
					t.Errorf("Spelunker.DigUp() error %v does not contain 'boom'", err)
				}
			} else {
				if got != tt.want {
					t.Errorf("Spelunker.DigUp() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}
