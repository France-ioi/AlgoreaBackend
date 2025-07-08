package logging

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus" //nolint:depguard
)

type consoleFormatter struct {
	textFormatter *textFormatter
}

func newConsoleFormatter() *consoleFormatter {
	textFormatter := newTextFormatter(true)
	textFormatter.logrusTextFormatter.DisableQuote = true
	return &consoleFormatter{textFormatter: textFormatter}
}

func (f *consoleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	buffer := &bytes.Buffer{}

	if fileLine, ok := entry.Data["fileline"]; ok {
		buffer.WriteString("\033[35m(")
		buffer.WriteString(fileLine.(string))
		buffer.WriteString(")\033[0m")
		delete(entry.Data, "fileline")
	}

	if duration, ok := entry.Data["duration"]; ok {
		buffer.WriteString(" \033[36;1m[")
		buffer.WriteString(duration.(string))
		buffer.WriteString("]\033[0m ")
		delete(entry.Data, "duration")
	}

	message := strings.TrimSpace(entry.Message)
	var messagePrefix, messageSuffix string
	if entry.Level < logrus.WarnLevel {
		messagePrefix = "\033[31m"
		messageSuffix = "\033[0m"
	}

	var renderMessageInTheEnd bool
	if dataType, ok := entry.Data["type"]; ok && dataType == "db" {
		renderMessageInTheEnd = true
	}

	if !renderMessageInTheEnd {
		buffer.WriteString(messagePrefix)
		buffer.WriteString(message)
		buffer.WriteString(messageSuffix)
	}

	var rowsAffectedString string
	if rowsAffected, ok := entry.Data["rows"]; ok {
		rowsAffectedString = fmt.Sprintf(" \t\033[32m[%v affected]\033[0m\n", rowsAffected)
		delete(entry.Data, "rows")
	}

	newEntry := logrus.NewEntry(entry.Logger).WithContext(entry.Context).WithFields(entry.Data).WithTime(entry.Time)
	newEntry.Level = entry.Level
	newEntry.Message = buffer.String()
	result, _ := f.textFormatter.Format(newEntry)

	if renderMessageInTheEnd {
		result = append(result, '\t')
		result = append(result, messagePrefix...)
		result = append(result, message...)
		result = append(result, messageSuffix...)
		result = append(result, '\n')
	}

	result = append(result, rowsAffectedString...)

	return result, nil
}

var _ logrus.Formatter = (*consoleFormatter)(nil)
