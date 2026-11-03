//go:build !prod

package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/monkey"
)

var nowRegexp = regexp.MustCompile(`(?i)\bNOW\s*\(\s*(?:(\d+)\s*)?\)`)

// DBTimePatch holds the information needed to restore the original behavior of the database time mocking.
// It is returned by MockDBTime and should be passed to RestoreDBTime to restore the original behavior.
type DBTimePatch struct {
	oldNow         *gorm.SqlExpr
	patchedMethods []*monkey.PatchGuard
}

// MockDBTime replaces the DB NOW() function call with a given constant value in all the queries.
func MockDBTime(timeStrRaw string) *DBTimePatch {
	patch := &DBTimePatch{}

	parsedTime, err := time.Parse(time.DateTime+".999999999", timeStrRaw)
	if err != nil {
		panic(err)
	}
	nowReplacer := getNowReplacer(parsedTime)

	patchDatabaseDBMethods(nowReplacer, patch)
	patch.oldNow = database.MockNow(parsedTime.Truncate(time.Second).Format(time.DateTime))
	patchGormMethods(nowReplacer, patch)
	patchDBMethods(nowReplacer, patch)
	return patch
}

func getNowReplacer(parsedTime time.Time) func(string) string {
	return func(nowStr string) string {
		layout := time.DateTime
		subMatches := nowRegexp.FindStringSubmatch(nowStr)
		precision := time.Second
		if subMatches[1] != "" {
			var err error
			fsp, err := strconv.Atoi(subMatches[1])
			if err != nil {
				panic(err)
			}
			layout += fmt.Sprintf(".%0*d", fsp, 0)
			for range fsp {
				precision /= 10
			}
		}
		return fmt.Sprintf("%q", parsedTime.Truncate(precision).Format(layout))
	}
}

// mockey Origin is not safe under concurrent entry (shared extraArgsGetter written in the
// MakeFunc trampoline before the replacement runs). MockDBTime therefore uses Unpatch/Restore
// like bou.ke/monkey. Per-method mutexes serialize Unpatch → real call → Restore so two
// goroutines cannot rewrite the same prologue at once; nested calls across methods are fine
// because each method has its own mutex and patch guard.
func patchDBMethods(nowReplacer func(string) string, patch *DBTimePatch) {
	var prepareContextMu sync.Mutex
	var prepareContextGuard *monkey.PatchGuard
	prepareContextGuard = monkey.PatchInstanceMethod(
		reflect.TypeFor[*sql.DB](), "PrepareContext",
		func(db *sql.DB, ctx context.Context, query string) (*sql.Stmt, error) {
			prepareContextMu.Lock()
			defer prepareContextMu.Unlock()
			prepareContextGuard.Unpatch()
			defer prepareContextGuard.Restore()
			query = nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
			return db.PrepareContext(ctx, query) //nolint:sqlclosecheck // the caller is responsible for closing the statement
		})
	patch.patchedMethods = append(patch.patchedMethods, prepareContextGuard)

	var queryContextMu sync.Mutex
	var queryContextGuard *monkey.PatchGuard
	queryContextGuard = monkey.PatchInstanceMethod(
		reflect.TypeFor[*sql.DB](), "QueryContext",
		func(db *sql.DB, ctx context.Context, query string, args ...any) (*sql.Rows, error) {
			queryContextMu.Lock()
			defer queryContextMu.Unlock()
			queryContextGuard.Unpatch()
			defer queryContextGuard.Restore()
			query = nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
			return db.QueryContext(ctx, query, args...) //nolint:sqlclosecheck // the caller is responsible for closing the rows
		})
	patch.patchedMethods = append(patch.patchedMethods, queryContextGuard)
}

