package tomlpath_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/internal/testutil"
	"github.com/detro/spelunk/plugin/modifier/tomlpath"
	"github.com/detro/spelunk/types"
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

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(&testutil.MockSource{Typ: "test", Val: tt.tomlPayload}),
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
