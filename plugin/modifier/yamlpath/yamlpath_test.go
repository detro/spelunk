package yamlpath_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/internal/testutil"
	"github.com/detro/spelunk/plugin/modifier/yamlpath"
	"github.com/detro/spelunk/types"
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

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(&testutil.MockSource{Typ: "test", Val: tt.yamlPayload}),
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
