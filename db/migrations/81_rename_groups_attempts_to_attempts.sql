-- +migrate Up
DROP TRIGGER `before_insert_groups_attempts`;
DROP TRIGGER `after_insert_items_items`;
DROP TRIGGER `before_update_items_items`;
DROP TRIGGER `before_delete_items_items`;

ALTER TABLE `groups_attempts`
    RENAME TO `attempts`,
    DROP FOREIGN KEY `fk_groups_attempts_creator_id_users_group_id`,
    ADD CONSTRAINT `fk_attempts_creator_id_users_group_id` FOREIGN KEY (`creator_id`)
        REFERENCES `users` (`group_id`) ON DELETE SET NULL,
    DROP CHECK `cs_groups_attempts_score_computed_is_valid`,
    ADD CONSTRAINT `cs_attempts_score_computed_is_valid` CHECK (`score_computed` BETWEEN 0 AND 100),
    DROP CHECK `cs_groups_attempts_score_edit_value_is_valid`,
    ADD CONSTRAINT `cs_attempts_score_edit_value_is_valid` CHECK ((ifnull(`score_edit_value`,0) between -(100) and 100));

-- +migrate StatementBegin
CREATE TRIGGER `before_insert_attempts` BEFORE INSERT ON `attempts` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
    FROM `permissions_generated`
    WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id`;

    UPDATE `attempts` SET `result_propagation_state` = 'changed'
    WHERE `item_id` = NEW.`child_item_id`;
END
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_items` BEFORE UPDATE ON `items_items` FOR EACH ROW BEGIN
    IF (OLD.child_item_id != NEW.child_item_id OR OLD.parent_item_id != NEW.parent_item_id) THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Unable to change immutable items_items.parent_item_id and/or items_items.child_item_id';
    END IF;
END
-- +migrate StatementEnd

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

    DELETE FROM `attempts` WHERE `item_id` = OLD.`parent_item_id` AND `started_at` IS NULL AND `score_edit_rule` IS NULL;

    UPDATE `attempts` SET `result_propagation_state` = 'to_be_recomputed'
    WHERE `item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd

ALTER TABLE `answers`
    DROP FOREIGN KEY `fk_answers_attempt_id_groups_attempts_id`,
    ADD CONSTRAINT `fk_answers_attempt_id_attempts_id`
        FOREIGN KEY (`attempt_id`) REFERENCES `attempts` (`id`) ON DELETE CASCADE;

ALTER TABLE `users_items`
    DROP FOREIGN KEY `fk_users_items_active_attempt_id_groups_attempts_id`,
    ADD CONSTRAINT `fk_users_items_active_attempt_id_attempts_id`
        FOREIGN KEY (`active_attempt_id`) REFERENCES `attempts` (`id`) ON DELETE CASCADE;

-- +migrate Down
DROP TRIGGER `before_insert_attempts`;
DROP TRIGGER `after_insert_items_items`;
DROP TRIGGER `before_update_items_items`;
DROP TRIGGER `before_delete_items_items`;

ALTER TABLE `attempts`
    RENAME TO `groups_attempts`,
    DROP FOREIGN KEY `fk_attempts_creator_id_users_group_id`,
    ADD CONSTRAINT `fk_groups_attempts_creator_id_users_group_id` FOREIGN KEY (`creator_id`)
        REFERENCES `users` (`group_id`) ON DELETE SET NULL,
    DROP CHECK `cs_attempts_score_computed_is_valid`,
    ADD CONSTRAINT `cs_groups_attempts_score_computed_is_valid` CHECK (`score_computed` BETWEEN 0 AND 100),
    DROP CHECK `cs_attempts_score_edit_value_is_valid`,
    ADD CONSTRAINT `cs_groups_attempts_score_edit_value_is_valid` CHECK ((ifnull(`score_edit_value`,0) between -(100) and 100));

-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_attempts` BEFORE INSERT ON `groups_attempts` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd

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

ALTER TABLE `answers`
    DROP FOREIGN KEY `fk_answers_attempt_id_attempts_id`,
    ADD CONSTRAINT `fk_answers_attempt_id_groups_attempts_id`
        FOREIGN KEY (`attempt_id`) REFERENCES `groups_attempts` (`id`) ON DELETE CASCADE;

ALTER TABLE `users_items`
    DROP FOREIGN KEY `fk_users_items_active_attempt_id_attempts_id`,
    ADD CONSTRAINT `fk_users_items_active_attempt_id_groups_attempts_id`
        FOREIGN KEY (`active_attempt_id`) REFERENCES `groups_attempts` (`id`) ON DELETE CASCADE;
