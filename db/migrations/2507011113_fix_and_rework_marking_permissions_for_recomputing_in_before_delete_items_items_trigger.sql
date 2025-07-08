-- +migrate Up
DROP TRIGGER `before_delete_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
  INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
  VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';

  REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
  SELECT `permissions_generated`.`group_id`, OLD.`child_item_id`, 'self' as `propagate_to`
  FROM `permissions_generated`
  WHERE `permissions_generated`.`item_id` = OLD.`parent_item_id`;

  -- Some results' ancestors should probably be removed
  -- DELETE FROM `results` WHERE ...

  INSERT IGNORE INTO `results_recompute_for_items` (`item_id`) VALUES (OLD.`parent_item_id`);
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `before_delete_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
  INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`)
  VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';

  INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
  SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
  FROM `permissions_generated`
  WHERE `permissions_generated`.`item_id` = OLD.`parent_item_id`;

  -- Some results' ancestors should probably be removed
  -- DELETE FROM `results` WHERE ...

  INSERT IGNORE INTO `results_recompute_for_items` (`item_id`) VALUES (OLD.`parent_item_id`);
END
-- +migrate StatementEnd
