package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

const (
	LevelTrace = slog.Level(-8) // More verbose than DEBUG (-4)
	LevelFatal = slog.Level(12) // More severe than ERROR (8)
)

func tweakLogsAttributes(_ []string, a slog.Attr) slog.Attr {
	// Render the additional levels we added
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		switch level {
		case LevelTrace:
			a.Value = slog.StringValue("TRACE")
		case LevelFatal:
			a.Value = slog.StringValue("FATAL")
		}
	}

	if a.Key == slog.TimeKey {
		a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339))
	}
	return a
}

type LogFormat string

const (
	LogFormatBoring = "boring"
	LogFormatJSON   = "json"
	LogFormatTinted = "tinted"
)

// newTextLogger returns a "vanilla" slog instance.
func newTextLogger(w io.Writer, l slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level:       l,
		ReplaceAttr: tweakLogsAttributes,
	}))
}

// newJSONLogger returns a slog logger, configured to write to a specific io.Writer in JSON format.
func newJSONLogger(w io.Writer, l slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       l,
		ReplaceAttr: tweakLogsAttributes,
	}))
}

// newColoredLogger returns a slog wrapping the Tint logger.
func newColoredLogger(w io.Writer, l slog.Level) *slog.Logger {
	return slog.New(
		tint.NewTextHandler(w, &tint.Options{Level: l, ReplaceAttr: tweakLogsAttributes}),
	)
}

// SetupDefaultLogger configures the default logger (i.e. slog.Default()).
// The verbosity is a value between `<= 0: Error` and `>= 4: Trace`.
// The format is one of the possible values of LogFormat.
//
// Returns the new default logger.
func SetupDefaultLogger(verbosity int8, format LogFormat) *slog.Logger {
	var level slog.Level
	switch {
	case verbosity <= 0:
		level = slog.LevelError
	case verbosity == 1:
		level = slog.LevelWarn
	case verbosity == 2:
		level = slog.LevelInfo
	case verbosity == 3:
		level = slog.LevelDebug
	case verbosity >= 4:
		level = LevelTrace
	}

	var newLogger *slog.Logger
	switch format {
	case LogFormatJSON:
		newLogger = newJSONLogger(os.Stderr, level)
	case LogFormatTinted:
		newLogger = newColoredLogger(os.Stderr, level)
	case LogFormatBoring:
	default:
		newLogger = newTextLogger(os.Stderr, level)
	}

	slog.SetDefault(newLogger)

	ctx := context.Background()
	slog.Log(ctx, LevelTrace, "Set default logger", "level", level, "format", format)
	return newLogger
}
