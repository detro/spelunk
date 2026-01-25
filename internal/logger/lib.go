package logger

import (
	"io"
	"log/slog"
	"os"
)

// Default returns the default slog logger
func Default() *slog.Logger {
	return slog.Default()
}

// DefaultJSON returns an slog logger, configured to write to STDOUT (default).
func DefaultJSON() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

// JSON returns an slog logger, configured to write to a specific io.Writer.
func JSON(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, nil))
}
