-- +migrate Up
CREATE TABLE `group_managers` (
    `group_id` BIGINT(20) NOT NULL,
    `manager_id` BIGINT(20) NOT NULL,
    `can_manage` ENUM('none', 'memberships', 'memberships_and_group') NOT NULL DEFAULT 'none',
    `can_grant_group_access` TINYINT(1) NOT NULL DEFAULT 0,
    `can_watch_members` TINYINT(1) NOT NULL DEFAULT 0,
    PRIMARY KEY (`group_id`, `manager_id`),
    CONSTRAINT `fk_group_managers_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_group_managers_manager_id_groups_id` FOREIGN KEY (`manager_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE
) COMMENT 'Group managers and their permissions' ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT IGNORE INTO `group_managers`
    SELECT `child_group_id`, `users`.`group_id`, 'memberships_and_group', 1, 1
    FROM `groups_groups`
    JOIN `users` ON `users`.`owned_group_id` = `groups_groups`.`parent_group_id`
    WHERE `role` IN ('manager', 'owner');

INSERT IGNORE INTO `group_managers`
    SELECT `child_group_id`, `users`.`group_id`, 'none', 0, 1
    FROM `groups_groups`
    JOIN `users` ON `users`.`owned_group_id` = `groups_groups`.`parent_group_id`
    WHERE `role` = 'observer';

-- +migrate Down
DROP TABLE `group_managers`;
