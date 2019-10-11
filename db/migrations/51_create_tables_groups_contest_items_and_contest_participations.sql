-- +migrate Up
CREATE TABLE `groups_contest_items` (
    `group_id` bigint(20) NOT NULL,
    `item_id` bigint(20) NOT NULL,
    `can_enter_from` datetime NOT NULL DEFAULT '9999-12-31 23:59:59' COMMENT 'Time from which the group can “enter” this contest',
    `can_enter_until` datetime NOT NULL DEFAULT '9999-12-31 23:59:59' COMMENT 'Time until which the group can “enter” this contest',
    `additional_time` time NOT NULL DEFAULT '00:00:00' COMMENT 'Time that was attributed (can be negative) to this group for this contest',
    PRIMARY KEY (`group_id`, `item_id`)
) COMMENT 'Group constraints on contest participations';

CREATE TABLE `contest_participations` (
    `group_id` bigint(20) NOT NULL,
    `item_id` bigint(20) NOT NULL,
    `entered_at` datetime NOT NULL COMMENT 'Time at which the group entered the contest',
    PRIMARY KEY (`group_id`, `item_id`)
) COMMENT 'Information on when teams or users entered contests';

INSERT INTO `contest_participations` (`group_id`, `item_id`, `entered_at`)
    SELECT users.self_group_id, users_items.item_id, users_items.contest_started_at
    FROM users_items
         JOIN users ON users.id = users_items.user_id
    WHERE users.self_group_id IS NOT NULL AND users_items.contest_started_at IS NOT NULL
ON DUPLICATE KEY UPDATE entered_at = users_items.contest_started_at;

INSERT INTO `groups_contest_items` (`group_id`, `item_id`, `additional_time`,`can_enter_from`, `can_enter_until`)
    SELECT groups_items.group_id, groups_items.item_id, groups_items.additional_time, '9999-12-31 23:59:59', '9999-12-31 23:59:59'
    FROM groups_items
    WHERE groups_items.additional_time IS NOT NULL
ON DUPLICATE KEY UPDATE additional_time = groups_items.additional_time;


-- +migrate Down
DROP TABLE IF EXISTS `groups_contest_items`;
DROP TABLE IF EXISTS `contest_participations`;
