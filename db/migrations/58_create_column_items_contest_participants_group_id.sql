-- +migrate Up
ALTER TABLE `items` ADD COLUMN `contest_participants_group_id` bigint(20) DEFAULT NULL
    COMMENT 'id of the group to which all the contest participants (users or teams) belong.'
    AFTER `qualified_group_id`;

-- +migrate Down
ALTER TABLE `items` DROP COLUMN `contest_participants_group_id`;
