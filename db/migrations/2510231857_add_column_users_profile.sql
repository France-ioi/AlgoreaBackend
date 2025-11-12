-- +goose Up
ALTER TABLE `users` ADD COLUMN `profile` JSON
  COMMENT 'A JSON object containing user profile information returned by the login module as the "profile" field'
  AFTER `login`;

-- +goose Down
ALTER TABLE `users` DROP COLUMN `profile`;
