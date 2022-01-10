-- +migrate Up
SET @id = FLOOR(RAND(1234) * 1000000000) + FLOOR(RAND(5678) * 1000000000) * 1000000000;

SET @old_fk_checks = @@SESSION.FOREIGN_KEY_CHECKS;
SET FOREIGN_KEY_CHECKS = 0;
INSERT INTO `items_strings` (`item_id`, `language_tag`, `title`)
    VALUES (@id, 'fr', 'Activités publiques'), (@id, 'en', 'Public activities');
INSERT INTO `items` (`id`, `type`, `default_language_tag`, `options`) VALUES (@id, 'Chapter', 'fr', '{}');
SET FOREIGN_KEY_CHECKS = @old_fk_checks;

INSERT INTO `permissions_granted` (`group_id`, `item_id`, `source_group_id`, `origin`, `can_view`)
    SELECT `groups`.`id`, @id, `groups`.`id`, 'group_membership', 'content'
    FROM `groups` WHERE `type` = 'Base' AND `name` = 'AllUsers';

UPDATE `groups` SET `root_activity_id` = @id WHERE `type` = 'Base' AND `name` = 'AllUsers';

INSERT INTO `items_items` (`parent_item_id`, `child_item_id`, `child_order`)
    SELECT @id, `id`, ROW_NUMBER() OVER () FROM `items` WHERE `is_root` ORDER BY `items`.`id`;

-- +migrate Down
UPDATE `items` SET `is_root` = 1 WHERE `id` IN (
    SELECT `child_item_id` FROM `items_items` WHERE `parent_item_id` IN (
        SELECT `root_activity_id` FROM `groups` WHERE `type` = 'Base' AND `name` = 'AllUsers'
    )
);

DELETE FROM `items_items` WHERE `parent_item_id` IN (
    SELECT `root_activity_id` FROM `groups` WHERE `type` = 'Base' AND `name` = 'AllUsers'
);

DELETE FROM `items_ancestors` WHERE `ancestor_item_id` IN (
    SELECT `root_activity_id` FROM `groups` WHERE `type` = 'Base' AND `name` = 'AllUsers'
);

DELETE FROM `results` WHERE `item_id` IN (
    SELECT `root_activity_id` FROM `groups` WHERE `type` = 'Base' AND `name` = 'AllUsers'
);

DELETE `permissions_granted` FROM `groups`
    JOIN `permissions_granted` ON `permissions_granted`.`group_id` = `groups`.`id` AND
                                  `permissions_granted`.`source_group_id` = `groups`.`id` AND
                                  `permissions_granted`.`item_id` = `groups`.`root_activity_id`
    WHERE `groups`.`type` = 'Base' AND `groups`.`name` = 'AllUsers' AND `permissions_granted`.`origin` = 'group_membership';

SET @old_fk_checks = @@SESSION.FOREIGN_KEY_CHECKS;
SET FOREIGN_KEY_CHECKS = 0;
DELETE `items_strings` FROM `groups`
    JOIN `items_strings` ON `items_strings`.`item_id` = `groups`.`root_activity_id`
    WHERE `groups`.`type` = 'Base' AND `groups`.`name` = 'AllUsers';
DELETE `items` FROM `groups`
    JOIN `items` ON `items`.`id` = `groups`.`root_activity_id`
    WHERE `groups`.`type` = 'Base' AND `groups`.`name` = 'AllUsers';
SET @id = FLOOR(RAND(1234) * 1000000000) + FLOOR(RAND(5678) * 1000000000) * 1000000000;
DELETE FROM `items_strings` WHERE `item_id`=@id;
DELETE FROM `items` WHERE `id`=@id;
SET FOREIGN_KEY_CHECKS = @old_fk_checks;

UPDATE `groups` SET `root_activity_id` = NULL WHERE `type` = 'Base' AND `name` = 'AllUsers';
