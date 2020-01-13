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
    DROP COLUMN `name`,
    DROP COLUMN `lang_prog`,
    DROP COLUMN `validated`,
    RENAME COLUMN `submitted_at` TO `created_at`,
    MODIFY COLUMN `type` enum('Submission','Saved','Current') NOT NULL DEFAULT 'Submission'
        COMMENT '\'Submission\' for answers submitted for grading, \'Saved\' for manual backups of answers, \'Current\' for automatic snapshots of the latest answers (unique for a user on an attempt)';

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
    ADD COLUMN `name` varchar(200) DEFAULT NULL AFTER `attempt_id`,
    ADD COLUMN `lang_prog` varchar(50) DEFAULT NULL COMMENT 'Programming language of this submission'
        AFTER `answer`,
    ADD COLUMN `validated` tinyint(1) DEFAULT NULL
        COMMENT 'Whether it is considered "validated" (above validation threshold for this item)'
        AFTER `score`,
    RENAME COLUMN `created_at` TO `submitted_at`,
    MODIFY COLUMN `type` enum('Submission','Saved','Current') NOT NULL DEFAULT 'Submission' COMMENT '';

UPDATE `users_answers` SET `validated` = (`score` = 100);
