package logging

import "github.com/sirupsen/logrus" //nolint:depguard

type jsonFormatter struct {
	logrusJSONFormatter *logrus.JSONFormatter
}

func newJSONFormatter() *jsonFormatter {
	return &jsonFormatter{logrusJSONFormatter: &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	}}
}

func (f *jsonFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return f.logrusJSONFormatter.Format(entry)
}

var _ logrus.Formatter = (*jsonFormatter)(nil)
