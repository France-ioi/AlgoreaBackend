-- +migrate Up
DROP TRIGGER IF EXISTS `before_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN
  SET NEW.is_team_membership = (SELECT type = 'Team' FROM `groups` WHERE id = NEW.parent_group_id FOR SHARE);
  IF NOT NEW.is_team_membership THEN
    INSERT INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_groups_groups` AFTER UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.expires_at != NEW.expires_at AND NOT NEW.`is_team_membership` THEN
        IF NEW.`expires_at` > NOW() THEN
            INSERT INTO `results_propagate`
            SELECT `participant_id`, `attempt_id`, `results`.`item_id`, 'to_be_propagated' AS `state`
            FROM (
                     SELECT `item_id`
                     FROM (
                              SELECT DISTINCT `item_id`
                              FROM `results`
                                       JOIN `groups_ancestors_active`
                                            ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                               `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
                              FOR SHARE
                          ) AS `result_items`
                     WHERE EXISTS(
                             SELECT 1
                             FROM `permissions_generated`
                                      JOIN `groups_ancestors_active` AS `grand_ancestors`
                                           ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                                              `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                                      JOIN `items_ancestors`
                                           ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                             WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
                               AND `permissions_generated`.`can_view_generated` != 'none'
                             FOR SHARE
                         )
                       AND NOT EXISTS(
                             SELECT 1
                             FROM `permissions_generated`
                                      JOIN `groups_ancestors_active` AS `child_ancestors`
                                           ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                                              `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                                      JOIN `items_ancestors`
                                           ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                             WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
                               AND `permissions_generated`.`can_view_generated` != 'none'
                             FOR SHARE
                         )
                     FOR SHARE
                 ) AS `result_items_filtered`
            JOIN `results` ON `results`.`item_id` = `result_items_filtered`.`item_id`
            JOIN `groups_ancestors_active`
              ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                 `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
            FOR SHARE
            ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
        END IF;

        INSERT INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
  IF NOT OLD.`is_team_membership` THEN
    INSERT INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `before_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_items` BEFORE INSERT ON `items_items` FOR EACH ROW BEGIN INSERT INTO `items_propagate` (id, ancestors_computation_state) VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd

DROP TRIGGER `before_delete_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
  INSERT INTO `items_propagate` (`id`, `ancestors_computation_state`)
  VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';

  REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
  SELECT `permissions_generated`.`group_id`, OLD.`child_item_id`, 'self' as `propagate_to`
  FROM `permissions_generated`
  WHERE `permissions_generated`.`item_id` = OLD.`parent_item_id`;

  -- Some results' ancestors should probably be removed
  -- DELETE FROM `results` WHERE ...

  INSERT INTO `results_recompute_for_items` (`item_id`) VALUES (OLD.`parent_item_id`) ON DUPLICATE KEY UPDATE `item_id`=`item_id`;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_generated` AFTER UPDATE ON `permissions_generated` FOR EACH ROW BEGIN
    IF OLD.`can_view_generated` = 'none' AND NEW.can_view_generated != 'none' THEN
      IF @synchronous_propagations_connection_id > 0 THEN
        INSERT INTO `results_propagate_sync`
        SELECT @synchronous_propagations_connection_id AS `connection_id`, `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
      ELSE
        INSERT INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
      END IF;
    END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER IF EXISTS `before_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN
  SET NEW.is_team_membership = (SELECT type = 'Team' FROM `groups` WHERE id = NEW.parent_group_id FOR SHARE);
  IF NOT NEW.is_team_membership THEN
    INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_groups_groups` AFTER UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.expires_at != NEW.expires_at AND NOT NEW.`is_team_membership` THEN
        IF NEW.`expires_at` > NOW() THEN
            INSERT INTO `results_propagate`
            SELECT `participant_id`, `attempt_id`, `results`.`item_id`, 'to_be_propagated' AS `state`
            FROM (
                     SELECT `item_id`
                     FROM (
                              SELECT DISTINCT `item_id`
                              FROM `results`
                                       JOIN `groups_ancestors_active`
                                            ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                               `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
                              FOR SHARE
                          ) AS `result_items`
                     WHERE EXISTS(
                             SELECT 1
                             FROM `permissions_generated`
                                      JOIN `groups_ancestors_active` AS `grand_ancestors`
                                           ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                                              `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                                      JOIN `items_ancestors`
                                           ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                             WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
                               AND `permissions_generated`.`can_view_generated` != 'none'
                             FOR SHARE
                         )
                       AND NOT EXISTS(
                             SELECT 1
                             FROM `permissions_generated`
                                      JOIN `groups_ancestors_active` AS `child_ancestors`
                                           ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                                              `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                                      JOIN `items_ancestors`
                                           ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                             WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
                               AND `permissions_generated`.`can_view_generated` != 'none'
                             FOR SHARE
                         )
                     FOR SHARE
                 ) AS `result_items_filtered`
            JOIN `results` ON `results`.`item_id` = `result_items_filtered`.`item_id`
            JOIN `groups_ancestors_active`
              ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                 `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
            FOR SHARE
            ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
        END IF;

        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
  IF NOT OLD.`is_team_membership` THEN
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `before_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_items` BEFORE INSERT ON `items_items` FOR EACH ROW BEGIN INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state) VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd

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

DROP TRIGGER `after_update_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_generated` AFTER UPDATE ON `permissions_generated` FOR EACH ROW BEGIN
    IF OLD.`can_view_generated` = 'none' AND NEW.can_view_generated != 'none' THEN
      IF @synchronous_propagations_connection_id > 0 THEN
        INSERT INTO `results_propagate_sync`
        SELECT @synchronous_propagations_connection_id AS `connection_id`, `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
      ELSE
        INSERT IGNORE INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
      END IF;
    END IF;
END
-- +migrate StatementEnd
