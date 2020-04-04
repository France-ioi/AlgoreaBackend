-- +migrate Up
RENAME TABLE `attempts` TO `results`;

ALTER TABLE `results`
    DROP FOREIGN KEY `fk_attempts_creator_id_users_group_id`,
    DROP CHECK `cs_attempts_order`,
    DROP CHECK `cs_attempts_score_computed_is_valid`,
    ADD CONSTRAINT `cs_results_score_computed_is_valid` CHECK ((`score_computed` between 0 and 100)),
    DROP CHECK `cs_attempts_score_edit_value_is_valid`,
    ADD CONSTRAINT `cs_results_score_edit_value_is_valid` CHECK ((ifnull(`score_edit_value`,0) between -(100) and 100)),
    RENAME COLUMN `group_id` TO `participant_id`,
    RENAME INDEX `group_item` TO `participant_id_item_id`,
    RENAME INDEX `group_item_score_desc_score_obtained_at` TO `participant_id_item_id_score_desc_score_obtained_at`,
    ADD COLUMN `attempt_id` BIGINT(20) NOT NULL DEFAULT 0 AFTER `participant_id`;

DROP TRIGGER `before_insert_attempts`;

CREATE TABLE `attempts` (
    `participant_id` BIGINT(20) NOT NULL,
    CONSTRAINT `fk_attempts_participant_id_groups_id` FOREIGN KEY (`participant_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    `id` BIGINT(20) NOT NULL
        COMMENT 'Identifier of this attempt for this participant, 0 is the default attempt for the participant, the next ones are sequentially assigned.',
    PRIMARY KEY (`participant_id`, `id`),
    `creator_id` BIGINT(20) DEFAULT NULL COMMENT 'The user who created this attempt',
    CONSTRAINT `fk_attempts_creator_id_users_group_id` FOREIGN KEY (`creator_id`) REFERENCES `users`(`group_id`) ON DELETE SET NULL,
    `parent_attempt_id` BIGINT(20) DEFAULT NULL COMMENT 'The attempt from which this one was forked. NULL for the default attempt.',
    `root_item_id` BIGINT(20) DEFAULT NULL COMMENT 'The item on which the attempt was created',
    CONSTRAINT `fk_attempts_root_item_id_items_id`FOREIGN KEY (`root_item_id`) REFERENCES `items`(`id`) ON DELETE SET NULL,
    `created_at` DATETIME NOT NULL DEFAULT NOW()
        COMMENT 'Time at which the attempt was manually created or was first marked as started (should be when it is first visited).',
    INDEX `participant_id_parent_attempt_id_root_item_id` (`participant_id`, `parent_attempt_id`, `root_item_id`),
    INDEX `participant_id_root_item_id` (`participant_id`, `root_item_id`)
    -- We cannot add the following constraint because it would set both `participant_id` and `parent_attempt_id` to NULL
    -- on deletion of the parent attempt.
    -- CONSTRAINT `attempts_participant_id_parent_attempt_id_attempts_participant_id_id`
    --    FOREIGN KEY (`participant_id`, `parent_attempt_id`) REFERENCES `attempts`(`participant_id`, `id`) SET NULL,
) ENGINE=InnoDB DEFAULT CHARSET=utf8
    COMMENT 'Attempts of participants (team or user) to solve a subtree of items. An attempt may have several answers for a same item. Every participant has a default attempt.';

INSERT INTO `attempts` (participant_id, id, creator_id, parent_attempt_id, root_item_id, created_at)
SELECT participant_id, 0, (
    SELECT creator_id FROM results AS r
    WHERE r.participant_id = results.participant_id AND creator_id IS NOT NULL
    ORDER BY r.started_at DESC
    LIMIT 1
), NULL, NULL, IFNULL(MIN(started_at), (SELECT created_at FROM `groups` WHERE groups.id = results.participant_id))
FROM `results` WHERE id IN (
    SELECT id FROM (
       SELECT id, ROW_NUMBER() OVER (PARTITION BY `participant_id`, `item_id` ORDER BY `order`) - 1 AS `order`
       FROM `results`
   ) ordered WHERE `order` = 0
)
GROUP BY participant_id;

INSERT INTO `attempts` (participant_id, id, creator_id, parent_attempt_id, root_item_id, created_at)
SELECT participant_id, ROW_NUMBER() OVER (PARTITION BY `participant_id` ORDER BY `item_id`, `order`, started_at) AS `id`,
       creator_id, 0, item_id, IFNULL(started_at, (SELECT created_at FROM `groups` WHERE groups.id = participant_id))
FROM `results` WHERE id IN (
    SELECT id FROM (
       SELECT id, ROW_NUMBER() OVER (PARTITION BY `participant_id`, `item_id` ORDER BY `order`) - 1 AS `order`
       FROM `results`
    ) ordered WHERE `order` > 0
);

INSERT IGNORE INTO `attempts` (participant_id, id, creator_id, created_at)
SELECT `id`, 0, IF(groups.type = 'User', groups.id, NULL), `created_at` FROM `groups` WHERE `type` IN ('User', 'Team');

INSERT INTO `results` (id, participant_id, item_id, attempt_id, `order`)
SELECT `id`, `participant_id`, `item_id`, ROW_NUMBER() OVER (PARTITION BY `participant_id` ORDER BY `item_id`, `order`, started_at), -1
FROM `results`
WHERE id IN (
    SELECT id FROM (
       SELECT id, ROW_NUMBER() OVER (PARTITION BY `participant_id`, `item_id` ORDER BY `order`) - 1 AS `order`
       FROM `results`
    ) ordered WHERE `order` > 0
)
ON DUPLICATE KEY UPDATE attempt_id = VALUES(attempt_id);

ALTER TABLE `answers`
    ADD COLUMN `participant_id` BIGINT(20) AFTER `id`,
    MODIFY COLUMN `attempt_id` BIGINT(20) NOT NULL AFTER `participant_id`,
    ADD COLUMN `item_id` BIGINT(20) AFTER `attempt_id`,
    DROP FOREIGN KEY `fk_answers_attempt_id_attempts_id`;

UPDATE `answers` JOIN `results` ON `results`.`id` = `answers`.`attempt_id`
SET `answers`.`participant_id` = `results`.`participant_id`,
    `answers`.`attempt_id` = `results`.`attempt_id`,
    `answers`.`item_id` = `results`.`item_id`;

ALTER TABLE `results`
    DROP PRIMARY KEY,
    DROP COLUMN `id`,
    DROP KEY `group_id_item_id_order`,
    ADD PRIMARY KEY (`participant_id`, `attempt_id`, `item_id`),
    DROP COLUMN `order`,
    DROP KEY `fk_attempts_creator_id_users_group_id`,
    DROP COLUMN `creator_id`,
    ADD CONSTRAINT `fk_results_participant_id_attempt_id_attempts_participant_id_id`
        FOREIGN KEY (participant_id, attempt_id) REFERENCES attempts(participant_id, id) ON DELETE CASCADE,
    ADD CONSTRAINT `fk_results_item_id_items_id` FOREIGN KEY (item_id) REFERENCES items(id);

ALTER TABLE `answers`
    MODIFY COLUMN `participant_id` BIGINT(20) NOT NULL,
    MODIFY COLUMN `item_id` BIGINT(20) NOT NULL,
    ADD CONSTRAINT `fk_answers_participant_id_attempt_id_item_id_results`
        FOREIGN KEY (`participant_id`, `attempt_id`, `item_id`)
            REFERENCES `results`(`participant_id`, `attempt_id`, `item_id`) ON DELETE CASCADE;

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

-- +migrate Down
ALTER TABLE `results`
    ADD COLUMN `id` BIGINT(20) FIRST,
    ADD COLUMN `creator_id` bigint(20) DEFAULT NULL
        COMMENT 'The user who created this attempt. NULL if created by propagation'
        AFTER `item_id`,
    DROP FOREIGN KEY `fk_results_participant_id_attempt_id_attempts_participant_id_id`;

UPDATE `results`
    JOIN `attempts` ON `attempts`.`participant_id` = `results`.`participant_id` AND
                       `attempts`.`id` = `results`.`attempt_id`
SET `results`.`id` = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000,
    `results`.`creator_id` = `attempts`.`creator_id`;

DROP TABLE `attempts`;

ALTER TABLE `answers` DROP FOREIGN KEY `fk_answers_participant_id_attempt_id_item_id_results`;

# 2,392,792 rows :/
UPDATE `answers` JOIN `results` USING(`participant_id`, `attempt_id`, `item_id`)
SET `answers`.`attempt_id` = `results`.`id`;

ALTER TABLE `results` DROP PRIMARY KEY;

UPDATE `results` SET `attempt_id` = `attempt_id` + 1;

ALTER TABLE `results`
    RENAME TO `attempts`,
    RENAME INDEX `participant_id_item_id` TO `group_item`,
    RENAME INDEX `participant_id_item_id_score_desc_score_obtained_at` TO `group_item_score_desc_score_obtained_at`,
    DROP FOREIGN KEY `fk_results_item_id_items_id`,
    DROP CHECK `cs_results_score_computed_is_valid`,
    DROP CHECK `cs_results_score_edit_value_is_valid`,
    MODIFY COLUMN `id` BIGINT(20) NOT NULL AUTO_INCREMENT,
    RENAME COLUMN `participant_id` TO `group_id`,
    ADD PRIMARY KEY (`id`),
    CHANGE COLUMN `attempt_id` `order` INT(11) NOT NULL AFTER `creator_id`,
    ADD UNIQUE KEY `group_id_item_id_order` (`group_id`,`item_id`,`order`),
    ADD CONSTRAINT `fk_attempts_creator_id_users_group_id`
        FOREIGN KEY (`creator_id`) REFERENCES `users` (`group_id`) ON DELETE SET NULL,
    ADD CONSTRAINT `cs_attempts_order` CHECK ((`order` > 0)),
    ADD CONSTRAINT `cs_attempts_score_computed_is_valid` CHECK (`score_computed` between 0 and 100),
    ADD CONSTRAINT `cs_attempts_score_edit_value_is_valid`
        CHECK (IFNULL(`score_edit_value`, 0) between -100 and 100);

ALTER TABLE `answers`
    DROP COLUMN `participant_id`,
    DROP COLUMN `item_id`,
    MODIFY COLUMN `attempt_id` BIGINT(20) NOT NULL AFTER `author_id`,
    ADD CONSTRAINT `fk_answers_attempt_id_attempts_id` FOREIGN KEY (`attempt_id`)
        REFERENCES `attempts` (`id`) ON DELETE CASCADE;

-- +migrate StatementBegin
CREATE TRIGGER `before_insert_attempts` BEFORE INSERT ON `attempts` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd

DROP TRIGGER `after_insert_groups_groups`;
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

DROP TRIGGER `after_insert_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_permissions_generated` AFTER INSERT ON `permissions_generated` FOR EACH ROW BEGIN
    IF NEW.`can_view_generated` != 'none' THEN
        UPDATE `attempts`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `attempts`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `attempts`.`group_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        SET `result_propagation_state` = 'to_be_propagated';
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `after_update_permissions_generated`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_generated` AFTER UPDATE ON `permissions_generated` FOR EACH ROW BEGIN
    IF OLD.`can_view_generated` = 'none' AND NEW.can_view_generated != 'none' THEN
        UPDATE `attempts`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `attempts`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `attempts`.`group_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        SET `result_propagation_state` = 'to_be_propagated';
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

    UPDATE `attempts` SET `result_propagation_state` = 'to_be_propagated'
    WHERE `item_id` = NEW.`child_item_id`;
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
