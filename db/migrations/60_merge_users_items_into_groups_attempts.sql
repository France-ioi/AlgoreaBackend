-- +migrate Up
ALTER TABLE `groups_attempts` ADD INDEX `group_id_item_id_order` (`group_id`, `item_id`, `order`);

-- Copy users data for all the items
INSERT INTO `groups_attempts` (
    `group_id`, `item_id`,
    `creator_user_id`,
    `order`,
    `score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,
    `submissions_attempts`,`tasks_tried`,`tasks_solved`,`children_validated`,`validated`,`finished`,
    `key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,
    `autonomy`,`started_at`,`validated_at`,`finished_at`,`latest_activity_at`,`thread_started_at`,
    `best_answer_at`,`latest_answer_at`,`latest_hint_at`,`ranked`,`all_lang_prog`)
SELECT
    `users`.`self_group_id`, `item_id`,
    `users`.`id`,
    (SELECT IFNULL(MAX(`order`)+1, 1) FROM `groups_attempts` WHERE `group_id` = `users`.`self_group_id` AND `groups_attempts`.`item_id` = `users_items`.`item_id`),
    `users_items`.`score`,`users_items`.`score_computed`,`users_items`.`score_reeval`,
    `users_items`.`score_diff_manual`,`users_items`.`score_diff_comment`,
    `users_items`.`submissions_attempts`,`users_items`.`tasks_tried`,`users_items`.`tasks_solved`,
    `users_items`.`children_validated`,`users_items`.`validated`,`users_items`.`finished`,
    `users_items`.`key_obtained`,`users_items`.`tasks_with_help`,`users_items`.`hints_requested`,
    `users_items`.`hints_cached`,`users_items`.`corrections_read`,`users_items`.`precision`,
    `users_items`.`autonomy`,`users_items`.`started_at`,`users_items`.`validated_at`,
    `users_items`.`finished_at`,`users_items`.`latest_activity_at`,`users_items`.`thread_started_at`,
    `users_items`.`best_answer_at`,`users_items`.`latest_answer_at`,`users_items`.`latest_hint_at`,
    `users_items`.`ranked`,`users_items`.`all_lang_prog`
FROM `users_items`
    JOIN `users` ON `users`.`id` = `users_items`.`user_id`
    JOIN `items` ON `items`.`id` = `users_items`.`item_id` AND NOT `items`.`has_attempts`;

ALTER TABLE `groups_attempts` DROP INDEX `group_id_item_id_order`;

ALTER TABLE `users_items`
    DROP COLUMN `score`,
    DROP COLUMN `score_computed`,
    DROP COLUMN `score_reeval`,
    DROP COLUMN `score_diff_manual`,
    DROP COLUMN `score_diff_comment`,
    DROP COLUMN `submissions_attempts`,
    DROP COLUMN `tasks_tried`,
    DROP COLUMN `tasks_solved`,
    DROP COLUMN `children_validated`,
    DROP COLUMN `validated`,
    DROP COLUMN `finished`,
    DROP COLUMN `key_obtained`,
    DROP COLUMN `tasks_with_help`,
    DROP COLUMN `hints_requested`,
    DROP COLUMN `hints_cached`,
    DROP COLUMN `corrections_read`,
    DROP COLUMN `precision`,
    DROP COLUMN `autonomy`,
    DROP COLUMN `started_at`,
    DROP COLUMN `validated_at`,
    DROP COLUMN `finished_at`,
    DROP COLUMN `latest_activity_at`,
    DROP COLUMN `thread_started_at`,
    DROP COLUMN `best_answer_at`,
    DROP COLUMN `latest_answer_at`,
    DROP COLUMN `latest_hint_at`,
    DROP COLUMN `ranked`,
    DROP COLUMN `all_lang_prog`,
    DROP COLUMN `ancestors_computation_state`,
    DROP COLUMN `state`,
    DROP COLUMN `answer`,
    DROP COLUMN `platform_data_removed`;

DROP VIEW `task_children_data_view`;

