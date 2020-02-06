package database

import (
	"database/sql"
	"errors"
	"os"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNewDBMock_ExitsOnGettingErrorFromSQLMockNew(t *testing.T) {
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

	var exitCode int
	var exitCalled bool
	monkey.Patch(os.Exit, func(code int) { exitCalled = true; exitCode = code; panic(someError) })
	defer monkey.UnpatchAll()

	assert.PanicsWithValue(t, someError, func() {
		_, _ = NewDBMock()
	})
	assert.True(t, exitCalled)
	assert.Equal(t, 1, exitCode)
}

func TestNewDBMock_ExitsOnGettingErrorFromOpen(t *testing.T) {
	someError := errors.New("some error")

	monkey.Patch(Open, func(interface{}) (*DB, error) { return nil, someError })
	var exitCode int
	var exitCalled bool
	monkey.Patch(os.Exit, func(code int) { exitCalled = true; exitCode = code; panic(someError) })
	defer monkey.UnpatchAll()

	assert.PanicsWithValue(t, someError, func() {
		_, _ = NewDBMock()
	})
	assert.True(t, exitCalled)
	assert.Equal(t, 1, exitCode)
}
