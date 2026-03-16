package spelunk_test

import (
	"context"
	"errors"
	"testing"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/internal/testutil"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
)

// mockModifier implements types.SecretModifier for testing.
// It takes the given mod string and appends it as `_<mod>` to the resulting secret.
type mockModifier struct {
	typ string
}

func (m *mockModifier) Type() string {
	return m.typ
}

func (m *mockModifier) Modify(_ context.Context, secretValue string, mod string) (string, error) {
	return secretValue + "_" + mod, nil
}

func TestSpelunker_DigUp(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		opts     []spelunk.SpelunkerOption
		coordStr string
		want     string
		errMatch error
	}{
		{
			name: "success with single source",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&testutil.MockSource{Typ: "test", Val: "secret-value"}),
			},
			coordStr: "test://loc",
			want:     "secret-value",
		},
		{
			name: "success with multiple sources",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&testutil.MockSource{Typ: "src1", Val: "val1"}),
				spelunk.WithSource(&testutil.MockSource{Typ: "src2", Val: "val2"}),
			},
			coordStr: "src2://loc",
			want:     "val2",
		},
		{
			name: "modifiers applied in order",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&testutil.MockSource{Typ: "src", Val: "val"}),
				spelunk.WithModifier(&mockModifier{typ: "mod1"}),
				spelunk.WithModifier(&mockModifier{typ: "mod2"}),
			},
			coordStr: "src://loc?mod1=a&mod2=b&mod1=c",
			want:     "val_a_b_c",
		},
		{
			name: "trim value by default",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&testutil.MockSource{Typ: "test", Val: "  secret  \n"}),
			},
			coordStr: "test://loc",
			want:     "secret",
		},
		{
			name: "disable trim value",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&testutil.MockSource{Typ: "test", Val: "  secret  \n"}),
				spelunk.WithoutTrimValue(),
			},
			coordStr: "test://loc",
			want:     "  secret  \n",
		},
		{
			name:     "unsupported source type",
			opts:     []spelunk.SpelunkerOption{},
			coordStr: "unknown://loc",
			want:     "",
			errMatch: spelunk.ErrUnsupportedSecretSourceType,
		},
		{
			name: "source returns error",
			opts: []spelunk.SpelunkerOption{
				spelunk.WithSource(&testutil.MockSource{Typ: "fail", Err: errors.New("boom")}),
			},
			coordStr: "fail://loc",
			want:     "",
			errMatch: spelunk.ErrFailedToDigUpSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			spelunker := spelunk.NewSpelunker(tt.opts...)
			got, err := spelunker.DigUp(ctx, coord)

			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
				if tt.name == "source returns error" {
					require.ErrorContains(t, err, "boom")
				}
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}
