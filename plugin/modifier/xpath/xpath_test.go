package xpath_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/internal/testutil"
	"github.com/detro/spelunk/plugin/modifier/xpath"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretModifierXPath_Type(t *testing.T) {
	mod := &xpath.SecretModifierXPath{}
	assert.Equal(t, "xp", mod.Type())
}

func TestSecretModifierXPath_Modify(t *testing.T) {
	ctx := context.Background()
	xmlData := `
<store>
	<book>
		<category>reference</category>
		<author>Nigel Rees</author>
		<title>Sayings of the Century</title>
		<price>8.95</price>
	</book>
	<book>
		<category>fiction</category>
		<author>Evelyn Waugh</author>
		<title>Sword of Honour</title>
		<price>12.99</price>
	</book>
</store>
`

	tests := []struct {
		name       string
		xmlPayload string
		coordStr   string
		expected   string
		errMatch   error
	}{
		{
			name:       "simple query",
			xmlPayload: xmlData,
			coordStr:   "test://loc?xp=//book[1]/title",
			expected:   "Sayings of the Century",
		},
		{
			name:       "number query",
			xmlPayload: xmlData,
			coordStr:   "test://loc?xp=//book[2]/price",
			expected:   "12.99",
		},
		{
			name:       "path not found",
			xmlPayload: xmlData,
			coordStr:   "test://loc?xp=//book[3]/title",
			errMatch:   xpath.ErrXPathMatchingFailed,
		},
		{
			name:       "invalid xml",
			xmlPayload: `{"not": "xml"}`,
			coordStr:   "test://loc?xp=//title",
			errMatch:   xpath.ErrSecretNotXML,
		},
		{
			name:       "invalid xpath",
			xmlPayload: xmlData,
			coordStr:   "test://loc?xp=//[invalid",
			errMatch:   xpath.ErrXPathFailed,
		},
		{
			name:       "comment node",
			xmlPayload: `<?xml version="1.0"?><!-- secret code is 1234 --><store></store>`,
			coordStr:   "test://loc?xp=//comment()",
			expected:   "secret code is 1234",
		},
		{
			name:       "processing instruction node",
			xmlPayload: `<?xml version="1.0"?><?test instruction?><store></store>`,
			coordStr:   "test://loc?xp=//processing-instruction('test')",
			expected:   "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(&testutil.MockSource{Typ: "test", Val: tt.xmlPayload}),
				xpath.WithXPath(),
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
