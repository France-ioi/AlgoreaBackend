package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
)

// StructuredLogger is a structured logrus Logger.
type StructuredLogger struct {
	Logger *log.Logger
}

// NewStructuredLogger implements a custom structured logrus Logger.
func NewStructuredLogger(logger *log.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{logger})
}

// NewLogEntry sets default request log fields.
func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{Logger: log.NewEntry(l.Logger)}
	logFields := log.Fields{}

	logFields["ts"] = time.Now().UTC().Format(time.RFC1123)

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	logFields["http_scheme"] = scheme
	logFields["http_proto"] = r.Proto
	logFields["http_method"] = r.Method

	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	entry.Logger = entry.Logger.WithFields(logFields)

	entry.Logger.Infoln("request started")

	return entry
}

// StructuredLoggerEntry is a log.FieldLogger.
type StructuredLoggerEntry struct {
	Logger log.FieldLogger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.Logger = l.Logger.WithFields(log.Fields{
		"resp_status":       status,
		"resp_bytes_length": bytes,
		"resp_elapsed_ms":   float64(elapsed.Nanoseconds()) / 1000000.0,
	})

	l.Logger.Infoln("request complete")
}

// Panic prints stack trace
func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(log.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
}

// Helper methods used by the application to get the request-scoped
// logger entry and set additional fields between handlers.

// GetLogEntry return the request scoped log.FieldLogger.
//noinspection GoUnusedExportedFunction
func GetLogEntry(r *http.Request) log.FieldLogger {
	entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
	return entry.Logger
}

// LogEntrySetField adds a field to the request scoped log.FieldLogger.
//noinspection GoUnusedExportedFunction
func LogEntrySetField(r *http.Request, key string, value interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
		entry.Logger = entry.Logger.WithField(key, value)
	}
}

// LogEntrySetFields adds multiple fields to the request scoped log.FieldLogger.
//noinspection GoUnusedExportedFunction
func LogEntrySetFields(r *http.Request, fields map[string]interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
		entry.Logger = entry.Logger.WithFields(fields)
	}
}
