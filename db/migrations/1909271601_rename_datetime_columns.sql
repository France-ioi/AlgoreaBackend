-- +migrate Up
ALTER TABLE `groups`
    RENAME COLUMN `date_created` TO `created_at`,
    RENAME COLUMN `code_end` TO `code_expires_at`,
    RENAME COLUMN `lock_user_deletion_date` TO `lock_user_deletion_until`;
ALTER TABLE `groups_attempts`
    RENAME COLUMN `start_date` TO `started_at`,
    RENAME COLUMN `validation_date` TO `validated_at`,
    RENAME COLUMN `finish_date` TO `finished_at`,
    RENAME COLUMN `last_activity_date` TO `latest_activity_at`,
    RENAME COLUMN `thread_start_date` TO `thread_started_at`,
    RENAME COLUMN `best_answer_date` TO `best_answer_at`,
    RENAME COLUMN `last_answer_date` TO `latest_answer_at`,
    RENAME COLUMN `last_hint_date` TO `latest_hint_at`,
    RENAME COLUMN `contest_start_date` TO `contest_started_at`;
ALTER TABLE `groups_groups`
    RENAME COLUMN `status_date` TO `type_changed_at`;
ALTER TABLE `groups_items`
    RENAME COLUMN `partial_access_date` TO `partial_access_since`,
    RENAME COLUMN `full_access_date` TO `full_access_since`,
    RENAME COLUMN `access_solutions_date` TO `solutions_access_since`,
    RENAME COLUMN `cached_full_access_date` TO `cached_full_access_since`,
    RENAME COLUMN `cached_partial_access_date` TO `cached_partial_access_since`,
    RENAME COLUMN `cached_access_solutions_date` TO `cached_solutions_access_since`,
    RENAME COLUMN `cached_grayed_access_date` TO `cached_grayed_access_since`;
ALTER TABLE `history_groups`
    RENAME COLUMN `date_created` TO `created_at`,
    RENAME COLUMN `code_end` TO `code_expires_at`,
    RENAME COLUMN `lock_user_deletion_date` TO `lock_user_deletion_until`;
ALTER TABLE `history_groups_attempts`
    RENAME COLUMN `start_date` TO `started_at`,
    RENAME COLUMN `validation_date` TO `validated_at`,
    RENAME COLUMN `finish_date` TO `finished_at`,
    RENAME COLUMN `last_activity_date` TO `latest_activity_at`,
    RENAME COLUMN `thread_start_date` TO `thread_started_at`,
    RENAME COLUMN `best_answer_date` TO `best_answer_at`,
    RENAME COLUMN `last_answer_date` TO `latest_answer_at`,
    RENAME COLUMN `last_hint_date` TO `latest_hint_at`,
    RENAME COLUMN `contest_start_date` TO `contest_started_at`;
ALTER TABLE `history_groups_groups`
    RENAME COLUMN `status_date` TO `type_changed_at`;
ALTER TABLE `history_groups_items`
    RENAME COLUMN `partial_access_date` TO `partial_access_since`,
    RENAME COLUMN `full_access_date` TO `full_access_since`,
    RENAME COLUMN `access_solutions_date` TO `solutions_access_since`,
    RENAME COLUMN `cached_full_access_date` TO `cached_full_access_since`,
    RENAME COLUMN `cached_partial_access_date` TO `cached_partial_access_since`,
    RENAME COLUMN `cached_access_solutions_date` TO `cached_solutions_access_since`,
    RENAME COLUMN `cached_grayed_access_date` TO `cached_grayed_access_since`;
ALTER TABLE `history_items`
    RENAME COLUMN `access_open_date` TO `contest_opens_at`,
    RENAME COLUMN `end_contest_date` TO `contest_closes_at`;
ALTER TABLE `history_messages`
    RENAME COLUMN `submission_date` TO `submitted_at`;
ALTER TABLE `history_threads`
    RENAME COLUMN `last_activity_date` TO `latest_activity_at`;
ALTER TABLE `history_users`
    RENAME COLUMN `registration_date` TO `registered_at`,
    RENAME COLUMN `last_login_date` TO `latest_login_at`,
    RENAME COLUMN `last_activity_date` TO `latest_activity_at`,
    RENAME COLUMN `notification_read_date` TO `notifications_read_at`;
ALTER TABLE `history_users_items`
    RENAME COLUMN `start_date` TO `started_at`,
    RENAME COLUMN `validation_date` TO `validated_at`,
    RENAME COLUMN `finish_date` TO `finished_at`,
    RENAME COLUMN `last_activity_date` TO `latest_activity_at`,
    RENAME COLUMN `thread_start_date` TO `thread_started_at`,
    RENAME COLUMN `best_answer_date` TO `best_answer_at`,
    RENAME COLUMN `last_answer_date` TO `latest_answer_at`,
    RENAME COLUMN `last_hint_date` TO `latest_hint_at`,
    RENAME COLUMN `contest_start_date` TO `contest_started_at`;
ALTER TABLE `history_users_threads`
    RENAME COLUMN `last_read_date` TO `lately_viewed_at`,
    RENAME COLUMN `last_write_date` TO `lately_posted_at`;
ALTER TABLE `items`
    RENAME COLUMN `access_open_date` TO `contest_opens_at`,
    RENAME COLUMN `end_contest_date` TO `contest_closes_at`;
ALTER TABLE `login_states`
    RENAME COLUMN `expiration_date` TO `expires_at`,
    RENAME INDEX `expiration_date` TO `expires_at`;
ALTER TABLE `messages`
    RENAME COLUMN `submission_date` TO `submitted_at`;
ALTER TABLE `sessions`
    RENAME COLUMN `expiration_date` TO `expires_at`,
    RENAME COLUMN `issued_at_date` TO `issued_at`,
    RENAME INDEX `expiration_date` TO `expires_at`;
ALTER TABLE `threads`
    RENAME COLUMN `last_activity_date` TO `latest_activity_at`;
ALTER TABLE `users`
    RENAME COLUMN `registration_date` TO `registered_at`,
    RENAME COLUMN `last_login_date` TO `latest_login_at`,
    RENAME COLUMN `last_activity_date` TO `latest_activity_at`,
    RENAME COLUMN `notification_read_date` TO `notifications_read_at`;
ALTER TABLE `users_answers`
    RENAME COLUMN `submission_date` TO `submitted_at`,
    RENAME COLUMN `grading_date` TO `graded_at`;
ALTER TABLE `users_items`
    RENAME COLUMN `start_date` TO `started_at`,
    RENAME COLUMN `validation_date` TO `validated_at`,
    RENAME COLUMN `finish_date` TO `finished_at`,
    RENAME COLUMN `last_activity_date` TO `latest_activity_at`,
    RENAME COLUMN `thread_start_date` TO `thread_started_at`,
    RENAME COLUMN `best_answer_date` TO `best_answer_at`,
    RENAME COLUMN `last_answer_date` TO `latest_answer_at`,
    RENAME COLUMN `last_hint_date` TO `latest_hint_at`,
    RENAME COLUMN `contest_start_date` TO `contest_started_at`;
ALTER TABLE `users_threads`
    RENAME COLUMN `last_read_date` TO `lately_viewed_at`,
    RENAME COLUMN `last_write_date` TO `lately_posted_at`;


