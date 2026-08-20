package lemon

import (
	"io"
	"log/slog"
	"strings"
)

// logLevels maps a level name to slog.Level. Accepted names
// (case-insensitive): debug, info, warn, error, critical.
var logLevels = map[string]slog.Level{
	"debug":    slog.LevelDebug,
	"info":     slog.LevelInfo,
	"warn":     slog.LevelWarn,
	"error":    slog.LevelError,
	"critical": slog.LevelError + 4,
}

// NewLogger builds a *slog.Logger writing to w at the level described by
// c.LogLevel. Unknown level names fall back to info.
func NewLogger(c Config, w io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: ParseLevel(c.LogLevel),
	}))
}

// ParseLevel resolves a level name to slog.Level. Unknown names fall back
// to slog.LevelInfo.
func ParseLevel(name string) slog.Level {
	if lvl, ok := logLevels[strings.ToLower(name)]; ok {
		return lvl
	}
	return slog.LevelInfo
}
