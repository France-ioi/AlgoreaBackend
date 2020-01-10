-- +migrate Up
INSERT INTO `groups` (`id`,`name`, `description`, `type`)
    VALUES (777, 'Former task owners',
            'Contains all the task owners from AlgoreaPlatform and has can_edit=children permission on former custom chapters',
            'Other');
SELECT @group_id := id FROM `groups`  WHERE `type` = 'Other' AND `name` = 'Former task owners' LIMIT 1;
INSERT INTO groups_groups (`id`, `parent_group_id`, `child_group_id`, `type`, `child_order`)
    SELECT (@group_id + `children`.`group_id`) % 9223372036854775806 + 1,
           @group_id AS `parent_group_id`,
           `children`.`group_id` AS `child_group_id`,
           'direct' AS `direct`,
           ROW_NUMBER() OVER() AS `child_order`
    FROM (SELECT DISTINCT `group_id` FROM `permissions_granted` WHERE `is_owner`) AS `children`;

INSERT INTO `permissions_granted` (`group_id`, `item_id`, `source_group_id`, `can_edit`)
    SELECT DISTINCT @group_id AS `group_id`,
           `id` AS `item_id`,
           -2 AS `source_group_id`,
           'children' AS `can_edit`
    FROM `items`
    WHERE `custom_chapter`;

UPDATE `items_items` JOIN `items` ON `items`.`id` = `items_items`.`parent_item_id` AND `items`.`custom_chapter`
    SET `items_items`.`content_view_propagation` = 'none',
        `items_items`.`upper_view_levels_propagation` = 'use_content_view_propagation',
        `items_items`.`grant_view_propagation` = 0,
        `items_items`.`watch_propagation` = 0,
        `items_items`.`edit_propagation` = 0;

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
