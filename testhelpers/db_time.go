//go:build !prod

package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"bou.ke/monkey"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
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
			for i := 0; i < fsp; i++ {
				precision /= 10
			}
		}
		return fmt.Sprintf("%q", parsedTime.Truncate(precision).Format(layout))
	}
}

func patchDBMethods(nowReplacer func(string) string, patch *DBTimePatch) {
	var prepareContextGuard, queryContextGuard *monkey.PatchGuard
	prepareContextGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&sql.DB{}), "PrepareContext",
		func(db *sql.DB, c context.Context, query string) (*sql.Stmt, error) {
			prepareContextGuard.Unpatch()
			defer prepareContextGuard.Restore()
			nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
			return db.PrepareContext(c, query) //nolint:sqlclosecheck // the caller is responsible for closing the statement
		})
	patch.patchedMethods = append(patch.patchedMethods, prepareContextGuard)
	queryContextGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&sql.DB{}), "QueryContext",
		func(db *sql.DB, c context.Context, query string, args ...interface{}) (*sql.Rows, error) {
			queryContextGuard.Unpatch()
			defer queryContextGuard.Restore()
			query = nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
			return db.QueryContext(c, query, args...) //nolint:sqlclosecheck // the caller is responsible for closing the rows
		})
	patch.patchedMethods = append(patch.patchedMethods, queryContextGuard)
}

func patchGormMethods(nowReplacer func(string) string, patch *DBTimePatch) {
	var execGuard, rawGuard *monkey.PatchGuard
	execGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&gorm.DB{}), "Exec",
		func(db *gorm.DB, query string, args ...interface{}) *gorm.DB {
			execGuard.Unpatch()
			defer execGuard.Restore()
			query = nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
			return db.Exec(query, args...)
		})
	patch.patchedMethods = append(patch.patchedMethods, execGuard)
	rawGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&gorm.DB{}), "Raw",
		func(db *gorm.DB, query string, args ...interface{}) *gorm.DB {
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
	var orderGuard *monkey.PatchGuard
	orderGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&database.DB{}), "Order",
		func(db *database.DB, value interface{}, reorder ...bool) *database.DB {
			orderGuard.Unpatch()
			defer orderGuard.Restore()
			if valueStr, ok := value.(string); ok {
				value = nowRegexp.ReplaceAllStringFunc(valueStr, nowReplacer)
			}
			return db.Order(value, reorder...)
		})
	patch.patchedMethods = append(patch.patchedMethods, orderGuard)

	var pluckGuard *monkey.PatchGuard
	pluckGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&database.DB{}), "Pluck",
		func(db *database.DB, column string, values interface{}) *database.DB {
			pluckGuard.Unpatch()
			defer pluckGuard.Restore()
			column = nowRegexp.ReplaceAllStringFunc(column, nowReplacer)
			return db.Pluck(column, values)
		})
	patch.patchedMethods = append(patch.patchedMethods, pluckGuard)

	var takeGuard *monkey.PatchGuard
	takeGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&database.DB{}), "Take",
		func(db *database.DB, out interface{}, where ...interface{}) *database.DB {
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

	var deleteGuard *monkey.PatchGuard
	deleteGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&database.DB{}), "Delete",
		func(db *database.DB, where ...interface{}) *database.DB {
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
	stringDBGuards := make(map[string]*monkey.PatchGuard, len(stringDBMethods))
	for _, methodName := range stringDBMethods {
		methodName := methodName
		stringDBGuards[methodName] = monkey.PatchInstanceMethod(
			reflect.TypeOf(&database.DB{}), methodName,
			func(db *database.DB, query string) *database.DB {
				stringDBGuards[methodName].Unpatch()
				defer stringDBGuards[methodName].Restore()
				query = nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
				reflMethod := reflect.ValueOf(db).MethodByName(methodName)
				reflArgs := []reflect.Value{reflect.ValueOf(query)}

				return reflMethod.Call(reflArgs)[0].Interface().(*database.DB)
			})
		patch.patchedMethods = append(patch.patchedMethods, stringDBGuards[methodName])
	}
}

func patchDatabaseDBMethodsWithStringQueryAndArgs(nowReplacer func(string) string, patch *DBTimePatch) {
	stringAndArgsDBMethods := [...]string{
		"Joins", "Raw", "Exec",
	}
	stringAndArgsDBGuards := make(map[string]*monkey.PatchGuard, len(stringAndArgsDBMethods))
	for _, methodName := range stringAndArgsDBMethods {
		methodName := methodName
		stringAndArgsDBGuards[methodName] = monkey.PatchInstanceMethod(
			reflect.TypeOf(&database.DB{}), methodName,
			func(db *database.DB, query string, args ...interface{}) *database.DB {
				stringAndArgsDBGuards[methodName].Unpatch()
				defer stringAndArgsDBGuards[methodName].Restore()
				query = nowRegexp.ReplaceAllStringFunc(query, nowReplacer)
				reflMethod := reflect.ValueOf(db).MethodByName(methodName)
				reflArgs := constructReflArgsForQueryAndArgs(query, args)

				return reflMethod.Call(reflArgs)[0].Interface().(*database.DB)
			})
		patch.patchedMethods = append(patch.patchedMethods, stringAndArgsDBGuards[methodName])
	}
}

func constructReflArgsForQueryAndArgs(query interface{}, args []interface{}) []reflect.Value {
	reflArgs := make([]reflect.Value, 0, len(args))
	reflArgs = append(reflArgs, reflect.ValueOf(query))
	nilReflValue := reflect.New(reflect.TypeOf((*string)(nil))).Elem()
	for _, arg := range args {
		arg := arg
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
	standardDBGuards := make(map[string]*monkey.PatchGuard, len(standardDBMethods))
	for _, methodName := range standardDBMethods {
		methodName := methodName
		standardDBGuards[methodName] = monkey.PatchInstanceMethod(
			reflect.TypeOf(&database.DB{}), methodName,
			func(db *database.DB, query interface{}, args ...interface{}) *database.DB {
				standardDBGuards[methodName].Unpatch()
				defer standardDBGuards[methodName].Restore()
				if queryStr, ok := query.(string); ok {
					query = nowRegexp.ReplaceAllStringFunc(queryStr, nowReplacer)
				}
				reflMethod := reflect.ValueOf(db).MethodByName(methodName)
				reflArgs := constructReflArgsForQueryAndArgs(query, args)

				return reflMethod.Call(reflArgs)[0].Interface().(*database.DB)
			})
		patch.patchedMethods = append(patch.patchedMethods, standardDBGuards[methodName])
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
