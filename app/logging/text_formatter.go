package logging

import (
	"sort"
	"strings"

	"github.com/sirupsen/logrus" //nolint:depguard

	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

type textFormatter struct {
	logrusTextFormatter *logrus.TextFormatter
}

func newTextFormatter(useColors bool) *textFormatter {
	return &textFormatter{logrusTextFormatter: &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000Z07:00",
		ForceColors:     useColors,
		SortingFunc:     logFieldsSorter,
	}}
}

func (f *textFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if !f.logrusTextFormatter.DisableQuote && f.logrusTextFormatter.ForceColors {
		entry.Message = spacesRegexp.ReplaceAllString(entry.Message, " ")
	}
	entry.Message = strings.TrimSpace(entry.Message)
	return f.logrusTextFormatter.Format(entry)
}

var _ logrus.Formatter = (*textFormatter)(nil)

func logFieldsSorter(keys []string) {
	newKeys := make([]string, 0, len(keys))

	keysSet := golang.NewSet(keys...)
	moveSetElementIntoSliceIfExists(keysSet, &newKeys, logrus.FieldKeyLevel)
	moveSetElementIntoSliceIfExists(keysSet, &newKeys, logrus.FieldKeyTime)
	moveSetElementIntoSliceIfExists(keysSet, &newKeys, logrus.FieldKeyMsg)
	moveSetElementIntoSliceIfExists(keysSet, &newKeys, logrus.FieldKeyLogrusError)
	moveSetElementIntoSliceIfExists(keysSet, &newKeys, "fileline")
	moveSetElementIntoSliceIfExists(keysSet, &newKeys, "duration")
	moveSetElementIntoSliceIfExists(keysSet, &newKeys, "type")

	otherValues := keysSet.Values()
	sort.Strings(otherValues)
	newKeys = append(newKeys, otherValues...)

	copy(keys, newKeys)
}

func moveSetElementIntoSliceIfExists(set *golang.Set[string], slice *[]string, element string) {
	if set.Contains(element) {
		*slice = append(*slice, element)
		set.Remove(element)
	}
}
