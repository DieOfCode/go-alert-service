package handler

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/DieOfCode/go-alert-service/internal/metrics"
	"github.com/go-chi/chi/v5/middleware"
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

func Decompress(l *zerolog.Logger) func(next http.Handler) http.Handler {
	gr := new(gzip.Reader)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			encodingHeaders := r.Header.Values("Content-Encoding")
			if !slices.Contains(encodingHeaders, "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			if err := gr.Reset(r.Body); err != nil {
				l.Error().Err(err).Msg("Reset gzip reader error")
			}
			defer gr.Close()

			r.Body = gr
			next.ServeHTTP(w, r)
		})
	}
}

func CheckHash(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hashSHA256 := r.Header.Get("HashSHA256")
			if hashSHA256 == "" {
				next.ServeHTTP(w, r)
				return
			}
			h := hmac.New(sha256.New, []byte(key))
			b, err := io.ReadAll(r.Body)
			if err != nil {
				writeResponse(w, http.StatusBadRequest, metrics.Error{Error: "Bad Request"})
				return
			}
			if _, err := h.Write(b); err != nil {
				writeResponse(w, http.StatusInternalServerError, metrics.Error{Error: "Internal Server Error"})
				return
			}
			d := h.Sum(nil)
			hh, err := hex.DecodeString(hashSHA256)
			if err != nil {
				writeResponse(w, http.StatusInternalServerError, metrics.Error{Error: "Internal Server Error"})
				return
			}
			if !hmac.Equal(d, hh) {
				writeResponse(w, http.StatusBadRequest, metrics.Error{Error: "Bad Request"})
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(b))
			next.ServeHTTP(w, r)
		})
	}
}
