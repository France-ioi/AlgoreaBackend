-- +migrate Up
ALTER TABLE `groups_attempts`
    ADD COLUMN `best_answer_id` BIGINT(20) DEFAULT NULL
        COMMENT 'First answer which scored the best score for this attempt'
        AFTER `best_answer_at`,
    ADD COLUMN `score_edit_rule` ENUM('set', 'diff') DEFAULT NULL
        COMMENT 'Whether the edit value replaces and adds up to the score of the best answer'
        AFTER `score_reeval`,
    ADD COLUMN `score_edit_value` FLOAT DEFAULT NULL
        COMMENT 'Score which overrides or adds up (depending on score_edit_rule) to the score obtained from best answer or propagation'
        AFTER `score_edit_rule`,
    ADD CONSTRAINT `cs_groups_attempts_score_edit_value_is_valid` CHECK (IFNULL(`score_edit_value`, 0) BETWEEN -100 AND 100),
    CHANGE COLUMN `score_diff_comment` `score_edit_comment` varchar(200) DEFAULT NULL
        COMMENT 'Explanation of the value set in score_edit_value',
    ADD CONSTRAINT `fk_groups_attempts_best_answer_id_users_answers_id`
        FOREIGN KEY (`best_answer_id`) REFERENCES `users_answers`(`id`) ON DELETE SET NULL;

ALTER TABLE `users_answers` ADD INDEX `best_for_attempt` (`attempt_id`, `type`, `score` DESC, `graded_at` ASC);
UPDATE groups_attempts ga SET best_answer_id = (
    SELECT ua.id FROM users_answers ua
    WHERE type = 'Submission' AND ua.attempt_id=ga.id AND ua.score IS NOT NULL AND
          graded_at IS NOT NULL
    ORDER BY score DESC, graded_at ASC
    LIMIT 1
);
ALTER TABLE `users_answers` DROP INDEX `best_for_attempt`;

# 79 rows
UPDATE `groups_attempts` SET `score_edit_rule` = 'set', `score_edit_value` = 0
WHERE `score_diff_manual` != 0 AND `score_diff_manual` = -`score_computed`;

# 1 row
UPDATE `groups_attempts` SET `score_edit_rule` = 'diff', `score_edit_value` = -40
WHERE `score_diff_manual` = -40;

# 83 rows
UPDATE `groups_attempts` SET `score` = 100 WHERE `score` > 100;

UPDATE `groups_attempts`
    JOIN `items` on `items`.`id` = `groups_attempts`.`item_id`
    LEFT JOIN `users_answers` AS `best_answer` ON `best_answer`.`id` = `groups_attempts`.`best_answer_id`
SET `groups_attempts`.`score_edit_value` = `groups_attempts`.`score`,
    `groups_attempts`.`score_edit_rule` = 'set',
    `groups_attempts`.`score_edit_comment` = 'Different scores in attempt and answer at migration'
WHERE `groups_attempts`.`score_edit_rule` IS NULL AND `items`.`type` = 'Task' AND
    ((`best_answer`.`id` IS NULL AND `groups_attempts`.`score` > 0) OR
     (`best_answer`.`id` IS NOT NULL AND `best_answer`.`score` != `groups_attempts`.`score`));

UPDATE `groups_attempts` SET `score_edit_comment` = NULL WHERE `score_edit_rule` IS NULL;

ALTER TABLE `groups_attempts`
    DROP COLUMN `score_diff_manual`,
    DROP COLUMN `score_computed`,
    CHANGE COLUMN `score` `score_computed` FLOAT NOT NULL DEFAULT '0'
        COMMENT 'Score computed from the best answer or by propagation, with score_edit_rule applied',
    ADD CONSTRAINT `cs_groups_attempts_score_computed_is_valid` CHECK (`score_computed` BETWEEN 0 AND 100),
    DROP INDEX `group_item_score_desc_best_answer_at`,
    ADD INDEX `group_item_score_computed_desc` (`group_id`,`item_id`,`score_computed` DESC),
    DROP COLUMN `best_answer_at`,
    DROP COLUMN `score_reeval`;

-- +migrate Down
UPDATE `groups_attempts` SET `score_edit_comment` = '' WHERE `score_edit_comment` IS NULL;

ALTER TABLE `groups_attempts`
    DROP FOREIGN KEY `fk_groups_attempts_best_answer_id_users_answers_id`,
    DROP CHECK `cs_groups_attempts_score_computed_is_valid`,
    ADD COLUMN `score_reeval` FLOAT DEFAULT '0'
        COMMENT 'Score computed during a reevaluation. This field allows to do a reevaluation, continue it if interrupted, and then apply the reevaluated score when/if we want.'
        AFTER `score_computed`,
    ADD COLUMN `score_diff_manual` FLOAT NOT NULL DEFAULT '0'
        COMMENT 'How much did we manually add to the computed score' AFTER `score_reeval`,
    CHANGE COLUMN `score_edit_comment` `score_diff_comment` VARCHAR(200) NOT NULL DEFAULT ''
        COMMENT 'Reason why the score was manually changed',
    ADD COLUMN `best_answer_at` DATETIME DEFAULT NULL COMMENT 'When this group obtained its best score so far on this attempt.'
        AFTER `latest_activity_at`,
    DROP INDEX `group_item_score_computed_desc`;

ALTER TABLE `groups_attempts`
    CHANGE COLUMN `score_computed` `score` FLOAT NOT NULL DEFAULT '0' COMMENT 'Current score for this attempt',
    ADD COLUMN `score_computed` FLOAT NOT NULL DEFAULT '0'
        COMMENT 'Score computed for this attempt (may be manually overriden)'
        AFTER `score`;

UPDATE `groups_attempts`
    LEFT JOIN `users_answers` ON `users_answers`.`id` = `groups_attempts`.`best_answer_id`
SET `groups_attempts`.`best_answer_at` = `users_answers`.`graded_at`,
    `groups_attempts`.`score_computed` = IF(
        `groups_attempts`.`score_edit_rule` IS NOT NULL, IFNULL(`users_answers`.`score`, 0), 0
    ),
    `groups_attempts`.`score` = CASE
        WHEN `groups_attempts`.`score_edit_rule` = 'diff'
            THEN `groups_attempts`.`score_computed` + `groups_attempts`.`score_edit_value`
        WHEN `groups_attempts`.`score_edit_rule` = 'set'
            THEN `groups_attempts`.`score_edit_value`
        ELSE
            `groups_attempts`.`score`
        END,
    `groups_attempts`.`score_diff_manual` = CASE
        WHEN `groups_attempts`.`score_edit_rule` = 'diff'
            THEN `groups_attempts`.`score_edit_value`
        WHEN `groups_attempts`.`score_edit_rule` = 'set'
            THEN `groups_attempts`.`score_computed` - `groups_attempts`.`score`
        ELSE
            0
        END;

ALTER TABLE `groups_attempts`
    DROP COLUMN `score_edit_rule`,
    DROP COLUMN `score_edit_value`,
    DROP COLUMN `best_answer_id`,
    ADD INDEX `group_item_score_desc_best_answer_at` (`group_id`, `item_id`, `score` DESC, `best_answer_at`);
