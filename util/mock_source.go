package util

import (
	"context"

	"github.com/detro/spelunk/v2/types"
)

// MockSource implements spelunk.SecretSource for testing
type MockSource struct {
	typ string
	Val string
	Err error
}

// NewMockSource creates a new MockSource with the required typ field
func NewMockSource(typ string) *MockSource {
	return &MockSource{
		typ: typ,
	}
}

func (m *MockSource) Type() string {
	return m.typ
}

func (m *MockSource) DigUp(_ context.Context, _ types.SecretCoord) (string, error) {
	return m.Val, m.Err
}

var _ types.SecretSource = (*MockSource)(nil)
