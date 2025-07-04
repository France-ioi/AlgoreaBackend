-- +migrate Up
DROP TRIGGER IF EXISTS `after_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_items_items` AFTER UPDATE ON `items_items` FOR EACH ROW BEGIN
    IF (OLD.`content_view_propagation` != NEW.`content_view_propagation` OR
        OLD.`upper_view_levels_propagation` != NEW.`upper_view_levels_propagation` OR
        OLD.`grant_view_propagation` != NEW.`grant_view_propagation` OR
        OLD.`watch_propagation` != NEW.`watch_propagation` OR
        OLD.`edit_propagation` != NEW.`edit_propagation`) THEN
        REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
        SELECT `permissions_generated`.`group_id`, NEW.`child_item_id`, 'self' as `propagate_to`
        FROM `permissions_generated`
        WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id`;
    END IF;
    IF (OLD.`category` != NEW.`category` OR OLD.`score_weight` != NEW.`score_weight`) THEN
        INSERT INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_recomputed' AS `state`
        FROM `results`
        WHERE `item_id` = NEW.`parent_item_id`
        ON DUPLICATE KEY UPDATE `state` = 'to_be_recomputed';
    END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER IF EXISTS `after_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_items_items` AFTER UPDATE ON `items_items` FOR EACH ROW BEGIN
    IF (OLD.`content_view_propagation` != NEW.`content_view_propagation` OR
        OLD.`upper_view_levels_propagation` != NEW.`upper_view_levels_propagation` OR
        OLD.`grant_view_propagation` != NEW.`grant_view_propagation` OR
        OLD.`watch_propagation` != NEW.`watch_propagation` OR
        OLD.`edit_propagation` != NEW.`edit_propagation`) THEN
        INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
        SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
        FROM `permissions_generated`
        WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id` OR `permissions_generated`.`item_id` = OLD.`parent_item_id`;
    END IF;
    IF (OLD.`category` != NEW.`category` OR OLD.`score_weight` != NEW.`score_weight`) THEN
        INSERT INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_recomputed' AS `state`
        FROM `results`
        WHERE `item_id` = NEW.`parent_item_id`
        ON DUPLICATE KEY UPDATE `state` = 'to_be_recomputed';
    END IF;
END
-- +migrate StatementEnd