func patchGormMethods(nowReplacer func(string) string, patch *DBTimePatch) {
	var execMu sync.Mutex
	var execGuard *monkey.PatchGuard
	execGuard = monkey.PatchInstanceMethod(
		reflect.TypeFor[*gorm.DB](), "Exec",
		func(db *gorm.DB, query string, args ...any) *gorm.DB {
			execMu.Lock()
			defer execMu.Unlock()
			execGuard.Unpatch()
			defer execGuard.Restore()
			query = nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
			return db.Exec(query, args...)
		})
	patch.patchedMethods = append(patch.patchedMethods, execGuard)

	var rawMu sync.Mutex
	var rawGuard *monkey.PatchGuard
	rawGuard = monkey.PatchInstanceMethod(
		reflect.TypeFor[*gorm.DB](), "Raw",
		func(db *gorm.DB, query string, args ...any) *gorm.DB {
			rawMu.Lock()
			defer rawMu.Unlock()
			rawGuard.Unpatch()
			defer rawGuard.Restore()
			query = nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
			return db.Raw(query, args...)
		})
	patch.patchedMethods = append(patch.patchedMethods, rawGuard)
}

func patchDatabaseDBMethods(nowReplacer func(string) string, patch *DBTimePatch) {
	patchDatabaseDBMethodsWithIntQueryAndArgs(nowReplacer, patch)
	patchDatabaseDBMethodsWithStringQueryAndArgs(nowReplacer, patch)
	patchDatabaseDBMethodsWithStringQuery(nowReplacer, patch)

	var orderMu sync.Mutex
	var orderGuard *monkey.PatchGuard
	orderGuard = monkey.PatchInstanceMethod(
		reflect.TypeFor[*database.DB](), "Order",
		func(db *database.DB, value any, reorder ...bool) *database.DB {
			orderMu.Lock()
			defer orderMu.Unlock()
			orderGuard.Unpatch()
			defer orderGuard.Restore()
			if valueStr, ok := value.(string); ok {
				value = nowRegexp.ReplaceAllStringFunc(valueStr, nowReplacer)
			}
			return db.Order(value, reorder...)
		})
	patch.patchedMethods = append(patch.patchedMethods, orderGuard)

	var pluckMu sync.Mutex
	var pluckGuard *monkey.PatchGuard
	pluckGuard = monkey.PatchInstanceMethod(
		reflect.TypeFor[*database.DB](), "Pluck",
		func(db *database.DB, column string, values any) *database.DB {
			pluckMu.Lock()
			defer pluckMu.Unlock()
			pluckGuard.Unpatch()
			defer pluckGuard.Restore()
			column = nowRegexp.ReplaceAllStringFunc(column, nowReplacer)
			return db.Pluck(column, values)
		})
	patch.patchedMethods = append(patch.patchedMethods, pluckGuard)

	var takeMu sync.Mutex
	var takeGuard *monkey.PatchGuard
	takeGuard = monkey.PatchInstanceMethod(
		reflect.TypeFor[*database.DB](), "Take",
		func(db *database.DB, out any, where ...any) *database.DB {
			takeMu.Lock()
			defer takeMu.Unlock()
			takeGuard.Unpatch()
			defer takeGuard.Restore()
			if len(where) > 0 {
				if whereStr, ok := where[0].(string); ok {
					where[0] = nowRegexp.ReplaceAllStringFunc(whereStr, nowReplacer)
				}
			}
			return db.Take(out, where...)
		})
	patch.patchedMethods = append(patch.patchedMethods, takeGuard)

	var deleteMu sync.Mutex
	var deleteGuard *monkey.PatchGuard
	deleteGuard = monkey.PatchInstanceMethod(
		reflect.TypeFor[*database.DB](), "Delete",
		func(db *database.DB, where ...any) *database.DB {
			deleteMu.Lock()
			defer deleteMu.Unlock()
			deleteGuard.Unpatch()
			defer deleteGuard.Restore()
			if len(where) > 0 {
				if whereStr, ok := where[0].(string); ok {
					where[0] = nowRegexp.ReplaceAllStringFunc(whereStr, nowReplacer)
				}
			}
			return db.Delete(where...)
		})
	patch.patchedMethods = append(patch.patchedMethods, deleteGuard)
}

