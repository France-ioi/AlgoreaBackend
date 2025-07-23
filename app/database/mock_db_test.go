package database

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestNewDBMock_ExitsOnGettingErrorFromSQLMockNew(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	for _, test := range []struct {
		name string
		f    func()
	}{
		{"NewDBMock", func() { _, _ = NewDBMock() }},
		{"NewDBMockWithLogConfig", func() {
			_, _ = NewDBMockWithLogConfig(ctx, LogConfig{}, false)
		}},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			someError := errors.New("some error")

			var patch *monkey.PatchGuard
			patch = monkey.Patch(sqlmock.New, reflect.MakeFunc(reflect.TypeOf(sqlmock.New),
				func(_ []reflect.Value) (results []reflect.Value) {
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
	testoutput.SuppressIfPasses(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	for _, test := range []struct {
		name string
		f    func()
	}{
		{"NewDBMock", func() { _, _ = NewDBMock() }},
		{"NewDBMockWithLogConfig", func() {
			_, _ = NewDBMockWithLogConfig(ctx, LogConfig{}, false)
		}},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			someError := errors.New("some error")

			monkey.Patch(OpenWithLogConfig, func(context.Context, interface{}, LogConfig, bool) (*DB, error) { return nil, someError })
			defer monkey.UnpatchAll()

			assert.PanicsWithError(t, "unable to create the gorm connection to the mock: some error", func() {
				test.f()
			})
		})
	}
}