DROP TRIGGER `after_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`) VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`created_at`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_timer`,NEW.`code_expires_at`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`); INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`name` <=> NEW.`name` AND OLD.`grade` <=> NEW.`grade` AND OLD.`grade_details` <=> NEW.`grade_details` AND OLD.`description` <=> NEW.`description` AND OLD.`created_at` <=> NEW.`created_at` AND OLD.`opened` <=> NEW.`opened` AND OLD.`free_access` <=> NEW.`free_access` AND OLD.`team_item_id` <=> NEW.`team_item_id` AND OLD.`team_participating` <=> NEW.`team_participating` AND OLD.`code` <=> NEW.`code` AND OLD.`code_timer` <=> NEW.`code_timer` AND OLD.`code_expires_at` <=> NEW.`code_expires_at` AND OLD.`redirect_path` <=> NEW.`redirect_path` AND OLD.`open_contest` <=> NEW.`open_contest` AND OLD.`type` <=> NEW.`type` AND OLD.`send_emails` <=> NEW.`send_emails`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`)       VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`created_at`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_timer`,NEW.`code_expires_at`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups` BEFORE DELETE ON `groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`name`,OLD.`grade`,OLD.`grade_details`,OLD.`description`,OLD.`created_at`,OLD.`opened`,OLD.`free_access`,OLD.`team_item_id`,OLD.`team_participating`,OLD.`code`,OLD.`code_timer`,OLD.`code_expires_at`,OLD.`redirect_path`,OLD.`open_contest`,OLD.`type`,OLD.`send_emails`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_attempts` AFTER INSERT ON `groups_attempts` FOR EACH ROW BEGIN INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`contest_started_at`,`ranked`,`all_lang_prog`) VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`order`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`started_at`,NEW.`validated_at`,NEW.`best_answer_at`,NEW.`latest_answer_at`,NEW.`thread_started_at`,NEW.`latest_hint_at`,NEW.`finished_at`,NEW.`latest_activity_at`,NEW.`contest_started_at`,NEW.`ranked`,NEW.`all_lang_prog`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_attempts` BEFORE UPDATE ON `groups_attempts` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`order` <=> NEW.`order` AND OLD.`score` <=> NEW.`score` AND OLD.`score_computed` <=> NEW.`score_computed` AND OLD.`score_reeval` <=> NEW.`score_reeval` AND OLD.`score_diff_manual` <=> NEW.`score_diff_manual` AND OLD.`score_diff_comment` <=> NEW.`score_diff_comment` AND OLD.`submissions_attempts` <=> NEW.`submissions_attempts` AND OLD.`tasks_tried` <=> NEW.`tasks_tried` AND OLD.`children_validated` <=> NEW.`children_validated` AND OLD.`validated` <=> NEW.`validated` AND OLD.`finished` <=> NEW.`finished` AND OLD.`key_obtained` <=> NEW.`key_obtained` AND OLD.`tasks_with_help` <=> NEW.`tasks_with_help` AND OLD.`hints_requested` <=> NEW.`hints_requested` AND OLD.`hints_cached` <=> NEW.`hints_cached` AND OLD.`corrections_read` <=> NEW.`corrections_read` AND OLD.`precision` <=> NEW.`precision` AND OLD.`autonomy` <=> NEW.`autonomy` AND OLD.`started_at` <=> NEW.`started_at` AND OLD.`validated_at` <=> NEW.`validated_at` AND OLD.`best_answer_at` <=> NEW.`best_answer_at` AND OLD.`latest_answer_at` <=> NEW.`latest_answer_at` AND OLD.`thread_started_at` <=> NEW.`thread_started_at` AND OLD.`latest_hint_at` <=> NEW.`latest_hint_at` AND OLD.`finished_at` <=> NEW.`finished_at` AND OLD.`contest_started_at` <=> NEW.`contest_started_at` AND OLD.`ranked` <=> NEW.`ranked` AND OLD.`all_lang_prog` <=> NEW.`all_lang_prog`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups_attempts` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`contest_started_at`,`ranked`,`all_lang_prog`)       VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`order`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`started_at`,NEW.`validated_at`,NEW.`best_answer_at`,NEW.`latest_answer_at`,NEW.`thread_started_at`,NEW.`latest_hint_at`,NEW.`finished_at`,NEW.`latest_activity_at`,NEW.`contest_started_at`,NEW.`ranked`,NEW.`all_lang_prog`) ; SET NEW.minus_score = -NEW.score; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_attempts` BEFORE DELETE ON `groups_attempts` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_attempts` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`contest_started_at`,`ranked`,`all_lang_prog`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`group_id`,OLD.`item_id`,OLD.`creator_user_id`,OLD.`order`,OLD.`score`,OLD.`score_computed`,OLD.`score_reeval`,OLD.`score_diff_manual`,OLD.`score_diff_comment`,OLD.`submissions_attempts`,OLD.`tasks_tried`,OLD.`children_validated`,OLD.`validated`,OLD.`finished`,OLD.`key_obtained`,OLD.`tasks_with_help`,OLD.`hints_requested`,OLD.`hints_cached`,OLD.`corrections_read`,OLD.`precision`,OLD.`autonomy`,OLD.`started_at`,OLD.`validated_at`,OLD.`best_answer_at`,OLD.`latest_answer_at`,OLD.`thread_started_at`,OLD.`latest_hint_at`,OLD.`finished_at`,OLD.`latest_activity_at`,OLD.`contest_started_at`,OLD.`ranked`,OLD.`all_lang_prog`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN INSERT INTO `history_groups_groups` (`id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`type_changed_at`,`inviting_user_id`) VALUES (NEW.`id`,@curVersion,NEW.`parent_group_id`,NEW.`child_group_id`,NEW.`child_order`,NEW.`type`,NEW.`role`,NEW.`type_changed_at`,NEW.`inviting_user_id`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF NEW.version <> OLD.version THEN
        SET @curVersion = NEW.version;
    ELSE
        SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    END IF;
    IF NOT (OLD.`id` = NEW.`id` AND OLD.`parent_group_id` <=> NEW.`parent_group_id` AND
            OLD.`child_group_id` <=> NEW.`child_group_id` AND OLD.`child_order` <=> NEW.`child_order`AND
            OLD.`type` <=> NEW.`type` AND OLD.`role` <=> NEW.`role` AND OLD.`type_changed_at` <=> NEW.`type_changed_at` AND
            OLD.`inviting_user_id` <=> NEW.`inviting_user_id`) THEN
        SET NEW.version = @curVersion;
        UPDATE `history_groups_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
        INSERT INTO `history_groups_groups` (
            `id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`type_changed_at`,`inviting_user_id`
        ) VALUES (
                     NEW.`id`,@curVersion,NEW.`parent_group_id`,NEW.`child_group_id`,NEW.`child_order`,NEW.`type`,NEW.`role`,
                     NEW.`type_changed_at`,NEW.`inviting_user_id`
                 );
    END IF;
    IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR OLD.type != NEW.type) THEN
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
            SELECT `groups_ancestors`.`child_group_id`, 'todo'
            FROM `groups_ancestors`
            WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
        ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        DELETE `groups_ancestors` FROM `groups_ancestors`
        WHERE `groups_ancestors`.`child_group_id` = OLD.`child_group_id` AND
                `groups_ancestors`.`ancestor_group_id` = OLD.`parent_group_id`;
        DELETE `bridges` FROM `groups_ancestors` `child_descendants`
                                  JOIN `groups_ancestors` `parent_ancestors`
                                  JOIN `groups_ancestors` `bridges`
                                       ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
                                           `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
        WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
                `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;
        DELETE `child_ancestors` FROM `groups_ancestors` `child_ancestors`
                                          JOIN `groups_ancestors` `parent_ancestors`
                                               ON (`child_ancestors`.`child_group_id` = OLD.`child_group_id` AND
                                                   `child_ancestors`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id`)
        WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id`;
        DELETE `parent_ancestors` FROM `groups_ancestors` `parent_ancestors`
                                           JOIN  `groups_ancestors` `child_ancestors`
                                                 ON (`parent_ancestors`.`ancestor_group_id` = OLD.`parent_group_id` AND
                                                     `child_ancestors`.`child_group_id` = `parent_ancestors`.`child_group_id`)
        WHERE `child_ancestors`.`ancestor_group_id` = OLD.`child_group_id`;
    END IF;
    IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR OLD.type != NEW.type) THEN
        INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    UPDATE `history_groups_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_groups_groups` (
        `id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`type_changed_at`,`inviting_user_id`,`deleted`
    ) VALUES (
                 OLD.`id`,@curVersion,OLD.`parent_group_id`,OLD.`child_group_id`,OLD.`child_order`,OLD.`type`,OLD.`role`,
                 OLD.`type_changed_at`,OLD.`inviting_user_id`, 1
             );
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
        SELECT `groups_ancestors`.`child_group_id`, 'todo'
        FROM `groups_ancestors`
        WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
    ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    DELETE `groups_ancestors` FROM `groups_ancestors`
    WHERE `groups_ancestors`.`child_group_id` = OLD.`child_group_id` AND
            `groups_ancestors`.`ancestor_group_id` = OLD.`parent_group_id`;
    DELETE `bridges`
    FROM `groups_ancestors` `child_descendants`
             JOIN `groups_ancestors` `parent_ancestors`
             JOIN `groups_ancestors` `bridges`
                  ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
                      `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
    WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
            `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;
    DELETE `child_ancestors`
    FROM `groups_ancestors` `child_ancestors`
             JOIN  `groups_ancestors` `parent_ancestors`
                   ON (`child_ancestors`.`child_group_id` = OLD.`child_group_id` AND
                       `child_ancestors`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id`)
    WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id`;
    DELETE `parent_ancestors`
    FROM `groups_ancestors` `parent_ancestors`
             JOIN  `groups_ancestors` `child_ancestors`
                   ON (`parent_ancestors`.`ancestor_group_id` = OLD.`parent_group_id` AND
                       `child_ancestors`.`child_group_id` = `parent_ancestors`.`child_group_id`)
    WHERE `child_ancestors`.`ancestor_group_id` = OLD.`child_group_id`;
END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN
    INSERT INTO `history_groups_items` (
        `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_since`,`full_access_since`,`access_reason`,
        `solutions_access_since`,`owner_access`,`manager_access`,`cached_partial_access_since`,`cached_full_access_since`,
        `cached_solutions_access_since`,`cached_grayed_access_since`,`cached_full_access`,`cached_partial_access`,
        `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`
    ) VALUES (
                 NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_since`,
                 NEW.`full_access_since`,NEW.`access_reason`,NEW.`solutions_access_since`,NEW.`owner_access`,
                 NEW.`manager_access`,NEW.`cached_partial_access_since`,NEW.`cached_full_access_since`,
                 NEW.`cached_solutions_access_since`,NEW.`cached_grayed_access_since`,NEW.`cached_full_access`,
                 NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,NEW.`cached_manager_access`
             );
    INSERT INTO `groups_items_propagate` (`id`, `propagate_access`) VALUE (NEW.`id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_access` = 'self';
END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_items` BEFORE UPDATE ON `groups_items` FOR EACH ROW BEGIN
    IF NEW.version <> OLD.version THEN
        SET @curVersion = NEW.version;
    ELSE
        SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    END IF;
    IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`item_id` <=> NEW.`item_id` AND
            OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`partial_access_since` <=> NEW.`partial_access_since` AND
            OLD.`full_access_since` <=> NEW.`full_access_since` AND OLD.`access_reason` <=> NEW.`access_reason` AND
            OLD.`solutions_access_since` <=> NEW.`solutions_access_since` AND OLD.`owner_access` <=> NEW.`owner_access` AND
            OLD.`manager_access` <=> NEW.`manager_access` AND OLD.`cached_partial_access_since` <=> NEW.`cached_partial_access_since` AND
            OLD.`cached_full_access_since` <=> NEW.`cached_full_access_since` AND OLD.`cached_solutions_access_since` <=> NEW.`cached_solutions_access_since` AND
            OLD.`cached_grayed_access_since` <=> NEW.`cached_grayed_access_since` AND OLD.`cached_full_access` <=> NEW.`cached_full_access` AND
            OLD.`cached_partial_access` <=> NEW.`cached_partial_access` AND OLD.`cached_access_solutions` <=> NEW.`cached_access_solutions` AND
            OLD.`cached_grayed_access` <=> NEW.`cached_grayed_access` AND OLD.`cached_manager_access` <=> NEW.`cached_manager_access`) THEN
        SET NEW.version = @curVersion;
        UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
        INSERT INTO `history_groups_items` (
            `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_since`,`full_access_since`,`access_reason`,
            `solutions_access_since`,`owner_access`,`manager_access`,`cached_partial_access_since`,`cached_full_access_since`,
            `cached_solutions_access_since`,`cached_grayed_access_since`,`cached_full_access`,`cached_partial_access`,
            `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`
        ) VALUES (
                     NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_since`,
                     NEW.`full_access_since`,NEW.`access_reason`,NEW.`solutions_access_since`,NEW.`owner_access`,
                     NEW.`manager_access`,NEW.`cached_partial_access_since`,NEW.`cached_full_access_since`,
                     NEW.`cached_solutions_access_since`,NEW.`cached_grayed_access_since`,NEW.`cached_full_access`,
                     NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,
                     NEW.`cached_manager_access`
                 );
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER `after_update_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_groups_items` AFTER UPDATE ON `groups_items` FOR EACH ROW BEGIN
    # As a date change may result in access change for descendants of the item, mark the entry as to be recomputed
    IF NOT (NEW.`full_access_since` <=> OLD.`full_access_since`AND NEW.`partial_access_since` <=> OLD.`partial_access_since`AND
            NEW.`solutions_access_since` <=> OLD.`solutions_access_since`AND NEW.`manager_access` <=> OLD.`manager_access`AND
            NEW.`access_reason` <=> OLD.`access_reason`) THEN
        INSERT INTO `groups_items_propagate` (`id`, `propagate_access`) VALUE (NEW.`id`, 'self')
        ON DUPLICATE KEY UPDATE `propagate_access` = 'self';
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_items` BEFORE DELETE ON `groups_items` FOR EACH ROW BEGIN
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_groups_items` (
        `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_since`,`full_access_since`,`access_reason`,
        `solutions_access_since`,`owner_access`,`manager_access`,`cached_partial_access_since`,`cached_full_access_since`,
        `cached_solutions_access_since`,`cached_grayed_access_since`,`cached_full_access`,`cached_partial_access`,
        `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`deleted`
    ) VALUES (
                 OLD.`id`,@curVersion,OLD.`group_id`,OLD.`item_id`,OLD.`creator_user_id`,OLD.`partial_access_since`,
                 OLD.`full_access_since`,OLD.`access_reason`,OLD.`solutions_access_since`,OLD.`owner_access`,
                 OLD.`manager_access`,OLD.`cached_partial_access_since`,OLD.`cached_full_access_since`,
                 OLD.`cached_solutions_access_since`,OLD.`cached_grayed_access_since`,OLD.`cached_full_access`,
                 OLD.`cached_partial_access`,OLD.`cached_access_solutions`,OLD.`cached_grayed_access`,
                 OLD.`cached_manager_access`, 1);
END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items` AFTER INSERT ON `items` FOR EACH ROW BEGIN INSERT INTO `history_items` (`id`,`version`,`url`,`options`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`team_mode`,`teams_editable`,`qualified_group_id`,`team_max_members`,`has_attempts`,`contest_opens_at`,`duration`,`contest_closes_at`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`) VALUES (NEW.`id`,@curVersion,NEW.`url`,NEW.`options`,NEW.`platform_id`,NEW.`text_id`,NEW.`repository_path`,NEW.`type`,NEW.`uses_api`,NEW.`read_only`,NEW.`full_screen`,NEW.`show_difficulty`,NEW.`show_source`,NEW.`hints_allowed`,NEW.`fixed_ranks`,NEW.`validation_type`,NEW.`validation_min`,NEW.`preparation_state`,NEW.`unlocked_item_ids`,NEW.`score_min_unlock`,NEW.`supported_lang_prog`,NEW.`default_language_id`,NEW.`team_mode`,NEW.`teams_editable`,NEW.`qualified_group_id`,NEW.`team_max_members`,NEW.`has_attempts`,NEW.`contest_opens_at`,NEW.`duration`,NEW.`contest_closes_at`,NEW.`show_user_infos`,NEW.`contest_phase`,NEW.`level`,NEW.`no_score`,NEW.`title_bar_visible`,NEW.`transparent_folder`,NEW.`display_details_in_parent`,NEW.`display_children_as_tabs`,NEW.`custom_chapter`,NEW.`group_code_enter`); INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items` BEFORE UPDATE ON `items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`url` <=> NEW.`url` AND OLD.`options` <=> NEW.`options` AND OLD.`platform_id` <=> NEW.`platform_id` AND OLD.`text_id` <=> NEW.`text_id` AND OLD.`repository_path` <=> NEW.`repository_path` AND OLD.`type` <=> NEW.`type` AND OLD.`uses_api` <=> NEW.`uses_api` AND OLD.`read_only` <=> NEW.`read_only` AND OLD.`full_screen` <=> NEW.`full_screen` AND OLD.`show_difficulty` <=> NEW.`show_difficulty` AND OLD.`show_source` <=> NEW.`show_source` AND OLD.`hints_allowed` <=> NEW.`hints_allowed` AND OLD.`fixed_ranks` <=> NEW.`fixed_ranks` AND OLD.`validation_type` <=> NEW.`validation_type` AND OLD.`validation_min` <=> NEW.`validation_min` AND OLD.`preparation_state` <=> NEW.`preparation_state` AND OLD.`unlocked_item_ids` <=> NEW.`unlocked_item_ids` AND OLD.`score_min_unlock` <=> NEW.`score_min_unlock` AND OLD.`supported_lang_prog` <=> NEW.`supported_lang_prog` AND OLD.`default_language_id` <=> NEW.`default_language_id` AND OLD.`team_mode` <=> NEW.`team_mode` AND OLD.`teams_editable` <=> NEW.`teams_editable` AND OLD.`qualified_group_id` <=> NEW.`qualified_group_id` AND OLD.`team_max_members` <=> NEW.`team_max_members` AND OLD.`has_attempts` <=> NEW.`has_attempts` AND OLD.`contest_opens_at` <=> NEW.`contest_opens_at` AND OLD.`duration` <=> NEW.`duration` AND OLD.`contest_closes_at` <=> NEW.`contest_closes_at` AND OLD.`show_user_infos` <=> NEW.`show_user_infos` AND OLD.`contest_phase` <=> NEW.`contest_phase` AND OLD.`level` <=> NEW.`level` AND OLD.`no_score` <=> NEW.`no_score` AND OLD.`title_bar_visible` <=> NEW.`title_bar_visible` AND OLD.`transparent_folder` <=> NEW.`transparent_folder` AND OLD.`display_details_in_parent` <=> NEW.`display_details_in_parent` AND OLD.`display_children_as_tabs` <=> NEW.`display_children_as_tabs` AND OLD.`custom_chapter` <=> NEW.`custom_chapter` AND OLD.`group_code_enter` <=> NEW.`group_code_enter`) THEN   SET NEW.version = @curVersion;   UPDATE `history_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_items` (`id`,`version`,`url`,`options`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`team_mode`,`teams_editable`,`qualified_group_id`,`team_max_members`,`has_attempts`,`contest_opens_at`,`duration`,`contest_closes_at`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`)       VALUES (NEW.`id`,@curVersion,NEW.`url`,NEW.`options`,NEW.`platform_id`,NEW.`text_id`,NEW.`repository_path`,NEW.`type`,NEW.`uses_api`,NEW.`read_only`,NEW.`full_screen`,NEW.`show_difficulty`,NEW.`show_source`,NEW.`hints_allowed`,NEW.`fixed_ranks`,NEW.`validation_type`,NEW.`validation_min`,NEW.`preparation_state`,NEW.`unlocked_item_ids`,NEW.`score_min_unlock`,NEW.`supported_lang_prog`,NEW.`default_language_id`,NEW.`team_mode`,NEW.`teams_editable`,NEW.`qualified_group_id`,NEW.`team_max_members`,NEW.`has_attempts`,NEW.`contest_opens_at`,NEW.`duration`,NEW.`contest_closes_at`,NEW.`show_user_infos`,NEW.`contest_phase`,NEW.`level`,NEW.`no_score`,NEW.`title_bar_visible`,NEW.`transparent_folder`,NEW.`display_details_in_parent`,NEW.`display_children_as_tabs`,NEW.`custom_chapter`,NEW.`group_code_enter`) ; END IF; SELECT platforms.id INTO @platformID FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1 ; SET NEW.platform_id=@platformID ; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items` BEFORE DELETE ON `items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_items` (`id`,`version`,`url`,`options`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`team_mode`,`teams_editable`,`qualified_group_id`,`team_max_members`,`has_attempts`,`contest_opens_at`,`duration`,`contest_closes_at`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`url`,OLD.`options`,OLD.`platform_id`,OLD.`text_id`,OLD.`repository_path`,OLD.`type`,OLD.`uses_api`,OLD.`read_only`,OLD.`full_screen`,OLD.`show_difficulty`,OLD.`show_source`,OLD.`hints_allowed`,OLD.`fixed_ranks`,OLD.`validation_type`,OLD.`validation_min`,OLD.`preparation_state`,OLD.`unlocked_item_ids`,OLD.`score_min_unlock`,OLD.`supported_lang_prog`,OLD.`default_language_id`,OLD.`team_mode`,OLD.`teams_editable`,OLD.`qualified_group_id`,OLD.`team_max_members`,OLD.`has_attempts`,OLD.`contest_opens_at`,OLD.`duration`,OLD.`contest_closes_at`,OLD.`show_user_infos`,OLD.`contest_phase`,OLD.`level`,OLD.`no_score`,OLD.`title_bar_visible`,OLD.`transparent_folder`,OLD.`display_details_in_parent`,OLD.`display_children_as_tabs`,OLD.`custom_chapter`,OLD.`group_code_enter`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_messages` AFTER INSERT ON `messages` FOR EACH ROW BEGIN INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submitted_at`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`) VALUES (NEW.`id`,@curVersion,NEW.`thread_id`,NEW.`user_id`,NEW.`submitted_at`,NEW.`published`,NEW.`title`,NEW.`body`,NEW.`trainers_only`,NEW.`archived`,NEW.`persistant`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_messages` BEFORE UPDATE ON `messages` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`thread_id` <=> NEW.`thread_id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`submitted_at` <=> NEW.`submitted_at` AND OLD.`published` <=> NEW.`published` AND OLD.`title` <=> NEW.`title` AND OLD.`body` <=> NEW.`body` AND OLD.`trainers_only` <=> NEW.`trainers_only` AND OLD.`archived` <=> NEW.`archived` AND OLD.`persistant` <=> NEW.`persistant`) THEN   SET NEW.version = @curVersion;   UPDATE `history_messages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submitted_at`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`)       VALUES (NEW.`id`,@curVersion,NEW.`thread_id`,NEW.`user_id`,NEW.`submitted_at`,NEW.`published`,NEW.`title`,NEW.`body`,NEW.`trainers_only`,NEW.`archived`,NEW.`persistant`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_messages` BEFORE DELETE ON `messages` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_messages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submitted_at`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`thread_id`,OLD.`user_id`,OLD.`submitted_at`,OLD.`published`,OLD.`title`,OLD.`body`,OLD.`trainers_only`,OLD.`archived`,OLD.`persistant`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_threads` AFTER INSERT ON `threads` FOR EACH ROW BEGIN INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`latest_activity_at`) VALUES (NEW.`id`,@curVersion,NEW.`type`,NEW.`creator_user_id`,NEW.`item_id`,NEW.`title`,NEW.`admin_help_asked`,NEW.`hidden`,NEW.`latest_activity_at`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_threads` BEFORE UPDATE ON `threads` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`type` <=> NEW.`type` AND OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`title` <=> NEW.`title` AND OLD.`admin_help_asked` <=> NEW.`admin_help_asked` AND OLD.`hidden` <=> NEW.`hidden` AND OLD.`latest_activity_at` <=> NEW.`latest_activity_at`) THEN   SET NEW.version = @curVersion;   UPDATE `history_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`latest_activity_at`)       VALUES (NEW.`id`,@curVersion,NEW.`type`,NEW.`creator_user_id`,NEW.`item_id`,NEW.`title`,NEW.`admin_help_asked`,NEW.`hidden`,NEW.`latest_activity_at`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_threads` BEFORE DELETE ON `threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`latest_activity_at`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`type`,OLD.`creator_user_id`,OLD.`item_id`,OLD.`title`,OLD.`admin_help_asked`,OLD.`hidden`,OLD.`latest_activity_at`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users` AFTER INSERT ON `users` FOR EACH ROW BEGIN INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registered_at`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`latest_login_at`,`latest_activity_at`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notifications_read_at`,`login_module_prefix`,`allow_subgroups`) VALUES (NEW.`id`,@curVersion,NEW.`login`,NEW.`open_id_identity`,NEW.`password_md5`,NEW.`salt`,NEW.`recover`,NEW.`registered_at`,NEW.`email`,NEW.`email_verified`,NEW.`first_name`,NEW.`last_name`,NEW.`country_code`,NEW.`time_zone`,NEW.`birth_date`,NEW.`graduation_year`,NEW.`grade`,NEW.`sex`,NEW.`student_id`,NEW.`address`,NEW.`zipcode`,NEW.`city`,NEW.`land_line_number`,NEW.`cell_phone_number`,NEW.`default_language`,NEW.`notify_news`,NEW.`notify`,NEW.`public_first_name`,NEW.`public_last_name`,NEW.`free_text`,NEW.`web_site`,NEW.`photo_autoload`,NEW.`lang_prog`,NEW.`latest_login_at`,NEW.`latest_activity_at`,NEW.`last_ip`,NEW.`basic_editor_mode`,NEW.`spaces_for_tab`,NEW.`member_state`,NEW.`godfather_user_id`,NEW.`step_level_in_site`,NEW.`is_admin`,NEW.`no_ranking`,NEW.`help_given`,NEW.`self_group_id`,NEW.`owned_group_id`,NEW.`access_group_id`,NEW.`notifications_read_at`,NEW.`login_module_prefix`,NEW.`allow_subgroups`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users` BEFORE UPDATE ON `users` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`login` <=> NEW.`login` AND OLD.`open_id_identity` <=> NEW.`open_id_identity` AND OLD.`password_md5` <=> NEW.`password_md5` AND OLD.`salt` <=> NEW.`salt` AND OLD.`recover` <=> NEW.`recover` AND OLD.`registered_at` <=> NEW.`registered_at` AND OLD.`email` <=> NEW.`email` AND OLD.`email_verified` <=> NEW.`email_verified` AND OLD.`first_name` <=> NEW.`first_name` AND OLD.`last_name` <=> NEW.`last_name` AND OLD.`country_code` <=> NEW.`country_code` AND OLD.`time_zone` <=> NEW.`time_zone` AND OLD.`birth_date` <=> NEW.`birth_date` AND OLD.`graduation_year` <=> NEW.`graduation_year` AND OLD.`grade` <=> NEW.`grade` AND OLD.`sex` <=> NEW.`sex` AND OLD.`student_id` <=> NEW.`student_id` AND OLD.`address` <=> NEW.`address` AND OLD.`zipcode` <=> NEW.`zipcode` AND OLD.`city` <=> NEW.`city` AND OLD.`land_line_number` <=> NEW.`land_line_number` AND OLD.`cell_phone_number` <=> NEW.`cell_phone_number` AND OLD.`default_language` <=> NEW.`default_language` AND OLD.`notify_news` <=> NEW.`notify_news` AND OLD.`notify` <=> NEW.`notify` AND OLD.`public_first_name` <=> NEW.`public_first_name` AND OLD.`public_last_name` <=> NEW.`public_last_name` AND OLD.`free_text` <=> NEW.`free_text` AND OLD.`web_site` <=> NEW.`web_site` AND OLD.`photo_autoload` <=> NEW.`photo_autoload` AND OLD.`lang_prog` <=> NEW.`lang_prog` AND OLD.`latest_login_at` <=> NEW.`latest_login_at` AND OLD.`latest_activity_at` <=> NEW.`latest_activity_at` AND OLD.`last_ip` <=> NEW.`last_ip` AND OLD.`basic_editor_mode` <=> NEW.`basic_editor_mode` AND OLD.`spaces_for_tab` <=> NEW.`spaces_for_tab` AND OLD.`member_state` <=> NEW.`member_state` AND OLD.`godfather_user_id` <=> NEW.`godfather_user_id` AND OLD.`step_level_in_site` <=> NEW.`step_level_in_site` AND OLD.`is_admin` <=> NEW.`is_admin` AND OLD.`no_ranking` <=> NEW.`no_ranking` AND OLD.`help_given` <=> NEW.`help_given` AND OLD.`self_group_id` <=> NEW.`self_group_id` AND OLD.`owned_group_id` <=> NEW.`owned_group_id` AND OLD.`access_group_id` <=> NEW.`access_group_id` AND OLD.`notifications_read_at` <=> NEW.`notifications_read_at` AND OLD.`login_module_prefix` <=> NEW.`login_module_prefix` AND OLD.`allow_subgroups` <=> NEW.`allow_subgroups`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registered_at`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`latest_login_at`,`latest_activity_at`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notifications_read_at`,`login_module_prefix`,`allow_subgroups`)       VALUES (NEW.`id`,@curVersion,NEW.`login`,NEW.`open_id_identity`,NEW.`password_md5`,NEW.`salt`,NEW.`recover`,NEW.`registered_at`,NEW.`email`,NEW.`email_verified`,NEW.`first_name`,NEW.`last_name`,NEW.`country_code`,NEW.`time_zone`,NEW.`birth_date`,NEW.`graduation_year`,NEW.`grade`,NEW.`sex`,NEW.`student_id`,NEW.`address`,NEW.`zipcode`,NEW.`city`,NEW.`land_line_number`,NEW.`cell_phone_number`,NEW.`default_language`,NEW.`notify_news`,NEW.`notify`,NEW.`public_first_name`,NEW.`public_last_name`,NEW.`free_text`,NEW.`web_site`,NEW.`photo_autoload`,NEW.`lang_prog`,NEW.`latest_login_at`,NEW.`latest_activity_at`,NEW.`last_ip`,NEW.`basic_editor_mode`,NEW.`spaces_for_tab`,NEW.`member_state`,NEW.`godfather_user_id`,NEW.`step_level_in_site`,NEW.`is_admin`,NEW.`no_ranking`,NEW.`help_given`,NEW.`self_group_id`,NEW.`owned_group_id`,NEW.`access_group_id`,NEW.`notifications_read_at`,NEW.`login_module_prefix`,NEW.`allow_subgroups`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users` BEFORE DELETE ON `users` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registered_at`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`latest_login_at`,`latest_activity_at`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notifications_read_at`,`login_module_prefix`,`allow_subgroups`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`login`,OLD.`open_id_identity`,OLD.`password_md5`,OLD.`salt`,OLD.`recover`,OLD.`registered_at`,OLD.`email`,OLD.`email_verified`,OLD.`first_name`,OLD.`last_name`,OLD.`country_code`,OLD.`time_zone`,OLD.`birth_date`,OLD.`graduation_year`,OLD.`grade`,OLD.`sex`,OLD.`student_id`,OLD.`address`,OLD.`zipcode`,OLD.`city`,OLD.`land_line_number`,OLD.`cell_phone_number`,OLD.`default_language`,OLD.`notify_news`,OLD.`notify`,OLD.`public_first_name`,OLD.`public_last_name`,OLD.`free_text`,OLD.`web_site`,OLD.`photo_autoload`,OLD.`lang_prog`,OLD.`latest_login_at`,OLD.`latest_activity_at`,OLD.`last_ip`,OLD.`basic_editor_mode`,OLD.`spaces_for_tab`,OLD.`member_state`,OLD.`godfather_user_id`,OLD.`step_level_in_site`,OLD.`is_admin`,OLD.`no_ranking`,OLD.`help_given`,OLD.`self_group_id`,OLD.`owned_group_id`,OLD.`access_group_id`,OLD.`notifications_read_at`,OLD.`login_module_prefix`,OLD.`allow_subgroups`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users_items` AFTER INSERT ON `users_items` FOR EACH ROW BEGIN INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`contest_started_at`,`ranked`,`all_lang_prog`,`state`,`answer`) VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`item_id`,NEW.`active_attempt_id`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`started_at`,NEW.`validated_at`,NEW.`best_answer_at`,NEW.`latest_answer_at`,NEW.`thread_started_at`,NEW.`latest_hint_at`,NEW.`finished_at`,NEW.`latest_activity_at`,NEW.`contest_started_at`,NEW.`ranked`,NEW.`all_lang_prog`,NEW.`state`,NEW.`answer`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users_items` BEFORE UPDATE ON `users_items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`active_attempt_id` <=> NEW.`active_attempt_id` AND OLD.`score` <=> NEW.`score` AND OLD.`score_computed` <=> NEW.`score_computed` AND OLD.`score_reeval` <=> NEW.`score_reeval` AND OLD.`score_diff_manual` <=> NEW.`score_diff_manual` AND OLD.`score_diff_comment` <=> NEW.`score_diff_comment` AND OLD.`tasks_tried` <=> NEW.`tasks_tried` AND OLD.`children_validated` <=> NEW.`children_validated` AND OLD.`validated` <=> NEW.`validated` AND OLD.`finished` <=> NEW.`finished` AND OLD.`key_obtained` <=> NEW.`key_obtained` AND OLD.`tasks_with_help` <=> NEW.`tasks_with_help` AND OLD.`hints_requested` <=> NEW.`hints_requested` AND OLD.`hints_cached` <=> NEW.`hints_cached` AND OLD.`corrections_read` <=> NEW.`corrections_read` AND OLD.`precision` <=> NEW.`precision` AND OLD.`autonomy` <=> NEW.`autonomy` AND OLD.`started_at` <=> NEW.`started_at` AND OLD.`validated_at` <=> NEW.`validated_at` AND OLD.`best_answer_at` <=> NEW.`best_answer_at` AND OLD.`latest_answer_at` <=> NEW.`latest_answer_at` AND OLD.`thread_started_at` <=> NEW.`thread_started_at` AND OLD.`latest_hint_at` <=> NEW.`latest_hint_at` AND OLD.`finished_at` <=> NEW.`finished_at` AND OLD.`contest_started_at` <=> NEW.`contest_started_at` AND OLD.`ranked` <=> NEW.`ranked` AND OLD.`all_lang_prog` <=> NEW.`all_lang_prog` AND OLD.`state` <=> NEW.`state` AND OLD.`answer` <=> NEW.`answer`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`contest_started_at`,`ranked`,`all_lang_prog`,`state`,`answer`)       VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`item_id`,NEW.`active_attempt_id`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`started_at`,NEW.`validated_at`,NEW.`best_answer_at`,NEW.`latest_answer_at`,NEW.`thread_started_at`,NEW.`latest_hint_at`,NEW.`finished_at`,NEW.`latest_activity_at`,NEW.`contest_started_at`,NEW.`ranked`,NEW.`all_lang_prog`,NEW.`state`,NEW.`answer`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users_items` BEFORE DELETE ON `users_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`contest_started_at`,`ranked`,`all_lang_prog`,`state`,`answer`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`user_id`,OLD.`item_id`,OLD.`active_attempt_id`,OLD.`score`,OLD.`score_computed`,OLD.`score_reeval`,OLD.`score_diff_manual`,OLD.`score_diff_comment`,OLD.`submissions_attempts`,OLD.`tasks_tried`,OLD.`children_validated`,OLD.`validated`,OLD.`finished`,OLD.`key_obtained`,OLD.`tasks_with_help`,OLD.`hints_requested`,OLD.`hints_cached`,OLD.`corrections_read`,OLD.`precision`,OLD.`autonomy`,OLD.`started_at`,OLD.`validated_at`,OLD.`best_answer_at`,OLD.`latest_answer_at`,OLD.`thread_started_at`,OLD.`latest_hint_at`,OLD.`finished_at`,OLD.`latest_activity_at`,OLD.`contest_started_at`,OLD.`ranked`,OLD.`all_lang_prog`,OLD.`state`,OLD.`answer`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users_threads` AFTER INSERT ON `users_threads` FOR EACH ROW BEGIN INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`lately_viewed_at`,`lately_posted_at`,`starred`) VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`thread_id`,NEW.`lately_viewed_at`,NEW.`lately_posted_at`,NEW.`starred`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users_threads` BEFORE UPDATE ON `users_threads` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`thread_id` <=> NEW.`thread_id` AND OLD.`lately_viewed_at` <=> NEW.`lately_viewed_at` AND OLD.`lately_posted_at` <=> NEW.`lately_posted_at` AND OLD.`starred` <=> NEW.`starred`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`lately_viewed_at`,`lately_posted_at`,`starred`)       VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`thread_id`,NEW.`lately_viewed_at`,NEW.`lately_posted_at`,NEW.`starred`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users_threads` BEFORE DELETE ON `users_threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`lately_viewed_at`,`lately_posted_at`,`starred`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`user_id`,OLD.`thread_id`,OLD.`lately_viewed_at`,OLD.`lately_posted_at`,OLD.`starred`, 1); END
-- +migrate StatementEnd

