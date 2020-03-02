-- +migrate Up
DROP TRIGGER `after_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_groups_groups` AFTER UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.expires_at != NEW.expires_at THEN
        IF NEW.`expires_at` > NOW() THEN
            UPDATE `attempts`
                JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `attempts`.`group_id` AND
                                                  `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
            SET `result_propagation_state` = 'to_be_propagated'
            WHERE EXISTS(
                SELECT 1 FROM `permissions_generated`
                    JOIN `groups_ancestors_active` AS `grand_ancestors`
                        ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                           `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                    JOIN `items_ancestors`
                        ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                WHERE `items_ancestors`.`child_item_id` = `attempts`.`item_id` AND
                      `permissions_generated`.`can_view_generated` != 'none'
            ) AND NOT EXISTS(
                SELECT 1 FROM `permissions_generated`
                    JOIN `groups_ancestors_active` AS `child_ancestors`
                        ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                           `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                    JOIN `items_ancestors`
                        ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                WHERE `items_ancestors`.`child_item_id` = `attempts`.`item_id` AND
                      `permissions_generated`.`can_view_generated` != 'none'
            );
        END IF;

        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';

        INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
END
-- +migrate StatementEnd

DROP TRIGGER `before_delete_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
    VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';

    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
    FROM `permissions_generated`
    WHERE `permissions_generated`.`item_id` = OLD.`parent_item_id`;

    -- Some attempts' ancestors should probably be removed
    -- DELETE FROM `attempts` WHERE ...

    UPDATE `attempts` SET `result_propagation_state` = 'to_be_recomputed'
    WHERE `item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `after_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_groups_groups` AFTER UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.expires_at != NEW.expires_at THEN
        IF NEW.`expires_at` > NOW() THEN
            UPDATE `attempts`
                JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `attempts`.`group_id` AND
                                                  `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
            SET `result_propagation_state` = 'to_be_propagated'
            WHERE EXISTS(
                SELECT 1 FROM `permissions_generated`
                    JOIN `groups_ancestors_active` AS `grand_ancestors`
                        ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                           `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                    JOIN `items_ancestors`
                        ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                WHERE `items_ancestors`.`child_item_id` = `attempts`.`item_id` AND
                      `permissions_generated`.`can_view_generated` != 'none'
            ) AND NOT EXISTS(
                SELECT 1 FROM `permissions_generated`
                    JOIN `groups_ancestors_active` AS `child_ancestors`
                        ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                           `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                    JOIN `items_ancestors`
                        ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                WHERE `items_ancestors`.`child_item_id` = `attempts`.`item_id` AND
                      `permissions_generated`.`can_view_generated` != 'none'
            );
        END IF;
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
            SELECT `groups_ancestors`.`child_group_id`, 'todo'
            FROM `groups_ancestors`
            WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
        ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        DELETE `bridges` FROM `groups_ancestors` `child_descendants`
                                  JOIN `groups_ancestors` `parent_ancestors`
                                  JOIN `groups_ancestors` `bridges`
                                       ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
                                           `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
        WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
                `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;

        INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
        SELECT `groups_ancestors`.`child_group_id`, 'todo'
        FROM `groups_ancestors`
        WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
    ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    DELETE `bridges`
    FROM `groups_ancestors` `child_descendants`
        JOIN `groups_ancestors` `parent_ancestors`
        JOIN `groups_ancestors` `bridges`
            ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
                `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
    WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
        `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;
END
-- +migrate StatementEnd

DROP TRIGGER `before_delete_items_items`;
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

    -- Some attempts' ancestors should probably be removed
    -- DELETE FROM `attempts` WHERE ...

    UPDATE `attempts` SET `result_propagation_state` = 'to_be_recomputed'
    WHERE `item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd
