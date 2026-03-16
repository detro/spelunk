package testutil

import (
	"context"

	"github.com/detro/spelunk/types"
)

// MockSource implements spelunk.SecretSource for testing
type MockSource struct {
	Typ string
	Val string
	Err error
}

func (m *MockSource) Type() string {
	return m.Typ
}

func (m *MockSource) DigUp(_ context.Context, _ types.SecretCoord) (string, error) {
	return m.Val, m.Err
}

var _ types.SecretSource = (*MockSource)(nil)