ALTER ALGORITHM=UNDEFINED
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



-- +migrate Down
ALTER TABLE `groups`
    RENAME COLUMN `created_at` TO `date_created`,
    RENAME COLUMN `code_expires_at` TO `code_end`,
    RENAME COLUMN `lock_user_deletion_until` TO `lock_user_deletion_date`;
ALTER TABLE `groups_attempts`
    RENAME COLUMN `started_at` TO `start_date`,
    RENAME COLUMN `validated_at` TO `validation_date`,
    RENAME COLUMN `finished_at` TO `finish_date`,
    RENAME COLUMN `latest_activity_at` TO `last_activity_date`,
    RENAME COLUMN `thread_started_at` TO `thread_start_date`,
    RENAME COLUMN `best_answer_at` TO `best_answer_date`,
    RENAME COLUMN `latest_answer_at` TO `last_answer_date`,
    RENAME COLUMN `latest_hint_at` TO `last_hint_date`,
    RENAME COLUMN `contest_started_at` TO `contest_start_date`;
ALTER TABLE `groups_groups`
    RENAME COLUMN `type_changed_at` TO `status_date`;
ALTER TABLE `groups_items`
    RENAME COLUMN `partial_access_since` TO `partial_access_date`,
    RENAME COLUMN `full_access_since` TO `full_access_date`,
    RENAME COLUMN `solutions_access_since` TO `access_solutions_date`,
    RENAME COLUMN `cached_full_access_since` TO `cached_full_access_date`,
    RENAME COLUMN `cached_partial_access_since` TO `cached_partial_access_date`,
    RENAME COLUMN `cached_solutions_access_since` TO `cached_access_solutions_date`,
    RENAME COLUMN `cached_grayed_access_since` TO `cached_grayed_access_date`;
