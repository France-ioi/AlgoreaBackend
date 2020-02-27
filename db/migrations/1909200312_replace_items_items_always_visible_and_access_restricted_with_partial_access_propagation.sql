-- +migrate Up
ALTER TABLE `items_items` ADD COLUMN `partial_access_propagation` enum('None','AsGrayed','AsPartial') NOT NULL DEFAULT 'None'
    COMMENT 'Specifies how to propagate partial access to the child item' AFTER `access_restricted`;
ALTER TABLE `history_items_items` ADD COLUMN `partial_access_propagation`
    enum('None','AsGrayed','AsPartial') NOT NULL DEFAULT 'None' AFTER `access_restricted`;

UPDATE `items_items` SET `partial_access_propagation` = IF(`access_restricted`, IF(`always_visible`, 'AsGrayed', 'None'), 'AsPartial');
UPDATE `history_items_items` SET `partial_access_propagation` = IF(`access_restricted`, IF(`always_visible`, 'AsGrayed', 'None'), 'AsPartial');

ALTER TABLE `items_items` DROP COLUMN `access_restricted`, DROP COLUMN `always_visible`;
ALTER TABLE `history_items_items` DROP COLUMN `access_restricted`, DROP COLUMN `always_visible`;

DROP TRIGGER IF EXISTS `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
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
DROP TRIGGER IF EXISTS `before_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_items` BEFORE UPDATE ON `items_items` FOR EACH ROW BEGIN
    IF NEW.version <> OLD.version THEN
        SET @curVersion = NEW.version;
    ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    END IF;
    IF NOT (OLD.`id` = NEW.`id` AND OLD.`parent_item_id` <=> NEW.`parent_item_id` AND
            OLD.`child_item_id` <=> NEW.`child_item_id` AND OLD.`child_order` <=> NEW.`child_order` AND
            OLD.`category` <=> NEW.`category` AND OLD.`partial_access_propagation` <=> NEW.`partial_access_propagation` AND
            OLD.`difficulty` <=> NEW.`difficulty`) THEN
        SET NEW.version = @curVersion;
        UPDATE `history_items_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
        INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                           `partial_access_propagation`,`difficulty`)
            VALUES (NEW.`id`,@curVersion,NEW.`parent_item_id`,NEW.`child_item_id`,NEW.`child_order`,NEW.`category`,
                    NEW.`partial_access_propagation`,NEW.`difficulty`);
    END IF;
    IF (OLD.child_item_id != NEW.child_item_id OR OLD.parent_item_id != NEW.parent_item_id) THEN
        INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
            VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
            VALUES (OLD.parent_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
            (
                SELECT `items_ancestors`.`child_item_id`, 'todo'
                FROM `items_ancestors`
                    WHERE `items_ancestors`.`ancestor_item_id` = OLD.`child_item_id`
            ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        DELETE `items_ancestors` from `items_ancestors`
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
    END IF;
    IF (OLD.child_item_id != NEW.child_item_id OR OLD.parent_item_id != NEW.parent_item_id) THEN
        INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state)
            VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
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

-- +migrate Down
ALTER TABLE `items_items`
    ADD COLUMN `always_visible` tinyint(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether the title of this child should always be visible within the parent (gray), even if the user did not unlock the access.'
        AFTER `category`,
    ADD COLUMN `access_restricted` tinyint(1) NOT NULL DEFAULT '1'
        COMMENT 'Whether this item is locked by default within the parent. If false, anyone who has access to the parent will also have access to the child.'
        AFTER `always_visible`;
ALTER TABLE `history_items_items`
    ADD COLUMN `always_visible` tinyint(1) NOT NULL DEFAULT '0' AFTER `category`,
    ADD COLUMN `access_restricted` tinyint(1) NOT NULL DEFAULT '1' AFTER `always_visible`;
UPDATE `items_items` SET `always_visible` = (`partial_access_propagation` = 'AsGrayed'),
                         `access_restricted` = (`partial_access_propagation` != 'AsPartial');
UPDATE `history_items_items` SET `always_visible` = (`partial_access_propagation` = 'AsGrayed'),
                         `access_restricted` = (`partial_access_propagation` != 'AsPartial');

DROP TRIGGER IF EXISTS `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                       `access_restricted`,`always_visible`,`difficulty`)
    VALUES (NEW.`id`,@curVersion,NEW.`parent_item_id`,NEW.`child_item_id`,NEW.`child_order`,NEW.`category`,
            NEW.`access_restricted`,NEW.`always_visible`,NEW.`difficulty`);
    INSERT IGNORE INTO `groups_items_propagate`
    SELECT `id`, 'children' as `propagate_access` FROM `groups_items`
    WHERE `groups_items`.`item_id` = NEW.`parent_item_id`
    ON DUPLICATE KEY UPDATE propagate_access='children';
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_items` BEFORE UPDATE ON `items_items` FOR EACH ROW BEGIN
    IF NEW.version <> OLD.version THEN
        SET @curVersion = NEW.version;
    ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    END IF;
    IF NOT (OLD.`id` = NEW.`id` AND OLD.`parent_item_id` <=> NEW.`parent_item_id` AND
            OLD.`child_item_id` <=> NEW.`child_item_id` AND OLD.`child_order` <=> NEW.`child_order` AND
            OLD.`category` <=> NEW.`category` AND OLD.`access_restricted` <=> NEW.`access_restricted` AND
            OLD.`always_visible` <=> NEW.`always_visible` AND OLD.`difficulty` <=> NEW.`difficulty`) THEN
        SET NEW.version = @curVersion;
        UPDATE `history_items_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
        INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                           `access_restricted`,`always_visible`,`difficulty`)
        VALUES (NEW.`id`,@curVersion,NEW.`parent_item_id`,NEW.`child_item_id`,NEW.`child_order`,NEW.`category`,
                NEW.`access_restricted`,NEW.`always_visible`,NEW.`difficulty`);
    END IF;
    IF (OLD.child_item_id != NEW.child_item_id OR OLD.parent_item_id != NEW.parent_item_id) THEN
        INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
        VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
        VALUES (OLD.parent_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
            (
                SELECT `items_ancestors`.`child_item_id`, 'todo'
                FROM `items_ancestors`
                WHERE `items_ancestors`.`ancestor_item_id` = OLD.`child_item_id`
            ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        DELETE `items_ancestors` from `items_ancestors`
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
    END IF;
    IF (OLD.child_item_id != NEW.child_item_id OR OLD.parent_item_id != NEW.parent_item_id) THEN
        INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state)
        VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    UPDATE `history_items_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                       `access_restricted`,`always_visible`,`difficulty`, `deleted`)
    VALUES (OLD.`id`,@curVersion,OLD.`parent_item_id`,OLD.`child_item_id`,OLD.`child_order`,OLD.`category`,
            OLD.`access_restricted`,OLD.`always_visible`,OLD.`difficulty`, 1);
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
ALTER TABLE `items_items` DROP COLUMN `partial_access_propagation`;
ALTER TABLE `history_items_items` DROP COLUMN `partial_access_propagation`;
