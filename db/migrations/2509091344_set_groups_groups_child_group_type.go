package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() { //nolint:gochecknoinits // register the migration
	goose.AddMigrationNoTxContext(SetGroupsGroupsChildGroupTypeUp, SetGroupsGroupsChildGroupTypeDown)
}

// SetGroupsGroupsChildGroupTypeUp sets groups_groups.child_group_type by chunks.
func SetGroupsGroupsChildGroupTypeUp(ctx context.Context, db *sql.DB) error {
	for {
		result, err := db.ExecContext(ctx, `
			UPDATE groups_groups
			JOIN (
				SELECT parent_group_id, child_group_id
				FROM groups_groups
				WHERE child_group_type IS NULL
				LIMIT 100
			) AS gg1 USING(parent_group_id, child_group_id)
			JOIN `+"`groups`"+` ON groups.id = groups_groups.child_group_id
			SET groups_groups.child_group_type = groups.type`)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return nil
		}
	}
}

// SetGroupsGroupsChildGroupTypeDown is a no-op migration reversing SetGroupsGroupsChildGroupTypeUp.
func SetGroupsGroupsChildGroupTypeDown(_ context.Context, _ *sql.DB) error {
	return nil
}
