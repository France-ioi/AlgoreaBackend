-- +migrate Up
INSERT INTO `groups` (`name`, `description`, `type`)
    VALUES ('Former task owners',
            'Contains all the task owners from AlgoreaPlatform and has can_edit=children permission on former custom chapters',
            'Other');
SELECT @group_id := id FROM `groups`  WHERE `type` = 'Other' AND `name` = 'Former task owners' LIMIT 1;
INSERT INTO groups_groups (`parent_group_id`, `child_group_id`, `type`, `child_order`)
    SELECT @group_id AS `parent_group_id`,
           `children`.`group_id` AS `child_group_id`,
           'direct' AS `direct`,
           ROW_NUMBER() OVER() AS `child_order`
    FROM (SELECT DISTINCT `group_id` FROM `permissions_granted` WHERE `is_owner`) AS `children`;

INSERT INTO `permissions_granted` (`group_id`, `item_id`, `giver_group_id`, `can_edit`)
    SELECT DISTINCT @group_id AS `group_id`,
           `id` AS `item_id`,
           -2 AS `giver_group_id`,
           'children' AS `can_edit`
    FROM `items`
    WHERE `custom_chapter`;

ALTER TABLE `items` DROP COLUMN `custom_chapter`;

-- +migrate Down
SELECT @group_id := id FROM `groups`  WHERE `type` = 'Other' AND `name` = 'Former task owners' LIMIT 1;

ALTER TABLE `items`
    ADD COLUMN `custom_chapter` tinyint(3) unsigned DEFAULT '0'
        COMMENT 'Whether it is a chapter where users can add their own content. Access to this chapter will not be propagated to its children'
        AFTER `display_details_in_parent`;

INSERT INTO `items` (`id`)
    SELECT `item_id` AS `id` FROM `permissions_granted` WHERE `group_id` = @group_id
ON DUPLICATE KEY UPDATE `custom_chapter` = 1;

DELETE FROM `groups_groups` WHERE `parent_group_id` = @group_id;
DELETE FROM `groups` WHERE id = @group_id;
