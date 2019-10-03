-- +migrate Up
ALTER TABLE `groups_items` DROP COLUMN `additional_time`;

-- +migrate Down
ALTER TABLE `groups_items`
    ADD COLUMN `additional_time` time DEFAULT NULL
        COMMENT 'Time that was attributed (can be negative) to this group for this item (typically, for  time-limited items)'
        AFTER `cached_manager_access`;

INSERT INTO `groups_items` (`group_id`, `item_id`, `additional_time`)
    SELECT group_id, groups_contest_items.item_id, groups_contest_items.additional_time
    FROM groups_contest_items
    WHERE groups_contest_items.additional_time != '00:00:00'
ON DUPLICATE KEY UPDATE additional_time = groups_contest_items.additional_time;
