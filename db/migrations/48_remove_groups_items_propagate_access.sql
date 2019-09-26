-- +migrate Up
UPDATE `groups_items_propagate` JOIN `groups_items` USING(`id`)
    SET `groups_items_propagate`.`propagate_access` = IF(
        `groups_items`.`propagate_access` = 'self',
        'self',
        `groups_items_propagate`.`propagate_access`
    );
INSERT IGNORE INTO `groups_items_propagate` SELECT id, propagate_access FROM `groups_items`;
DELETE FROM `groups_items_propagate` WHERE `propagate_access` = 'done';

ALTER TABLE `groups_items` DROP INDEX `propagate_access`, DROP COLUMN `propagate_access`;
ALTER TABLE `history_groups_items` DROP COLUMN `propagate_access`;
ALTER TABLE `groups_items_propagate` MODIFY COLUMN `propagate_access` enum('self','children') NOT NULL;

DROP TRIGGER `before_insert_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_insert_groups_items` BEFORE INSERT ON `groups_items` FOR EACH ROW BEGIN
    IF (NEW.id IS NULL OR NEW.id = 0) THEN
        SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;
    END IF;
    SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion;
    SET NEW.version = @curVersion;
END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN
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
CREATE DEFINER=`algorea`@`%` TRIGGER `before_update_groups_items` BEFORE UPDATE ON `groups_items` FOR EACH ROW BEGIN
    IF NEW.version <> OLD.version THEN
        SET @curVersion = NEW.version;
    ELSE
        SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion;
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
DROP TRIGGER IF EXISTS `after_update_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_update_groups_items` AFTER UPDATE ON `groups_items` FOR EACH ROW BEGIN
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
CREATE DEFINER=`algorea`@`%` TRIGGER `before_delete_groups_items` BEFORE DELETE ON `groups_items` FOR EACH ROW BEGIN
    SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion;
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

DROP TRIGGER IF EXISTS `after_insert_items_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                       `partial_access_propagation`,`difficulty`)
        VALUES (NEW.`id`,@curVersion,NEW.`parent_item_id`,NEW.`child_item_id`,NEW.`child_order`,NEW.`category`,
                NEW.`partial_access_propagation`,NEW.`difficulty`);
    INSERT IGNORE INTO `groups_items_propagate`
        SELECT `id`, 'children' as `propagate_access` FROM `groups_items`
        WHERE `groups_items`.`item_id` = NEW.`parent_item_id`;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_items_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_update_items_items` AFTER UPDATE ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `groups_items_propagate`
        SELECT `id`, 'children' as `propagate_access`
        FROM `groups_items`
        WHERE `groups_items`.`item_id` = NEW.`parent_item_id` OR `groups_items`.`item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS `before_delete_items_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
    SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion;
    UPDATE `history_items_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                       `partial_access_propagation`,`difficulty`, `deleted`)
        VALUES (OLD.`id`,@curVersion,OLD.`parent_item_id`,OLD.`child_item_id`,OLD.`child_order`,OLD.`category`,
                OLD.`partial_access_propagation`,OLD.`difficulty`, 1);
    INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
        VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
        VALUES (OLD.parent_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
        (
            SELECT `items_ancestors`.`child_item_id`, 'todo' FROM `items_ancestors`
            WHERE `items_ancestors`.`ancestor_item_id` = OLD.`child_item_id`
        ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    DELETE `items_ancestors` FROM `items_ancestors`
        WHERE `items_ancestors`.`child_item_id` = OLD.`child_item_id` AND
            `items_ancestors`.`ancestor_item_id` = OLD.`parent_item_id`;
    DELETE `bridges` FROM `items_ancestors` `child_descendants`
        JOIN `items_ancestors` `parent_ancestors`
        JOIN `items_ancestors` `bridges` ON (
            `bridges`.`ancestor_item_id` = `parent_ancestors`.`ancestor_item_id` AND
            `bridges`.`child_item_id` = `child_descendants`.`child_item_id`
        )
    WHERE `parent_ancestors`.`child_item_id` = OLD.`parent_item_id` AND
            `child_descendants`.`ancestor_item_id` = OLD.`child_item_id`;
    DELETE `child_ancestors` FROM `items_ancestors` `child_ancestors`
        JOIN  `items_ancestors` `parent_ancestors` ON (
            `child_ancestors`.`child_item_id` = OLD.`child_item_id` AND
            `child_ancestors`.`ancestor_item_id` = `parent_ancestors`.`ancestor_item_id`
        )
    WHERE `parent_ancestors`.`child_item_id` = OLD.`parent_item_id`;
    DELETE `parent_ancestors` FROM `items_ancestors` `parent_ancestors`
        JOIN  `items_ancestors` `child_ancestors` ON (
            `parent_ancestors`.`ancestor_item_id` = OLD.`parent_item_id` AND
            `child_ancestors`.`child_item_id` = `parent_ancestors`.`child_item_id`
        )
        WHERE `child_ancestors`.`ancestor_item_id` = OLD.`child_item_id`;
    INSERT IGNORE INTO `groups_items_propagate`
        SELECT `id`, 'children' as `propagate_access` FROM `groups_items`
        WHERE `groups_items`.`item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd

