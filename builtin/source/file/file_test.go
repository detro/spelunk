package file_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/builtin/source/file"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
)

func TestSecretSourceFile_Type(t *testing.T) {
	s := &file.SecretSourceFile{}
	if got := s.Type(); got != "file" {
		t.Errorf("SecretSourceFile.Type() = %v, want %v", got, "file")
	}
}

func TestSecretSourceFile_DigUp(t *testing.T) {
	absSecretTxtFile, err := filepath.Abs("testdata/secret.txt")
	require.NoError(t, err)

	nonReadable, _ := os.CreateTemp("", "non-readable")
	defer func() { _ = os.Remove(nonReadable.Name()) }()
	_ = os.Chmod(nonReadable.Name(), 0o200)

	tests := []struct {
		name     string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "absolute path",
			coordStr: "file://" + absSecretTxtFile,
			want:     "This is a secret file content.",
		},
		{
			name:     "relative path",
			coordStr: "file://testdata/secret.txt",
			want:     "This is a secret file content.",
		},
		{
			name:     "relative (local) path",
			coordStr: "file://./testdata/secret.txt",
			want:     "This is a secret file content.",
		},
		{
			name:     "non-existent file",
			coordStr: "file:///path/to/non/existent/file",
			errMatch: types.ErrSecretNotFound,
		},
		{
			name:     "non-readable file",
			coordStr: "file://" + nonReadable.Name(),
			errMatch: file.ErrSecretSourceFileFailedOpen,
		},
	}

	spelunker := spelunk.NewSpelunker(
		spelunk.WithSource(&file.SecretSourceFile{}),
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			got, err := spelunker.DigUp(context.Background(), coord)
			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}
