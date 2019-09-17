-- +migrate Up
ALTER TABLE `history_groups_attempts` MODIFY COLUMN `bDeleted` tinyint(1) NOT NULL DEFAULT '0';

-- +migrate Down
ALTER TABLE `history_groups_attempts` MODIFY COLUMN `bDeleted` tinyint(1) NOT NULL;
