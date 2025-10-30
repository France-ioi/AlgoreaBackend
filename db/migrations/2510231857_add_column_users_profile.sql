-- +goose Up
SET @query = IF(
  NOT EXISTS(
    SELECT *
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE table_name = 'users'
      AND table_schema = DATABASE()
      AND column_name = 'profile'
  ),
  "ALTER TABLE `users` ADD COLUMN `profile` JSON
COMMENT 'A JSON object containing user profile information returned by the login module as the \"profile\" field'
AFTER `login`",
  'DO TRUE'
);
PREPARE stmt FROM @query;
EXECUTE stmt;

-- +goose Down
ALTER TABLE `users` DROP COLUMN `profile`;
