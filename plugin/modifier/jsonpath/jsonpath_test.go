package jsonpath_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk/plugin/modifier/jsonpath/v2"
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/types"
	"github.com/detro/spelunk/v2/util"
	"github.com/stretchr/testify/require"
)

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
			name:     "invalid jsonpath syntax",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.[invalid",
			errMatch: jsonpath.ErrJSONPathInvalid,
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

			src := util.NewMockSource("test")
			src.Val = tt.val

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(src),
				jsonpath.WithJSONPath(),
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
