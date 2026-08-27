package yamlpath_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk/plugin/modifier/yamlpath/v2"
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/types"
	"github.com/detro/spelunk/v2/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretModifierYAMLPath_Type(t *testing.T) {
	mod := &yamlpath.SecretModifierYAMLPath{}
	assert.Equal(t, "yp", mod.Type())
}

func TestSecretModifierYAMLPath_Modify(t *testing.T) {
	ctx := context.Background()
	yamlData := `
store:
  book:
    - category: reference
      author: Nigel Rees
      title: Sayings of the Century
      price: 8.95
    - category: fiction
      author: Evelyn Waugh
      title: Sword of Honour
      price: 12.99
`

	tests := []struct {
		name        string
		yamlPayload string
		coordStr    string
		expected    string
		errMatch    error
	}{
		{
			name:        "simple string",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.store.book[0].title",
			expected:    "Sayings of the Century",
		},
		{
			name:        "number formatting",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.store.book[1].price",
			expected:    "12.99",
		},
		{
			name:        "path not found",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.store.book[3].title",
			errMatch:    yamlpath.ErrYAMLPathFailed,
		},
		{
			name:        "invalid path syntax",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.store.[invalid",
			errMatch:    yamlpath.ErrYAMLPathInvalid,
		},
		{
			name:        "invalid yaml",
			yamlPayload: `invalid: yaml: : [}`,
			coordStr:    "test://loc?yp=$.title",
			errMatch:    yamlpath.ErrSecretNotYAML,
		},
		{
			name:        "list return first element",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.store.book[*].title",
			expected:    "Sayings of the Century",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			src := util.NewMockSource("test")
			src.Val = tt.yamlPayload

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(src),
				yamlpath.WithYAMLPath(),
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

func TestSecretModifierYAMLPath_Modify_EdgeCases(t *testing.T) {
	ctx := context.Background()

	yamlData := `
nullval: null
tilde: ~
valueless_key:
float: 1.50000
whole: 10.000
negative: -42
big: 10000000000000000000000
intval: 123
empty: ""
boolfalse: false
unquoted_yes: yes
dotted.key: dotval
emptylist: []
emptymap: {}
nested:
  deep:
    key: value
matrix:
  - [1, 2]
  - [3, 4]
unicode: "héllo wörld"
quoted: 'say "hi"'
multiline: |
  line1
  line2
anchored: &anchor anchored value
aliased: *anchor
users:
  - name: alice
    age: 30
  - name: bob
    age: 25
`

	tests := []struct {
		name        string
		yamlPayload string
		coordStr    string
		expected    string
		errMatch    error
	}{
		{
			name:        "explicit null",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.nullval",
			errMatch:    yamlpath.ErrYAMLPathMatchingFailed,
		},
		{
			name:        "tilde is null",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.tilde",
			errMatch:    yamlpath.ErrYAMLPathMatchingFailed,
		},
		{
			name:        "key without a value is null",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.valueless_key",
			errMatch:    yamlpath.ErrYAMLPathMatchingFailed,
		},
		{
			name:        "float drops trailing zeros",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.float",
			expected:    "1.5",
		},
		{
			name:        "whole float has no decimal part",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.whole",
			expected:    "10",
		},
		{
			name:        "negative number",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.negative",
			expected:    "-42",
		},
		{
			name:        "large number avoids scientific notation",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.big",
			expected:    "10000000000000000000000",
		},
		{
			name:        "integer",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.intval",
			expected:    "123",
		},
		{
			name:        "empty string",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.empty",
			expected:    "",
		},
		{
			name:        "boolean false",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.boolfalse",
			expected:    "false",
		},
		{
			name:        "unquoted yes is a string, not a boolean",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.unquoted_yes",
			expected:    "yes",
		},
		{
			name:        "bracket notation for key containing a dot",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$['dotted.key']",
			expected:    "dotval",
		},
		{
			name:        "empty list",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.emptylist",
			expected:    "[]",
		},
		{
			name:        "empty map",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.emptymap",
			expected:    "{}",
		},
		{
			name:        "deeply nested key",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.nested.deep.key",
			expected:    "value",
		},
		{
			name:        "nested map is marshalled to JSON",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.nested",
			expected:    `{"deep":{"key":"value"}}`,
		},
		{
			name:        "list of lists (return first element)",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.matrix",
			expected:    "[1,2]",
		},
		{
			name:        "filter expression",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.users[?(@.age>26)].name",
			expected:    "alice",
		},
		{
			name:        "recursive descent (return first match)",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$..name",
			expected:    "alice",
		},
		{
			name:        "unicode is preserved",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.unicode",
			expected:    "héllo wörld",
		},
		{
			name:        "embedded quotes are preserved",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.quoted",
			expected:    `say "hi"`,
		},
		{
			name:        "multiline block scalar",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.multiline",
			expected:    "line1\nline2",
		},
		{
			name:        "anchor alias is resolved",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.aliased",
			expected:    "anchored value",
		},
		{
			name:        "index out of range",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=$.users[99]",
			errMatch:    yamlpath.ErrYAMLPathFailed,
		},
		{
			name:        "empty YAMLPath",
			yamlPayload: yamlData,
			coordStr:    "test://loc?yp=",
			errMatch:    yamlpath.ErrYAMLPathFailed,
		},
		{
			name:        "YAML list at the root",
			yamlPayload: "- first\n- second\n",
			coordStr:    "test://loc?yp=$[1]",
			expected:    "second",
		},
		{
			name:        "YAML scalar at the root",
			yamlPayload: `just a string`,
			coordStr:    "test://loc?yp=$",
			expected:    "just a string",
		},
		{
			name:        "empty secret is valid YAML with nothing to match",
			yamlPayload: "",
			coordStr:    "test://loc?yp=$.foo",
			errMatch:    yamlpath.ErrYAMLPathFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			src := util.NewMockSource("test")
			src.Val = tt.yamlPayload

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(src),
				yamlpath.WithYAMLPath(),
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
