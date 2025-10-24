-- +goose Up
ALTER TABLE `users` DROP INDEX `profile_is_null`;

-- +goose Down
ALTER TABLE `users` ADD INDEX `profile_is_null` ((NOT temp_user AND `profile` IS NULL));
