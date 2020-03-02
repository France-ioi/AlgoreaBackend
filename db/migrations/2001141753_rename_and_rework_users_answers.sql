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
    MODIFY COLUMN `type` enum('Submission','Saved','Current') NOT NULL
        COMMENT '\'Submission\' for answers submitted for grading, \'Saved\' for manual backups of answers, \'Current\' for automatic snapshots of the latest answers (unique for a user on an attempt)';

CREATE TABLE `gradings` (
    `answer_id` BIGINT(20) NOT NULL,
    `score` FLOAT NOT NULL COMMENT 'Score obtained',
    `graded_at` DATETIME NOT NULL COMMENT 'When was it last graded',
    PRIMARY KEY (`answer_id`),
    CONSTRAINT `fk_submissions_answer_id_answers_id`
       FOREIGN KEY (`answer_id`) REFERENCES `answers`(`id`) ON DELETE CASCADE
) COMMENT 'Grading results for answers' ENGINE=InnoDB DEFAULT CHARSET=utf8;;

INSERT INTO `gradings` (answer_id, score, graded_at)
SELECT `id`, `score`, IFNULL(`graded_at`, `created_at`) FROM `answers`
WHERE `type` = 'Submission' AND `score` IS NOT NULL;

ALTER TABLE `answers`
    DROP COLUMN `score`,
    DROP COLUMN `graded_at`;

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
    ADD COLUMN `score` FLOAT DEFAULT NULL COMMENT 'Score obtained' AFTER `submitted_at`,
    ADD COLUMN `graded_at` DATETIME DEFAULT NULL COMMENT 'When was it last graded' AFTER `score`,
    ADD COLUMN `validated` tinyint(1) DEFAULT NULL
        COMMENT 'Whether it is considered "validated" (above validation threshold for this item)'
        AFTER `score`,
    RENAME COLUMN `created_at` TO `submitted_at`,
    MODIFY COLUMN `type` enum('Submission','Saved','Current') NOT NULL DEFAULT 'Submission' COMMENT '';

UPDATE `users_answers` JOIN `gradings` ON `gradings`.`answer_id` = `users_answers`.`id`
SET `users_answers`.`score` = `gradings`.`score`,
    `users_answers`.`graded_at` = `gradings`.`graded_at`,
    `users_answers`.`validated` = (`users_answers`.`score` = 100);

DROP TABLE `gradings`;
