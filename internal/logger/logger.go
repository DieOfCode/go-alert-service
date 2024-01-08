package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

type LogFormatter struct {
	*zerolog.Logger
}

func (l *LogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	logger := l.With().
		Str("URI", r.RequestURI).
		Str("method", r.Method).
		Logger()

	return &LogEntry{&logger}
}

type LogEntry struct {
	*zerolog.Logger
}

func (l *LogEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ any) {
	l.Info().
		Str("elapsed", elapsed.String()).
		Int("bytes", bytes).
		Int("status", status).
		Msg("Request handled")
}

func (l *LogEntry) Panic(v any, stack []byte) {
	l.Info().
		Interface("panic", v).
		Bytes("stack", stack).
		Msg("Panic handled")
}
