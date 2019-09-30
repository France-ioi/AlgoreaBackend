-- +migrate Up
ALTER TABLE `groups`
    RENAME COLUMN `code_timer` TO `code_lifetime`;
ALTER TABLE `history_groups`
    RENAME COLUMN `code_timer` TO `code_lifetime`;

DROP TRIGGER `after_insert_groups`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_lifetime`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`) VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`created_at`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_lifetime`,NEW.`code_expires_at`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`); INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`name` <=> NEW.`name` AND OLD.`grade` <=> NEW.`grade` AND OLD.`grade_details` <=> NEW.`grade_details` AND OLD.`description` <=> NEW.`description` AND OLD.`created_at` <=> NEW.`created_at` AND OLD.`opened` <=> NEW.`opened` AND OLD.`free_access` <=> NEW.`free_access` AND OLD.`team_item_id` <=> NEW.`team_item_id` AND OLD.`team_participating` <=> NEW.`team_participating` AND OLD.`code` <=> NEW.`code` AND OLD.`code_lifetime` <=> NEW.`code_lifetime` AND OLD.`code_expires_at` <=> NEW.`code_expires_at` AND OLD.`redirect_path` <=> NEW.`redirect_path` AND OLD.`open_contest` <=> NEW.`open_contest` AND OLD.`type` <=> NEW.`type` AND OLD.`send_emails` <=> NEW.`send_emails`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_lifetime`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`)       VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`created_at`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_lifetime`,NEW.`code_expires_at`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_delete_groups` BEFORE DELETE ON `groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_lifetime`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`name`,OLD.`grade`,OLD.`grade_details`,OLD.`description`,OLD.`created_at`,OLD.`opened`,OLD.`free_access`,OLD.`team_item_id`,OLD.`team_participating`,OLD.`code`,OLD.`code_lifetime`,OLD.`code_expires_at`,OLD.`redirect_path`,OLD.`open_contest`,OLD.`type`,OLD.`send_emails`, 1); END
-- +migrate StatementEnd


-- +migrate Down
ALTER TABLE `groups`
RENAME COLUMN `code_lifetime` TO `code_timer`;
ALTER TABLE `history_groups`
    RENAME COLUMN `code_lifetime` TO `code_timer`;

DROP TRIGGER `after_insert_groups`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`) VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`created_at`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_timer`,NEW.`code_expires_at`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`); INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`name` <=> NEW.`name` AND OLD.`grade` <=> NEW.`grade` AND OLD.`grade_details` <=> NEW.`grade_details` AND OLD.`description` <=> NEW.`description` AND OLD.`created_at` <=> NEW.`created_at` AND OLD.`opened` <=> NEW.`opened` AND OLD.`free_access` <=> NEW.`free_access` AND OLD.`team_item_id` <=> NEW.`team_item_id` AND OLD.`team_participating` <=> NEW.`team_participating` AND OLD.`code` <=> NEW.`code` AND OLD.`code_timer` <=> NEW.`code_timer` AND OLD.`code_expires_at` <=> NEW.`code_expires_at` AND OLD.`redirect_path` <=> NEW.`redirect_path` AND OLD.`open_contest` <=> NEW.`open_contest` AND OLD.`type` <=> NEW.`type` AND OLD.`send_emails` <=> NEW.`send_emails`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`)       VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`created_at`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_timer`,NEW.`code_expires_at`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_delete_groups` BEFORE DELETE ON `groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`name`,OLD.`grade`,OLD.`grade_details`,OLD.`description`,OLD.`created_at`,OLD.`opened`,OLD.`free_access`,OLD.`team_item_id`,OLD.`team_participating`,OLD.`code`,OLD.`code_timer`,OLD.`code_expires_at`,OLD.`redirect_path`,OLD.`open_contest`,OLD.`type`,OLD.`send_emails`, 1); END
-- +migrate StatementEnd