ALTER TABLE `history_groups`
    RENAME COLUMN `created_at` TO `date_created`,
    RENAME COLUMN `code_expires_at` TO `code_end`,
    RENAME COLUMN `lock_user_deletion_until` TO `lock_user_deletion_date`;
ALTER TABLE `history_groups_attempts`
    RENAME COLUMN `started_at` TO `start_date`,
    RENAME COLUMN `validated_at` TO `validation_date`,
    RENAME COLUMN `finished_at` TO `finish_date`,
    RENAME COLUMN `latest_activity_at` TO `last_activity_date`,
    RENAME COLUMN `thread_started_at` TO `thread_start_date`,
    RENAME COLUMN `best_answer_at` TO `best_answer_date`,
    RENAME COLUMN `latest_answer_at` TO `last_answer_date`,
    RENAME COLUMN `latest_hint_at` TO `last_hint_date`,
    RENAME COLUMN `contest_started_at` TO `contest_start_date`;
ALTER TABLE `history_groups_groups`
    RENAME COLUMN `type_changed_at` TO `status_date`;
ALTER TABLE `history_groups_items`
    RENAME COLUMN `partial_access_since` TO `partial_access_date`,
    RENAME COLUMN `full_access_since` TO `full_access_date`,
    RENAME COLUMN `solutions_access_since` TO `access_solutions_date`,
    RENAME COLUMN `cached_full_access_since` TO `cached_full_access_date`,
    RENAME COLUMN `cached_partial_access_since` TO `cached_partial_access_date`,
    RENAME COLUMN `cached_solutions_access_since` TO `cached_access_solutions_date`,
    RENAME COLUMN `cached_grayed_access_since` TO `cached_grayed_access_date`;
