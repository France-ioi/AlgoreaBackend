-- +migrate Up
# Delete 'RootAdmin' `groups` (cannot be undone)
CREATE TEMPORARY TABLE root_admin_groups
    SELECT DISTINCT `parent_group_id`
    FROM `groups_groups` AS gg
    JOIN `groups` AS `parent`
        ON `parent`.`id` = gg.`parent_group_id` AND
            `parent`.`type`='Base'
    JOIN `groups` AS child
        ON child.id = gg.child_group_id AND child.type = 'UserAdmin';

DELETE `groups_ancestors` FROM `groups_ancestors`
    JOIN `root_admin_groups` ON `root_admin_groups`.`parent_group_id` = `groups_ancestors`.`ancestor_group_id`;
DELETE `groups_ancestors` FROM `groups_ancestors`
    JOIN `root_admin_groups` ON `root_admin_groups`.`parent_group_id` = `groups_ancestors`.`child_group_id`;
DELETE `groups` FROM `groups`
    JOIN `root_admin_groups` ON `root_admin_groups`.`parent_group_id` = `groups`.`id`;
DELETE `groups_propagate` FROM `groups_propagate`
    JOIN `root_admin_groups` ON `root_admin_groups`.`parent_group_id` = `groups_propagate`.`id`;

DELETE `groups_groups` FROM `groups_groups`
    JOIN `root_admin_groups` ON `root_admin_groups`.`parent_group_id` = `groups_groups`.`child_group_id`;

DELETE `groups_groups` FROM `groups_groups`
    JOIN `root_admin_groups` ON `root_admin_groups`.`parent_group_id` = `groups_groups`.`parent_group_id`;

DROP TEMPORARY TABLE root_admin_groups;

CREATE TEMPORARY TABLE user_admin_groups
SELECT `id` FROM `groups` WHERE `groups`.`type` = 'UserAdmin';

# Delete `groups` matching with type = 'UserAdmin'
DELETE `filters` FROM `filters` JOIN `user_admin_groups` ON `user_admin_groups`.`id` = `filters`.`group_id`;
DELETE `groups_ancestors` FROM `groups_ancestors`
    JOIN `user_admin_groups` ON `user_admin_groups`.`id` = `groups_ancestors`.`ancestor_group_id`;
DELETE `groups_ancestors` FROM `groups_ancestors`
    JOIN `user_admin_groups` ON `user_admin_groups`.`id` = `groups_ancestors`.`child_group_id`;
DELETE `groups_attempts` FROM `groups_attempts`
    JOIN `user_admin_groups` ON `user_admin_groups`.`id` = `groups_attempts`.`group_id`;
DELETE `groups_contest_items` FROM `groups_contest_items`
    JOIN `user_admin_groups` ON `user_admin_groups`.`id` = `groups_contest_items`.`group_id`;
DELETE `groups_groups` FROM `groups_groups`
    JOIN `user_admin_groups` ON `user_admin_groups`.`id` = `groups_groups`.`parent_group_id`;
DELETE `groups_groups` FROM `groups_groups`
    JOIN `user_admin_groups` ON `user_admin_groups`.`id` = `groups_groups`.`child_group_id`;
DELETE `groups_login_prefixes` FROM `groups_login_prefixes`
    JOIN `user_admin_groups` ON `user_admin_groups`.`id` = `groups_login_prefixes`.`group_id`;
DELETE `groups` FROM `groups` JOIN `user_admin_groups` USING(`id`);
DELETE `groups_propagate` FROM `groups_propagate`
    JOIN `user_admin_groups` USING(`id`);

DROP TEMPORARY TABLE `user_admin_groups`;

# Remove UserAdmin from `groups`.`type`
ALTER TABLE `groups` MODIFY COLUMN `type` enum('Class','Team','Club','Friends','Other','UserSelf','Base') NOT NULL;

ALTER TABLE `users`
    DROP INDEX `owned_group_id`,
    DROP COLUMN `owned_group_id`;

-- +migrate Down
ALTER TABLE `users`
    ADD COLUMN `owned_group_id` bigint(20) DEFAULT NULL
        COMMENT 'Group that will contain groups that this user manages'
        AFTER `help_given`,
    ADD UNIQUE KEY `owned_group_id` (`owned_group_id`);

# Add UserAdmin into `groups`.`type`
ALTER TABLE `groups` MODIFY COLUMN `type` enum('Class','Team','Club','Friends','Other','UserSelf','UserAdmin','Base') NOT NULL;

# Restore `groups` with `type` = 'UserAdmin' (use `team_item_id` as a temporary storage)
INSERT INTO `groups` (`name`, `type`, `team_item_id`, `created_at`)
    SELECT CONCAT(`name`, '-admin'), 'UserAdmin', `user_groups`.`id`, NOW()
    FROM `groups` AS `user_groups` WHERE `user_groups`.`type` = 'UserSelf';

UPDATE `users`
    JOIN `groups` AS `admin_groups`
        ON `admin_groups`.`type` = 'UserAdmin' AND `admin_groups`.`team_item_id` = `users`.`group_id`
SET `owned_group_id` = `admin_groups`.`id`;

UPDATE `groups` SET `team_item_id` = NULL WHERE `type` = 'UserAdmin';

INSERT INTO `groups_groups` (`child_group_id`, `parent_group_id`, `role`, `child_order`)
    SELECT `group_managers`.`group_id`, `users`.`owned_group_id`, 'owner',
           (SELECT IFNULL(MAX(`child_order`),0)+1 FROM `groups_groups` WHERE `parent_group_id` = `users`.`owned_group_id`)
    FROM `group_managers`
    JOIN `users` ON `users`.`group_id` = `group_managers`.`manager_id`
    WHERE `can_manage` = 'memberships_and_group' AND `can_grant_group_access` AND `can_watch_members`;

INSERT INTO `groups_groups` (`child_group_id`, `parent_group_id`, `role`, `child_order`)
    SELECT `group_managers`.`group_id`, `users`.`owned_group_id`, 'observer',
           (SELECT IFNULL(MAX(`child_order`),0)+1 FROM `groups_groups` WHERE `parent_group_id` = `users`.`owned_group_id`)
    FROM `group_managers`
    JOIN `users` ON `users`.`group_id` = `group_managers`.`manager_id`
    WHERE `can_manage` = 'none' AND NOT `can_grant_group_access` AND `can_watch_members`;
