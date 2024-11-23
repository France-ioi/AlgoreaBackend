package logging

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus" //nolint:depguard
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_jsonFormatter_Format(t *testing.T) {
	f := newJSONFormatter()

	expectedTime := time.Date(2021, 1, 2, 3, 4, 5, 123456789, time.FixedZone("MyZone", 3*60*60))
	entry := logrus.NewEntry(logrus.New()).WithTime(expectedTime).WithFields(map[string]interface{}{
		"key1": "line1\nline2",
		"key2": "line3\nline4",
	})
	entry.Level = logrus.WarnLevel
	entry.Message = formatterTestLogMessage
	got, err := f.Format(entry)
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"level":"warning",
		"msg":"\nmessage line1\nmessage line2\n",
		"time":"2021-01-02T03:04:05.123+03:00",
		"key1":"line1\nline2",
		"key2":"line3\nline4"
	}`, string(got))
}