ALTER TABLE `history_items`
    RENAME COLUMN `contest_opens_at` TO `access_open_date`,
    RENAME COLUMN `contest_closes_at` TO `end_contest_date`;
ALTER TABLE `history_messages`
    RENAME COLUMN `submitted_at` TO `submission_date`;
ALTER TABLE `history_threads`
    RENAME COLUMN `latest_activity_at` TO `last_activity_date`;
ALTER TABLE `history_users`
    RENAME COLUMN `registered_at` TO `registration_date`,
    RENAME COLUMN `latest_login_at` TO `last_login_date`,
    RENAME COLUMN `latest_activity_at` TO `last_activity_date`,
    RENAME COLUMN `notifications_read_at` TO `notification_read_date`;
ALTER TABLE `history_users_items`
    RENAME COLUMN `started_at` TO `start_date`,
    RENAME COLUMN `validated_at` TO `validation_date`,
    RENAME COLUMN `finished_at` TO `finish_date`,
    RENAME COLUMN `latest_activity_at` TO `last_activity_date`,
    RENAME COLUMN `thread_started_at` TO `thread_start_date`,
    RENAME COLUMN `best_answer_at` TO `best_answer_date`,
    RENAME COLUMN `latest_answer_at` TO `last_answer_date`,
    RENAME COLUMN `latest_hint_at` TO `last_hint_date`,
    RENAME COLUMN `contest_started_at` TO `contest_start_date`;
ALTER TABLE `history_users_threads`
    RENAME COLUMN `lately_viewed_at` TO `last_read_date`,
    RENAME COLUMN `lately_posted_at` TO `last_write_date`;
ALTER TABLE `items`
    RENAME COLUMN `contest_opens_at` TO `access_open_date`,
    RENAME COLUMN `contest_closes_at` TO `end_contest_date`;
ALTER TABLE `login_states`
    RENAME COLUMN `expires_at` TO `expiration_date`,
    RENAME INDEX `expires_at` TO `expiration_date`;
ALTER TABLE `messages`
    RENAME COLUMN `submitted_at` TO `submission_date`;
ALTER TABLE `sessions`
    RENAME COLUMN `expires_at` TO `expiration_date`,
    RENAME COLUMN `issued_at` TO `issued_at_date`,
    RENAME INDEX `expires_at` TO `expiration_date`;
ALTER TABLE `threads`
    RENAME COLUMN `latest_activity_at` TO `last_activity_date`;
ALTER TABLE `users`
    RENAME COLUMN `registered_at` TO `registration_date`,
    RENAME COLUMN `latest_login_at` TO `last_login_date`,
    RENAME COLUMN `latest_activity_at` TO `last_activity_date`,
    RENAME COLUMN `notifications_read_at` TO `notification_read_date`;
ALTER TABLE `users_answers`
    RENAME COLUMN `submitted_at` TO `submission_date`,
    RENAME COLUMN `graded_at` TO `grading_date`;
ALTER TABLE `users_items`
    RENAME COLUMN `started_at` TO `start_date`,
    RENAME COLUMN `validated_at` TO `validation_date`,
    RENAME COLUMN `finished_at` TO `finish_date`,
    RENAME COLUMN `latest_activity_at` TO `last_activity_date`,
    RENAME COLUMN `thread_started_at` TO `thread_start_date`,
    RENAME COLUMN `best_answer_at` TO `best_answer_date`,
    RENAME COLUMN `latest_answer_at` TO `last_answer_date`,
    RENAME COLUMN `latest_hint_at` TO `last_hint_date`,
    RENAME COLUMN `contest_started_at` TO `contest_start_date`;
ALTER TABLE `users_threads`
    RENAME COLUMN `lately_viewed_at` TO `last_read_date`,
    RENAME COLUMN `lately_posted_at` TO `last_write_date`;

