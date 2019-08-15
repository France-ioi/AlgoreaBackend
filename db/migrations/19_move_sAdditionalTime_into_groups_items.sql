-- +migrate Up
ALTER TABLE `groups_attempts` DROP COLUMN `sAdditionalTime`;
ALTER TABLE `users_items` DROP COLUMN `sAdditionalTime`;
ALTER TABLE `groups_items` ADD COLUMN `sAdditionalTime` datetime DEFAULT NULL AFTER `sPropagateAccess`;

-- +migrate Down
ALTER TABLE `groups_items` DROP COLUMN `sAdditionalTime`;
ALTER TABLE `users_items` ADD COLUMN `sAdditionalTime` datetime DEFAULT NULL AFTER `sLastHintDate`;
ALTER TABLE `groups_attempts` ADD COLUMN `sAdditionalTime` datetime DEFAULT NULL AFTER `sLastHintDate`;
