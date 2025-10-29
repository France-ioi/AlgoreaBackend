package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() { //nolint:gochecknoinits // register the migration
	goose.AddMigrationNoTxContext(SetUsersProfileUp, SetUsersProfileDown)
}

// SetUsersProfileUp sets users.profile by chunks.
func SetUsersProfileUp(ctx context.Context, db *sql.DB) error {
	for {
		result, err := db.ExecContext(ctx, `
			UPDATE users
			SET profile = JSON_OBJECT("first_name", first_name, "last_name", last_name)
			WHERE NOT temp_user AND profile IS NULL
			LIMIT 100`)
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

// SetUsersProfileDown is a no-op migration reversing SetUsersProfileUp.
func SetUsersProfileDown(_ context.Context, _ *sql.DB) error {
	return nil
}
