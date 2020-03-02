-- +migrate Up
ALTER TABLE `groups_attempts`
    CHANGE COLUMN `ancestors_computation_state`
        `result_propagation_state` ENUM('done','processing','todo','temp','to_be_propagated','to_be_recomputed'),
    RENAME INDEX `ancestors_computation_state` TO `result_propagation_state`;

UPDATE `groups_attempts` SET `result_propagation_state` = 'done' WHERE `result_propagation_state` = 'temp';
UPDATE `groups_attempts` SET `result_propagation_state` = 'to_be_recomputed' WHERE `result_propagation_state` = 'todo';

ALTER TABLE `groups_attempts`
    MODIFY COLUMN `result_propagation_state` ENUM('done','processing','to_be_propagated','to_be_recomputed') NOT NULL DEFAULT 'done'
        COMMENT 'Used by the algorithm that computes results for items that have children and unlocks items if needed ("to_be_propagated" means that ancestors should be recomputed).';

DROP TRIGGER IF EXISTS `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
    FROM `permissions_generated`
    WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id`;

    UPDATE `groups_attempts` SET `result_propagation_state` = 'to_be_recomputed'
    WHERE `item_id` = NEW.`parent_item_id`;
END
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS `before_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_items` BEFORE UPDATE ON `items_items` FOR EACH ROW BEGIN
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

        INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state)
        VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';

        UPDATE `groups_attempts` SET `result_propagation_state` = 'to_be_recomputed'
            WHERE `item_id` = OLD.`parent_item_id`;

        IF (OLD.`parent_item_id` != NEW.`parent_item_id`) THEN
            UPDATE `groups_attempts` SET `result_propagation_state` = 'to_be_recomputed'
                WHERE `item_id` = NEW.`parent_item_id`;
        END IF;
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS `before_delete_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
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

    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
        SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
        FROM `permissions_generated`
        WHERE `permissions_generated`.`item_id` = OLD.`parent_item_id`;

    UPDATE `groups_attempts` SET `result_propagation_state` = 'to_be_recomputed'
        WHERE `item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd

-- +migrate Down
ALTER TABLE `groups_attempts`
    CHANGE COLUMN `result_propagation_state`
        `ancestors_computation_state` ENUM('done','processing','to_be_propagated','to_be_recomputed','todo','temp'),
    RENAME INDEX `result_propagation_state` TO `ancestors_computation_state`;

UPDATE `groups_attempts` SET `ancestors_computation_state` = 'todo'
WHERE `ancestors_computation_state` IN ('to_be_propagated', 'to_be_recomputed');

ALTER TABLE `groups_attempts`
    MODIFY COLUMN `ancestors_computation_state` enum('done','processing','todo','temp') NOT NULL DEFAULT 'done'
        COMMENT 'Whether the data was propagated to the users'' individual users_items.';

DROP TRIGGER IF EXISTS `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
    FROM `permissions_generated`
    WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id`;
END
-- +migrate StatementEnd

-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_items` BEFORE UPDATE ON `items_items` FOR EACH ROW BEGIN
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

    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
        SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
        FROM `permissions_generated`
        WHERE `permissions_generated`.`item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd
