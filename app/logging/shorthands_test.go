package logging

import (
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_ShorthandsBasic(t *testing.T) {
	hook, restoreFct := MockSharedLoggerHook()
	defer restoreFct()
	SharedLogger.Level = logrus.DebugLevel

	fakeExit := func(int) {
		panic("os.Exit called")
	}
	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	tests := []struct {
		name  string
		fct   func(args ...interface{})
		level logrus.Level
		panic bool
	}{
		{"Debug", Debug, logrus.DebugLevel, false},
		{"Info", Info, logrus.InfoLevel, false},
		{"Warn", Warn, logrus.WarnLevel, false},
		{"Error", Error, logrus.ErrorLevel, false},
		{"Fatal", Fatal, logrus.FatalLevel, true}, // does `os.exit(1)`, monkey-patched to panic
		{"Panic", Panic, logrus.PanicLevel, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()
			if tt.panic {
				assert.Panics(t, func() { tt.fct("a message") })
			} else {
				tt.fct("a message")
			}
			assert.Len(t, hook.AllEntries(), 1)
			assert.Equal(t, tt.level, hook.LastEntry().Level)
			assert.Equal(t, "a message", hook.LastEntry().Message)
		})
	}
}

func Test_ShorthandsFormatted(t *testing.T) {
	hook, restoreFct := MockSharedLoggerHook()
	defer restoreFct()
	SharedLogger.Level = logrus.DebugLevel

	fakeExit := func(int) {
		panic("os.Exit called")
	}
	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	tests := []struct {
		name  string
		fct   func(format string, args ...interface{})
		level logrus.Level
		panic bool
	}{
		{"Debugf", Debugf, logrus.DebugLevel, false},
		{"Infof", Infof, logrus.InfoLevel, false},
		{"Warnf", Warnf, logrus.WarnLevel, false},
		{"Errorf", Errorf, logrus.ErrorLevel, false},
		{"Fatalf", Fatalf, logrus.FatalLevel, true}, // does `os.exit(1)`, monkey-patched to panic
		{"Panicf", Panicf, logrus.PanicLevel, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()
			if tt.panic {
				assert.Panics(t, func() { tt.fct("msg: %d", 2) })
			} else {
				tt.fct("msg: %d", 2)
			}
			assert.Len(t, hook.AllEntries(), 1)
			assert.Equal(t, tt.level, hook.LastEntry().Level)
			assert.Equal(t, "msg: 2", hook.LastEntry().Message)
		})
	}
}

func Test_WithField(t *testing.T) {
	hook, restoreFct := MockSharedLoggerHook()
	defer restoreFct()
	WithField("foo", "bar").Error("error msg")
	assert.Len(t, hook.AllEntries(), 1)
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(t, "error msg", hook.LastEntry().Message)
	assert.Equal(t, "bar", hook.LastEntry().Data["foo"])
}

func Test_WithFields(t *testing.T) {
	hook, restoreFct := MockSharedLoggerHook()
	defer restoreFct()
	WithFields(map[string]interface{}{"foo": "bar", "foo2": "bar2"}).Error("error msg")
	assert.Len(t, hook.AllEntries(), 1)
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(t, "error msg", hook.LastEntry().Message)
	assert.Equal(t, "bar", hook.LastEntry().Data["foo"])
	assert.Equal(t, "bar2", hook.LastEntry().Data["foo2"])
}
