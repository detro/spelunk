package xpath

import (
	"context"
	"fmt"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
)

var (
	ErrXPathFailed         = fmt.Errorf("failed to apply XPath")
	ErrXPathMatchingFailed = fmt.Errorf("failed to match XPath")
	ErrSecretNotXML        = fmt.Errorf("secret is not a valid XML")
)

// SecretModifierXPath is a modifier that can extract a specific field out of an XML stored in a secret value.
// After the secret has been dug-up, the modifier digs further at the provided XPath, and returns
// the inner text found there.
type SecretModifierXPath struct{}

var _ types.SecretModifier = (*SecretModifierXPath)(nil)

func (s *SecretModifierXPath) Type() string {
	return "xp"
}

func (s *SecretModifierXPath) Modify(
	_ context.Context,
	secretValue string,
	mod string,
) (string, error) {
	xmlDoc, err := xmlquery.Parse(strings.NewReader(secretValue))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSecretNotXML, err)
	}

	node, err := xmlquery.Query(xmlDoc, mod)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrXPathFailed, mod, err)
	}
	if node == nil {
		return "", fmt.Errorf("%w: result is null for path %q", ErrXPathMatchingFailed, mod)
	}

	switch node.Type {
	case xmlquery.CommentNode,
		xmlquery.DeclarationNode,
		xmlquery.ProcessingInstruction,
		xmlquery.NotationNode:
		return node.Data, nil
	default:
		return node.InnerText(), nil
	}
}

// WithXPath adds the XPath modifier to a Spelunker.
func WithXPath() spelunk.SpelunkerOption {
	return spelunk.WithModifier(&SecretModifierXPath{})
}