-- +migrate Down
ALTER TABLE `groups_items_propagate` MODIFY COLUMN `propagate_access` enum('self','children','done') NOT NULL;
ALTER TABLE `history_groups_items`
    ADD COLUMN `propagate_access` enum('self','children','done') NOT NULL DEFAULT 'self' AFTER `cached_manager_access`;
ALTER TABLE `groups_items`
    ADD COLUMN `propagate_access` enum('self','children','done') NOT NULL DEFAULT 'self'
        COMMENT 'Internal state used for access propagation' AFTER `cached_manager_access`,
    ADD INDEX `propagate_access` (`propagate_access`);
UPDATE `groups_items` JOIN `groups_items_propagate` USING(`id`)
    SET `groups_items`.`propagate_access` = IF(`groups_items_propagate`.`propagate_access` = 'self', 'self', 'done');

DROP TRIGGER `before_insert_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_insert_groups_items` BEFORE INSERT ON `groups_items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion;SET NEW.version = @curVersion; SET NEW.`propagate_access`='self' ; END
-- +migrate StatementEnd

DROP TRIGGER `after_insert_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN INSERT INTO `history_groups_items` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_date`,`full_access_date`,`access_reason`,`access_solutions_date`,`owner_access`,`manager_access`,`cached_partial_access_date`,`cached_full_access_date`,`cached_access_solutions_date`,`cached_grayed_access_date`,`cached_full_access`,`cached_partial_access`,`cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`propagate_access`) VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_date`,NEW.`full_access_date`,NEW.`access_reason`,NEW.`access_solutions_date`,NEW.`owner_access`,NEW.`manager_access`,NEW.`cached_partial_access_date`,NEW.`cached_full_access_date`,NEW.`cached_access_solutions_date`,NEW.`cached_grayed_access_date`,NEW.`cached_full_access`,NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,NEW.`cached_manager_access`,NEW.`propagate_access`); END
-- +migrate StatementEnd

DROP TRIGGER `before_update_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_update_groups_items` BEFORE UPDATE ON `groups_items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`partial_access_date` <=> NEW.`partial_access_date` AND OLD.`full_access_date` <=> NEW.`full_access_date` AND OLD.`access_reason` <=> NEW.`access_reason` AND OLD.`access_solutions_date` <=> NEW.`access_solutions_date` AND OLD.`owner_access` <=> NEW.`owner_access` AND OLD.`manager_access` <=> NEW.`manager_access` AND OLD.`cached_partial_access_date` <=> NEW.`cached_partial_access_date` AND OLD.`cached_full_access_date` <=> NEW.`cached_full_access_date` AND OLD.`cached_access_solutions_date` <=> NEW.`cached_access_solutions_date` AND OLD.`cached_grayed_access_date` <=> NEW.`cached_grayed_access_date` AND OLD.`cached_full_access` <=> NEW.`cached_full_access` AND OLD.`cached_partial_access` <=> NEW.`cached_partial_access` AND OLD.`cached_access_solutions` <=> NEW.`cached_access_solutions` AND OLD.`cached_grayed_access` <=> NEW.`cached_grayed_access` AND OLD.`cached_manager_access` <=> NEW.`cached_manager_access`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups_items` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_date`,`full_access_date`,`access_reason`,`access_solutions_date`,`owner_access`,`manager_access`,`cached_partial_access_date`,`cached_full_access_date`,`cached_access_solutions_date`,`cached_grayed_access_date`,`cached_full_access`,`cached_partial_access`,`cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`propagate_access`)       VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_date`,NEW.`full_access_date`,NEW.`access_reason`,NEW.`access_solutions_date`,NEW.`owner_access`,NEW.`manager_access`,NEW.`cached_partial_access_date`,NEW.`cached_full_access_date`,NEW.`cached_access_solutions_date`,NEW.`cached_grayed_access_date`,NEW.`cached_full_access`,NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,NEW.`cached_manager_access`,NEW.`propagate_access`) ; END IF; IF NOT (NEW.`full_access_date` <=> OLD.`full_access_date`AND NEW.`partial_access_date` <=> OLD.`partial_access_date`AND NEW.`access_solutions_date` <=> OLD.`access_solutions_date`AND NEW.`manager_access` <=> OLD.`manager_access`AND NEW.`access_reason` <=> OLD.`access_reason`)THEN SET NEW.`propagate_access` = 'self'; END IF; END
-- +migrate StatementEnd

DROP TRIGGER `after_update_groups_items`;

