-- +migrate Up
DROP TRIGGER `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    SELECT `permissions_generated`.`group_id`, NEW.`child_item_id`, 'self' as `propagate_to`
    FROM `permissions_generated`
    WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id`;

    INSERT IGNORE INTO `results_propagate`
    SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
    FROM `results`
    WHERE `item_id` = NEW.`child_item_id`;

    INSERT INTO `results_propagate`
    SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_recomputed' AS `state`
    FROM `results`
    WHERE `item_id` = NEW.`parent_item_id`
    ON DUPLICATE KEY UPDATE `state` = 'to_be_recomputed';
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
    FROM `permissions_generated`
    WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id`;

    INSERT IGNORE INTO `results_propagate`
    SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
    FROM `results`
    WHERE `item_id` = NEW.`child_item_id`;

    INSERT INTO `results_propagate`
    SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_recomputed' AS `state`
    FROM `results`
    WHERE `item_id` = NEW.`parent_item_id`
    ON DUPLICATE KEY UPDATE `state` = 'to_be_recomputed';
END
-- +migrate StatementEnd
