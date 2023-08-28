// Package configdb makes sure the database has all the mandatory rows expected from the configuration.
package configdb

import (
	"fmt"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
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
			{name: "TempUsers", id: domainConfig.TempUsersGroup},
		} {
			hasRows, err := groupStore.ByID(spec.id).HasRows()
			if err != nil {
				return err
			}
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
			{parentName: "AllUsers", childName: "TempUsers", parentID: domainConfig.AllUsersGroup, childID: domainConfig.TempUsersGroup},
		} {
			hasRows, err := groupGroupStore.
				Where("parent_group_id = ?", spec.parentID).
				Where("child_group_id = ?", spec.childID).
				Select("1").Limit(1).HasRows()
			if err != nil {
				return err
			}
			if !hasRows {
				return fmt.Errorf("no %s -> %s link in groups_groups for domain %q",
					spec.parentName, spec.childName, domainConfig.Domains[0])
			}
		}
	}

	// There must be an entry in propagations table with id 1 to handle propagation.
	propagationStore := datastore.Propagations()
	hasRows, err := propagationStore.
		Where("propagation_id = 1").
		Select("1").Limit(1).HasRows()
	if err != nil {
		return err
	}
	if !hasRows {
		return fmt.Errorf("missing entry in propagations table with id 1")
	}

	return nil
}
