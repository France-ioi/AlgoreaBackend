-- +migrate Up
UPDATE `groups` SET `team_item_id` = NULL WHERE `team_item_id` = 0;

UPDATE `groups` SET activity_id = team_item_id WHERE team_item_id IS NOT NULL;
ALTER TABLE `groups`
    DROP COLUMN `team_item_id`,
    DROP COLUMN `team_participating`;

-- +migrate Down
ALTER TABLE `groups`
    ADD COLUMN `team_item_id` BIGINT(20) DEFAULT NULL
        COMMENT 'If this group is a team, what item is it attached to?'
            AFTER `is_public`,
    ADD COLUMN `team_participating` tinyint(1) NOT NULL DEFAULT '0'
        COMMENT 'Did the team start the item it is associated to (from team_item_id)?'
            AFTER `team_item_id`;

UPDATE `groups` SET team_item_id = activity_id WHERE activity_id IS NOT NULL;
