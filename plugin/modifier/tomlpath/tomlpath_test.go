package tomlpath_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk/plugin/modifier/tomlpath/v2"
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/types"
	"github.com/detro/spelunk/v2/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretModifierTOMLPath_Type(t *testing.T) {
	mod := &tomlpath.SecretModifierTOMLPath{}
	assert.Equal(t, "tp", mod.Type())
}

func TestSecretModifierTOMLPath_Modify(t *testing.T) {
	ctx := context.Background()
	tomlData := `
[store]
[[store.book]]
category = "reference"
author = "Nigel Rees"
title = "Sayings of the Century"
price = 8.95

[[store.book]]
category = "fiction"
author = "Evelyn Waugh"
title = "Sword of Honour"
price = 12.99
`

	tests := []struct {
		name        string
		tomlPayload string
		coordStr    string
		expected    string
		errMatch    error
	}{
		{
			name:        "simple string",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.store.book[0].title",
			expected:    "Sayings of the Century",
		},
		{
			name:        "number formatting",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.store.book[1].price",
			expected:    "12.99",
		},
		{
			name:        "path not found",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.store.book[3].title",
			errMatch:    tomlpath.ErrTOMLPathFailed,
		},
		{
			name:        "invalid path syntax",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.store.[invalid",
			errMatch:    tomlpath.ErrTOMLPathInvalid,
		},
		{
			name:        "invalid toml",
			tomlPayload: `invalid = toml = `,
			coordStr:    "test://loc?tp=$.title",
			errMatch:    tomlpath.ErrSecretNotTOML,
		},
		{
			name:        "list return first element",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.store.book[*].title",
			expected:    "Sayings of the Century",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			src := util.NewMockSource("test")
			src.Val = tt.tomlPayload

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(src),
				tomlpath.WithTOMLPath(),
			)

			res, err := spelunker.DigUp(ctx, coord)
			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestSecretModifierTOMLPath_Modify_EdgeCases(t *testing.T) {
	ctx := context.Background()

	tomlData := `
float = 1.50000
whole = 10.000
negative = -42
big = 9223372036854775807
intval = 123
empty = ""
boolfalse = false
"dotted.key" = "dotval"
emptylist = []
unicode = "héllo wörld"
quoted = "say \"hi\""
multiline = """
line1
line2
"""
date = 2024-01-15
datetime = 2024-01-15T10:30:00Z
matrix = [[1, 2], [3, 4]]

[emptytable]

[nested.deep]
key = "value"

[[users]]
name = "alice"
age = 30

[[users]]
name = "bob"
age = 25
`

	tests := []struct {
		name        string
		tomlPayload string
		coordStr    string
		expected    string
		errMatch    error
	}{
		{
			name:        "float drops trailing zeros",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.float",
			expected:    "1.5",
		},
		{
			name:        "whole float has no decimal part",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.whole",
			expected:    "10",
		},
		{
			name:        "negative number",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.negative",
			expected:    "-42",
		},
		{
			name:        "large integer keeps full precision",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.big",
			expected:    "9223372036854775807",
		},
		{
			name:        "integer",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.intval",
			expected:    "123",
		},
		{
			name:        "empty string",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.empty",
			expected:    "",
		},
		{
			name:        "boolean false",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.boolfalse",
			expected:    "false",
		},
		{
			name:        "bracket notation for key containing a dot",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$['dotted.key']",
			expected:    "dotval",
		},
		{
			name:        "empty list",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.emptylist",
			expected:    "[]",
		},
		{
			name:        "empty table",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.emptytable",
			expected:    "{}",
		},
		{
			name:        "deeply nested key",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.nested.deep.key",
			expected:    "value",
		},
		{
			name:        "nested table is marshalled to JSON",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.nested",
			expected:    `{"deep":{"key":"value"}}`,
		},
		{
			name:        "list of lists (return first element)",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.matrix",
			expected:    "[1,2]",
		},
		{
			name:        "filter expression",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.users[?(@.age>26)].name",
			expected:    "alice",
		},
		{
			name:        "recursive descent (return first match)",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$..name",
			expected:    "alice",
		},
		{
			name:        "unicode is preserved",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.unicode",
			expected:    "héllo wörld",
		},
		{
			name:        "embedded quotes are preserved",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.quoted",
			expected:    `say "hi"`,
		},
		{
			name:        "multiline basic string",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.multiline",
			expected:    "line1\nline2",
		},
		{
			name:        "local date is marshalled as a JSON string",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.date",
			expected:    `"2024-01-15"`,
		},
		{
			name:        "offset datetime is marshalled as a JSON string",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.datetime",
			expected:    `"2024-01-15T10:30:00Z"`,
		},
		{
			name:        "index out of range",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.users[99]",
			errMatch:    tomlpath.ErrTOMLPathFailed,
		},
		{
			name:        "key not found",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=$.missing",
			errMatch:    tomlpath.ErrTOMLPathFailed,
		},
		{
			name:        "table holding a value that cannot be marshalled to JSON",
			tomlPayload: "[tbl]\nnotanumber = nan\n",
			coordStr:    "test://loc?tp=$.tbl",
			errMatch:    tomlpath.ErrTOMLPathMatchingFailed,
		},
		{
			name:        "empty TOMLPath",
			tomlPayload: tomlData,
			coordStr:    "test://loc?tp=",
			errMatch:    tomlpath.ErrTOMLPathFailed,
		},
		{
			name:        "empty secret is valid TOML with nothing to match",
			tomlPayload: "",
			coordStr:    "test://loc?tp=$.foo",
			errMatch:    tomlpath.ErrTOMLPathFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			src := util.NewMockSource("test")
			src.Val = tt.tomlPayload

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(src),
				tomlpath.WithTOMLPath(),
			)

			res, err := spelunker.DigUp(ctx, coord)
			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, res)
		})
	}
}
