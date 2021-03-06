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
		urlParameters        string
		acceptedFields       map[string]*FieldSortingParams
		defaultRules         string
		tieBreakerFieldNames []string
		skipSortParameters   bool
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
				defaultRules:         "-name,id",
				tieBreakerFieldNames: []string{"id"},
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
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL:      "SELECT id FROM `users` ORDER BY name ASC, id DESC",
			wantAPIError: NoError},
		{name: "sorting (request rules are skipped)",
			args: args{
				urlParameters: "?sort=name,-id",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				tieBreakerFieldNames: []string{"id"},
				skipSortParameters:   true,
				defaultRules:         "id",
			},
			wantSQL:      "SELECT id FROM `users` ORDER BY id ASC",
			wantAPIError: NoError},
		{name: "sorting (custom column name for ordering)",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", ColumnNameForOrdering: "LOWER(name)"},
					"id":   {ColumnName: "id", FieldType: "int64", ColumnNameForOrdering: "-id"},
				},
				tieBreakerFieldNames: []string{"id"},
				defaultRules:         "-name,id",
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
				defaultRules:         "-name,id",
				tieBreakerFieldNames: []string{"id"},
			},
			wantAPIError: ErrInvalidRequest(errors.New(`a field cannot be a sorting parameter more than once: "name"`))},
		{name: "unallowed field",
			args: args{
				urlParameters: "?sort=class",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules:         "-name,id",
				tieBreakerFieldNames: []string{"id"},
			},
			wantAPIError: ErrInvalidRequest(errors.New(`unallowed field in sorting parameters: "class"`))},
		{name: "allows ignored field in sort",
			args: args{
				urlParameters: "?sort=name",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {Ignore: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
		},
		{name: "'null last' for a non-nullable field",
			args: args{
				urlParameters: "?sort=name$",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules:         "-name,id",
				tieBreakerFieldNames: []string{"id"},
			},
			wantAPIError: ErrInvalidRequest(errors.New(`'null last' sorting cannot be applied to a non-nullable field: "name"`))},
		{name: "applies default null sorting for nullable fields",
			args: args{
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", Nullable: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules:         "-name,id",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL:      "SELECT id FROM `users` ORDER BY name IS NOT NULL, name DESC, id ASC",
			wantAPIError: NoError},
		{name: "'null last'",
			args: args{
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", Nullable: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules:         "id,-name$",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL:      "SELECT id FROM `users` ORDER BY id ASC, name IS NULL, name DESC",
			wantAPIError: NoError},
		{name: "add a tie-breaker field",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules:         "-name",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL:          "SELECT id FROM `users` ORDER BY name DESC, id ASC",
			wantSQLArguments: nil,
			wantAPIError:     NoError},
		{name: "add multiple tie-breaker fields",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name":     {ColumnName: "name", FieldType: "string"},
					"group_id": {ColumnName: "group_id", FieldType: "int64"},
					"item_id":  {ColumnName: "item_id", FieldType: "int64"},
				},
				defaultRules:         "-name",
				tieBreakerFieldNames: []string{"group_id", "item_id"},
			},
			wantSQL:          "SELECT id FROM `users` ORDER BY name DESC, group_id ASC, item_id ASC",
			wantSQLArguments: nil,
			wantAPIError:     NoError},
		{name: "add some of tie-breaker fields",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name":     {ColumnName: "name", FieldType: "string"},
					"group_id": {ColumnName: "group_id", FieldType: "int64"},
					"item_id":  {ColumnName: "item_id", FieldType: "int64"},
				},
				defaultRules:         "-name,item_id",
				tieBreakerFieldNames: []string{"group_id", "item_id"},
			},
			wantSQL:          "SELECT id FROM `users` ORDER BY name DESC, item_id ASC, group_id ASC",
			wantSQLArguments: nil,
			wantAPIError:     NoError},
		{name: "do not add a tie-breaker field if a sorting field is unique",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", Unique: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules:         "-name",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL:          "SELECT id FROM `users` ORDER BY name DESC",
			wantSQLArguments: nil,
			wantAPIError:     NoError},
		{name: "sorting + paging",
			args: args{
				urlParameters: "?from.id=1&from.name=Joe&from.flag=1",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
					"flag": {ColumnName: "bFlag", FieldType: "bool"},
				},
				defaultRules:         "-name,id,flag",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL: "SELECT id FROM `users` " +
				"WHERE ((name < ?) OR (name = ? AND id > ?) OR (name = ? AND id = ? AND bFlag > ?)) " +
				"ORDER BY name DESC, id ASC, bFlag ASC",
			wantSQLArguments: []driver.Value{"Joe", "Joe", 1, "Joe", 1, true},
			wantAPIError:     NoError},
		{name: "sorting + paging (ignored fields are skipped)",
			args: args{
				urlParameters: "?from.id=1&from.name=Joe&from.flag=1",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", Ignore: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
					"flag": {ColumnName: "bFlag", FieldType: "bool", Ignore: true},
				},
				defaultRules:         "-name,id,flag",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL:          "SELECT id FROM `users` WHERE ((id > ?)) ORDER BY id ASC",
			wantSQLArguments: []driver.Value{1},
			wantAPIError:     NoError},
		{name: "sorting + paging by a nullable field",
			args: args{
				urlParameters: "?from.id=1&from.name=Joe&from.flag=1",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", Nullable: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
					"flag": {ColumnName: "bFlag", FieldType: "bool"},
				},
				defaultRules:         "-name,id,flag",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL: "SELECT id FROM `users` " +
				"WHERE ((name < ?) OR (name = ? AND id > ?) OR (name = ? AND id = ? AND bFlag > ?)) " +
				"ORDER BY name IS NOT NULL, name DESC, id ASC, bFlag ASC",
			wantSQLArguments: []driver.Value{"Joe", "Joe", 1, "Joe", 1, true},
			wantAPIError:     NoError},
		{name: "sorting + paging by a nullable field (null last)",
			args: args{
				urlParameters: "?from.id=1&from.name=Joe&from.flag=1",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", Nullable: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
					"flag": {ColumnName: "bFlag", FieldType: "bool"},
				},
				defaultRules:         "-name$,id,flag",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL: "SELECT id FROM `users` " +
				"WHERE (((name < ? OR name IS NULL)) OR (name = ? AND id > ?) OR (name = ? AND id = ? AND bFlag > ?)) " +
				"ORDER BY name IS NULL, name DESC, id ASC, bFlag ASC",
			wantSQLArguments: []driver.Value{"Joe", "Joe", 1, "Joe", 1, true},
			wantAPIError:     NoError},
		{name: "sorting + paging by a nullable field (null last, nullable field is in the middle)",
			args: args{
				urlParameters: "?from.id=1&from.name=Joe&from.flag=1",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", Nullable: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
					"flag": {ColumnName: "bFlag", FieldType: "bool"},
				},
				defaultRules:         "flag,-name$,id",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL: "SELECT id FROM `users` " +
				"WHERE ((bFlag > ?) OR (bFlag = ? AND (name < ? OR name IS NULL)) OR (bFlag = ? AND name = ? AND id > ?)) " +
				"ORDER BY bFlag ASC, name IS NULL, name DESC, id ASC",
			wantSQLArguments: []driver.Value{true, true, "Joe", true, "Joe", 1},
			wantAPIError:     NoError},
		{name: "sorting + paging by a nullable field (from value is null, null first)",
			args: args{
				urlParameters: "?from.id=1&from.name=[NULL]&from.flag=1",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", Nullable: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
					"flag": {ColumnName: "bFlag", FieldType: "bool"},
				},
				defaultRules:         "-name,id,flag",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL: "SELECT id FROM `users` " +
				"WHERE ((name IS NOT NULL) OR (name IS NULL AND id > ?) OR (name IS NULL AND id = ? AND bFlag > ?)) " +
				"ORDER BY name IS NOT NULL, name DESC, id ASC, bFlag ASC",
			wantSQLArguments: []driver.Value{1, 1, true},
			wantAPIError:     NoError},
		{name: "sorting + paging by a nullable field (from value is null, null last)",
			args: args{
				urlParameters: "?from.id=1&from.name=[NULL]&from.flag=1",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", Nullable: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
					"flag": {ColumnName: "bFlag", FieldType: "bool"},
				},
				defaultRules:         "-name$,id,flag",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL: "SELECT id FROM `users` " +
				"WHERE ((name IS NULL AND id > ?) OR (name IS NULL AND id = ? AND bFlag > ?)) " +
				"ORDER BY name IS NULL, name DESC, id ASC, bFlag ASC",
			wantSQLArguments: []driver.Value{1, 1, true},
			wantAPIError:     NoError},
		{name: "sorting + paging by a nullable field (from value is null, null last, nullable field is the last field)",
			args: args{
				urlParameters: "?from.id=1&from.name=[NULL]&from.flag=1",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string", Nullable: true},
					"id":   {ColumnName: "id", FieldType: "int64"},
					"flag": {ColumnName: "bFlag", FieldType: "bool"},
				},
				defaultRules:         "id,flag,name$",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL: "SELECT id FROM `users` " +
				"WHERE ((id > ?) OR (id = ? AND bFlag > ?) OR (id = ? AND bFlag = ? AND name IS NULL)) " +
				"ORDER BY id ASC, bFlag ASC, name IS NULL, name ASC",
			wantSQLArguments: []driver.Value{1, 1, true},
			wantAPIError:     NoError},
		{name: "wrong value in from.id field",
			args: args{
				urlParameters: "?from.id=abc&from.name=Joe",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "string"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules:         "-name,id",
				tieBreakerFieldNames: []string{"id"},
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
				defaultRules:         "-name,id",
				tieBreakerFieldNames: []string{"id"},
			},
			wantAPIError: ErrInvalidRequest(errors.New(`all 'from' parameters (from.name, from.id) or none of them must be present`))},
		{name: "unsupported field type",
			args: args{
				urlParameters: "?from.name=Joe&from.id=2",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "name", FieldType: "interface{}"},
					"id":   {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules:         "-name,id",
				tieBreakerFieldNames: []string{"id"},
			},
			shouldPanic: errors.New(`unsupported type "interface{}" for field "name"`)},
		{name: "unallowed from fields",
			args: args{
				urlParameters: "?from.field=Joe&from.version=2&from.name=Jane&from.id=1234",
				acceptedFields: map[string]*FieldSortingParams{
					"id":   {ColumnName: "id", FieldType: "int64"},
					"name": {ColumnName: "name", FieldType: "string", Ignore: true},
				},
				defaultRules:         "name,id",
				tieBreakerFieldNames: []string{"id"},
			},
			wantAPIError: ErrInvalidRequest(errors.New(`unallowed paging parameters (from.field, from.version)`))},
		{name: "paging by time",
			args: args{
				urlParameters: "?from.submitted_at=" + url.QueryEscape("2006-01-02T15:04:05+03:00") + "&from.id=1",
				acceptedFields: map[string]*FieldSortingParams{
					"submitted_at": {ColumnName: "submitted_at", FieldType: "time"},
					"id":           {ColumnName: "id", FieldType: "int64"},
				},
				defaultRules:         "submitted_at,id",
				tieBreakerFieldNames: []string{"id"},
			},
			wantSQL: "SELECT id FROM `users`  WHERE ((submitted_at > ?) OR (submitted_at = ? AND id > ?)) " +
				"ORDER BY submitted_at ASC, id ASC",
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

			query, gotAPIError := ApplySortingAndPaging(request, query, tt.args.acceptedFields, tt.args.defaultRules,
				tt.args.tieBreakerFieldNames, tt.args.skipSortParameters)
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
