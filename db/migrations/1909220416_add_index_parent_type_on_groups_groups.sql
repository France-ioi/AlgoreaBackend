-- +migrate Up
ALTER TABLE `groups_groups` ADD INDEX  `parent_type` (`parent_group_id`, `type`);

-- +migrate Down
ALTER TABLE `groups_groups` DROP INDEX `parent_type`;
