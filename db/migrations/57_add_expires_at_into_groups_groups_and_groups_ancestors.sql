-- +migrate Up
ALTER TABLE `groups_groups` ADD COLUMN `expires_at` datetime  NOT NULL DEFAULT '9999-12-31 23:59:59'
    COMMENT 'The group membership expires at the specified time';
ALTER TABLE `groups_ancestors` ADD COLUMN `expires_at` datetime NOT NULL DEFAULT '9999-12-31 23:59:59'
    COMMENT 'The group relation expires at the specified time';

DROP VIEW IF EXISTS groups_ancestors_active;
CREATE VIEW groups_ancestors_active AS SELECT * FROM groups_ancestors WHERE NOW() < expires_at;

DROP VIEW IF EXISTS groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR OLD.type != NEW.type OR NOT OLD.expires_at <=> NEW.expires_at) THEN
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
CREATE DEFINER=`algorea`@`%` TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
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

-- +migrate Down
DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR OLD.type != NEW.type) THEN
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
            SELECT `groups_ancestors`.`child_group_id`, 'todo'
            FROM `groups_ancestors`
            WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
        ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        DELETE `groups_ancestors` FROM `groups_ancestors`
        WHERE `groups_ancestors`.`child_group_id` = OLD.`child_group_id` AND
                `groups_ancestors`.`ancestor_group_id` = OLD.`parent_group_id`;
        DELETE `bridges` FROM `groups_ancestors` `child_descendants`
                                  JOIN `groups_ancestors` `parent_ancestors`
                                  JOIN `groups_ancestors` `bridges`
                                       ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
                                           `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
        WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
                `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;
        DELETE `child_ancestors` FROM `groups_ancestors` `child_ancestors`
                                          JOIN `groups_ancestors` `parent_ancestors`
                                               ON (`child_ancestors`.`child_group_id` = OLD.`child_group_id` AND
                                                   `child_ancestors`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id`)
        WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id`;
        DELETE `parent_ancestors` FROM `groups_ancestors` `parent_ancestors`
                                           JOIN  `groups_ancestors` `child_ancestors`
                                                 ON (`parent_ancestors`.`ancestor_group_id` = OLD.`parent_group_id` AND
                                                     `child_ancestors`.`child_group_id` = `parent_ancestors`.`child_group_id`)
        WHERE `child_ancestors`.`ancestor_group_id` = OLD.`child_group_id`;
    END IF;
    IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR OLD.type != NEW.type) THEN
        INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd

DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE DEFINER=`algorea`@`%` TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
        SELECT `groups_ancestors`.`child_group_id`, 'todo'
        FROM `groups_ancestors`
        WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
    ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    DELETE `groups_ancestors` FROM `groups_ancestors`
    WHERE `groups_ancestors`.`child_group_id` = OLD.`child_group_id` AND
            `groups_ancestors`.`ancestor_group_id` = OLD.`parent_group_id`;
    DELETE `bridges`
    FROM `groups_ancestors` `child_descendants`
             JOIN `groups_ancestors` `parent_ancestors`
             JOIN `groups_ancestors` `bridges`
                  ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
                      `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
    WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
            `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;
    DELETE `child_ancestors`
    FROM `groups_ancestors` `child_ancestors`
             JOIN  `groups_ancestors` `parent_ancestors`
                   ON (`child_ancestors`.`child_group_id` = OLD.`child_group_id` AND
                       `child_ancestors`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id`)
    WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id`;
    DELETE `parent_ancestors`
    FROM `groups_ancestors` `parent_ancestors`
             JOIN  `groups_ancestors` `child_ancestors`
                   ON (`parent_ancestors`.`ancestor_group_id` = OLD.`parent_group_id` AND
                       `child_ancestors`.`child_group_id` = `parent_ancestors`.`child_group_id`)
    WHERE `child_ancestors`.`ancestor_group_id` = OLD.`child_group_id`;
END
-- +migrate StatementEnd

ALTER TABLE `groups_groups` DROP COLUMN `expires_at`;
ALTER TABLE `groups_ancestors` DROP COLUMN `expires_at`;

DROP VIEW groups_ancestors_active;
DROP VIEW groups_groups_active;
