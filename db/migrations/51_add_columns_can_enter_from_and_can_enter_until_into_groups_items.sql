-- +migrate Up
ALTER TABLE `groups_items`
    ADD COLUMN `can_enter_from` datetime DEFAULT NULL COMMENT 'Time from which the group can “enter” this time-limited item',
    ADD COLUMN `can_enter_until` datetime DEFAULT NULL COMMENT 'Time until which the group can “enter” this time-limited item',
    ADD COLUMN `contest_started_at` datetime DEFAULT NULL COMMENT 'Time at which the group entered the contest';
ALTER TABLE `history_groups_items`
    ADD COLUMN `can_enter_from` datetime DEFAULT NULL,
    ADD COLUMN `can_enter_until` datetime DEFAULT NULL,
    ADD COLUMN `contest_started_at` datetime DEFAULT NULL;

DROP TRIGGER `after_insert_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN
    INSERT INTO `history_groups_items` (
        `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_since`,`full_access_since`,`access_reason`,
        `solutions_access_since`,`owner_access`,`manager_access`,`cached_partial_access_since`,`cached_full_access_since`,
        `cached_solutions_access_since`,`cached_grayed_access_since`,`cached_full_access`,`cached_partial_access`,
        `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`can_enter_from`,`can_enter_until`,
        `contest_started_at`
    ) VALUES (
                 NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_since`,
                 NEW.`full_access_since`,NEW.`access_reason`,NEW.`solutions_access_since`,NEW.`owner_access`,
                 NEW.`manager_access`,NEW.`cached_partial_access_since`,NEW.`cached_full_access_since`,
                 NEW.`cached_solutions_access_since`,NEW.`cached_grayed_access_since`,NEW.`cached_full_access`,
                 NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,NEW.`cached_manager_access`,
                 NEW.`can_enter_from`,NEW.`can_enter_until`,NEW.`contest_started_at`
             );
    INSERT INTO `groups_items_propagate` (`id`, `propagate_access`) VALUE (NEW.`id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_access` = 'self';
END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_update_groups_items` BEFORE UPDATE ON `groups_items` FOR EACH ROW BEGIN
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
            OLD.`cached_grayed_access` <=> NEW.`cached_grayed_access` AND OLD.`cached_manager_access` <=> NEW.`cached_manager_access` AND
            OLD.`can_enter_from` <=> NEW.`can_enter_from` AND OLD.`can_enter_until` <=> NEW.`can_enter_until` AND
            OLD.`contest_started_at` <=> NEW.`contest_started_at`) THEN
        SET NEW.version = @curVersion;
        UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
        INSERT INTO `history_groups_items` (
            `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_since`,`full_access_since`,`access_reason`,
            `solutions_access_since`,`owner_access`,`manager_access`,`cached_partial_access_since`,`cached_full_access_since`,
            `cached_solutions_access_since`,`cached_grayed_access_since`,`cached_full_access`,`cached_partial_access`,
            `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`can_enter_from`,`can_enter_until`,
            `contest_started_at`
        ) VALUES (
                     NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_since`,
                     NEW.`full_access_since`,NEW.`access_reason`,NEW.`solutions_access_since`,NEW.`owner_access`,
                     NEW.`manager_access`,NEW.`cached_partial_access_since`,NEW.`cached_full_access_since`,
                     NEW.`cached_solutions_access_since`,NEW.`cached_grayed_access_since`,NEW.`cached_full_access`,
                     NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,
                     NEW.`cached_manager_access`,NEW.`can_enter_from`,NEW.`can_enter_until`,NEW.`contest_started_at`
                 );
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_delete_groups_items` BEFORE DELETE ON `groups_items` FOR EACH ROW BEGIN
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_groups_items` (
        `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_since`,`full_access_since`,`access_reason`,
        `solutions_access_since`,`owner_access`,`manager_access`,`cached_partial_access_since`,`cached_full_access_since`,
        `cached_solutions_access_since`,`cached_grayed_access_since`,`cached_full_access`,`cached_partial_access`,
        `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`can_enter_from`,`can_enter_until`,
        `contest_started_at`,`deleted`
    ) VALUES (
                 OLD.`id`,@curVersion,OLD.`group_id`,OLD.`item_id`,OLD.`creator_user_id`,OLD.`partial_access_since`,
                 OLD.`full_access_since`,OLD.`access_reason`,OLD.`solutions_access_since`,OLD.`owner_access`,
                 OLD.`manager_access`,OLD.`cached_partial_access_since`,OLD.`cached_full_access_since`,
                 OLD.`cached_solutions_access_since`,OLD.`cached_grayed_access_since`,OLD.`cached_full_access`,
                 OLD.`cached_partial_access`,OLD.`cached_access_solutions`,OLD.`cached_grayed_access`,
                 OLD.`cached_manager_access`,OLD.`can_enter_from`,OLD.`can_enter_until`,OLD.`contest_started_at`, 1);
END
-- +migrate StatementEnd

INSERT INTO `groups_items` (`group_id`, `item_id`, `contest_started_at`)
    SELECT users.self_group_id, users.self_group_id, users_items.contest_started_at
    FROM users_items
         JOIN users ON users.id = users_items.user_id
    WHERE users.self_group_id IS NOT NULL
ON DUPLICATE KEY UPDATE contest_started_at = users_items.contest_started_at;


-- +migrate Down
ALTER TABLE `groups_items`
    DROP COLUMN `can_enter_from`,
    DROP COLUMN `can_enter_until`,
    DROP COLUMN `contest_started_at`;
ALTER TABLE `history_groups_items`
    DROP COLUMN `can_enter_from`,
    DROP COLUMN `can_enter_until`,
    DROP COLUMN `contest_started_at`;

DROP TRIGGER `after_insert_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN
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
CREATE DEFINER=`algorea`@`%` TRIGGER `before_update_groups_items` BEFORE UPDATE ON `groups_items` FOR EACH ROW BEGIN
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
DROP TRIGGER `before_delete_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_delete_groups_items` BEFORE DELETE ON `groups_items` FOR EACH ROW BEGIN
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
