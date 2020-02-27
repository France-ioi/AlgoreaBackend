-- +migrate Up

/* 2000897 rows */
UPDATE `users_answers`
    JOIN `groups_attempts` ON `groups_attempts`.`group_id` = `users_answers`.`user_id` AND
                              `groups_attempts`.item_id = `users_answers`.`item_id`
SET `users_answers`.`attempt_id` = `groups_attempts`.`id`
WHERE `users_answers`.`attempt_id` IS NULL;

/* 35548 rows */
UPDATE `users_answers`
    JOIN `users_items` ON `users_items`.`user_id` = `users_answers`.`user_id` AND
                          `users_items`.`item_id` = `users_answers`.`item_id`
SET `users_answers`.`attempt_id` = `users_items`.`active_attempt_id`
WHERE `users_answers`.`attempt_id` IS NULL;

/* 2414 rows */
UPDATE `users_answers`
    JOIN `groups_groups` ON `groups_groups`.`child_group_id` = `users_answers`.`user_id`
    JOIN `groups` ON `groups`.`id` = `groups_groups`.`parent_group_id`
    JOIN `items_ancestors` ON `items_ancestors`.`ancestor_item_id` = `groups`.`team_item_id` AND
                              `items_ancestors`.`child_item_id` = `users_answers`.`item_id`
    JOIN `groups_attempts` ON `groups_attempts`.`group_id` = `groups`.`id` AND
                              `groups_attempts`.`item_id` = `items_ancestors`.`child_item_id`
SET `users_answers`.`attempt_id` = `groups_attempts`.`id`
WHERE `users_answers`.`attempt_id` IS NULL;

/* 457848 rows :~( */
DELETE FROM `users_answers` WHERE `attempt_id` IS NULL;

/* 878 rows */
DELETE `users_answers` FROM `users_answers`
    LEFT JOIN `groups_attempts` ON `groups_attempts`.`id` = `users_answers`.`attempt_id`
WHERE `groups_attempts`.`id` IS NULL;

ALTER TABLE `users_answers`
    MODIFY COLUMN `attempt_id` bigint(20) NOT NULL,
    DROP INDEX `item_id`,
    DROP COLUMN `item_id`,
    ADD CONSTRAINT `fk_users_answers_attempt_id_groups_attempts_id`
        FOREIGN KEY (`attempt_id`) REFERENCES `groups_attempts`(`id`) ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE `users_answers`
    DROP FOREIGN KEY `fk_users_answers_attempt_id_groups_attempts_id`,
    ADD COLUMN `item_id` bigint(20) NOT NULL AFTER `user_id`,
    ADD INDEX `item_id` (`item_id`),
    MODIFY COLUMN `attempt_id` bigint(20) DEFAULT NULL;

UPDATE `users_answers` JOIN `groups_attempts` ON `groups_attempts`.`id` = `users_answers`.`attempt_id`
SET `users_answers`.`item_id` = `groups_attempts`.`item_id`;
