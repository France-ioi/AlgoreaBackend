package service

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"regexp"
	"testing"

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
	}{
		{name: "sorting (default rules)",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "sName", FieldType: "string"},
					"id":   {ColumnName: "ID", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
			wantSQL:          "SELECT ID FROM `users` ORDER BY sName DESC, ID ASC",
			wantSQLArguments: nil,
			wantAPIError:     NoError},
		{name: "sorting (request rules)",
			args: args{
				urlParameters: "?sort=name,-id",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "sName", FieldType: "string"},
					"id":   {ColumnName: "ID", FieldType: "int64"},
				},
			},
			wantSQL:          "SELECT ID FROM `users` ORDER BY sName ASC, ID DESC",
			wantSQLArguments: nil,
			wantAPIError:     NoError},
		{name: "repeated field",
			args: args{
				urlParameters: "?sort=name,name",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "sName", FieldType: "string"},
					"id":   {ColumnName: "ID", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
			wantSQL:          "",
			wantSQLArguments: nil,
			wantAPIError:     ErrInvalidRequest(errors.New(`a field cannot be a sorting parameter more than once: "name"`))},
		{name: "unknown field",
			args: args{
				urlParameters: "?sort=class",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "sName", FieldType: "string"},
					"id":   {ColumnName: "ID", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
			wantSQL:          "",
			wantSQLArguments: nil,
			wantAPIError:     ErrInvalidRequest(errors.New(`unknown field in sorting parameters: "class"`))},
		{name: "add id field",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "sName", FieldType: "string"},
				},
				defaultRules: "-name",
			},
			wantSQL:          "SELECT ID FROM `users` ORDER BY sName DESC, ID ASC",
			wantSQLArguments: nil,
			wantAPIError:     NoError},
		{name: "no rules (adds id)",
			args: args{
				urlParameters: "",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "sName", FieldType: "string"},
					"id":   {ColumnName: "ID", FieldType: "int64"},
				},
			},
			wantSQL:          "SELECT ID FROM `users` ORDER BY ID ASC",
			wantSQLArguments: nil,
			wantAPIError:     NoError},
		{name: "sorting + paging",
			args: args{
				urlParameters: "?from.id=1&from.name=Joe",
				acceptedFields: map[string]*FieldSortingParams{
					"name": {ColumnName: "sName", FieldType: "string"},
					"id":   {ColumnName: "ID", FieldType: "int64"},
				},
				defaultRules: "-name,id",
			},
			wantSQL:          "SELECT ID FROM `users` WHERE ((sName < ?) OR (sName = ? AND ID > ?)) ORDER BY sName DESC, ID ASC",
			wantSQLArguments: []driver.Value{"Joe", "Joe", 1},
			wantAPIError:     NoError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, dbMock := database.NewDBMock()
			defer func() { _ = db.Close() }()
			if tt.wantSQL != "" {
				dbMock.ExpectQuery(regexp.QuoteMeta(tt.wantSQL)).WithArgs(tt.wantSQLArguments...).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			}

			request, _ := http.NewRequest("GET", "/"+tt.args.urlParameters, nil)
			query := db.Table("users").Select("ID")
			query, gotAPIError := ApplySorting(request, query, tt.args.acceptedFields, tt.args.defaultRules)
			assert.Equal(t, tt.wantAPIError, gotAPIError)

			if gotAPIError == NoError {
				var result []struct{}
				query.Scan(&result)
			}

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}
