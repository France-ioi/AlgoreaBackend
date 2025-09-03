package logging

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus" //nolint:depguard
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeLogEntryForTextFormatterTests() *logrus.Entry {
	entry := logrus.NewEntry(logrus.New()).
		WithTime(time.Date(2021, 1, 2, 3, 4, 5, 123456789, time.FixedZone("MyZone", 3*60*60))).
		WithFields(map[string]interface{}{
			"key1":     "line1\nline2",
			"key2":     "line3\nline4",
			"fileline": "myfile.go:12",
			"duration": "1.234s",
			"type":     "db",
		})
	entry.Level = logrus.WarnLevel
	entry.Message = formatterTestLogMessage
	return entry
}

func Test_textFormatter_Format_Colored(t *testing.T) {
	f := newTextFormatter(true)

	got, err := f.Format(makeLogEntryForTextFormatterTests())
	require.NoError(t, err)

	assert.Equal(t,
		"\033[33mWARN\033[0m[2021-01-02 03:04:05.123+03:00] message line1 message line2                   "+
			"\033[33mfileline\033[0m=\"myfile.go:12\" \033[33mduration\033[0m=1.234s \033[33mtype\033[0m=db "+""+
			"\033[33mkey1\033[0m=\"line1\\nline2\" \033[33mkey2\033[0m=\"line3\\nline4\"\n",
		string(got))
}

func Test_textFormatter_Format_NotColored(t *testing.T) {
	f := newTextFormatter(false)

	got, err := f.Format(makeLogEntryForTextFormatterTests())
	require.NoError(t, err)

	assert.Equal(t,
		"level=warning time=\"2021-01-02 03:04:05.123+03:00\" msg=\"message line1\\nmessage line2\" "+
			"fileline=\"myfile.go:12\" duration=1.234s type=db "+""+
			"key1=\"line1\\nline2\" key2=\"line3\\nline4\"\n",
		string(got))
}
