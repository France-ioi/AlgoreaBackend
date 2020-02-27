-- +migrate Up
ALTER TABLE `permissions_granted`
    ADD COLUMN `can_enter_from` DATETIME NOT NULL DEFAULT '9999-12-31 23:59:59'
        COMMENT 'Time from which the group can “enter” this item, superseded by `items.entering_time_min`' AFTER `can_view`,
    ADD COLUMN `can_enter_until` DATETIME NOT NULL DEFAULT '9999-12-31 23:59:59'
        COMMENT 'Time until which the group can “enter” this item, superseded by `items.entering_time_max`' AFTER `can_enter_from`,
    MODIFY COLUMN `can_grant_view` ENUM('none','enter','content','content_with_descendants','solution','solution_with_grant') NOT NULL DEFAULT 'none'
        COMMENT 'The level of visibility that the group can give on this item to other groups on which it has the right to';

ALTER TABLE `permissions_generated`
    MODIFY COLUMN `can_grant_view_generated`
        ENUM('none','enter','content','content_with_descendants','solution','solution_with_grant') NOT NULL DEFAULT 'none'
        COMMENT 'The aggregated level of visibility that the group can give on this item to other groups on which it has the right to';

ALTER TABLE `groups_contest_items`
    DROP COLUMN can_enter_from,
    DROP COLUMN can_enter_until;

ALTER TABLE `items`
    ADD COLUMN `entering_time_min` DATETIME DEFAULT NULL
        COMMENT 'Lower bound on the entering time. Has the priority over given can_enter_from/until permissions.'
        AFTER `contest_entering_condition`,
    ADD COLUMN `entering_time_max` DATETIME DEFAULT NULL
        COMMENT 'Upper bound on the entering time. Has the priority over given can_enter_from/until permissions.'
        AFTER `entering_time_min`;

-- +migrate Down
ALTER TABLE `groups_contest_items`
    ADD COLUMN `can_enter_from` DATETIME NOT NULL DEFAULT '9999-12-31 23:59:59'
        COMMENT 'Time from which the group can “enter” this contest' AFTER `item_id`,
    ADD COLUMN `can_enter_until` DATETIME NOT NULL DEFAULT '9999-12-31 23:59:59'
        COMMENT 'Time until which the group can “enter” this contest' AFTER `can_enter_from`;

INSERT INTO `groups_contest_items` (`group_id`, `item_id`, `can_enter_from`, `can_enter_until`)
SELECT `group_id`, `item_id`, `can_enter_from`, `can_enter_until` FROM `permissions_granted`
WHERE `permissions_granted`.`can_enter_from` IS NOT NULL OR `permissions_granted`.`can_enter_until` IS NOT NULL
ON DUPLICATE KEY UPDATE
    `groups_contest_items`.`can_enter_from` = IFNULL(VALUES(`can_enter_from`), `groups_contest_items`.`can_enter_from`),
    `groups_contest_items`.`can_enter_until` = IFNULL(VALUES(`can_enter_until`), `groups_contest_items`.`can_enter_until`);

UPDATE `permissions_granted` SET `can_grant_view` = 'none' WHERE `can_grant_view` = 'enter';
ALTER TABLE `permissions_granted`
    DROP COLUMN `can_enter_from`,
    DROP COLUMN `can_enter_until`,
    MODIFY COLUMN `can_grant_view` ENUM('none','content','content_with_descendants','solution','solution_with_grant') NOT NULL DEFAULT 'none'
        COMMENT 'The level of visibility that the group can give on this item to other groups on which it has the right to';

UPDATE `permissions_generated` SET `can_grant_view_generated` = 'none' WHERE `can_grant_view_generated` = 'enter';
ALTER TABLE `permissions_generated`
    MODIFY COLUMN `can_grant_view_generated`
        ENUM('none','content','content_with_descendants','solution','solution_with_grant') NOT NULL DEFAULT 'none'
        COMMENT 'The aggregated level of visibility that the group can give on this item to other groups on which it has the right to';

ALTER TABLE `items`
    DROP COLUMN `entering_time_min`,
    DROP COLUMN `entering_time_max`;