-- +migrate Down
ALTER TABLE `users_items`
    ADD COLUMN `score` float NOT NULL DEFAULT '0'
        COMMENT 'Current score for this attempt ; can be a cached computation'
        AFTER `active_attempt_id`,
    ADD COLUMN `score_computed` float NOT NULL DEFAULT '0' COMMENT 'Deprecated'
        AFTER `score`,
    ADD COLUMN `score_reeval` float DEFAULT '0' COMMENT 'Deprecated'
        AFTER `score_computed`,
    ADD COLUMN `score_diff_manual` float NOT NULL DEFAULT '0'
        COMMENT 'How much did we manually add to the computed score'
        AFTER `score_reeval`,
    ADD COLUMN `score_diff_comment` varchar(200) NOT NULL DEFAULT ''
        COMMENT 'Why was the score manually changed ?'
        AFTER `score_diff_manual`,
    ADD COLUMN `submissions_attempts` int(11) NOT NULL DEFAULT '0'
        COMMENT 'How many submissions in total for this item and its children?'
        AFTER `score_diff_comment`,
    ADD COLUMN `tasks_tried` int(11) NOT NULL DEFAULT '0' COMMENT 'Deprecated'
        AFTER `submissions_attempts`,
    ADD COLUMN `tasks_solved` int(11) NOT NULL DEFAULT '0' COMMENT 'Deprecated'
        AFTER `tasks_tried`,
    ADD COLUMN `children_validated` int(11) NOT NULL DEFAULT '0' COMMENT 'Deprecated'
        AFTER `tasks_solved`,
    ADD COLUMN `validated` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Deprecated'
        AFTER `children_validated`,
    ADD COLUMN `finished` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Deprecated'
        AFTER `validated`,
    ADD COLUMN `key_obtained` tinyint(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether the user obtained the key on this item. Changed to 1 if the user gets a score >= items.score_min_unlock, will grant access to new item from items.unlocked_item_ids. This information is propagated to users_items.'
        AFTER `finished`,
    ADD COLUMN `tasks_with_help` int(11) NOT NULL DEFAULT '0'
        COMMENT 'For how many of this item''s descendants tasks within this attempts, did the user ask for hints (or help on the forum - not implemented)?'
        AFTER `key_obtained`,
    ADD COLUMN `hints_requested` mediumtext COMMENT 'Deprecated'
        AFTER `tasks_with_help`,
    ADD COLUMN `hints_cached` int(11) NOT NULL DEFAULT '0' COMMENT 'Deprecated'
        AFTER `hints_requested`,
    ADD COLUMN `corrections_read` int(11) NOT NULL DEFAULT '0'
        COMMENT 'Number of solutions the user read among the descendants of this item.'
        AFTER `hints_cached`,
    ADD COLUMN `precision` int(11) NOT NULL DEFAULT '0'
        COMMENT 'Precision (based on a formula to be defined) of the user recently, when working on this item and its descendants.'
        AFTER `corrections_read`,
    ADD COLUMN `autonomy` int(11) NOT NULL DEFAULT '0'
        COMMENT 'Autonomy (based on a formula to be defined) of the user was recently, when working on this item and its descendants: how much help / hints he used.'
        AFTER `precision`,
    ADD COLUMN `started_at` datetime DEFAULT NULL COMMENT 'Deprecated'
        AFTER `autonomy`,
    ADD COLUMN `validated_at` datetime DEFAULT NULL COMMENT 'Deprecated'
        AFTER `started_at`,
    ADD COLUMN `finished_at` datetime DEFAULT NULL COMMENT 'Deprecated'
        AFTER `validated_at`,
    ADD COLUMN `latest_activity_at` datetime DEFAULT NULL
        COMMENT 'When was the last activity within this task.'
        AFTER `finished_at`,
    ADD COLUMN `thread_started_at` datetime DEFAULT NULL
        COMMENT 'When was a discussion thread started by this user/group on the forum'
        AFTER `latest_activity_at`,
    ADD COLUMN `best_answer_at` datetime DEFAULT NULL COMMENT 'Deprecated'
        AFTER `thread_started_at`,
    ADD COLUMN `latest_answer_at` datetime DEFAULT NULL COMMENT 'Deprecated'
        AFTER `best_answer_at`,
    ADD COLUMN `latest_hint_at` datetime DEFAULT NULL COMMENT 'Deprecated'
        AFTER `latest_answer_at`,
    ADD COLUMN `ranked` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Deprecated'
        AFTER `latest_hint_at`,
    ADD COLUMN `all_lang_prog` varchar(200) DEFAULT NULL
        COMMENT 'List of programming languages used'
        AFTER `ranked`,
    ADD COLUMN `ancestors_computation_state` enum('done','processing','todo','temp') NOT NULL DEFAULT 'todo'
        COMMENT 'Used to denote whether the ancestors data have to be recomputed (after this item''s score was changed for instance)'
        AFTER `all_lang_prog`,
    ADD COLUMN `state` mediumtext COMMENT 'Deprecated'
        AFTER `ancestors_computation_state`,
    ADD COLUMN `answer` mediumtext COMMENT 'Deprecated'
        AFTER `state`,
    ADD COLUMN `platform_data_removed` tinyint(4) NOT NULL DEFAULT '0'
        AFTER `answer`;

INSERT INTO `users_items` (
    `user_id`, `item_id`,
    `score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,
    `submissions_attempts`,`tasks_tried`,`tasks_solved`,`children_validated`,`validated`,`finished`,
    `key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,
    `autonomy`,`started_at`,`validated_at`,`finished_at`,`latest_activity_at`,`thread_started_at`,
    `best_answer_at`,`latest_answer_at`,`latest_hint_at`,`ranked`,`all_lang_prog`)
SELECT
    `users`.`id`, `item_id`,
    `groups_attempts`.`score`,`groups_attempts`.`score_computed`,`groups_attempts`.`score_reeval`,
    `groups_attempts`.`score_diff_manual`,`groups_attempts`.`score_diff_comment`,
    `groups_attempts`.`submissions_attempts`,`groups_attempts`.`tasks_tried`,
    `groups_attempts`.`tasks_solved`,`groups_attempts`.`children_validated`,
    `groups_attempts`.`validated`,`groups_attempts`.`finished`,
    `groups_attempts`.`key_obtained`,`groups_attempts`.`tasks_with_help`,
    `groups_attempts`.`hints_requested`,`groups_attempts`.`hints_cached`,
    `groups_attempts`.`corrections_read`,`groups_attempts`.`precision`,
    `groups_attempts`.`autonomy`,`groups_attempts`.`started_at`,`groups_attempts`.`validated_at`,
    `groups_attempts`.`finished_at`,`groups_attempts`.`latest_activity_at`,
    `groups_attempts`.`thread_started_at`,`groups_attempts`.`best_answer_at`,
    `groups_attempts`.`latest_answer_at`,`groups_attempts`.`latest_hint_at`,
    `groups_attempts`.`ranked`,`groups_attempts`.`all_lang_prog`
FROM `groups_attempts`
    JOIN `users` ON `users`.`id` = `groups_attempts`.`creator_user_id`
ON DUPLICATE KEY UPDATE
    `score`=VALUES(`score`),`score_computed`=VALUES(`score_computed`),
    `score_reeval`=VALUES(`score_reeval`),`score_diff_manual`=VALUES(`score_diff_manual`),
    `score_diff_comment`=VALUES(`score_diff_comment`),
    `submissions_attempts`=VALUES(`submissions_attempts`),`tasks_tried`=VALUES(`tasks_tried`),
    `tasks_solved`=VALUES(`tasks_solved`),`children_validated`=VALUES(`children_validated`),
    `validated`=VALUES(`validated`),`finished`=VALUES(`finished`),
    `key_obtained`=VALUES(`key_obtained`),`tasks_with_help`=VALUES(`tasks_with_help`),
    `hints_requested`=VALUES(`hints_requested`),`hints_cached`=VALUES(`hints_cached`),
    `corrections_read`=VALUES(`corrections_read`),`precision`=VALUES(`precision`),
    `autonomy`=VALUES(`autonomy`),`started_at`=VALUES(`started_at`),
    `validated_at`=VALUES(`validated_at`),`finished_at`=VALUES(`finished_at`),
    `latest_activity_at`=VALUES(`latest_activity_at`),`thread_started_at`=VALUES(`thread_started_at`),
    `best_answer_at`=VALUES(`best_answer_at`),`latest_answer_at`=VALUES(`latest_answer_at`),
    `latest_hint_at`=VALUES(`latest_hint_at`),`ranked`=VALUES(`ranked`),
    `all_lang_prog`=VALUES(`all_lang_prog`);

ALTER TABLE `users_answers` ADD INDEX `user_item_submitted_at_desc` (`user_id`, `item_id`, `submitted_at` DESC);
UPDATE `users_items`
JOIN LATERAL (
    SELECT `users_answers`.`answer`, `users_answers`.`state`
    FROM `users_answers`
    WHERE `users_answers`.`user_id` = `users_items`.`user_id` AND `users_answers`.`item_id` = `users_items`.`item_id`
    ORDER BY `users_answers`.`user_id`, `users_answers`.`item_id`, `users_answers`.`submitted_at`
    LIMIT 1
) AS `users_answers`
SET `users_items`.`answer`=`users_answers`.`answer`, `users_items`.`state`=`users_answers`.`state`;
ALTER TABLE `users_answers` DROP INDEX `user_item_submitted_at_desc`;

DELETE `groups_attempts` FROM `groups_attempts`
    JOIN `users` ON `users`.`self_group_id` = `groups_attempts`.`group_id`;

CREATE ALGORITHM=UNDEFINED
    SQL SECURITY DEFINER
    VIEW `task_children_data_view` AS
SELECT
    `parent_users_items`.`id` AS `user_item_id`,
    SUM(IF(`task_children`.`id` IS NOT NULL AND `task_children`.`validated`, 1, 0)) AS `children_validated`,
    SUM(IF(`task_children`.`id` IS NOT NULL AND `task_children`.`validated`, 0, 1)) AS `children_non_validated`,
    SUM(IF(`items_items`.`category` = 'Validation' AND
        (ISNULL(`task_children`.`id`) OR `task_children`.`validated` != 1), 1, 0)) AS `children_category`,
    MAX(`task_children`.`validated_at`) AS `max_validated_at`,
    MAX(IF(`items_items`.`category` = 'Validation', `task_children`.`validated_at`, NULL)) AS `max_validated_at_categories`
FROM `users_items` AS `parent_users_items`
    JOIN `items_items` ON(
        `parent_users_items`.`item_id` = `items_items`.`parent_item_id`
    )
    LEFT JOIN `users_items` AS `task_children` ON(
        `items_items`.`child_item_id` = `task_children`.`item_id` AND
        `task_children`.`user_id` = `parent_users_items`.`user_id`
    )
    JOIN `items` ON(
        `items`.`ID` = `items_items`.`child_item_id`
    )
WHERE `items`.`type` <> 'Course' AND `items`.`no_score` = 0
GROUP BY `user_item_id`;
