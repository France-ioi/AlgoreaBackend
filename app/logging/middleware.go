package logging

// Code inspired by https://github.com/dhax/go-base/ under MIT License:
//
// MIT License
//
// Copyright (c) 2017 Dikton Haxhijaj
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus" //nolint:depguard
)

// StructuredLogger is a structured logrus Logger.
type StructuredLogger struct{}

// NewStructuredLogger implements a custom structured logger using the global one.
func NewStructuredLogger() func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{})
}

// NewLogEntry sets default request log fields.
func (l *StructuredLogger) NewLogEntry(httpRequest *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{Logger: EntryFromContext(httpRequest.Context())}
	logFields := logrus.Fields{}

	logFields["type"] = "web"

	scheme := "http"
	if httpRequest.TLS != nil {
		scheme = "https"
	}

	logFields["http_scheme"] = scheme
	logFields["http_proto"] = httpRequest.Proto
	logFields["http_method"] = httpRequest.Method

	logFields["remote_addr"] = httpRequest.RemoteAddr
	logFields["user_agent"] = httpRequest.UserAgent()

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, httpRequest.Host, httpRequest.RequestURI)

	entry.Logger = entry.Logger.WithFields(logFields)

	entry.Logger.Infoln("request started")

	return entry
}

var _ middleware.LogFormatter = &StructuredLogger{}

// StructuredLoggerEntry wraps a logrus.FieldLogger.
type StructuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"resp_status":       status,
		"resp_bytes_length": bytes,
		"resp_elapsed_ms":   float64(elapsed.Nanoseconds()) / float64(time.Millisecond),
	})

	l.Logger.Infoln("request complete")
}

// Panic prints stack trace.
func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
}

// Helper methods used by the application to get the request-scoped
// logger entry and set additional fields between handlers.

// GetLogEntry return the request scoped logrus.FieldLogger.
// noinspection GoUnusedExportedFunction.
func GetLogEntry(r *http.Request) logrus.FieldLogger {
	entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
	return entry.Logger
}

// LogEntrySetField adds a field to the request scoped logrus.FieldLogger.
// noinspection GoUnusedExportedFunction.
func LogEntrySetField(r *http.Request, key string, value interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
		entry.Logger = entry.Logger.WithField(key, value)
	}
}

// LogEntrySetFields adds multiple fields to the request scoped logrus.FieldLogger.
// noinspection GoUnusedExportedFunction.
func LogEntrySetFields(r *http.Request, fields map[string]interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
		entry.Logger = entry.Logger.WithFields(fields)
	}
}
