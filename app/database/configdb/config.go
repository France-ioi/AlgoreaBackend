// Package configdb makes sure the database has all the mandatory rows expected from the configuration.
package configdb

import (
	"fmt"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
)

// CheckConfig checks that the database contains all the data needed by the config_check.
func CheckConfig(datastore *database.DataStore, domainsConfig []domain.ConfigItem) error {
	groupStore := datastore.Groups()
	groupGroupStore := datastore.ActiveGroupGroups()

	for _, domainConfig := range domainsConfig {
		for _, spec := range []struct {
			name string
			id   int64
		}{
			{name: "AllUsers", id: domainConfig.AllUsersGroup},
			{name: "NonTempUsers", id: domainConfig.NonTempUsersGroup},
			{name: "TempUsers", id: domainConfig.TempUsersGroup},
		} {
			hasRows, err := groupStore.ByID(spec.id).HasRows()
			mustNotBeError(err)

			if !hasRows {
				return fmt.Errorf("no %s group for domain %q", spec.name, domainConfig.Domains[0])
			}
		}

		for _, spec := range []struct {
			parentName string
			childName  string
			parentID   int64
			childID    int64
		}{
			{parentName: "AllUsers", childName: "NonTempUsers", parentID: domainConfig.AllUsersGroup, childID: domainConfig.NonTempUsersGroup},
			{parentName: "AllUsers", childName: "TempUsers", parentID: domainConfig.AllUsersGroup, childID: domainConfig.TempUsersGroup},
		} {
			hasRows, err := groupGroupStore.
				Where("parent_group_id = ?", spec.parentID).
				Where("child_group_id = ?", spec.childID).
				Select("1").Limit(1).HasRows()
			mustNotBeError(err)

			if !hasRows {
				return fmt.Errorf("no %s -> %s link in groups_groups for domain %q",
					spec.parentName, spec.childName, domainConfig.Domains[0])
			}
		}
	}

	return nil
}

// CreateMissingData fills the database with required data (if missing).
func CreateMissingData(datastore *database.DataStore, domainsConfig []domain.ConfigItem) error {
	return datastore.InTransaction(func(store *database.DataStore) error {
		return insertRootGroupsAndRelations(store, domainsConfig)
	})
}

func insertRootGroupsAndRelations(store *database.DataStore, domainsConfig []domain.ConfigItem) error {
	groupStore := store.Groups()
	groupGroupStore := store.GroupGroups()
	var relationsToCreate []map[string]interface{}
	for _, domainConfig := range domainsConfig {
		domainConfig := domainConfig
		insertRootGroups(groupStore, &domainConfig)
		for _, spec := range []map[string]interface{}{
			{"parent_group_id": domainConfig.AllUsersGroup, "child_group_id": domainConfig.NonTempUsersGroup},
			{"parent_group_id": domainConfig.AllUsersGroup, "child_group_id": domainConfig.TempUsersGroup},
		} {
			found, err := groupGroupStore.
				WithExclusiveWriteLock().
				Where("parent_group_id = ?", spec["parent_group_id"]).
				Where("child_group_id = ?", spec["child_group_id"]).
				Limit(1).
				HasRows()
			mustNotBeError(err)

			if !found {
				relationsToCreate = append(relationsToCreate, spec)
			}
		}
	}

	if len(relationsToCreate) > 0 {
		return groupStore.GroupGroups().CreateRelationsWithoutChecking(relationsToCreate)
	}

	return nil
}

func insertRootGroups(groupStore *database.GroupStore, domainConfig *domain.ConfigItem) {
	for _, spec := range []struct {
		name string
		id   int64
	}{
		{name: "AllUsers", id: domainConfig.AllUsersGroup},
		{name: "NonTempUsers", id: domainConfig.NonTempUsersGroup},
		{name: "TempUsers", id: domainConfig.TempUsersGroup},
	} {
		found, err := groupStore.
			ByID(spec.id).
			WithExclusiveWriteLock().
			Where("type = 'Base'").
			Where("name = ?", spec.name).
			Where("text_id = ?", spec.name).
			Limit(1).
			HasRows()
		mustNotBeError(err)

		if !found {
			mustNotBeError(groupStore.InsertMap(map[string]interface{}{
				"id": spec.id, "type": "Base", "name": spec.name, "text_id": spec.name,
			}))
		}
	}
}

func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}