DROP TRIGGER `after_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`date_created`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_end`,`redirect_path`,`open_contest`,`type`,`send_emails`) VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`date_created`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_timer`,NEW.`code_end`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`); INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`name` <=> NEW.`name` AND OLD.`grade` <=> NEW.`grade` AND OLD.`grade_details` <=> NEW.`grade_details` AND OLD.`description` <=> NEW.`description` AND OLD.`date_created` <=> NEW.`date_created` AND OLD.`opened` <=> NEW.`opened` AND OLD.`free_access` <=> NEW.`free_access` AND OLD.`team_item_id` <=> NEW.`team_item_id` AND OLD.`team_participating` <=> NEW.`team_participating` AND OLD.`code` <=> NEW.`code` AND OLD.`code_timer` <=> NEW.`code_timer` AND OLD.`code_end` <=> NEW.`code_end` AND OLD.`redirect_path` <=> NEW.`redirect_path` AND OLD.`open_contest` <=> NEW.`open_contest` AND OLD.`type` <=> NEW.`type` AND OLD.`send_emails` <=> NEW.`send_emails`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`date_created`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_end`,`redirect_path`,`open_contest`,`type`,`send_emails`)       VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`date_created`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_timer`,NEW.`code_end`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups` BEFORE DELETE ON `groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`date_created`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_end`,`redirect_path`,`open_contest`,`type`,`send_emails`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`name`,OLD.`grade`,OLD.`grade_details`,OLD.`description`,OLD.`date_created`,OLD.`opened`,OLD.`free_access`,OLD.`team_item_id`,OLD.`team_participating`,OLD.`code`,OLD.`code_timer`,OLD.`code_end`,OLD.`redirect_path`,OLD.`open_contest`,OLD.`type`,OLD.`send_emails`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_attempts` AFTER INSERT ON `groups_attempts` FOR EACH ROW BEGIN INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`) VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`order`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`start_date`,NEW.`validation_date`,NEW.`best_answer_date`,NEW.`last_answer_date`,NEW.`thread_start_date`,NEW.`last_hint_date`,NEW.`finish_date`,NEW.`last_activity_date`,NEW.`contest_start_date`,NEW.`ranked`,NEW.`all_lang_prog`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_attempts` BEFORE UPDATE ON `groups_attempts` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`order` <=> NEW.`order` AND OLD.`score` <=> NEW.`score` AND OLD.`score_computed` <=> NEW.`score_computed` AND OLD.`score_reeval` <=> NEW.`score_reeval` AND OLD.`score_diff_manual` <=> NEW.`score_diff_manual` AND OLD.`score_diff_comment` <=> NEW.`score_diff_comment` AND OLD.`submissions_attempts` <=> NEW.`submissions_attempts` AND OLD.`tasks_tried` <=> NEW.`tasks_tried` AND OLD.`children_validated` <=> NEW.`children_validated` AND OLD.`validated` <=> NEW.`validated` AND OLD.`finished` <=> NEW.`finished` AND OLD.`key_obtained` <=> NEW.`key_obtained` AND OLD.`tasks_with_help` <=> NEW.`tasks_with_help` AND OLD.`hints_requested` <=> NEW.`hints_requested` AND OLD.`hints_cached` <=> NEW.`hints_cached` AND OLD.`corrections_read` <=> NEW.`corrections_read` AND OLD.`precision` <=> NEW.`precision` AND OLD.`autonomy` <=> NEW.`autonomy` AND OLD.`start_date` <=> NEW.`start_date` AND OLD.`validation_date` <=> NEW.`validation_date` AND OLD.`best_answer_date` <=> NEW.`best_answer_date` AND OLD.`last_answer_date` <=> NEW.`last_answer_date` AND OLD.`thread_start_date` <=> NEW.`thread_start_date` AND OLD.`last_hint_date` <=> NEW.`last_hint_date` AND OLD.`finish_date` <=> NEW.`finish_date` AND OLD.`contest_start_date` <=> NEW.`contest_start_date` AND OLD.`ranked` <=> NEW.`ranked` AND OLD.`all_lang_prog` <=> NEW.`all_lang_prog`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups_attempts` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`)       VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`order`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`start_date`,NEW.`validation_date`,NEW.`best_answer_date`,NEW.`last_answer_date`,NEW.`thread_start_date`,NEW.`last_hint_date`,NEW.`finish_date`,NEW.`last_activity_date`,NEW.`contest_start_date`,NEW.`ranked`,NEW.`all_lang_prog`) ; SET NEW.minus_score = -NEW.score; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_attempts` BEFORE DELETE ON `groups_attempts` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_attempts` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`group_id`,OLD.`item_id`,OLD.`creator_user_id`,OLD.`order`,OLD.`score`,OLD.`score_computed`,OLD.`score_reeval`,OLD.`score_diff_manual`,OLD.`score_diff_comment`,OLD.`submissions_attempts`,OLD.`tasks_tried`,OLD.`children_validated`,OLD.`validated`,OLD.`finished`,OLD.`key_obtained`,OLD.`tasks_with_help`,OLD.`hints_requested`,OLD.`hints_cached`,OLD.`corrections_read`,OLD.`precision`,OLD.`autonomy`,OLD.`start_date`,OLD.`validation_date`,OLD.`best_answer_date`,OLD.`last_answer_date`,OLD.`thread_start_date`,OLD.`last_hint_date`,OLD.`finish_date`,OLD.`last_activity_date`,OLD.`contest_start_date`,OLD.`ranked`,OLD.`all_lang_prog`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN INSERT INTO `history_groups_groups` (`id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`status_date`,`inviting_user_id`) VALUES (NEW.`id`,@curVersion,NEW.`parent_group_id`,NEW.`child_group_id`,NEW.`child_order`,NEW.`type`,NEW.`role`,NEW.`status_date`,NEW.`inviting_user_id`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF NEW.version <> OLD.version THEN
        SET @curVersion = NEW.version;
    ELSE
        SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    END IF;
    IF NOT (OLD.`id` = NEW.`id` AND OLD.`parent_group_id` <=> NEW.`parent_group_id` AND
            OLD.`child_group_id` <=> NEW.`child_group_id` AND OLD.`child_order` <=> NEW.`child_order`AND
            OLD.`type` <=> NEW.`type` AND OLD.`role` <=> NEW.`role` AND OLD.`status_date` <=> NEW.`status_date` AND
            OLD.`inviting_user_id` <=> NEW.`inviting_user_id`) THEN
        SET NEW.version = @curVersion;
        UPDATE `history_groups_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
        INSERT INTO `history_groups_groups` (
            `id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`status_date`,`inviting_user_id`
        ) VALUES (
                     NEW.`id`,@curVersion,NEW.`parent_group_id`,NEW.`child_group_id`,NEW.`child_order`,NEW.`type`,NEW.`role`,
                     NEW.`status_date`,NEW.`inviting_user_id`
                 );
    END IF;
    IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR OLD.type != NEW.type) THEN
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
            SELECT `groups_ancestors`.`child_group_id`, 'todo'
            FROM `groups_ancestors`
            WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
        ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        DELETE `groups_ancestors` FROM `groups_ancestors`
        WHERE `groups_ancestors`.`child_group_id` = OLD.`child_group_id` AND
                `groups_ancestors`.`ancestor_group_id` = OLD.`parent_group_id`;
        DELETE `bridges` FROM `groups_ancestors` `child_descendants`
                                  JOIN `groups_ancestors` `parent_ancestors`
                                  JOIN `groups_ancestors` `bridges`
                                       ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
                                           `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
        WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
                `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;
        DELETE `child_ancestors` FROM `groups_ancestors` `child_ancestors`
                                          JOIN `groups_ancestors` `parent_ancestors`
                                               ON (`child_ancestors`.`child_group_id` = OLD.`child_group_id` AND
                                                   `child_ancestors`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id`)
        WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id`;
        DELETE `parent_ancestors` FROM `groups_ancestors` `parent_ancestors`
                                           JOIN  `groups_ancestors` `child_ancestors`
                                                 ON (`parent_ancestors`.`ancestor_group_id` = OLD.`parent_group_id` AND
                                                     `child_ancestors`.`child_group_id` = `parent_ancestors`.`child_group_id`)
        WHERE `child_ancestors`.`ancestor_group_id` = OLD.`child_group_id`;
    END IF;
    IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR OLD.type != NEW.type) THEN
        INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    UPDATE `history_groups_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_groups_groups` (
        `id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`status_date`,`inviting_user_id`,`deleted`
    ) VALUES (
                 OLD.`id`,@curVersion,OLD.`parent_group_id`,OLD.`child_group_id`,OLD.`child_order`,OLD.`type`,OLD.`role`,
                 OLD.`status_date`,OLD.`inviting_user_id`, 1
             );
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
        SELECT `groups_ancestors`.`child_group_id`, 'todo'
        FROM `groups_ancestors`
        WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
    ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    DELETE `groups_ancestors` FROM `groups_ancestors`
    WHERE `groups_ancestors`.`child_group_id` = OLD.`child_group_id` AND
            `groups_ancestors`.`ancestor_group_id` = OLD.`parent_group_id`;
    DELETE `bridges`
    FROM `groups_ancestors` `child_descendants`
             JOIN `groups_ancestors` `parent_ancestors`
             JOIN `groups_ancestors` `bridges`
                  ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
                      `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
    WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
            `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;
    DELETE `child_ancestors`
    FROM `groups_ancestors` `child_ancestors`
             JOIN  `groups_ancestors` `parent_ancestors`
                   ON (`child_ancestors`.`child_group_id` = OLD.`child_group_id` AND
                       `child_ancestors`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id`)
    WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id`;
    DELETE `parent_ancestors`
    FROM `groups_ancestors` `parent_ancestors`
             JOIN  `groups_ancestors` `child_ancestors`
                   ON (`parent_ancestors`.`ancestor_group_id` = OLD.`parent_group_id` AND
                       `child_ancestors`.`child_group_id` = `parent_ancestors`.`child_group_id`)
    WHERE `child_ancestors`.`ancestor_group_id` = OLD.`child_group_id`;
END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN
    INSERT INTO `history_groups_items` (
        `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_date`,`full_access_date`,`access_reason`,
        `access_solutions_date`,`owner_access`,`manager_access`,`cached_partial_access_date`,`cached_full_access_date`,
        `cached_access_solutions_date`,`cached_grayed_access_date`,`cached_full_access`,`cached_partial_access`,
        `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`
    ) VALUES (
                 NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_date`,
                 NEW.`full_access_date`,NEW.`access_reason`,NEW.`access_solutions_date`,NEW.`owner_access`,
                 NEW.`manager_access`,NEW.`cached_partial_access_date`,NEW.`cached_full_access_date`,
                 NEW.`cached_access_solutions_date`,NEW.`cached_grayed_access_date`,NEW.`cached_full_access`,
                 NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,NEW.`cached_manager_access`
             );
    INSERT INTO `groups_items_propagate` (`id`, `propagate_access`) VALUE (NEW.`id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_access` = 'self';
END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_items` BEFORE UPDATE ON `groups_items` FOR EACH ROW BEGIN
    IF NEW.version <> OLD.version THEN
        SET @curVersion = NEW.version;
    ELSE
        SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    END IF;
    IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`item_id` <=> NEW.`item_id` AND
            OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`partial_access_date` <=> NEW.`partial_access_date` AND
            OLD.`full_access_date` <=> NEW.`full_access_date` AND OLD.`access_reason` <=> NEW.`access_reason` AND
            OLD.`access_solutions_date` <=> NEW.`access_solutions_date` AND OLD.`owner_access` <=> NEW.`owner_access` AND
            OLD.`manager_access` <=> NEW.`manager_access` AND OLD.`cached_partial_access_date` <=> NEW.`cached_partial_access_date` AND
            OLD.`cached_full_access_date` <=> NEW.`cached_full_access_date` AND OLD.`cached_access_solutions_date` <=> NEW.`cached_access_solutions_date` AND
            OLD.`cached_grayed_access_date` <=> NEW.`cached_grayed_access_date` AND OLD.`cached_full_access` <=> NEW.`cached_full_access` AND
            OLD.`cached_partial_access` <=> NEW.`cached_partial_access` AND OLD.`cached_access_solutions` <=> NEW.`cached_access_solutions` AND
            OLD.`cached_grayed_access` <=> NEW.`cached_grayed_access` AND OLD.`cached_manager_access` <=> NEW.`cached_manager_access`) THEN
        SET NEW.version = @curVersion;
        UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
        INSERT INTO `history_groups_items` (
            `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_date`,`full_access_date`,`access_reason`,
            `access_solutions_date`,`owner_access`,`manager_access`,`cached_partial_access_date`,`cached_full_access_date`,
            `cached_access_solutions_date`,`cached_grayed_access_date`,`cached_full_access`,`cached_partial_access`,
            `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`
        ) VALUES (
                     NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_date`,
                     NEW.`full_access_date`,NEW.`access_reason`,NEW.`access_solutions_date`,NEW.`owner_access`,
                     NEW.`manager_access`,NEW.`cached_partial_access_date`,NEW.`cached_full_access_date`,
                     NEW.`cached_access_solutions_date`,NEW.`cached_grayed_access_date`,NEW.`cached_full_access`,
                     NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,
                     NEW.`cached_manager_access`
                 );
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER `after_update_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_groups_items` AFTER UPDATE ON `groups_items` FOR EACH ROW BEGIN
    # As a date change may result in access change for descendants of the item, mark the entry as to be recomputed
    IF NOT (NEW.`full_access_date` <=> OLD.`full_access_date`AND NEW.`partial_access_date` <=> OLD.`partial_access_date`AND
            NEW.`access_solutions_date` <=> OLD.`access_solutions_date`AND NEW.`manager_access` <=> OLD.`manager_access`AND
            NEW.`access_reason` <=> OLD.`access_reason`) THEN
        INSERT INTO `groups_items_propagate` (`id`, `propagate_access`) VALUE (NEW.`id`, 'self')
        ON DUPLICATE KEY UPDATE `propagate_access` = 'self';
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_items` BEFORE DELETE ON `groups_items` FOR EACH ROW BEGIN
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_groups_items` (
        `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_date`,`full_access_date`,`access_reason`,
        `access_solutions_date`,`owner_access`,`manager_access`,`cached_partial_access_date`,`cached_full_access_date`,
        `cached_access_solutions_date`,`cached_grayed_access_date`,`cached_full_access`,`cached_partial_access`,
        `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`deleted`
    ) VALUES (
                 OLD.`id`,@curVersion,OLD.`group_id`,OLD.`item_id`,OLD.`creator_user_id`,OLD.`partial_access_date`,
                 OLD.`full_access_date`,OLD.`access_reason`,OLD.`access_solutions_date`,OLD.`owner_access`,
                 OLD.`manager_access`,OLD.`cached_partial_access_date`,OLD.`cached_full_access_date`,
                 OLD.`cached_access_solutions_date`,OLD.`cached_grayed_access_date`,OLD.`cached_full_access`,
                 OLD.`cached_partial_access`,OLD.`cached_access_solutions`,OLD.`cached_grayed_access`,
                 OLD.`cached_manager_access`, 1);
END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items` AFTER INSERT ON `items` FOR EACH ROW BEGIN INSERT INTO `history_items` (`id`,`version`,`url`,`options`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`team_mode`,`teams_editable`,`qualified_group_id`,`team_max_members`,`has_attempts`,`access_open_date`,`duration`,`end_contest_date`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`) VALUES (NEW.`id`,@curVersion,NEW.`url`,NEW.`options`,NEW.`platform_id`,NEW.`text_id`,NEW.`repository_path`,NEW.`type`,NEW.`uses_api`,NEW.`read_only`,NEW.`full_screen`,NEW.`show_difficulty`,NEW.`show_source`,NEW.`hints_allowed`,NEW.`fixed_ranks`,NEW.`validation_type`,NEW.`validation_min`,NEW.`preparation_state`,NEW.`unlocked_item_ids`,NEW.`score_min_unlock`,NEW.`supported_lang_prog`,NEW.`default_language_id`,NEW.`team_mode`,NEW.`teams_editable`,NEW.`qualified_group_id`,NEW.`team_max_members`,NEW.`has_attempts`,NEW.`access_open_date`,NEW.`duration`,NEW.`end_contest_date`,NEW.`show_user_infos`,NEW.`contest_phase`,NEW.`level`,NEW.`no_score`,NEW.`title_bar_visible`,NEW.`transparent_folder`,NEW.`display_details_in_parent`,NEW.`display_children_as_tabs`,NEW.`custom_chapter`,NEW.`group_code_enter`); INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items` BEFORE UPDATE ON `items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`url` <=> NEW.`url` AND OLD.`options` <=> NEW.`options` AND OLD.`platform_id` <=> NEW.`platform_id` AND OLD.`text_id` <=> NEW.`text_id` AND OLD.`repository_path` <=> NEW.`repository_path` AND OLD.`type` <=> NEW.`type` AND OLD.`uses_api` <=> NEW.`uses_api` AND OLD.`read_only` <=> NEW.`read_only` AND OLD.`full_screen` <=> NEW.`full_screen` AND OLD.`show_difficulty` <=> NEW.`show_difficulty` AND OLD.`show_source` <=> NEW.`show_source` AND OLD.`hints_allowed` <=> NEW.`hints_allowed` AND OLD.`fixed_ranks` <=> NEW.`fixed_ranks` AND OLD.`validation_type` <=> NEW.`validation_type` AND OLD.`validation_min` <=> NEW.`validation_min` AND OLD.`preparation_state` <=> NEW.`preparation_state` AND OLD.`unlocked_item_ids` <=> NEW.`unlocked_item_ids` AND OLD.`score_min_unlock` <=> NEW.`score_min_unlock` AND OLD.`supported_lang_prog` <=> NEW.`supported_lang_prog` AND OLD.`default_language_id` <=> NEW.`default_language_id` AND OLD.`team_mode` <=> NEW.`team_mode` AND OLD.`teams_editable` <=> NEW.`teams_editable` AND OLD.`qualified_group_id` <=> NEW.`qualified_group_id` AND OLD.`team_max_members` <=> NEW.`team_max_members` AND OLD.`has_attempts` <=> NEW.`has_attempts` AND OLD.`access_open_date` <=> NEW.`access_open_date` AND OLD.`duration` <=> NEW.`duration` AND OLD.`end_contest_date` <=> NEW.`end_contest_date` AND OLD.`show_user_infos` <=> NEW.`show_user_infos` AND OLD.`contest_phase` <=> NEW.`contest_phase` AND OLD.`level` <=> NEW.`level` AND OLD.`no_score` <=> NEW.`no_score` AND OLD.`title_bar_visible` <=> NEW.`title_bar_visible` AND OLD.`transparent_folder` <=> NEW.`transparent_folder` AND OLD.`display_details_in_parent` <=> NEW.`display_details_in_parent` AND OLD.`display_children_as_tabs` <=> NEW.`display_children_as_tabs` AND OLD.`custom_chapter` <=> NEW.`custom_chapter` AND OLD.`group_code_enter` <=> NEW.`group_code_enter`) THEN   SET NEW.version = @curVersion;   UPDATE `history_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_items` (`id`,`version`,`url`,`options`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`team_mode`,`teams_editable`,`qualified_group_id`,`team_max_members`,`has_attempts`,`access_open_date`,`duration`,`end_contest_date`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`)       VALUES (NEW.`id`,@curVersion,NEW.`url`,NEW.`options`,NEW.`platform_id`,NEW.`text_id`,NEW.`repository_path`,NEW.`type`,NEW.`uses_api`,NEW.`read_only`,NEW.`full_screen`,NEW.`show_difficulty`,NEW.`show_source`,NEW.`hints_allowed`,NEW.`fixed_ranks`,NEW.`validation_type`,NEW.`validation_min`,NEW.`preparation_state`,NEW.`unlocked_item_ids`,NEW.`score_min_unlock`,NEW.`supported_lang_prog`,NEW.`default_language_id`,NEW.`team_mode`,NEW.`teams_editable`,NEW.`qualified_group_id`,NEW.`team_max_members`,NEW.`has_attempts`,NEW.`access_open_date`,NEW.`duration`,NEW.`end_contest_date`,NEW.`show_user_infos`,NEW.`contest_phase`,NEW.`level`,NEW.`no_score`,NEW.`title_bar_visible`,NEW.`transparent_folder`,NEW.`display_details_in_parent`,NEW.`display_children_as_tabs`,NEW.`custom_chapter`,NEW.`group_code_enter`) ; END IF; SELECT platforms.id INTO @platformID FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1 ; SET NEW.platform_id=@platformID ; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items` BEFORE DELETE ON `items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_items` (`id`,`version`,`url`,`options`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`team_mode`,`teams_editable`,`qualified_group_id`,`team_max_members`,`has_attempts`,`access_open_date`,`duration`,`end_contest_date`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`url`,OLD.`options`,OLD.`platform_id`,OLD.`text_id`,OLD.`repository_path`,OLD.`type`,OLD.`uses_api`,OLD.`read_only`,OLD.`full_screen`,OLD.`show_difficulty`,OLD.`show_source`,OLD.`hints_allowed`,OLD.`fixed_ranks`,OLD.`validation_type`,OLD.`validation_min`,OLD.`preparation_state`,OLD.`unlocked_item_ids`,OLD.`score_min_unlock`,OLD.`supported_lang_prog`,OLD.`default_language_id`,OLD.`team_mode`,OLD.`teams_editable`,OLD.`qualified_group_id`,OLD.`team_max_members`,OLD.`has_attempts`,OLD.`access_open_date`,OLD.`duration`,OLD.`end_contest_date`,OLD.`show_user_infos`,OLD.`contest_phase`,OLD.`level`,OLD.`no_score`,OLD.`title_bar_visible`,OLD.`transparent_folder`,OLD.`display_details_in_parent`,OLD.`display_children_as_tabs`,OLD.`custom_chapter`,OLD.`group_code_enter`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_messages` AFTER INSERT ON `messages` FOR EACH ROW BEGIN INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submission_date`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`) VALUES (NEW.`id`,@curVersion,NEW.`thread_id`,NEW.`user_id`,NEW.`submission_date`,NEW.`published`,NEW.`title`,NEW.`body`,NEW.`trainers_only`,NEW.`archived`,NEW.`persistant`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_messages` BEFORE UPDATE ON `messages` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`thread_id` <=> NEW.`thread_id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`submission_date` <=> NEW.`submission_date` AND OLD.`published` <=> NEW.`published` AND OLD.`title` <=> NEW.`title` AND OLD.`body` <=> NEW.`body` AND OLD.`trainers_only` <=> NEW.`trainers_only` AND OLD.`archived` <=> NEW.`archived` AND OLD.`persistant` <=> NEW.`persistant`) THEN   SET NEW.version = @curVersion;   UPDATE `history_messages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submission_date`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`)       VALUES (NEW.`id`,@curVersion,NEW.`thread_id`,NEW.`user_id`,NEW.`submission_date`,NEW.`published`,NEW.`title`,NEW.`body`,NEW.`trainers_only`,NEW.`archived`,NEW.`persistant`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_messages` BEFORE DELETE ON `messages` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_messages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submission_date`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`thread_id`,OLD.`user_id`,OLD.`submission_date`,OLD.`published`,OLD.`title`,OLD.`body`,OLD.`trainers_only`,OLD.`archived`,OLD.`persistant`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_threads` AFTER INSERT ON `threads` FOR EACH ROW BEGIN INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`last_activity_date`) VALUES (NEW.`id`,@curVersion,NEW.`type`,NEW.`creator_user_id`,NEW.`item_id`,NEW.`title`,NEW.`admin_help_asked`,NEW.`hidden`,NEW.`last_activity_date`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_threads` BEFORE UPDATE ON `threads` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`type` <=> NEW.`type` AND OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`title` <=> NEW.`title` AND OLD.`admin_help_asked` <=> NEW.`admin_help_asked` AND OLD.`hidden` <=> NEW.`hidden` AND OLD.`last_activity_date` <=> NEW.`last_activity_date`) THEN   SET NEW.version = @curVersion;   UPDATE `history_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`last_activity_date`)       VALUES (NEW.`id`,@curVersion,NEW.`type`,NEW.`creator_user_id`,NEW.`item_id`,NEW.`title`,NEW.`admin_help_asked`,NEW.`hidden`,NEW.`last_activity_date`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_threads` BEFORE DELETE ON `threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`last_activity_date`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`type`,OLD.`creator_user_id`,OLD.`item_id`,OLD.`title`,OLD.`admin_help_asked`,OLD.`hidden`,OLD.`last_activity_date`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users` AFTER INSERT ON `users` FOR EACH ROW BEGIN INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registration_date`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`last_login_date`,`last_activity_date`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notification_read_date`,`login_module_prefix`,`allow_subgroups`) VALUES (NEW.`id`,@curVersion,NEW.`login`,NEW.`open_id_identity`,NEW.`password_md5`,NEW.`salt`,NEW.`recover`,NEW.`registration_date`,NEW.`email`,NEW.`email_verified`,NEW.`first_name`,NEW.`last_name`,NEW.`country_code`,NEW.`time_zone`,NEW.`birth_date`,NEW.`graduation_year`,NEW.`grade`,NEW.`sex`,NEW.`student_id`,NEW.`address`,NEW.`zipcode`,NEW.`city`,NEW.`land_line_number`,NEW.`cell_phone_number`,NEW.`default_language`,NEW.`notify_news`,NEW.`notify`,NEW.`public_first_name`,NEW.`public_last_name`,NEW.`free_text`,NEW.`web_site`,NEW.`photo_autoload`,NEW.`lang_prog`,NEW.`last_login_date`,NEW.`last_activity_date`,NEW.`last_ip`,NEW.`basic_editor_mode`,NEW.`spaces_for_tab`,NEW.`member_state`,NEW.`godfather_user_id`,NEW.`step_level_in_site`,NEW.`is_admin`,NEW.`no_ranking`,NEW.`help_given`,NEW.`self_group_id`,NEW.`owned_group_id`,NEW.`access_group_id`,NEW.`notification_read_date`,NEW.`login_module_prefix`,NEW.`allow_subgroups`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users` BEFORE UPDATE ON `users` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`login` <=> NEW.`login` AND OLD.`open_id_identity` <=> NEW.`open_id_identity` AND OLD.`password_md5` <=> NEW.`password_md5` AND OLD.`salt` <=> NEW.`salt` AND OLD.`recover` <=> NEW.`recover` AND OLD.`registration_date` <=> NEW.`registration_date` AND OLD.`email` <=> NEW.`email` AND OLD.`email_verified` <=> NEW.`email_verified` AND OLD.`first_name` <=> NEW.`first_name` AND OLD.`last_name` <=> NEW.`last_name` AND OLD.`country_code` <=> NEW.`country_code` AND OLD.`time_zone` <=> NEW.`time_zone` AND OLD.`birth_date` <=> NEW.`birth_date` AND OLD.`graduation_year` <=> NEW.`graduation_year` AND OLD.`grade` <=> NEW.`grade` AND OLD.`sex` <=> NEW.`sex` AND OLD.`student_id` <=> NEW.`student_id` AND OLD.`address` <=> NEW.`address` AND OLD.`zipcode` <=> NEW.`zipcode` AND OLD.`city` <=> NEW.`city` AND OLD.`land_line_number` <=> NEW.`land_line_number` AND OLD.`cell_phone_number` <=> NEW.`cell_phone_number` AND OLD.`default_language` <=> NEW.`default_language` AND OLD.`notify_news` <=> NEW.`notify_news` AND OLD.`notify` <=> NEW.`notify` AND OLD.`public_first_name` <=> NEW.`public_first_name` AND OLD.`public_last_name` <=> NEW.`public_last_name` AND OLD.`free_text` <=> NEW.`free_text` AND OLD.`web_site` <=> NEW.`web_site` AND OLD.`photo_autoload` <=> NEW.`photo_autoload` AND OLD.`lang_prog` <=> NEW.`lang_prog` AND OLD.`last_login_date` <=> NEW.`last_login_date` AND OLD.`last_activity_date` <=> NEW.`last_activity_date` AND OLD.`last_ip` <=> NEW.`last_ip` AND OLD.`basic_editor_mode` <=> NEW.`basic_editor_mode` AND OLD.`spaces_for_tab` <=> NEW.`spaces_for_tab` AND OLD.`member_state` <=> NEW.`member_state` AND OLD.`godfather_user_id` <=> NEW.`godfather_user_id` AND OLD.`step_level_in_site` <=> NEW.`step_level_in_site` AND OLD.`is_admin` <=> NEW.`is_admin` AND OLD.`no_ranking` <=> NEW.`no_ranking` AND OLD.`help_given` <=> NEW.`help_given` AND OLD.`self_group_id` <=> NEW.`self_group_id` AND OLD.`owned_group_id` <=> NEW.`owned_group_id` AND OLD.`access_group_id` <=> NEW.`access_group_id` AND OLD.`notification_read_date` <=> NEW.`notification_read_date` AND OLD.`login_module_prefix` <=> NEW.`login_module_prefix` AND OLD.`allow_subgroups` <=> NEW.`allow_subgroups`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registration_date`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`last_login_date`,`last_activity_date`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notification_read_date`,`login_module_prefix`,`allow_subgroups`)       VALUES (NEW.`id`,@curVersion,NEW.`login`,NEW.`open_id_identity`,NEW.`password_md5`,NEW.`salt`,NEW.`recover`,NEW.`registration_date`,NEW.`email`,NEW.`email_verified`,NEW.`first_name`,NEW.`last_name`,NEW.`country_code`,NEW.`time_zone`,NEW.`birth_date`,NEW.`graduation_year`,NEW.`grade`,NEW.`sex`,NEW.`student_id`,NEW.`address`,NEW.`zipcode`,NEW.`city`,NEW.`land_line_number`,NEW.`cell_phone_number`,NEW.`default_language`,NEW.`notify_news`,NEW.`notify`,NEW.`public_first_name`,NEW.`public_last_name`,NEW.`free_text`,NEW.`web_site`,NEW.`photo_autoload`,NEW.`lang_prog`,NEW.`last_login_date`,NEW.`last_activity_date`,NEW.`last_ip`,NEW.`basic_editor_mode`,NEW.`spaces_for_tab`,NEW.`member_state`,NEW.`godfather_user_id`,NEW.`step_level_in_site`,NEW.`is_admin`,NEW.`no_ranking`,NEW.`help_given`,NEW.`self_group_id`,NEW.`owned_group_id`,NEW.`access_group_id`,NEW.`notification_read_date`,NEW.`login_module_prefix`,NEW.`allow_subgroups`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users` BEFORE DELETE ON `users` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registration_date`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`last_login_date`,`last_activity_date`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notification_read_date`,`login_module_prefix`,`allow_subgroups`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`login`,OLD.`open_id_identity`,OLD.`password_md5`,OLD.`salt`,OLD.`recover`,OLD.`registration_date`,OLD.`email`,OLD.`email_verified`,OLD.`first_name`,OLD.`last_name`,OLD.`country_code`,OLD.`time_zone`,OLD.`birth_date`,OLD.`graduation_year`,OLD.`grade`,OLD.`sex`,OLD.`student_id`,OLD.`address`,OLD.`zipcode`,OLD.`city`,OLD.`land_line_number`,OLD.`cell_phone_number`,OLD.`default_language`,OLD.`notify_news`,OLD.`notify`,OLD.`public_first_name`,OLD.`public_last_name`,OLD.`free_text`,OLD.`web_site`,OLD.`photo_autoload`,OLD.`lang_prog`,OLD.`last_login_date`,OLD.`last_activity_date`,OLD.`last_ip`,OLD.`basic_editor_mode`,OLD.`spaces_for_tab`,OLD.`member_state`,OLD.`godfather_user_id`,OLD.`step_level_in_site`,OLD.`is_admin`,OLD.`no_ranking`,OLD.`help_given`,OLD.`self_group_id`,OLD.`owned_group_id`,OLD.`access_group_id`,OLD.`notification_read_date`,OLD.`login_module_prefix`,OLD.`allow_subgroups`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users_items` AFTER INSERT ON `users_items` FOR EACH ROW BEGIN INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`,`state`,`answer`) VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`item_id`,NEW.`active_attempt_id`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`start_date`,NEW.`validation_date`,NEW.`best_answer_date`,NEW.`last_answer_date`,NEW.`thread_start_date`,NEW.`last_hint_date`,NEW.`finish_date`,NEW.`last_activity_date`,NEW.`contest_start_date`,NEW.`ranked`,NEW.`all_lang_prog`,NEW.`state`,NEW.`answer`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users_items` BEFORE UPDATE ON `users_items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`active_attempt_id` <=> NEW.`active_attempt_id` AND OLD.`score` <=> NEW.`score` AND OLD.`score_computed` <=> NEW.`score_computed` AND OLD.`score_reeval` <=> NEW.`score_reeval` AND OLD.`score_diff_manual` <=> NEW.`score_diff_manual` AND OLD.`score_diff_comment` <=> NEW.`score_diff_comment` AND OLD.`tasks_tried` <=> NEW.`tasks_tried` AND OLD.`children_validated` <=> NEW.`children_validated` AND OLD.`validated` <=> NEW.`validated` AND OLD.`finished` <=> NEW.`finished` AND OLD.`key_obtained` <=> NEW.`key_obtained` AND OLD.`tasks_with_help` <=> NEW.`tasks_with_help` AND OLD.`hints_requested` <=> NEW.`hints_requested` AND OLD.`hints_cached` <=> NEW.`hints_cached` AND OLD.`corrections_read` <=> NEW.`corrections_read` AND OLD.`precision` <=> NEW.`precision` AND OLD.`autonomy` <=> NEW.`autonomy` AND OLD.`start_date` <=> NEW.`start_date` AND OLD.`validation_date` <=> NEW.`validation_date` AND OLD.`best_answer_date` <=> NEW.`best_answer_date` AND OLD.`last_answer_date` <=> NEW.`last_answer_date` AND OLD.`thread_start_date` <=> NEW.`thread_start_date` AND OLD.`last_hint_date` <=> NEW.`last_hint_date` AND OLD.`finish_date` <=> NEW.`finish_date` AND OLD.`contest_start_date` <=> NEW.`contest_start_date` AND OLD.`ranked` <=> NEW.`ranked` AND OLD.`all_lang_prog` <=> NEW.`all_lang_prog` AND OLD.`state` <=> NEW.`state` AND OLD.`answer` <=> NEW.`answer`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`,`state`,`answer`)       VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`item_id`,NEW.`active_attempt_id`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`start_date`,NEW.`validation_date`,NEW.`best_answer_date`,NEW.`last_answer_date`,NEW.`thread_start_date`,NEW.`last_hint_date`,NEW.`finish_date`,NEW.`last_activity_date`,NEW.`contest_start_date`,NEW.`ranked`,NEW.`all_lang_prog`,NEW.`state`,NEW.`answer`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users_items` BEFORE DELETE ON `users_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`,`state`,`answer`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`user_id`,OLD.`item_id`,OLD.`active_attempt_id`,OLD.`score`,OLD.`score_computed`,OLD.`score_reeval`,OLD.`score_diff_manual`,OLD.`score_diff_comment`,OLD.`submissions_attempts`,OLD.`tasks_tried`,OLD.`children_validated`,OLD.`validated`,OLD.`finished`,OLD.`key_obtained`,OLD.`tasks_with_help`,OLD.`hints_requested`,OLD.`hints_cached`,OLD.`corrections_read`,OLD.`precision`,OLD.`autonomy`,OLD.`start_date`,OLD.`validation_date`,OLD.`best_answer_date`,OLD.`last_answer_date`,OLD.`thread_start_date`,OLD.`last_hint_date`,OLD.`finish_date`,OLD.`last_activity_date`,OLD.`contest_start_date`,OLD.`ranked`,OLD.`all_lang_prog`,OLD.`state`,OLD.`answer`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users_threads` AFTER INSERT ON `users_threads` FOR EACH ROW BEGIN INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`last_read_date`,`last_write_date`,`starred`) VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`thread_id`,NEW.`last_read_date`,NEW.`last_write_date`,NEW.`starred`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users_threads` BEFORE UPDATE ON `users_threads` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`thread_id` <=> NEW.`thread_id` AND OLD.`last_read_date` <=> NEW.`last_read_date` AND OLD.`last_write_date` <=> NEW.`last_write_date` AND OLD.`starred` <=> NEW.`starred`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`last_read_date`,`last_write_date`,`starred`)       VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`thread_id`,NEW.`last_read_date`,NEW.`last_write_date`,NEW.`starred`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users_threads` BEFORE DELETE ON `users_threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`last_read_date`,`last_write_date`,`starred`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`user_id`,OLD.`thread_id`,OLD.`last_read_date`,OLD.`last_write_date`,OLD.`starred`, 1); END
-- +migrate StatementEnd

ALTER ALGORITHM=UNDEFINED
  SQL SECURITY DEFINER
  VIEW `task_children_data_view` AS
SELECT
    `parent_users_items`.`id` AS `user_item_id`,
    SUM(IF(`task_children`.`id` IS NOT NULL AND `task_children`.`validated`, 1, 0)) AS `children_validated`,
    SUM(IF(`task_children`.`id` IS NOT NULL AND `task_children`.`validated`, 0, 1)) AS `children_non_validated`,
    SUM(IF(`items_items`.`category` = 'Validation' AND
           (ISNULL(`task_children`.`id`) OR `task_children`.`validated` != 1), 1, 0)) AS `children_category`,
    MAX(`task_children`.`validation_date`) AS `max_validation_date`,
    MAX(IF(`items_items`.`category` = 'Validation', `task_children`.`validation_date`, NULL)) AS `max_validation_date_categories`
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
