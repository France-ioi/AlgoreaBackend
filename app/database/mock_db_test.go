package database

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestNewDBMock_ExitsOnGettingErrorFromSQLMockNew(t *testing.T) {
	for _, test := range []struct {
		name string
		f    func()
	}{
		{"NewDBMock", func() { _, _ = NewDBMock() }},
		{"NewDBMockWithLogConfig", func() { _, _ = NewDBMockWithLogConfig(LogConfig{}, false) }},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			someError := errors.New("some error")

			var patch *monkey.PatchGuard
			patch = monkey.Patch(sqlmock.New, reflect.MakeFunc(reflect.TypeOf(sqlmock.New),
				func(args []reflect.Value) (results []reflect.Value) {
					patch.Unpatch()
					_, mock, _ := sqlmock.New()
					patch.Restore()
					return []reflect.Value{
						reflect.ValueOf((*sql.DB)(nil)),
						reflect.ValueOf(mock),
						reflect.ValueOf(someError),
					}
				}).Interface())
			defer monkey.UnpatchAll()

			assert.PanicsWithError(t, "unable to create the mock db: some error", func() {
				test.f()
			})
		})
	}
}

func TestNewDBMock_ExitsOnGettingErrorFromOpen(t *testing.T) {
	for _, test := range []struct {
		name string
		f    func()
	}{
		{"NewDBMock", func() { _, _ = NewDBMock() }},
		{"NewDBMockWithLogConfig", func() { _, _ = NewDBMockWithLogConfig(LogConfig{}, false) }},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			someError := errors.New("some error")

			monkey.Patch(OpenWithLogConfig, func(interface{}, LogConfig, bool) (*DB, error) { return nil, someError })
			defer monkey.UnpatchAll()

			assert.PanicsWithError(t, "unable to create the gorm connection to the mock: some error", func() {
				test.f()
			})
		})
	}
}
