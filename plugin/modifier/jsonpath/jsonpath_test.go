package jsonpath_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk/plugin/modifier/jsonpath/v2"
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/types"
	"github.com/detro/spelunk/v2/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretModifierJSONPath_Type(t *testing.T) {
	mod := &jsonpath.SecretModifierJSONPath{}
	assert.Equal(t, "jp", mod.Type())
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

func TestSecretModifier_JSONPath_EdgeCases(t *testing.T) {
	ctx := context.Background()

	jsonSecret := `{
		"nullval": null,
		"float": 1.50000,
		"whole": 10.000,
		"negative": -42,
		"big": 10000000000000000000000,
		"empty": "",
		"boolfalse": false,
		"dotted.key": "dotval",
		"emptylist": [],
		"nested": {"deep": {"key": "value"}},
		"matrix": [[1, 2], [3, 4]],
		"unicode": "héllo wörld",
		"quoted": "say \"hi\"",
		"users": [{"name": "alice", "age": 30}, {"name": "bob", "age": 25}]
	}`

	tests := []struct {
		name     string
		val      string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "null value",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.nullval",
			errMatch: jsonpath.ErrJSONPathMatchingFailed,
		},
		{
			name:     "float drops trailing zeros",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.float",
			want:     "1.5",
		},
		{
			name:     "whole float has no decimal part",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.whole",
			want:     "10",
		},
		{
			name:     "negative number",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.negative",
			want:     "-42",
		},
		{
			name:     "large number avoids scientific notation",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.big",
			want:     "10000000000000000000000",
		},
		{
			name:     "empty string",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.empty",
			want:     "",
		},
		{
			name:     "boolean false",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.boolfalse",
			want:     "false",
		},
		{
			name:     "bracket notation for key containing a dot",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$['dotted.key']",
			want:     "dotval",
		},
		{
			name:     "empty list",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.emptylist",
			want:     "[]",
		},
		{
			name:     "deeply nested key",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.nested.deep.key",
			want:     "value",
		},
		{
			name:     "nested object is marshalled back to JSON",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.nested",
			want:     `{"deep":{"key":"value"}}`,
		},
		{
			name:     "list of lists (return first element)",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.matrix",
			want:     "[1,2]",
		},
		{
			name:     "filter expression",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.users[?(@.age>26)].name",
			want:     "alice",
		},
		{
			name:     "recursive descent (return first match)",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$..name",
			want:     "alice",
		},
		{
			name:     "unicode is preserved",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.unicode",
			want:     "héllo wörld",
		},
		{
			name:     "embedded quotes are preserved",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.quoted",
			want:     `say "hi"`,
		},
		{
			name:     "index out of range",
			val:      jsonSecret,
			coordStr: "test://loc?jp=$.users[99]",
			errMatch: jsonpath.ErrJSONPathFailed,
		},
		{
			name:     "empty JSONPath",
			val:      jsonSecret,
			coordStr: "test://loc?jp=",
			errMatch: jsonpath.ErrJSONPathFailed,
		},
		{
			name:     "empty JSON object",
			val:      "{}",
			coordStr: "test://loc?jp=$.foo",
			errMatch: jsonpath.ErrJSONPathFailed,
		},
		{
			name:     "JSON array at the root",
			val:      `["first", "second"]`,
			coordStr: "test://loc?jp=$[1]",
			want:     "second",
		},
		{
			name:     "JSON scalar at the root",
			val:      `"just a string"`,
			coordStr: "test://loc?jp=$",
			want:     "just a string",
		},
		{
			name:     "empty secret",
			val:      "",
			coordStr: "test://loc?jp=$.foo",
			errMatch: jsonpath.ErrSecretNotJSON,
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
