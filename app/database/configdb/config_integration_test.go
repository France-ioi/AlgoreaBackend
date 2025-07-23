package configdb

import (
	"errors"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestCheckConfig_Integration(t *testing.T) {
	testoutput.SuppressIfPasses(t)

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
					AllUsersGroup: 2, NonTempUsersGroup: 3, TempUsersGroup: 4,
				},
				{
					Domains:       []string{"www.france-ioi.org"},
					AllUsersGroup: 6, NonTempUsersGroup: 7, TempUsersGroup: 8,
				},
			},
			fixtures: `
groups:
	- {id: 2}
	- {id: 3}
	- {id: 4}
	- {id: 6}
	- {id: 7}
	- {id: 8}
groups_groups:
	- {parent_group_id: 2, child_group_id: 3}
	- {parent_group_id: 2, child_group_id: 4}
	- {parent_group_id: 6, child_group_id: 7}
	- {parent_group_id: 6, child_group_id: 8}
			`,
		},
		{
			name: "AllUsers is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"192.168.0.1"},
					AllUsersGroup: 2, NonTempUsersGroup: 3, TempUsersGroup: 4,
				},
			},
			fixtures: `
groups:
	- {id: 3}
	- {id: 4}
			`,
			expectedError: errors.New("no AllUsers group for domain \"192.168.0.1\""),
		},
		{
			name: "NonTempUsers is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, NonTempUsersGroup: 3, TempUsersGroup: 4,
				},
			},
			fixtures: `
groups:
	- {id: 2}
	- {id: 4}
			`,
			expectedError: errors.New("no NonTempUsers group for domain \"127.0.0.1\""),
		},
		{
			name: "TempUsers is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, NonTempUsersGroup: 3, TempUsersGroup: 4,
				},
			},
			fixtures: `
groups:
	- {id: 2}
	- {id: 3}
			`,
			expectedError: errors.New("no TempUsers group for domain \"127.0.0.1\""),
		},
		{
			name: "AllUsers -> NonTempUsers relation is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, NonTempUsersGroup: 3, TempUsersGroup: 4,
				},
			},
			fixtures: `
groups:
	- {id: 2}
	- {id: 3}
	- {id: 4}
groups_groups:
	- {parent_group_id: 2, child_group_id: 4}
			`,
			expectedError: errors.New("no AllUsers -> NonTempUsers link in groups_groups for domain \"127.0.0.1\""),
		},
		{
			name: "AllUsers -> TempUsers relation is missing",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, NonTempUsersGroup: 3, TempUsersGroup: 4,
				},
			},
			fixtures: `
groups:
	- {id: 2}
	- {id: 3}
	- {id: 4}
groups_groups:
	- {parent_group_id: 2, child_group_id: 3}
			`,
			expectedError: errors.New("no AllUsers -> TempUsers link in groups_groups for domain \"127.0.0.1\""),
		},
	}

	ctx := testhelpers.CreateTestContext()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(ctx, tt.fixtures)
			defer func() { _ = db.Close() }()

			conf := viper.New()
			conf.Set("domains", tt.config)

			domainConfig, err := app.DomainsConfig(conf)
			require.NoError(t, err)

			err = CheckConfig(database.NewDataStore(db), domainConfig)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestCreateMissingData_Integration(t *testing.T) {
	testoutput.SuppressIfPasses(t)

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
					AllUsersGroup: 2, NonTempUsersGroup: 3, TempUsersGroup: 4,
				},
			},
			fixtures: `
groups:
	- {id: 2, type: "Base", name: "AllUsers", text_id: "AllUsers"}
	- {id: 3, type: "Base", name: "NonTempUsers", text_id: "NonTempUsers"}
	- {id: 4, type: "Base", name: "TempUsers", text_id: "TempUsers"}
groups_groups:
	- {parent_group_id: 2, child_group_id: 3}
	- {parent_group_id: 2, child_group_id: 4}
			`,
			checkConfigPassBefore: true,
		},
		{
			name: "create all rows, nothing exists yet",
			config: []domain.ConfigItem{
				{
					Domains:       []string{"127.0.0.1"},
					AllUsersGroup: 2, NonTempUsersGroup: 3, TempUsersGroup: 4,
				},
			},
			checkConfigPassBefore: false,
		},
	}

	ctx := testhelpers.CreateTestContext()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(ctx, tt.fixtures)
			defer func() { _ = db.Close() }()

			conf := viper.New()
			conf.Set("domains", tt.config)

			domainConfig, err := app.DomainsConfig(conf)
			require.NoError(t, err)

			// Verify that CheckConfig() passes or fails as expected before we call CreateMissingData().
			errCheckConfigBefore := CheckConfig(database.NewDataStore(db), domainConfig)
			if tt.checkConfigPassBefore {
				require.NoError(t, errCheckConfigBefore)
			} else {
				require.Error(t, errCheckConfigBefore)
			}

			err = CreateMissingData(database.NewDataStore(db), domainConfig)
			require.NoError(t, err)

			// Once CreateMissingData() have been called, CheckConfig() should pass.
			err = CheckConfig(database.NewDataStore(db), domainConfig)
			assert.NoError(t, err)
		})
	}
}
