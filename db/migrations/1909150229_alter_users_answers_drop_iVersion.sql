-- +migrate Up
ALTER TABLE `users_answers` DROP COLUMN `iVersion`;

-- +migrate Down
ALTER TABLE `users_answers` ADD COLUMN `iVersion` bigint(20) NOT NULL;
