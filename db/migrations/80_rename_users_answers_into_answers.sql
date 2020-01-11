-- +migrate Up
ALTER TABLE `users_answers`
    RENAME `answers`,
    DROP FOREIGN KEY `fk_users_answers_attempt_id_groups_attempts_id`,
    DROP FOREIGN KEY `fk_users_answers_user_id_users_group_id`,
    RENAME COLUMN `user_id` TO `author_id`,
    ADD CONSTRAINT `fk_answers_attempt_id_groups_attempts_id`
        FOREIGN KEY (`attempt_id`) REFERENCES `groups_attempts` (`id`) ON DELETE CASCADE,
    ADD CONSTRAINT `fk_answers_author_id_users_group_id`
        FOREIGN KEY (`author_id`) REFERENCES `users` (`group_id`) ON DELETE CASCADE,
    DROP COLUMN `name`;

-- +migrate Down
ALTER TABLE `answers`
    RENAME `users_answers`,
    DROP FOREIGN KEY `fk_answers_attempt_id_groups_attempts_id`,
    DROP FOREIGN KEY `fk_answers_author_id_users_group_id`,
    RENAME COLUMN `author_id` TO `user_id`,
    ADD CONSTRAINT `fk_users_answers_attempt_id_groups_attempts_id`
        FOREIGN KEY (`attempt_id`) REFERENCES `groups_attempts` (`id`) ON DELETE CASCADE,
    ADD CONSTRAINT `fk_users_answers_user_id_users_group_id`
        FOREIGN KEY (`user_id`) REFERENCES `users` (`group_id`) ON DELETE CASCADE,
    ADD COLUMN `name` varchar(200) DEFAULT NULL AFTER `attempt_id`;