func patchDatabaseDBMethodsWithStringQuery(nowReplacer func(string) string, patch *DBTimePatch) {
	stringDBMethods := [...]string{
		"Table", "Group",
	}
	for _, methodName := range stringDBMethods {
		var methodMu sync.Mutex
		var guard *monkey.PatchGuard
		guard = monkey.PatchInstanceMethod(
			reflect.TypeFor[*database.DB](), methodName,
			func(db *database.DB, query string) *database.DB {
				methodMu.Lock()
				defer methodMu.Unlock()
				guard.Unpatch()
				defer guard.Restore()
				query = nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
				//nolint:forcetypeassert // both methods always return *database.DB
				return reflect.ValueOf(db).MethodByName(methodName).
					Call([]reflect.Value{reflect.ValueOf(query)})[0].
					Interface().(*database.DB)
			})
		patch.patchedMethods = append(patch.patchedMethods, guard)
	}
}

func patchDatabaseDBMethodsWithStringQueryAndArgs(nowReplacer func(string) string, patch *DBTimePatch) {
	stringAndArgsDBMethods := [...]string{
		"Joins", "Raw", "Exec",
	}
	for _, methodName := range stringAndArgsDBMethods {
		var methodMu sync.Mutex
		var guard *monkey.PatchGuard
		guard = monkey.PatchInstanceMethod(
			reflect.TypeFor[*database.DB](), methodName,
			func(db *database.DB, query string, args ...any) *database.DB {
				methodMu.Lock()
				defer methodMu.Unlock()
				guard.Unpatch()
				defer guard.Restore()
				query = nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
				//nolint:forcetypeassert // all the methods always return *database.DB
				return reflect.ValueOf(db).MethodByName(methodName).
					Call(constructReflArgsForQueryAndArgs(query, args))[0].
					Interface().(*database.DB)
			})
		patch.patchedMethods = append(patch.patchedMethods, guard)
	}
}

func constructReflArgsForQueryAndArgs(query any, args []any) []reflect.Value {
	reflArgs := make([]reflect.Value, 0, len(args)+1)
	reflArgs = append(reflArgs, reflect.ValueOf(query))
	// Untyped nil becomes the zero reflect.Value, which panics in Call; use a typed *string nil instead.
	nilReflValue := reflect.New(reflect.TypeFor[*string]()).Elem()
	for _, arg := range args {
		var reflValue reflect.Value
		if arg == nil {
			reflValue = nilReflValue
		} else {
			reflValue = reflect.ValueOf(arg)
		}
		reflArgs = append(reflArgs, reflValue)
	}
	return reflArgs
}

func patchDatabaseDBMethodsWithIntQueryAndArgs(nowReplacer func(string) string, patch *DBTimePatch) {
	standardDBMethods := [...]string{
		"Where", "Select", "Having",
	}
	for _, methodName := range standardDBMethods {
		var methodMu sync.Mutex
		var guard *monkey.PatchGuard
		guard = monkey.PatchInstanceMethod(
			reflect.TypeFor[*database.DB](), methodName,
			func(db *database.DB, query any, args ...any) *database.DB {
				methodMu.Lock()
				defer methodMu.Unlock()
				guard.Unpatch()
				defer guard.Restore()
				if queryStr, ok := query.(string); ok {
					query = nowRegexp.ReplaceAllStringFunc(queryStr, nowReplacer)
				}
				//nolint:forcetypeassert // all the methods always return *database.DB
				return reflect.ValueOf(db).MethodByName(methodName).
					Call(constructReflArgsForQueryAndArgs(query, args))[0].
					Interface().(*database.DB)
			})
		patch.patchedMethods = append(patch.patchedMethods, guard)
	}
}

// RestoreDBTime restores the usual behavior of the NOW() function.
func RestoreDBTime(patch *DBTimePatch) {
	database.RestoreNow(patch.oldNow)
	for i := len(patch.patchedMethods) - 1; i >= 0; i-- {
		patch.patchedMethods[i].Unpatch()
	}
	patch.patchedMethods = nil
}
