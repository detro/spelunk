package jsonpath_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/builtin/modifier/jsonpath"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
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

func (m *mockSource) DigUp(_ context.Context, _ types.SecretCoord) (string, error) {
	return m.val, m.err
}

func TestSecretModifier_JSONPath(t *testing.T) {
	ctx := context.Background()

	jsonSecret := `{
		"foo": "bar",
		"num": 123,
		"bool": true,
		"list": ["a", "b"],
		"nested": {"key": "value"},
		"users": [{"name": "alice"}, {"name": "bob"}]
	}`

	tests := []struct {
		name     string
		val      string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "simple string",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.foo",
			want:     "bar",
		},
		{
			name:     "number",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.num",
			want:     "123",
		},
		{
			name:     "boolean",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.bool",
			want:     "true",
		},
		{
			name:     "nested object",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.nested",
			want:     `{"key":"value"}`,
		},
		{
			name:     "list (return first element)",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.list",
			want:     "a",
		},
		{
			name:     "list explicit index",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.list[1]",
			want:     "b",
		},
		{
			name:     "invalid json",
			val:      "not json",
			coordStr: "test://loc?jp=$.foo",
			errMatch: jsonpath.ErrSecretNotJSON,
		},
		{
			name:     "path not found",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.missing",
			errMatch: jsonpath.ErrJSONPathFailed,
		},
		{
			name:     "multiple matches (return first)",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.users[*].name",
			want:     "alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(&mockSource{typ: "test", val: tt.val}),
				spelunk.WithModifier(&jsonpath.SecretModifierJSONPath{}),
			)

			got, err := spelunker.DigUp(ctx, coord)
			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}
