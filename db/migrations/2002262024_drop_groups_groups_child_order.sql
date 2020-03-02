-- +migrate Up
ALTER TABLE `groups_groups`
    DROP INDEX `parent_order`,
    DROP COLUMN `child_order`;

DROP VIEW groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

-- +migrate Down
ALTER TABLE `groups_groups`
    ADD COLUMN `child_order` int(11) NOT NULL DEFAULT '0' COMMENT 'Position of this child among its siblings.'
        AFTER `child_group_id`,
    ADD INDEX `parent_order` (`parent_group_id`,`child_order`);

UPDATE `groups_groups` JOIN (
    SELECT `parent_group_id`, `child_group_id`,
           ROW_NUMBER() OVER (PARTITION BY `parent_group_id` ORDER BY `child_group_id`) AS `child_order`
    FROM `groups_groups`
) AS `gg` USING (`parent_group_id`, `child_group_id`)
SET `groups_groups`.`child_order` = `gg`.`child_order`;

DROP VIEW groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;
