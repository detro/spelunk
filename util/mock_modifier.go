package util

import (
	"context"

	"github.com/detro/spelunk/v2/types"
)

// MockModifier implements types.SecretModifier for testing
type MockModifier struct {
	typ      string
	Val      string
	Err      error
	ArgToVal map[string]string
}

// NewMockModifier creates a new MockModifier with the required typ field
func NewMockModifier(typ string) *MockModifier {
	return &MockModifier{
		typ: typ,
	}
}

func (m *MockModifier) Type() string {
	return m.typ
}

func (m *MockModifier) Modify(_ context.Context, value string, modArg string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	if m.Val != "" {
		return m.Val, nil
	}
	if m.ArgToVal != nil {
		if val, ok := m.ArgToVal[modArg]; ok {
			return val, nil
		}
	}
	if modArg == "key" {
		return "mysecret", nil
	}
	return value, nil
}

var _ types.SecretModifier = (*MockModifier)(nil)