DROP TRIGGER `before_delete_groups_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_delete_groups_items` BEFORE DELETE ON `groups_items` FOR EACH ROW BEGIN SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion; UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups_items` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_date`,`full_access_date`,`access_reason`,`access_solutions_date`,`owner_access`,`manager_access`,`cached_partial_access_date`,`cached_full_access_date`,`cached_access_solutions_date`,`cached_grayed_access_date`,`cached_full_access`,`cached_partial_access`,`cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`propagate_access`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`group_id`,OLD.`item_id`,OLD.`creator_user_id`,OLD.`partial_access_date`,OLD.`full_access_date`,OLD.`access_reason`,OLD.`access_solutions_date`,OLD.`owner_access`,OLD.`manager_access`,OLD.`cached_partial_access_date`,OLD.`cached_full_access_date`,OLD.`cached_access_solutions_date`,OLD.`cached_grayed_access_date`,OLD.`cached_full_access`,OLD.`cached_partial_access`,OLD.`cached_access_solutions`,OLD.`cached_grayed_access`,OLD.`cached_manager_access`,OLD.`propagate_access`, 1); END
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS `after_insert_items_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                       `partial_access_propagation`,`difficulty`)
    VALUES (NEW.`id`,@curVersion,NEW.`parent_item_id`,NEW.`child_item_id`,NEW.`child_order`,NEW.`category`,
            NEW.`partial_access_propagation`,NEW.`difficulty`);
    INSERT IGNORE INTO `groups_items_propagate`
    SELECT `id`, 'children' as `propagate_access` FROM `groups_items`
    WHERE `groups_items`.`item_id` = NEW.`parent_item_id`
    ON DUPLICATE KEY UPDATE propagate_access='children';
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_items_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `after_update_items_items` AFTER UPDATE ON `items_items` FOR EACH ROW BEGIN INSERT IGNORE INTO `groups_items_propagate` SELECT `id`, 'children' as `propagate_access` FROM `groups_items` WHERE `groups_items`.`item_id` = NEW.`parent_item_id` OR `groups_items`.`item_id` = OLD.`parent_item_id` ON DUPLICATE KEY UPDATE propagate_access='children' ; END
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS `before_delete_items_items`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
    SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion;
    UPDATE `history_items_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                       `partial_access_propagation`,`difficulty`, `deleted`)
    VALUES (OLD.`id`,@curVersion,OLD.`parent_item_id`,OLD.`child_item_id`,OLD.`child_order`,OLD.`category`,
            OLD.`partial_access_propagation`,OLD.`difficulty`, 1);
    INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
    VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
    VALUES (OLD.parent_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
        (
            SELECT `items_ancestors`.`child_item_id`, 'todo' FROM `items_ancestors`
            WHERE `items_ancestors`.`ancestor_item_id` = OLD.`child_item_id`
        ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    DELETE `items_ancestors` FROM `items_ancestors`
    WHERE `items_ancestors`.`child_item_id` = OLD.`child_item_id` AND
            `items_ancestors`.`ancestor_item_id` = OLD.`parent_item_id`;
    DELETE `bridges` FROM `items_ancestors` `child_descendants`
                              JOIN `items_ancestors` `parent_ancestors`
                              JOIN `items_ancestors` `bridges` ON (
                `bridges`.`ancestor_item_id` = `parent_ancestors`.`ancestor_item_id` AND
                `bridges`.`child_item_id` = `child_descendants`.`child_item_id`
        )
    WHERE `parent_ancestors`.`child_item_id` = OLD.`parent_item_id` AND
            `child_descendants`.`ancestor_item_id` = OLD.`child_item_id`;
    DELETE `child_ancestors` FROM `items_ancestors` `child_ancestors`
                                      JOIN  `items_ancestors` `parent_ancestors` ON (
                `child_ancestors`.`child_item_id` = OLD.`child_item_id` AND
                `child_ancestors`.`ancestor_item_id` = `parent_ancestors`.`ancestor_item_id`
        )
    WHERE `parent_ancestors`.`child_item_id` = OLD.`parent_item_id`;
    DELETE `parent_ancestors` FROM `items_ancestors` `parent_ancestors`
                                       JOIN  `items_ancestors` `child_ancestors` ON (
                `parent_ancestors`.`ancestor_item_id` = OLD.`parent_item_id` AND
                `child_ancestors`.`child_item_id` = `parent_ancestors`.`child_item_id`
        )
    WHERE `child_ancestors`.`ancestor_item_id` = OLD.`child_item_id`;
    INSERT IGNORE INTO `groups_items_propagate`
    SELECT `id`, 'children' as `propagate_access` FROM `groups_items`
    WHERE `groups_items`.`item_id` = OLD.`parent_item_id`
    ON DUPLICATE KEY UPDATE propagate_access='children';
END
-- +migrate StatementEnd
