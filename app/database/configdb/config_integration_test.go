package configdb

import (
	"errors"
	"testing"

	"github.com/spf13/viper"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestCheckConfig_Integration(t *testing.T) {
	tests := []struct {
		name          string
		config        []domain.ConfigItem
		fixtures      string
		expectedError error
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
			fixtures: `
groups:
	- {id: 2}
	- {id: 4}
	- {id: 6}
	- {id: 8}
groups_groups:
	- {parent_group_id: 2, child_group_id: 4}
	- {parent_group_id: 6, child_group_id: 8}
propagations:
	- {propagation_id: 1}
			`,
		},
		{
			name: "AllUsers is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"192.168.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			fixtures: `
groups:
	- {id: 4}
propagations:
	- {propagation_id: 1}
			`,
			expectedError: errors.New("no AllUsers group for domain \"192.168.0.1\""),
		},
		{
			name: "TempUsers is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			fixtures: `
groups:
	- {id: 2}
propagations:
	- {propagation_id: 1}
			`,
			expectedError: errors.New("no TempUsers group for domain \"127.0.0.1\""),
		},
		{
			name: "AllUsers -> TempUsers relation is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			fixtures: `
groups:
	- {id: 2}
	- {id: 4}
propagations:
	- {propagation_id: 1}
			`,
			expectedError: errors.New("no AllUsers -> TempUsers link in groups_groups for domain \"127.0.0.1\""),
		},
		{
			name:          "propagations entry with id=1 missing",
			fixtures:      "",
			expectedError: errors.New("missing entry in propagations table with id 1"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(tt.fixtures)
			defer func() { _ = db.Close() }()

			conf := viper.New()
			conf.Set("domains", tt.config)

			domainConfig, err := app.DomainsConfig(conf)
			assertlib.NoError(t, err)

			err = CheckConfig(database.NewDataStore(db), domainConfig)

			assertlib.Equal(t, tt.expectedError, err)
		})
	}
}

func TestCreateMissingData_Integration(t *testing.T) {
	tests := []struct {
		name                  string
		config                []domain.ConfigItem
		fixtures              string
		checkConfigPassBefore bool
	}{
		{
			name: "nothing to insert, everything already exists",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			fixtures: `
groups:
	- {id: 2, type: "Base", name: "AllUsers", text_id: "AllUsers"}
	- {id: 4, type: "Base", name: "TempUsers", text_id: "TempUsers"}
groups_groups:
	- {parent_group_id: 2, child_group_id: 4}
propagations:
	- {propagation_id: 1}
			`,
			checkConfigPassBefore: true,
		},
		{
			name: "create all rows, nothing exists yet",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			checkConfigPassBefore: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(tt.fixtures)
			defer func() { _ = db.Close() }()

			conf := viper.New()
			conf.Set("domains", tt.config)

			domainConfig, err := app.DomainsConfig(conf)
			assertlib.NoError(t, err)

			// Verify that CheckConfig() passes or fails as expected before we call CreateMissingData().
			errCheckConfigBefore := CheckConfig(database.NewDataStore(db), domainConfig)
			if tt.checkConfigPassBefore {
				assertlib.NoError(t, errCheckConfigBefore)
			} else {
				assertlib.Error(t, errCheckConfigBefore)
			}

			err = CreateMissingData(database.NewDataStore(db), domainConfig)
			assertlib.NoError(t, err)

			// Once CreateMissingData() have been called, CheckConfig() should pass.
			err = CheckConfig(database.NewDataStore(db), domainConfig)
			assertlib.NoError(t, err)
		})
	}
}
