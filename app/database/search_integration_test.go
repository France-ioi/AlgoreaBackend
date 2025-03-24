//go:build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestDB_WhereSearchStringMatches_Integration(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups:
			- {id: 111, name: 'précédente'}
			- {id: 222, name: 'jusqu''à'}
			- {id: 333, name: "précédente jusqu'à"}
			- {id: 444, name: 'éléphant'}
			- {id: 555, name: 'Task #1'}
			- {id: 666, name: 'Task #2'}
			- {id: 777, name: "with'quote"}
			- {id: 888, name: "with_underscore"}
		items_strings:
			- {item_id: 111, language_tag: fr, title: 'précédente'}
			- {item_id: 222, language_tag: fr, title: 'jusqu''à'}
			- {item_id: 333, language_tag: fr, title: "précédente jusqu'à"}
			- {item_id: 444, language_tag: fr, title: 'éléphant'}
			- {item_id: 555, language_tag: fr, title: 'Task #1'}
			- {item_id: 666, language_tag: fr, title: 'Task #2'}
			- {item_id: 777, language_tag: fr, title: "with'quote"}
			- {item_id: 888, language_tag: fr, title: "with_underscore"}
`)
	defer func() { _ = db.Close() }()

	for _, test := range []struct {
		searchString string
		expectedIDs  []int64
	}{
		{
			searchString: "précédente jusqu'à",
			expectedIDs:  []int64{333},
		},
		{
			searchString: "éléphant",
			expectedIDs:  []int64{444},
		},
		{
			searchString: "à",
			expectedIDs:  []int64{222, 333},
		},
		{
			searchString: "Task-1*",
			expectedIDs:  []int64{555},
		},
		{
			searchString: "Task",
			expectedIDs:  []int64{555, 666},
		},
		{
			searchString: "Task ",
			expectedIDs:  []int64{555, 666},
		},
		{
			searchString: "Task 1",
			expectedIDs:  []int64{555},
		},
		{
			searchString: "Task #1",
			expectedIDs:  []int64{555},
		},
		{
			searchString: "with'quote",
			expectedIDs:  []int64{777},
		},
		{
			searchString: "with_underscore",
			expectedIDs:  []int64{888},
		},
		{
			searchString: "underscore",
			expectedIDs:  []int64{},
		},
	} {
		test := test
		t.Run(test.searchString, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			for _, tableColumns := range []struct {
				table        string
				searchColumn string
				idColumn     string
			}{
				{
					table:        "groups",
					searchColumn: "name",
					idColumn:     "id",
				},
				{
					table:        "items_strings",
					searchColumn: "title",
					idColumn:     "item_id",
				},
			} {
				t.Run(tableColumns.table, func(t *testing.T) {
					testoutput.SuppressIfPasses(t)
					var ids []int64
					require.NoError(t, database.NewDataStore(db).Table(tableColumns.table).
						WhereSearchStringMatches(tableColumns.searchColumn, "", test.searchString).
						Pluck(tableColumns.idColumn, &ids).Error())
					assert.ElementsMatch(t, test.expectedIDs, ids)
				})
			}
		})
	}
}
