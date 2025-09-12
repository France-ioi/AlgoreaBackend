-- +goose Up
ALTER TABLE `groups_ancestors` ADD INDEX `child_group_type` (`child_group_type`);

-- +goose Down
ALTER TABLE `groups_ancestors` DROP INDEX `child_group_type`;
