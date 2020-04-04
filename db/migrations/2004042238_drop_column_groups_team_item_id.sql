-- +migrate Up
UPDATE `groups` SET activity_id = team_item_id WHERE team_item_id IS NOT NULL;
ALTER TABLE `groups` DROP COLUMN `team_item_id`;

-- +migrate Down
ALTER TABLE `groups` ADD COLUMN `team_item_id` BIGINT(20) DEFAULT NULL
    COMMENT 'If this group is a team, what item is it attached to?'
        AFTER `is_public`;
UPDATE `groups` SET team_item_id = activity_id WHERE activity_id IS NOT NULL;
