-- +migrate Up
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN
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
END
-- +migrate StatementEnd

DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.`parent_group_id` != NEW.`parent_group_id` OR OLD.`child_group_id` != NEW.`child_group_id` THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Unable to change immutable groups_groups.parent_group_id and/or groups_groups.child_group_id';
    END IF;
END
-- +migrate StatementEnd

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

-- +migrate Down
DROP TRIGGER `after_insert_groups_groups`;

DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR NOT OLD.expires_at <=> NEW.expires_at) THEN
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

DROP TRIGGER `after_update_groups_groups`;
