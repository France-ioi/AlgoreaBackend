package database

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestDB_WhereSearchStringMatches(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	type args struct {
		field         string
		fallbackField string
		searchString  string
	}
	tests := []struct {
		name          string
		args          args
		wantCondition string
		wantArgs      []driver.Value
	}{
		{
			name: "words with accents",
			args: args{field: "title2", fallbackField: "title3", searchString: "précédente jusqu'à"},
			wantCondition: "((title2 IS NOT NULL AND MATCH(title2) AGAINST(? IN BOOLEAN MODE)) OR " +
				"(title2 IS NULL AND MATCH(title3) AGAINST(? IN BOOLEAN MODE)))",
			wantArgs: []driver.Value{"+précédente* +jusqu'à*", "+précédente* +jusqu'à*"},
		},
		{
			name: "no letters",
			args: args{field: "title", fallbackField: "title2", searchString: "```"},
			wantCondition: "((title IS NOT NULL AND MATCH(title) AGAINST(? IN BOOLEAN MODE)) OR " +
				"(title IS NULL AND MATCH(title2) AGAINST(? IN BOOLEAN MODE)))",
			wantArgs: []driver.Value{"", ""},
		},
		{
			name:          "no fallback field",
			args:          args{field: "title", searchString: "abc def"},
			wantCondition: "((title IS NOT NULL AND MATCH(title) AGAINST(? IN BOOLEAN MODE)))",
			wantArgs:      []driver.Value{"+abc* +def*"},
		},
		{
			name:          "filters out special characters",
			args:          args{field: "title", searchString: "~!@#$%^&*()_+`-=[]\\{}|;':\",./<>?"},
			wantCondition: "((title IS NOT NULL AND MATCH(title) AGAINST(? IN BOOLEAN MODE)))",
			wantArgs:      []driver.Value{"+'*"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, sqlMock := NewDBMock()
			defer func() { _ = db.Close() }()

			sqlMock.ExpectQuery("^" + regexp.QuoteMeta("SELECT * FROM `items` WHERE "+tt.wantCondition) + "$").
				WithArgs(tt.wantArgs...).
				WillReturnRows(sqlMock.NewRows([]string{"id"}))

			var result []map[string]interface{}
			require.NoError(t, db.Table("items").
				WhereSearchStringMatches(tt.args.field, tt.args.fallbackField, tt.args.searchString).
				ScanIntoSliceOfMaps(&result).Error())

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}
