-- +migrate Up
SELECT count(*) INTO @cnt FROM `permissions_granted` AS pg WHERE `origin` = 'self';

/* all 9319 rows on the first run, but can be reapplied */
UPDATE `permissions_granted` SET `origin` = 'other' WHERE @cnt = 0;

/* 3213 rows */
UPDATE `permissions_granted` SET `origin` = 'self' WHERE `source_group_id` = `group_id` AND `origin` = 'other';

/* 2837 rows */
UPDATE `permissions_granted`
    JOIN `item_dependencies` ON `item_dependencies`.`dependent_item_id` = `permissions_granted`.`item_id`
    JOIN `groups` ON `groups`.`id` = `permissions_granted`.`group_id` AND `groups`.`type` = 'UserSelf'
    JOIN `groups_attempts`
        ON `groups_attempts`.`group_id` = `permissions_granted`.`group_id` AND
           `groups_attempts`.`item_id` = `item_dependencies`.`item_id` AND
           `groups_attempts`.`score` >= `item_dependencies`.`score`
SET `permissions_granted`.`origin` = 'item_unlocking',
    `permissions_granted`.`source_group_id` = `permissions_granted`.`group_id`
WHERE `origin` = 'other' AND `can_view` = 'content' AND `can_grant_view` = 'none' AND
      `can_edit` = 'none' AND `can_watch` = 'none' AND NOT `is_owner` AND `source_group_id` = -1;

/* 1368 rows (many of them look like given to groups on contest start, but items.duration=NULL) */
UPDATE `permissions_granted`
    JOIN `groups` ON `groups`.`id` = `permissions_granted`.`group_id` AND `groups`.`type` != 'UserSelf'
SET `origin` = 'group_membership'
WHERE `origin` = 'other';

/* many of 'other' rows look like given to users on contest start, but items.duration=NULL
 (see item_id = 1060839842388488342, group_id = 973991116442908368) */

UPDATE `permissions_granted` SET `source_group_id` = `group_id` WHERE `source_group_id` < 0;

ALTER TABLE `permissions_granted`
    ADD CONSTRAINT `fk_permissions_granted_source_group_id_groups_id` FOREIGN KEY (`source_group_id`) REFERENCES `groups`(`id`)
        ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE `permissions_granted`
    DROP FOREIGN KEY `fk_permissions_granted_source_group_id_groups_id`;
