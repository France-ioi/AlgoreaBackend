-- +migrate Up
ALTER TABLE `results`
    DROP KEY `result_propagation_state`,
    DROP COLUMN `result_propagation_state`;

DROP TRIGGER `after_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN
    IF NEW.`expires_at` > NOW() THEN
        INSERT IGNORE INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
        WHERE EXISTS(
            SELECT 1 FROM `permissions_generated`
                JOIN `groups_ancestors_active` AS `grand_ancestors`
                    ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                       `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                JOIN `items_ancestors`
                    ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
            WHERE `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                  `permissions_generated`.`can_view_generated` != 'none'
        ) AND NOT EXISTS(
            SELECT 1 FROM `permissions_generated`
                JOIN `groups_ancestors_active` AS `child_ancestors`
                    ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                       `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                JOIN `items_ancestors`
                    ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
            WHERE `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                  `permissions_generated`.`can_view_generated` != 'none'
        );
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_groups_groups` AFTER UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.expires_at != NEW.expires_at THEN
        IF NEW.`expires_at` > NOW() THEN
            INSERT IGNORE INTO `results_propagate`
            SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
            FROM `results`
                JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                                  `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
            WHERE EXISTS(
                SELECT 1 FROM `permissions_generated`
                    JOIN `groups_ancestors_active` AS `grand_ancestors`
                        ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                           `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                    JOIN `items_ancestors`
                        ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                WHERE `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                      `permissions_generated`.`can_view_generated` != 'none'
            ) AND NOT EXISTS(
                SELECT 1 FROM `permissions_generated`
                    JOIN `groups_ancestors_active` AS `child_ancestors`
                        ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                           `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                    JOIN `items_ancestors`
                        ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                WHERE `items_ancestors`.`child_item_id` = `results`.`item_id` AND
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
END
-- +migrate StatementEnd

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

    INSERT INTO `results_propagate`
    SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
    FROM `results`
    WHERE `item_id` = OLD.`parent_item_id`
    ON DUPLICATE KEY UPDATE `state` = 'to_be_recomputed';
END
-- +migrate StatementEnd

DROP TRIGGER `after_insert_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_permissions_generated` AFTER INSERT ON `permissions_generated` FOR EACH ROW BEGIN
    IF NEW.`can_view_generated` != 'none' THEN
        INSERT IGNORE INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`;
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_generated` AFTER UPDATE ON `permissions_generated` FOR EACH ROW BEGIN
    IF OLD.`can_view_generated` = 'none' AND NEW.can_view_generated != 'none' THEN
        INSERT IGNORE INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`;
    END IF;
END
-- +migrate StatementEnd

-- +migrate Down
ALTER TABLE `results`
    ADD COLUMN `result_propagation_state` enum('done','to_be_propagated','to_be_recomputed') NOT NULL DEFAULT 'done'
        COMMENT 'Used by the algorithm that computes results for items that have children and unlocks items if needed ("to_be_propagated" means that ancestors should be recomputed).'
        AFTER `latest_hint_at`,
    ADD KEY `result_propagation_state` (`result_propagation_state`);

DROP TRIGGER `after_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN
    IF NEW.`expires_at` > NOW() THEN
        UPDATE `results`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
        SET `result_propagation_state` = 'to_be_propagated'
        WHERE EXISTS(
            SELECT 1 FROM `permissions_generated`
                JOIN `groups_ancestors_active` AS `grand_ancestors`
                    ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                       `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                JOIN `items_ancestors`
                    ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
            WHERE `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                  `permissions_generated`.`can_view_generated` != 'none'
        ) AND NOT EXISTS(
            SELECT 1 FROM `permissions_generated`
                JOIN `groups_ancestors_active` AS `child_ancestors`
                    ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                       `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                JOIN `items_ancestors`
                    ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
            WHERE `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                  `permissions_generated`.`can_view_generated` != 'none'
        );
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_groups_groups` AFTER UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.expires_at != NEW.expires_at THEN
        IF NEW.`expires_at` > NOW() THEN
            UPDATE `results`
                JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                                  `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
            SET `result_propagation_state` = 'to_be_propagated'
            WHERE EXISTS(
                SELECT 1 FROM `permissions_generated`
                    JOIN `groups_ancestors_active` AS `grand_ancestors`
                        ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                           `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                    JOIN `items_ancestors`
                        ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                WHERE `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                      `permissions_generated`.`can_view_generated` != 'none'
            ) AND NOT EXISTS(
                SELECT 1 FROM `permissions_generated`
                    JOIN `groups_ancestors_active` AS `child_ancestors`
                        ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                           `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                    JOIN `items_ancestors`
                        ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                WHERE `items_ancestors`.`child_item_id` = `results`.`item_id` AND
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

DROP TRIGGER `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
    FROM `permissions_generated`
    WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id`;

    UPDATE `results` SET `result_propagation_state` = 'to_be_propagated'
    WHERE `item_id` = NEW.`child_item_id`;
END
-- +migrate StatementEnd

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
        UPDATE `results` SET `result_propagation_state` = 'to_be_recomputed' WHERE `item_id` = NEW.`parent_item_id`;
    END IF;
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

    -- Some results' ancestors should probably be removed
    -- DELETE FROM `results` WHERE ...

    UPDATE `results` SET `result_propagation_state` = 'to_be_recomputed'
    WHERE `item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd

DROP TRIGGER `after_insert_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_permissions_generated` AFTER INSERT ON `permissions_generated` FOR EACH ROW BEGIN
    IF NEW.`can_view_generated` != 'none' THEN
        UPDATE `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        SET `result_propagation_state` = 'to_be_propagated';
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_generated` AFTER UPDATE ON `permissions_generated` FOR EACH ROW BEGIN
    IF OLD.`can_view_generated` = 'none' AND NEW.can_view_generated != 'none' THEN
        UPDATE `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        SET `result_propagation_state` = 'to_be_propagated';
    END IF;
END
-- +migrate StatementEnd
