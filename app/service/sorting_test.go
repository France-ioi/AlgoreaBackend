package service

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestApplySorting(t *testing.T) {
	type args struct {
		urlParameters  string
		acceptedFields map[string]*FieldSortingParams
		defaultRules   string
	}
	tests := []struct {
		name             string
		args             args
		wantSQL          string
		wantSQLArguments []driver.Value
		wantAPIError     APIError
		shouldPanic      error
	}{
		{name: "sorting (default rules)",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
			wantSQL:      "SELECT id FROM `users` ORDER BY name DESC, id ASC",
			wantAPIError: NoError},
		{name: "sorting (request rules)",
			args: args{
				urlParameters: "?sort=name,-id",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
			},
			wantSQL:      "SELECT id FROM `users` ORDER BY name ASC, id DESC",
			wantAPIError: NoError},
		{name: "sorting (custom column name for ordering)",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", ColumnNameForOrdering: "LOWER(name)"},
					"id":   {ColumnName: "id", FieldType: "int64", ColumnNameForOrdering: "-id"},
				},
				defaultRules: "-name,id",
			},
			wantSQL:      "SELECT id FROM `users` ORDER BY LOWER(name) DESC, -id ASC",
			wantAPIError: NoError},
		{name: "repeated field",
			args: args{
				urlParameters: "?sort=name,name",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
			wantAPIError: ErrInvalidRequest(errors.New(`a field cannot be a sorting parameter more than once: "name"`))},
		{name: "unallowed field",
			args: args{
				urlParameters: "?sort=class",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
			wantAPIError: ErrInvalidRequest(errors.New(`unallowed field in sorting parameters: "class"`))},
		{name: "add id field",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
				},
				defaultRules: "-name",
			},
			wantSQL:          "SELECT id FROM `users` ORDER BY name DESC, id ASC",
			wantSQLArguments: nil,
			wantAPIError:     NoError},
		{name: "no rules (adds id)",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
			},
			wantSQL:      "SELECT id FROM `users` ORDER BY id ASC",
			wantAPIError: NoError},
		{name: "sorting + paging",
			args: args{
				urlParameters: "?from.id=1&from.name=Joe&from.flag=1",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
					"flag": {ColumnName: "bFlag", FieldType: "bool"},
				},
				defaultRules: "-name,id,flag",
			},
			wantSQL: "SELECT id FROM `users` " +
				"WHERE ((name < ?) OR (name = ? AND id > ?) OR (name = ? AND id = ? AND bFlag > ?)) " +
				"ORDER BY name DESC, id ASC, bFlag ASC",
			wantSQLArguments: []driver.Value{"Joe", "Joe", 1, "Joe", 1, true},
			wantAPIError:     NoError},
		{name: "wrong value in from.id field",
			args: args{
				urlParameters: "?from.id=abc&from.name=Joe",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
			wantAPIError: ErrInvalidRequest(errors.New(`wrong value for from.id (should be int64)`))},
		{name: "one of the from. fields is skipped",
			args: args{
				urlParameters: "?from.id=2",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"type": {ColumnName: "type", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
			wantAPIError: ErrInvalidRequest(errors.New(`all 'from' parameters (from.name, from.id) or none of them must be present`))},
		{name: "unsupported field type",
			args: args{
				urlParameters: "?from.name=Joe&from.id=2",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "interface{}"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
			shouldPanic: errors.New(`unsupported type "interface{}" for field "name"`)},
		{name: "unallowed from fields",
			args: args{
				urlParameters: "?from.field=Joe&from.version=2&from.name=Jane",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
				},
				defaultRules: "id",
			},
			wantAPIError: ErrInvalidRequest(errors.New(`unallowed paging parameters (from.field, from.name, from.version)`))},
		{name: "paging by time",
			args: args{
				urlParameters: "?from.submission_date=" + url.QueryEscape("2006-01-02T15:04:05+03:00") + "&from.id=1",
				acceptedFields: map[string]*FieldSortingParams{
					"submission_date": {ColumnName: "submission_date", FieldType: "time"},
				},
				defaultRules: "submission_date",
			},
			wantSQL: "SELECT id FROM `users`  WHERE ((submission_date > ?) OR (submission_date = ? AND id > ?)) " +
				"ORDER BY submission_date ASC, id ASC",
			wantSQLArguments: []driver.Value{
				sqlMockTime{time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("MSK", 3*3600))},
				sqlMockTime{time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("MSK", 3*3600))},
				1},
			wantAPIError: NoError},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					if tt.shouldPanic == nil {
						assert.Fail(t, "unexpected panic() was called with value %+v", p)
					} else {
						assert.Equal(t, tt.shouldPanic, p, "panic() value mismatched")
					}
				} else if tt.shouldPanic != nil {
					assert.Fail(t, "expected the test to panic(), but it did not")
				}
			}()
			db, dbMock := database.NewDBMock()
			defer func() { _ = db.Close() }()
			if tt.wantSQL != "" {
				dbMock.ExpectQuery("^" + regexp.QuoteMeta(tt.wantSQL) + "$").WithArgs(tt.wantSQLArguments...).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			}

			request, _ := http.NewRequest("GET", "/"+tt.args.urlParameters, nil)
			query := db.Table("users").Select("id")

			query, gotAPIError := ApplySortingAndPaging(request, query, tt.args.acceptedFields, tt.args.defaultRules)
			assert.Equal(t, tt.wantAPIError, gotAPIError)

			if gotAPIError == NoError {
				var result []struct{}
				query.Scan(&result)
			}

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

type sqlMockTime struct {
	time.Time
}

// Match satisfies sqlmock.Argument interface
func (a sqlMockTime) Match(v driver.Value) bool {
	var secondValue time.Time
	switch value := v.(type) {
	case sqlMockTime:
		secondValue = value.Time
	case time.Time:
		secondValue = value
	default:
		return false
	}
	return a.Time.Equal(secondValue)
}
