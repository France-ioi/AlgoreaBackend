-- +migrate Up
DELETE `users_items` FROM `users_items`
LEFT JOIN `groups_attempts` ON `groups_attempts`.`id` = `users_items`.`active_attempt_id`
WHERE `groups_attempts`.`id` IS NULL;

ALTER TABLE `users_items`
    MODIFY COLUMN `active_attempt_id` bigint(20) NOT NULL COMMENT 'Current attempt selected by this user.',
    DROP INDEX `active_attempt_id`,
    DROP PRIMARY KEY,
    DROP COLUMN `id`,
    DROP INDEX `user_item`,
    ADD PRIMARY KEY `user_id_item_id` (`user_id`, `item_id`),
    ADD CONSTRAINT `fk_user_id` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    ADD CONSTRAINT `fk_item_id` FOREIGN KEY (`item_id`) REFERENCES `items`(`id`) ON DELETE CASCADE,
    ADD CONSTRAINT `fk_active_attempt_id` FOREIGN KEY (`active_attempt_id`) REFERENCES `groups_attempts` (`id`) ON DELETE CASCADE;
DROP TRIGGER `before_insert_users_items`;

-- +migrate Down
ALTER TABLE `users_items`
    DROP FOREIGN KEY `fk_active_attempt_id`,
    DROP FOREIGN KEY `fk_user_id`,
    DROP FOREIGN KEY `fk_item_id`;
ALTER TABLE `users_items`
    DROP PRIMARY KEY,
    ADD COLUMN `id` bigint(20) NOT NULL FIRST,
    MODIFY COLUMN `active_attempt_id` bigint(20) DEFAULT NULL COMMENT 'Current attempt selected by this user.',
    ADD INDEX `active_attempt_id` (`active_attempt_id`),
    ADD UNIQUE INDEX `user_item` (`user_id`, `item_id`);

UPDATE `users_items` SET `id` = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;

ALTER TABLE `users_items` MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT PRIMARY KEY;

-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users_items` BEFORE INSERT ON `users_items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
