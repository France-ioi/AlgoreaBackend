package logging

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus" //nolint:depguard
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const formatterTestLogMessage = "\nmessage line1\nmessage line2\n"

func Test_consoleFormatter_Format_TextTest(t *testing.T) {
	f := newConsoleFormatter()

	got, err := f.Format(makeLogEntryForTextFormatterTests())
	require.NoError(t, err)

	assert.Equal(t,
		"\033[33mWARN\033[0m[2021-01-02 03:04:05.123+03:00] "+
			"\033[35m(myfile.go:12)\033[0m "+
			"\033[36;1m[1.234s]\033[0m "+
			"  \033[33mtype\033[0m=db "+""+
			"\033[33mkey1\033[0m=line1\nline2 \033[33mkey2\033[0m=line3\nline4"+
			"\n\tmessage line1\nmessage line2\n",
		string(got))
}

func Test_consoleFormatter_Format_TextTest_WithAffectedRows(t *testing.T) {
	f := newConsoleFormatter()

	entry := makeLogEntryForTextFormatterTests()
	entry.Data["rows"] = 123

	got, err := f.Format(entry)
	require.NoError(t, err)

	assert.Equal(t,
		"\033[33mWARN\033[0m[2021-01-02 03:04:05.123+03:00] "+
			"\033[35m(myfile.go:12)\033[0m "+
			"\033[36;1m[1.234s]\033[0m "+
			"  \033[33mtype\033[0m=db "+""+
			"\033[33mkey1\033[0m=line1\nline2 \033[33mkey2\033[0m=line3\nline4"+
			"\n\tmessage line1\nmessage line2\n"+
			" \t\033[32m[123 affected]\033[0m\n",
		string(got))
}

func Test_consoleFormatter_Format_ErrorLevel(t *testing.T) {
	f := newConsoleFormatter()

	entry := logrus.NewEntry(logrus.New()).
		WithTime(time.Date(2021, 1, 2, 3, 4, 5, 123456789, time.FixedZone("MyZone", 3*60*60))).
		WithFields(map[string]interface{}{
			"key1":     "line1\nline2",
			"key2":     "line3\nline4",
			"fileline": "myfile.go:12",
			"duration": "1.234s",
			"type":     "db",
		})
	entry.Level = logrus.ErrorLevel
	entry.Message = formatterTestLogMessage

	got, err := f.Format(entry)
	require.NoError(t, err)

	assert.Equal(t,
		"\033[31mERRO\033[0m[2021-01-02 03:04:05.123+03:00] "+
			"\033[35m(myfile.go:12)\033[0m "+
			"\033[36;1m[1.234s]\033[0m "+
			"  \033[31mtype\033[0m=db "+""+
			"\033[31mkey1\033[0m=line1\nline2 \033[31mkey2\033[0m=line3\nline4"+
			"\n\t\033[31mmessage line1\nmessage line2\033[0m\n",
		string(got))
}

func Test_consoleFormatter_Format_MessageInTheMiddle(t *testing.T) {
	f := newConsoleFormatter()

	entry := logrus.NewEntry(logrus.New()).
		WithTime(time.Date(2021, 1, 2, 3, 4, 5, 123456789, time.FixedZone("MyZone", 3*60*60))).
		WithFields(map[string]interface{}{
			"key1":     "line1\nline2",
			"key2":     "line3\nline4",
			"fileline": "myfile.go:12",
			"duration": "1.234s",
		})
	entry.Level = logrus.InfoLevel
	entry.Message = formatterTestLogMessage

	got, err := f.Format(entry)
	require.NoError(t, err)

	assert.Equal(t,
		"\033[36mINFO\033[0m[2021-01-02 03:04:05.123+03:00] "+
			"\033[35m(myfile.go:12)\033[0m "+
			"\033[36;1m[1.234s]\033[0m "+
			"message line1\nmessage line2 "+
			" \033[36mkey1\033[0m=line1\nline2 \033[36mkey2\033[0m=line3\nline4\n",
		string(got))
}
