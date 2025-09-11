-- +goose Up
ALTER TABLE `groups_groups` ADD INDEX `child_group_type` (`child_group_type`);

-- +goose Down
ALTER TABLE `groups_groups` DROP INDEX `child_group_type`;
