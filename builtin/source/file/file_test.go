package file_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/builtin/source/file"
	"github.com/detro/spelunk/v2/types"
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

	// Temp file with spaces in the name for URL-encoded path test
	tempDir := t.TempDir()
	spaceFilePath := filepath.Join(tempDir, "my secret.txt")
	require.NoError(t, os.WriteFile(spaceFilePath, []byte("space content"), 0o600))
	spaceCoordStr := "file://" + filepath.ToSlash(filepath.Join(tempDir, "my%20secret.txt"))

	tests := []struct {
		name     string
		opts     []spelunk.SpelunkerOption
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
			name:     "parent directory traversal path",
			coordStr: "file://../file/testdata/secret.txt",
			want:     "This is a secret file content.",
		},
		{
			name:     "url-encoded path",
			coordStr: spaceCoordStr,
			want:     "space content",
		},
		{
			name:     "file with whitespace trimmed",
			coordStr: "file://testdata/secret_whitespace.txt",
			want:     "secret with whitespace",
		},
		{
			name:     "file with whitespace untrimmed",
			opts:     []spelunk.SpelunkerOption{spelunk.WithoutTrimValue()},
			coordStr: "file://testdata/secret_whitespace.txt",
			want:     "  secret with whitespace  \n",
		},
		{
			name:     "directory read failure",
			coordStr: "file://testdata",
			errMatch: file.ErrSecretSourceFileFailedRead,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := append(
				[]spelunk.SpelunkerOption{spelunk.WithSource(&file.SecretSourceFile{})},
				tt.opts...)
			spelunker := spelunk.NewSpelunker(opts...)

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
