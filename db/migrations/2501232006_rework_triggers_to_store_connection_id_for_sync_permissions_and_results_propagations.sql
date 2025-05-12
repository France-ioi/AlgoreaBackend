-- +migrate Up
DROP TRIGGER `after_insert_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_permissions_generated` AFTER INSERT ON `permissions_generated` FOR EACH ROW BEGIN
    IF NEW.`can_view_generated` != 'none' THEN
      IF @synchronous_propagations_connection_id > 0 THEN
        INSERT IGNORE INTO `results_propagate_sync`
        SELECT @synchronous_propagations_connection_id AS `connection_id`, `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`;
      ELSE
        INSERT IGNORE INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`;
      END IF;
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_generated` AFTER UPDATE ON `permissions_generated` FOR EACH ROW BEGIN
    IF OLD.`can_view_generated` = 'none' AND NEW.can_view_generated != 'none' THEN
      IF @synchronous_propagations_connection_id > 0 THEN
        INSERT IGNORE INTO `results_propagate_sync`
        SELECT @synchronous_propagations_connection_id AS `connection_id`, `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`;
      ELSE
        INSERT IGNORE INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`;
      END IF;
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_insert_permissions_granted`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_permissions_granted` AFTER INSERT ON `permissions_granted` FOR EACH ROW BEGIN
  IF @synchronous_propagations_connection_id > 0 THEN
    REPLACE INTO `permissions_propagate_sync` (`connection_id`, `group_id`, `item_id`, `propagate_to`)
      VALUE (@synchronous_propagations_connection_id, NEW.`group_id`, NEW.`item_id`, 'self');
  ELSE
    REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
      VALUE (NEW.`group_id`, NEW.`item_id`, 'self');
  END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_permissions_granted`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_granted` AFTER UPDATE ON `permissions_granted` FOR EACH ROW BEGIN
    IF NOT (NEW.`can_view` <=> OLD.`can_view` AND NEW.`can_grant_view` <=> OLD.`can_grant_view` AND
            NEW.`can_watch` <=> OLD.`can_watch` AND NEW.`can_edit` <=> OLD.`can_edit` AND
            NEW.`is_owner` <=> OLD.`is_owner`) THEN
      IF @synchronous_propagations_connection_id > 0 THEN
        REPLACE INTO `permissions_propagate_sync` (`connection_id`, `group_id`, `item_id`, `propagate_to`)
          VALUE (@synchronous_propagations_connection_id, NEW.`group_id`, NEW.`item_id`, 'self');
      ELSE
        REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
          VALUE (NEW.`group_id`, NEW.`item_id`, 'self');
      END IF;
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_delete_permissions_granted`;
-- +migrate StatementBegin
CREATE TRIGGER `after_delete_permissions_granted` AFTER DELETE ON `permissions_granted` FOR EACH ROW BEGIN
  IF @synchronous_propagations_connection_id > 0 THEN
    REPLACE INTO `permissions_propagate_sync` (`connection_id`, `group_id`, `item_id`, `propagate_to`)
      VALUE (@synchronous_propagations_connection_id, OLD.`group_id`, OLD.`item_id`, 'self');
  ELSE
    REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
      VALUE (OLD.`group_id`, OLD.`item_id`, 'self');
  END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `after_insert_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_permissions_generated` AFTER INSERT ON `permissions_generated` FOR EACH ROW BEGIN
  IF NEW.`can_view_generated` != 'none' THEN
    IF @synchronous_propagations THEN
      INSERT IGNORE INTO `results_propagate_sync`
      SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
      FROM `results`
             JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                       `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
             JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                               `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`;
    ELSE
      INSERT IGNORE INTO `results_propagate`
      SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
      FROM `results`
             JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                       `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
             JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                               `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`;
    END IF;
  END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_generated` AFTER UPDATE ON `permissions_generated` FOR EACH ROW BEGIN
  IF OLD.`can_view_generated` = 'none' AND NEW.can_view_generated != 'none' THEN
    IF @synchronous_propagations THEN
      INSERT IGNORE INTO `results_propagate_sync`
      SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
      FROM `results`
             JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                       `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
             JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                               `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`;
    ELSE
      INSERT IGNORE INTO `results_propagate`
      SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
      FROM `results`
             JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                       `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
             JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                               `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`;
    END IF;
  END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_insert_permissions_granted`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_permissions_granted` AFTER INSERT ON `permissions_granted` FOR EACH ROW BEGIN
  IF @synchronous_propagations THEN
    INSERT INTO `permissions_propagate_sync` (`group_id`, `item_id`, `propagate_to`)
      VALUE (NEW.`group_id`, NEW.`item_id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_to` = 'self';
  ELSE
    INSERT INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
      VALUE (NEW.`group_id`, NEW.`item_id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_to` = 'self';
  END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_permissions_granted`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_granted` AFTER UPDATE ON `permissions_granted` FOR EACH ROW BEGIN
  IF NOT (NEW.`can_view` <=> OLD.`can_view` AND NEW.`can_grant_view` <=> OLD.`can_grant_view` AND
          NEW.`can_watch` <=> OLD.`can_watch` AND NEW.`can_edit` <=> OLD.`can_edit` AND
          NEW.`is_owner` <=> OLD.`is_owner`) THEN
    IF @synchronous_propagations THEN
      REPLACE INTO `permissions_propagate_sync` (`group_id`, `item_id`, `propagate_to`) VALUE (NEW.`group_id`, NEW.`item_id`, 'self');
    ELSE
      INSERT INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`) VALUE (NEW.`group_id`, NEW.`item_id`, 'self')
      ON DUPLICATE KEY UPDATE `propagate_to` = 'self';
    END IF;
  END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_delete_permissions_granted`;
-- +migrate StatementBegin
CREATE TRIGGER `after_delete_permissions_granted` AFTER DELETE ON `permissions_granted` FOR EACH ROW BEGIN
  IF @synchronous_propagations THEN
    REPLACE INTO `permissions_propagate_sync` (`group_id`, `item_id`, `propagate_to`) VALUE (OLD.`group_id`, OLD.`item_id`, 'self');
  ELSE
    INSERT INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`) VALUE (OLD.`group_id`, OLD.`item_id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_to` = 'self';
  END IF;
END
-- +migrate StatementEnd
