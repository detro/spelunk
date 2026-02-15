package file

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/detro/spelunk/types"
)

var (
	ErrSecretSourceFileDoesNotExist = fmt.Errorf("secret file does not exist")
	ErrSecretSourceFileFailedOpen   = fmt.Errorf("failed to open secret file")
	ErrSecretSourceFileFailedRead   = fmt.Errorf("failed to read secret file")
)

// SecretSourceFile digs up secrets from local files.
// The URI scheme for this source is "file". Examples:
//
//	file:///path/to/secret.txt
//	file://relative/path/to/secret.txt
//	file://./path/to/secret/from/this/directory.txt
//
// This types.SecretSource is built-in to spelunker.Spelunker.
type SecretSourceFile struct{}

var _ types.SecretSource = (*SecretSourceFile)(nil)

func (s *SecretSourceFile) Type() string {
	return "file"
}

func (s *SecretSourceFile) DigUp(_ context.Context, coord types.SecretCoord) (string, error) {
	if _, err := os.Stat(coord.Location); os.IsNotExist(err) {
		return "", fmt.Errorf("%w: %q", ErrSecretSourceFileDoesNotExist, coord.Location)
	}

	f, err := os.Open(coord.Location)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrSecretSourceFileFailedOpen, coord.Location, err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrSecretSourceFileFailedRead, coord.Location, err)
	}

	return string(content), nil
}
