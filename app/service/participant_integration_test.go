//go:build !unit

package service_test

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestGetParticipantIDFromRequest(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 1, type: Class}, {id: 2, type: Team}, {id: 3, type: Team}, {id: 4, type: User}, {id: 5, type: User}]
		users: [{group_id: 4, login: john}, {group_id: 5, login: jane}]
		groups_groups:
			- {parent_group_id: 1, child_group_id: 4}
			- {parent_group_id: 2, child_group_id: 5}
			- {parent_group_id: 3, child_group_id: 4}
	`)
	defer func() { _ = db.Close() }()
	store := database.NewDataStore(db)
	require.NoError(t, store.InTransaction(func(trStore *database.DataStore) error {
		return trStore.GroupGroups().CreateNewAncestors()
	}))

	tests := []struct {
		name           string
		query          string
		expectedResult int64
		expectedError  error
	}{
		{
			name:          "no team",
			query:         "as_team_id=404",
			expectedError: service.ErrForbidden(errors.New("can't use given as_team_id as a user's team")),
		},
		{
			name:          "as_team_id is not a team",
			query:         "param&as_team_id=1",
			expectedError: service.ErrForbidden(errors.New("can't use given as_team_id as a user's team")),
		},
		{
			name:          "the current user is not a member of as_team_id",
			query:         "as_team_id=2",
			expectedError: service.ErrForbidden(errors.New("can't use given as_team_id as a user's team")),
		},
		{
			name:           "okay",
			query:          "param&as_team_id=3",
			expectedResult: 3,
		},
	}
	for _, test := range tests {
		testoutput.SuppressIfPasses(t)

		test := test
		participantID, err := service.GetParticipantIDFromRequest(
			&http.Request{URL: &url.URL{RawQuery: test.query}}, &database.User{GroupID: 4}, store)
		assert.Equal(t, test.expectedResult, participantID)
		assert.Equal(t, test.expectedError, err)
	}
}
