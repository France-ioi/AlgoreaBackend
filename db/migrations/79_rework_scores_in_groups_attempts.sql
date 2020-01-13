-- +migrate Up
ALTER TABLE `groups_attempts`
    CHANGE COLUMN `best_answer_at` `score_obtained_at` DATETIME DEFAULT NULL
        COMMENT 'Submission time of the first answer which led to the best score',
    RENAME INDEX `group_item_score_desc_best_answer_at` TO `group_item_score_desc_score_obtained_at`,
    ADD COLUMN `score_edit_rule` ENUM('set', 'diff') DEFAULT NULL
        COMMENT 'Whether the edit value replaces and adds up to the score of the best answer'
        AFTER `score_reeval`,
    ADD COLUMN `score_edit_value` FLOAT DEFAULT NULL
        COMMENT 'Score which overrides or adds up (depending on score_edit_rule) to the score obtained from best answer or propagation'
        AFTER `score_edit_rule`,
    ADD CONSTRAINT `cs_groups_attempts_score_edit_value_is_valid` CHECK (IFNULL(`score_edit_value`, 0) BETWEEN -100 AND 100),
    CHANGE COLUMN `score_diff_comment` `score_edit_comment` varchar(200) DEFAULT NULL
        COMMENT 'Explanation of the value set in score_edit_value',
    MODIFY COLUMN `validated_at` DATETIME DEFAULT NULL
        COMMENT 'Submission time of the first answer that made the item validated within this attempt (validation criteria depends on item)';

# 79 rows
UPDATE `groups_attempts` SET `score_edit_rule` = 'set', `score_edit_value` = 0
WHERE `score_diff_manual` != 0 AND `score_diff_manual` = -`score_computed`;

# 1 row
UPDATE `groups_attempts` SET `score_edit_rule` = 'diff', `score_edit_value` = -40
WHERE `score_diff_manual` = -40 AND `score_computed` = 100 AND `score` = 60;

# 83 rows
UPDATE `groups_attempts` SET `score` = 100 WHERE `score` > 100;

# 168,905 rows :/
UPDATE `groups_attempts`
    JOIN `items` on `items`.`id` = `groups_attempts`.`item_id`
    LEFT JOIN LATERAL (
        SELECT MAX(`score`) AS `score`
        FROM `users_answers`
        WHERE `users_answers`.`attempt_id` = `groups_attempts`.`id` AND `users_answers`.`type` = 'Submission' AND
              `users_answers`.`graded_at` IS NOT NULL
    ) AS `best_score` ON 1
SET `groups_attempts`.`score_edit_value` = `groups_attempts`.`score`,
    `groups_attempts`.`score_edit_rule` = 'set',
    `groups_attempts`.`score_edit_comment` = 'Different scores in attempt and answer at migration'
WHERE `groups_attempts`.`score_edit_rule` IS NULL AND `items`.`type` = 'Task' AND
    IFNULL(`best_score`.`score`, 0) != `groups_attempts`.`score`;

UPDATE `groups_attempts` SET `score_edit_comment` = NULL WHERE `score_edit_rule` IS NULL;

ALTER TABLE `groups_attempts`
    DROP COLUMN `score_diff_manual`,
    DROP COLUMN `score_computed`,
    CHANGE COLUMN `score` `score_computed` FLOAT NOT NULL DEFAULT '0'
        COMMENT 'Score computed from the best answer or by propagation, with score_edit_rule applied',
    ADD CONSTRAINT `cs_groups_attempts_score_computed_is_valid` CHECK (`score_computed` BETWEEN 0 AND 100),
    DROP COLUMN `score_reeval`;

-- +migrate Down
UPDATE `groups_attempts` SET `score_edit_comment` = '' WHERE `score_edit_comment` IS NULL;

ALTER TABLE `groups_attempts`
    CHANGE COLUMN `score_obtained_at` `best_answer_at` DATETIME DEFAULT NULL
        COMMENT 'When this group obtained its best score so far on this attempt.',
    RENAME INDEX `group_item_score_desc_score_obtained_at` TO `group_item_score_desc_best_answer_at`,
    DROP CHECK `cs_groups_attempts_score_computed_is_valid`,
    ADD COLUMN `score_reeval` FLOAT DEFAULT '0'
        COMMENT 'Score computed during a reevaluation. This field allows to do a reevaluation, continue it if interrupted, and then apply the reevaluated score when/if we want.'
        AFTER `score_computed`,
    ADD COLUMN `score_diff_manual` FLOAT NOT NULL DEFAULT '0'
        COMMENT 'How much did we manually add to the computed score' AFTER `score_reeval`,
    CHANGE COLUMN `score_edit_comment` `score_diff_comment` VARCHAR(200) NOT NULL DEFAULT ''
        COMMENT 'Reason why the score was manually changed',
    MODIFY COLUMN `validated_at` DATETIME DEFAULT NULL
        COMMENT 'When the item was validated within this attempt (validation criteria depends on item)';

ALTER TABLE `groups_attempts`
    CHANGE COLUMN `score_computed` `score` FLOAT NOT NULL DEFAULT '0' COMMENT 'Current score for this attempt',
    ADD COLUMN `score_computed` FLOAT NOT NULL DEFAULT '0'
        COMMENT 'Score computed for this attempt (may be manually overriden)'
        AFTER `score`;

UPDATE `groups_attempts`
    LEFT JOIN LATERAL (
        SELECT MAX(`score`) AS `score`
        FROM `users_answers`
        WHERE `users_answers`.`attempt_id` = `groups_attempts`.`id` AND `users_answers`.`type` = 'Submission' AND
            `users_answers`.`graded_at` IS NOT NULL
        ) AS `best_score` ON 1
SET `groups_attempts`.`score_computed` = IF(
        `groups_attempts`.`score_edit_rule` IS NOT NULL, IFNULL(`best_score`.`score`, 0), 0
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
    DROP COLUMN `score_edit_value`;
