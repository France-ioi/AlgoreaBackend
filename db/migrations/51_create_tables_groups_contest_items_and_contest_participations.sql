-- +migrate Up
CREATE TABLE `groups_contest_items` (
    `group_id` bigint(20) NOT NULL,
    `contest_item_id` bigint(20) NOT NULL,
    `can_enter_from` datetime NOT NULL DEFAULT '9999-12-31 23:59:59' COMMENT 'Time from which the group can “enter” this time-limited item',
    `can_enter_until` datetime NOT NULL DEFAULT '9999-12-31 23:59:59' COMMENT 'Time until which the group can “enter” this time-limited item',
    `additional_time` time NOT NULL DEFAULT '00:00:00' COMMENT 'Time that was attributed (can be negative) to this group for this time-limited item',
    PRIMARY KEY (`group_id`, `contest_item_id`)
);

CREATE TABLE `contest_participations` (
    `group_id` bigint(20) NOT NULL,
    `contest_item_id` bigint(20) NOT NULL,
    `contest_started_at` datetime DEFAULT NULL COMMENT 'Time at which the group entered the contest',
    PRIMARY KEY (`group_id`, `contest_item_id`)
);

INSERT INTO `contest_participations` (`group_id`, `contest_item_id`, `contest_started_at`)
    SELECT users.self_group_id, users_items.item_id, users_items.contest_started_at
    FROM users_items
         JOIN users ON users.id = users_items.user_id
    WHERE users.self_group_id IS NOT NULL AND users_items.contest_started_at IS NOT NULL
ON DUPLICATE KEY UPDATE contest_started_at = users_items.contest_started_at;

INSERT INTO `groups_contest_items` (`group_id`, `contest_item_id`, `additional_time`,`can_enter_from`, `can_enter_until`)
    SELECT groups_items.group_id, groups_items.item_id, groups_items.additional_time, '9999-12-31 23:59:59', '9999-12-31 23:59:59'
    FROM groups_items
    WHERE groups_items.additional_time IS NOT NULL
ON DUPLICATE KEY UPDATE additional_time = groups_items.additional_time;


-- +migrate Down
DROP TABLE IF EXISTS `groups_contest_items`;
DROP TABLE IF EXISTS `contest_participations`;
