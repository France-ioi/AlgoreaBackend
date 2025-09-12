package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() { //nolint:gochecknoinits // register the migration
	goose.AddMigrationNoTxContext(SetGroupsAncestorsChildGroupTypeUp, SetGroupsAncestorsChildGroupTypeDown)
}

// SetGroupsAncestorsChildGroupTypeUp sets groups_ancestors.child_group_type by chunks.
func SetGroupsAncestorsChildGroupTypeUp(ctx context.Context, db *sql.DB) error {
	for {
		result, err := db.ExecContext(ctx, `
			UPDATE groups_ancestors
			JOIN (
				SELECT ancestor_group_id, child_group_id
				FROM groups_ancestors
				WHERE child_group_type IS NULL
				LIMIT 100
			) AS ga1 USING(ancestor_group_id, child_group_id)
			JOIN `+"`groups`"+` ON groups.id = groups_ancestors.child_group_id
			SET groups_ancestors.child_group_type = groups.type`)
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

// SetGroupsAncestorsChildGroupTypeDown is a no-op migration reversing SetGroupsAncestorsChildGroupTypeUp.
func SetGroupsAncestorsChildGroupTypeDown(_ context.Context, _ *sql.DB) error {
	return nil
}
