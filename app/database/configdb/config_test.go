package configdb

import (
	"errors"
	"regexp"
	"testing"

	"github.com/spf13/viper"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
)

func Test_mustNotBeError(t *testing.T) {
	mustNotBeError(nil)

	expectedError := errors.New("some error")
	assertlib.PanicsWithValue(t, expectedError, func() {
		mustNotBeError(expectedError)
	})
}

//nolint:gocyclo
func TestCheckConfig(t *testing.T) { //nolint:gocognit Should be refactored.
	type relationSpec struct {
		database.ParentChild
		exists bool
		error  bool
	}
	type groupSpec struct {
		id     int64
		exists bool
		error  bool
	}
	type propagationSpec struct {
		exists bool
		error  bool
	}

	tests := []struct {
		name                     string
		config                   []domain.ConfigItem
		expectedGroupsToCheck    []groupSpec
		expectedRelationsToCheck []relationSpec
		propagationID1           *propagationSpec
		expectedError            error
	}{
		{
			name: "everything is okay",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1", "192.168.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
				{
					Domains:       []string{"www.france-ioi.org"},
					AllUsersGroup: 6, TempUsersGroup: 8,
				},
			},
			expectedGroupsToCheck: []groupSpec{
				{id: 2, exists: true},
				{id: 4, exists: true},
				{id: 6, exists: true},
				{id: 8, exists: true},
			},
			expectedRelationsToCheck: []relationSpec{
				{ParentChild: database.ParentChild{ParentID: 2, ChildID: 4}, exists: true},
				{ParentChild: database.ParentChild{ParentID: 6, ChildID: 8}, exists: true},
			},
			propagationID1: &propagationSpec{exists: true},
		},
		{
			name: "AllUsers is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"192.168.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			expectedGroupsToCheck: []groupSpec{
				{id: 2},
			},
			propagationID1: nil,
			expectedError:  errors.New("no AllUsers group for domain \"192.168.0.1\""),
		},
		{
			name: "TempUsers is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			expectedGroupsToCheck: []groupSpec{
				{id: 2, exists: true},
				{id: 4},
			},
			propagationID1: nil,
			expectedError:  errors.New("no TempUsers group for domain \"127.0.0.1\""),
		},
		{
			name: "AllUsers -> TempUsers relation is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			expectedGroupsToCheck: []groupSpec{
				{id: 2, exists: true},
				{id: 4, exists: true},
			},
			expectedRelationsToCheck: []relationSpec{
				{ParentChild: database.ParentChild{ParentID: 2, ChildID: 4}},
			},
			propagationID1: nil,
			expectedError:  errors.New("no AllUsers -> TempUsers link in groups_groups for domain \"127.0.0.1\""),
		},
		{
			name:           "propagation with id=1 is missing",
			propagationID1: &propagationSpec{exists: false},
			expectedError:  errors.New("missing entry in propagations table with id 1"),
		},
		{
			name: "error on group checking",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			expectedGroupsToCheck: []groupSpec{
				{id: 2, error: true},
			},
			propagationID1: nil,
			expectedError:  errors.New("some error"),
		},
		{
			name: "error on relation checking",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			expectedGroupsToCheck: []groupSpec{
				{id: 2, exists: true},
				{id: 4, exists: true},
			},
			expectedRelationsToCheck: []relationSpec{
				{ParentChild: database.ParentChild{ParentID: 2, ChildID: 4}, error: true},
			},
			propagationID1: nil,
			expectedError:  errors.New("some error"),
		},
		{
			name:           "error on propagation entry checking",
			propagationID1: &propagationSpec{error: true},
			expectedError:  errors.New("some error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock := database.NewDBMock()
			defer func() { _ = db.Close() }()
			mock.MatchExpectationsInOrder(false)

			var expectedError error

			for _, expectedGroupToCheck := range tt.expectedGroupsToCheck {
				queryMock := mock.ExpectQuery("^" + regexp.QuoteMeta(
					"SELECT 1 FROM `groups`  WHERE (groups.id = ?) LIMIT 1",
				) + "$").WithArgs(expectedGroupToCheck.id)
				if !expectedGroupToCheck.error {
					rowsToReturn := mock.NewRows([]string{"1"})
					if expectedGroupToCheck.exists {
						rowsToReturn.AddRow(1)
					}
					queryMock.WillReturnRows(rowsToReturn)
				} else {
					expectedError = errors.New("some error")
					queryMock.WillReturnError(expectedError)
				}
			}
			if expectedError == nil {
				for _, expectedRelationToCheck := range tt.expectedRelationsToCheck {
					rowsToReturn := mock.NewRows([]string{"1"})
					if expectedRelationToCheck.exists {
						rowsToReturn.AddRow(1)
					}
					queryMock := mock.ExpectQuery("^"+regexp.QuoteMeta(
						"SELECT 1 FROM `groups_groups_active` WHERE (parent_group_id = ?) AND (child_group_id = ?) LIMIT 1",
					)+"$").
						WithArgs(expectedRelationToCheck.ParentID, expectedRelationToCheck.ChildID)
					if !expectedRelationToCheck.error {
						queryMock.WillReturnRows(rowsToReturn)
					} else {
						expectedError = errors.New("some error")
						queryMock.WillReturnError(expectedError)
					}
				}
			}

			// Propagation.
			if tt.propagationID1 != nil {
				queryMock := mock.ExpectQuery("^" + regexp.QuoteMeta(
					"SELECT 1 FROM `propagations`  WHERE (propagation_id = 1) LIMIT 1",
				) + "$")
				if !tt.propagationID1.error {
					rowsToReturn := mock.NewRows([]string{"1"})
					if tt.propagationID1.exists {
						rowsToReturn.AddRow(1)
					}
					queryMock.WillReturnRows(rowsToReturn)
				} else {
					expectedError = errors.New("some error")
					queryMock.WillReturnError(expectedError)
				}
			}

			conf := viper.New()
			conf.Set("domains", tt.config)

			domainConfig, err := app.DomainsConfig(conf)
			assertlib.NoError(t, err)

			err = CheckConfig(database.NewDataStore(db), domainConfig)
			assertlib.Equal(t, tt.expectedError, err)
			assertlib.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
