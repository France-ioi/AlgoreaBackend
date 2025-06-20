package service

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestApplySorting(t *testing.T) {
	type args struct {
		urlParameters              string
		sortingAndPagingParameters *SortingAndPagingParameters
	}
	tests := []struct {
		name             string
		args             args
		wantSQL          string
		wantSQLArguments []driver.Value
		wantError        error
		shouldPanic      error
	}{
		{
			name: "sorting (default rules)",
			args: args{
				urlParameters: "",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "-name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantSQL: "SELECT id FROM `users` ORDER BY name DESC, id ASC",
		},
		{
			name: "sorting (request rules)",
			args: args{
				urlParameters: "?sort=name,-id",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
					},
					TieBreakers: SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantSQL: "SELECT id FROM `users` ORDER BY name ASC, id DESC",
		},
		{
			name: "sorting (request rules are skipped)",
			args: args{
				urlParameters: "?sort=name,-id",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
					},
					TieBreakers:         SortingAndPagingTieBreakers{"id": FieldTypeInt64},
					IgnoreSortParameter: true,
					DefaultRules:        "id",
				},
			},
			wantSQL: "SELECT id FROM `users` ORDER BY id ASC",
		},
		{
			name: "repeated field",
			args: args{
				urlParameters: "?sort=name,name",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "-name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantError: ErrInvalidRequest(errors.New(`a field cannot be a sorting parameter more than once: "name"`)),
		},
		{
			name: "unallowed field",
			args: args{
				urlParameters: "?sort=class",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "-name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantError: ErrInvalidRequest(errors.New(`unallowed field in sorting parameters: "class"`)),
		},
		{
			name: "'null last' for a non-nullable field",
			args: args{
				urlParameters: "?sort=name$",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "-name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantError: ErrInvalidRequest(errors.New(`'null last' sorting cannot be applied to a non-nullable field: "name"`)),
		},
		{
			name: "applies default null sorting for nullable fields",
			args: args{
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name", Nullable: true},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "-name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantSQL: "SELECT id FROM `users` ORDER BY name IS NOT NULL, name DESC, id ASC",
		},
		{
			name: "'null last'",
			args: args{
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name", Nullable: true},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "id,-name$",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantSQL: "SELECT id FROM `users` ORDER BY id ASC, name IS NULL, name DESC",
		},
		{
			name: "add a tie-breaker field",
			args: args{
				urlParameters: "",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "-name",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantSQL:          "SELECT id FROM `users` ORDER BY name DESC, id ASC",
			wantSQLArguments: nil,
		},
		{
			name: "add multiple tie-breaker fields",
			args: args{
				urlParameters: "",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name":     {ColumnName: "name"},
						"group_id": {ColumnName: "group_id"},
						"item_id":  {ColumnName: "item_id"},
					},
					DefaultRules: "-name",
					TieBreakers:  SortingAndPagingTieBreakers{"group_id": FieldTypeInt64, "item_id": FieldTypeInt64},
				},
			},
			wantSQL:          "SELECT id FROM `users` ORDER BY name DESC, group_id ASC, item_id ASC",
			wantSQLArguments: nil,
		},
		{
			name: "add some of tie-breaker fields",
			args: args{
				urlParameters: "",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name":     {ColumnName: "name"},
						"group_id": {ColumnName: "group_id"},
						"item_id":  {ColumnName: "item_id"},
					},
					DefaultRules: "-name,item_id",
					TieBreakers:  SortingAndPagingTieBreakers{"group_id": FieldTypeInt64, "item_id": FieldTypeInt64},
				},
			},
			wantSQL:          "SELECT id FROM `users` ORDER BY name DESC, item_id ASC, group_id ASC",
			wantSQLArguments: nil,
		},
		{
			name: "sorting + paging",
			args: args{
				urlParameters: "?from.id=1&from.name=Joe&from.flag=1",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
						"flag": {ColumnName: "bFlag"},
					},
					DefaultRules: "-name,id,flag",
					TieBreakers: SortingAndPagingTieBreakers{
						"id":   FieldTypeInt64,
						"name": FieldTypeString,
						"flag": FieldTypeBool,
					},
				},
			},
			wantSQL: "SELECT id FROM `users` " +
				"WHERE ((name < ?) OR (name = ? AND id > ?) OR (name = ? AND id = ? AND bFlag > ?)) " +
				"ORDER BY name DESC, id ASC, bFlag ASC",
			wantSQLArguments: []driver.Value{"Joe", "Joe", 1, "Joe", 1, true},
		},
		{
			name: "sorting + paging with a nullable field",
			args: args{
				urlParameters: "?from.id=1",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name", Nullable: true},
						"id":   {ColumnName: "id"},
						"flag": {ColumnName: "bFlag"},
					},
					DefaultRules: "-name,id,flag",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantSQL: "SELECT id FROM `users` " +
				"JOIN (SELECT name AS name, id AS id, bFlag AS flag FROM `users` WHERE (id <=> ?) " +
				"  ORDER BY name IS NOT NULL, name DESC, id ASC, bFlag ASC LIMIT 1) AS from_page " +
				"WHERE ((IF(from_page.name IS NULL, name IS NOT NULL, name < from_page.name)) OR " +
				"  (name <=> from_page.name AND id > from_page.id) OR " +
				"  (name <=> from_page.name AND id <=> from_page.id AND bFlag > from_page.flag)) " +
				"ORDER BY name IS NOT NULL, name DESC, id ASC, bFlag ASC",
			wantSQLArguments: []driver.Value{1},
		},
		{
			name: "sorting + paging with a nullable field (null last)",
			args: args{
				urlParameters: "?from.id=1",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name", Nullable: true},
						"id":   {ColumnName: "id"},
						"flag": {ColumnName: "bFlag"},
					},
					DefaultRules: "-name$,id,flag",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantSQL: "SELECT id FROM `users` " +
				"JOIN (SELECT name AS name, id AS id, bFlag AS flag FROM `users` WHERE (id <=> ?) " +
				"  ORDER BY name IS NULL, name DESC, id ASC, bFlag ASC LIMIT 1) AS from_page " +
				"WHERE ((IF(from_page.name IS NULL, FALSE, name IS NULL OR name < from_page.name)) OR " +
				"  (name <=> from_page.name AND id > from_page.id) OR " +
				"  (name <=> from_page.name AND id <=> from_page.id AND bFlag > from_page.flag)) " +
				"ORDER BY name IS NULL, name DESC, id ASC, bFlag ASC",
			wantSQLArguments: []driver.Value{1},
		},
		{
			name: "sorting + paging with a nullable field (null last, nullable field is in the middle)",
			args: args{
				urlParameters: "?from.id=1",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name", Nullable: true},
						"id":   {ColumnName: "id"},
						"flag": {ColumnName: "bFlag"},
					},
					DefaultRules: "flag,-name$,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantSQL: "SELECT id FROM `users` " +
				"JOIN (SELECT bFlag AS flag, name AS name, id AS id FROM `users` WHERE (id <=> ?) " +
				"  ORDER BY bFlag ASC, name IS NULL, name DESC, id ASC LIMIT 1) AS from_page " +
				"WHERE ((bFlag > from_page.flag) OR " +
				"  (bFlag <=> from_page.flag AND IF(from_page.name IS NULL, FALSE, name IS NULL OR name < from_page.name)) OR " +
				"  (bFlag <=> from_page.flag AND name <=> from_page.name AND id > from_page.id)) " +
				"ORDER BY bFlag ASC, name IS NULL, name DESC, id ASC",
			wantSQLArguments: []driver.Value{1},
		},
		{
			name: "does not do paging when StartFromRowQuery = FromFirstRow",
			args: args{
				urlParameters: "?from.id=1",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name", Nullable: true},
						"id":   {ColumnName: "id"},
						"flag": {ColumnName: "bFlag"},
					},
					DefaultRules:         "flag,-name$,id",
					TieBreakers:          SortingAndPagingTieBreakers{"id": FieldTypeInt64},
					StartFromRowSubQuery: FromFirstRow,
				},
			},
			wantSQL: "SELECT id FROM `users` " +
				"ORDER BY bFlag ASC, name IS NULL, name DESC, id ASC",
		},
		{
			name: "uses a custom sub-query for the first row if StartFromRowQuery is given",
			args: args{
				urlParameters: "?from.id=1",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name", Nullable: true},
						"id":   {ColumnName: "id"},
						"flag": {ColumnName: "bFlag"},
					},
					DefaultRules:         "flag,-name$,id",
					TieBreakers:          SortingAndPagingTieBreakers{"id": FieldTypeInt64},
					StartFromRowSubQuery: gorm.Expr("(SELECT '1' AS name, 2 AS id, 3 AS flag)"),
				},
			},
			wantSQL: "SELECT id FROM `users` " +
				"JOIN (SELECT '1' AS name, 2 AS id, 3 AS flag) AS from_page " +
				"WHERE ((bFlag > from_page.flag) OR " +
				"  (bFlag <=> from_page.flag AND IF(from_page.name IS NULL, FALSE, name IS NULL OR name < from_page.name)) OR " +
				"  (bFlag <=> from_page.flag AND name <=> from_page.name AND id > from_page.id)) " +
				"ORDER BY bFlag ASC, name IS NULL, name DESC, id ASC",
		},
		{
			name: "wrong value in from.id field",
			args: args{
				urlParameters: "?from.id=abc",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "-name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			wantError: ErrInvalidRequest(errors.New(`wrong value for from.id (should be int64)`)),
		},
		{
			name: "one of the 'from.*' fields is skipped",
			args: args{
				urlParameters: "?from.id=2",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"type": {ColumnName: "type"},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "-name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64, "name": FieldTypeString},
				},
			},
			wantError: ErrInvalidRequest(errors.New(`all 'from' parameters (from.id, from.name) or none of them must be present`)),
		},
		{
			name: "unsupported field type",
			args: args{
				urlParameters: "?from.name=Joe&from.id=2",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "-name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": "interface{}"},
				},
			},
			shouldPanic: errors.New(`unsupported type "interface{}" for field "id"`),
		},
		{
			name: "unknown tie-breaker field",
			args: args{
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id"},
					},
					DefaultRules: "-name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"flag": FieldTypeInt64},
				},
			},
			shouldPanic: errors.New(`no such field "flag", cannot use it as a tie-breaker field`),
		},
		{
			name: "nullable field as a tie-breaker field",
			args: args{
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"name": {ColumnName: "name"},
						"id":   {ColumnName: "id", Nullable: true},
					},
					DefaultRules: "-name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64},
				},
			},
			shouldPanic: errors.New(`a nullable field "id" cannot be a tie-breaker field`),
		},
		{
			name: "unallowed from fields",
			args: args{
				urlParameters: "?from.field=Joe&from.version=2&from.name=Jane&from.id=1234",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"id":   {ColumnName: "id"},
						"name": {ColumnName: "name"},
					},
					DefaultRules: "name,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64, "name": FieldTypeString},
				},
			},
			wantError: ErrInvalidRequest(errors.New(`unallowed paging parameters (from.field, from.version)`)),
		},
		{
			name: "paging by time",
			args: args{
				urlParameters: "?from.submitted_at=" + url.QueryEscape("2006-01-02T15:04:05+03:00") + "&from.id=1",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"submitted_at": {ColumnName: "submitted_at"},
						"id":           {ColumnName: "id"},
					},
					DefaultRules: "submitted_at,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64, "submitted_at": FieldTypeTime},
				},
			},
			wantSQL: "SELECT id FROM `users`  WHERE ((submitted_at > ?) OR (submitted_at = ? AND id > ?)) " +
				"ORDER BY submitted_at ASC, id ASC",
			wantSQLArguments: []driver.Value{"2006-01-02 12:04:05", "2006-01-02 12:04:05", 1},
		},
		{
			name: "paging by time with milliseconds",
			args: args{
				urlParameters: "?from.submitted_at=" + url.QueryEscape("2006-01-02T15:04:05.001+03:00") + "&from.id=1",
				sortingAndPagingParameters: &SortingAndPagingParameters{
					Fields: map[string]*FieldSortingParams{
						"submitted_at": {ColumnName: "submitted_at"},
						"id":           {ColumnName: "id"},
					},
					DefaultRules: "submitted_at,id",
					TieBreakers:  SortingAndPagingTieBreakers{"id": FieldTypeInt64, "submitted_at": FieldTypeTime},
				},
			},
			wantSQL: "SELECT id FROM `users`  WHERE ((submitted_at > ?) OR (submitted_at = ? AND id > ?)) " +
				"ORDER BY submitted_at ASC, id ASC",
			wantSQLArguments: []driver.Value{"2006-01-02 12:04:05.001", "2006-01-02 12:04:05.001", 1},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

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

			request, _ := http.NewRequest("GET", "/"+tt.args.urlParameters, http.NoBody)
			query := db.Table("users").Select("id")

			query, gotError := ApplySortingAndPaging(request, query, tt.args.sortingAndPagingParameters)
			assert.Equal(t, tt.wantError, gotError)

			if gotError == nil {
				var result []struct{}
				query.Scan(&result)
			}

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}
