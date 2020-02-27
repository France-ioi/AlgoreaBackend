-- +migrate Up
DELETE `attempts` FROM `attempts` LEFT JOIN `groups` ON `groups`.`id` = `attempts`.`group_id`
WHERE `groups`.`id` IS NULL;

UPDATE `attempts` SET `latest_activity_at` = NULLIF(GREATEST(
        IFNULL(`started_at`, '1000-01-01 00:00:00'),
        IFNULL(`latest_answer_at`, '1000-01-01 00:00:00'),
        IFNULL(`latest_hint_at`, '1000-01-01 00:00:00'),
        IFNULL(`entered_at`, '1000-01-01 00:00:00')
    ), '1000-01-01 00:00:00')
WHERE `latest_activity_at` IS NULL;

CREATE TEMPORARY TABLE `max_latest_activity_at` (
    `id` BIGINT(20) NOT NULL PRIMARY KEY,
    `max_of_descendants` DATETIME NOT NULL
)
SELECT `attempts`.`id`, (
    SELECT MAX(`latest_activity_at`)
    FROM `attempts` AS `descendants`
        JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `descendants`.`item_id`
    WHERE `items_ancestors`.`ancestor_item_id` = `attempts`.`item_id` AND
          `descendants`.`group_id` = `attempts`.`group_id`
    ) AS `max_of_descendants` FROM `attempts`
WHERE `latest_activity_at` IS NULL HAVING `max_of_descendants` IS NOT NULL;

UPDATE `attempts`
    JOIN `max_latest_activity_at` ON `max_latest_activity_at`.`id` = `attempts`.`id`
SET `latest_activity_at` = `max_latest_activity_at`.`max_of_descendants`
WHERE `latest_activity_at` IS NULL;

DROP TEMPORARY TABLE `max_latest_activity_at`;

CREATE TEMPORARY TABLE `max_latest_activity_at` (
    `id` BIGINT(20) NOT NULL PRIMARY KEY,
    `max_of_ancestors` DATETIME NOT NULL
)
SELECT `attempts`.`id`, (
    SELECT MAX(`latest_activity_at`)
    FROM `attempts` AS `ancestors`
        JOIN `items_ancestors` ON `items_ancestors`.`ancestor_item_id` = `ancestors`.`item_id`
    WHERE `items_ancestors`.`child_item_id` = `attempts`.`item_id` AND
          `ancestors`.`group_id` = `attempts`.`group_id`
    ) AS `max_of_ancestors` FROM `attempts`
WHERE `latest_activity_at` IS NULL HAVING `max_of_ancestors` IS NOT NULL;

UPDATE `attempts`
    JOIN `max_latest_activity_at` ON `max_latest_activity_at`.`id` = `attempts`.`id`
SET `latest_activity_at` = `max_latest_activity_at`.`max_of_ancestors`
WHERE `latest_activity_at` IS NULL;

DROP TEMPORARY TABLE `max_latest_activity_at`;

UPDATE `attempts` SET `latest_activity_at` = '2020-01-01 00:00:00' WHERE `latest_activity_at` IS NULL;

ALTER TABLE `attempts`
    MODIFY COLUMN `creator_id` BIGINT(20) DEFAULT NULL
        COMMENT 'The user who created this attempt. NULL if created by propagation',
    MODIFY COLUMN `started_at` DATETIME DEFAULT NULL
        COMMENT 'Time at which the attempt was manually created or was first marked as started (should be when it is first visited). Not propagated',
    MODIFY COLUMN `validated_at` DATETIME DEFAULT NULL
        COMMENT 'Submission time of the first answer that made the attempt validated',
    MODIFY COLUMN `validated` TINYINT(1) GENERATED ALWAYS AS (`validated_at` is not null) VIRTUAL NOT NULL
        COMMENT 'Auto-generated from `validated_at`',
    MODIFY COLUMN `latest_hint_at` DATETIME DEFAULT NULL
        COMMENT 'Time of the last request for a hint. Only for tasks, not propagated',
    CHANGE COLUMN `latest_answer_at` `latest_submission_at` DATETIME DEFAULT NULL
        COMMENT 'Time of the latest submission. Only for tasks, not propagated',
    MODIFY COLUMN `submissions` INT(11) NOT NULL DEFAULT '0'
        COMMENT 'Number of submissions. Only for tasks, not propagated',
    MODIFY COLUMN `latest_activity_at` DATETIME NOT NULL DEFAULT NOW()
        COMMENT 'Time of the latest activity (attempt creation, submission, hint request) of a user on this attempt or its children',
    DROP COLUMN `tasks_solved`,
    DROP COLUMN `finished`,
    DROP COLUMN `children_validated`,
    DROP COLUMN `entered_at`;

-- +migrate Down
ALTER TABLE `attempts`
    MODIFY COLUMN `creator_id` BIGINT(20) DEFAULT NULL,
    MODIFY COLUMN `started_at` DATETIME DEFAULT NULL
        COMMENT 'When the item was started for this attempt. Can be different from when the attempt was created.',
    MODIFY COLUMN `validated_at` DATETIME DEFAULT NULL
        COMMENT 'Submission time of the first answer that made the item validated within this attempt (validation criteria depends on item)',
    MODIFY COLUMN `validated` TINYINT(1) GENERATED ALWAYS AS (`validated_at` is not null) VIRTUAL NOT NULL
        COMMENT 'See `validated_at`',
    MODIFY COLUMN `latest_hint_at` DATETIME DEFAULT NULL
        COMMENT 'When the last hint has been obtained for this attempt',
    CHANGE COLUMN `latest_submission_at` `latest_answer_at` DATETIME DEFAULT NULL
        COMMENT 'When the last answer was provided by this group on this attempt',
    MODIFY COLUMN `submissions` INT(11) NOT NULL DEFAULT '0'
        COMMENT 'Number of submissions in total for this item and its children',
    MODIFY COLUMN `latest_activity_at` DATETIME DEFAULT NULL
        COMMENT 'When the last activity within this task.',
    ADD COLUMN `tasks_solved` INT(11) NOT NULL DEFAULT '0'
        COMMENT 'Number of tasks which have been solved among this item''s descendants (at least one submission), within this attempt'
            AFTER `tasks_tried`,
    ADD COLUMN `finished` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether this item is finished within this attempt (max score obtained for a task, every child finished for a chapter)'
            AFTER `validated`,
    ADD COLUMN `children_validated` INT(11) NOT NULL DEFAULT '0'
        COMMENT 'Number of items which have been validated among this item''s direct children (solved, or read, or chapters marked as validated), within this attempt'
            AFTER `tasks_solved`,
    ADD COLUMN `entered_at` DATETIME DEFAULT NULL
        COMMENT 'Time at which the group entered the contest'
            AFTER `hints_cached`;
