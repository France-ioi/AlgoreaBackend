-- +migrate Up
DROP TABLE `history_filters`;
ALTER TABLE `filters` DROP INDEX `version`, DROP COLUMN `version`;
DROP TRIGGER IF EXISTS `before_insert_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_filters` BEFORE INSERT ON `filters` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_filters`;
DROP TRIGGER IF EXISTS `before_update_filters`;
DROP TRIGGER IF EXISTS `before_delete_filters`;

ALTER TABLE `groups` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_groups`;
DROP TRIGGER IF EXISTS `before_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups` BEFORE INSERT ON `groups` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_groups`;
DROP TRIGGER IF EXISTS `before_delete_groups`;

ALTER TABLE `groups_ancestors` DROP COLUMN `version`;
DROP TABLE `history_groups_ancestors`;
DROP TRIGGER IF EXISTS `before_insert_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_ancestors` BEFORE INSERT ON `groups_ancestors` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups_ancestors`;
DROP TRIGGER IF EXISTS `before_update_groups_ancestors`;
DROP TRIGGER IF EXISTS `before_delete_groups_ancestors`;

ALTER TABLE `groups_attempts` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_groups_attempts`;
DROP TRIGGER IF EXISTS `before_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_attempts` BEFORE INSERT ON `groups_attempts` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SET NEW.minus_score = -NEW.score; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups_attempts`;
DROP TRIGGER IF EXISTS `before_update_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_attempts` BEFORE UPDATE ON `groups_attempts` FOR EACH ROW BEGIN IF NOT (OLD.`score` <=> NEW.`score`) THEN SET NEW.minus_score = -NEW.score; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_groups_attempts`;

ALTER TABLE `groups_groups` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_groups_groups`;
DROP TRIGGER IF EXISTS `before_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups_groups`;
DROP TRIGGER IF EXISTS `before_update_groups_groups`;
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
DROP TRIGGER IF EXISTS `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
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

ALTER TABLE `groups_items` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_groups_items`;
DROP TRIGGER IF EXISTS `before_insert_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_items` BEFORE INSERT ON `groups_items` FOR EACH ROW BEGIN
    IF (NEW.id IS NULL OR NEW.id = 0) THEN
        SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN
    INSERT INTO `groups_items_propagate` (`id`, `propagate_access`) VALUE (NEW.`id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_access` = 'self';
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_groups_items`;
DROP TRIGGER IF EXISTS `before_delete_groups_items`;

ALTER TABLE `groups_login_prefixes` DROP COLUMN `version`;
DROP TABLE `history_groups_login_prefixes`;
DROP TRIGGER IF EXISTS `before_insert_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_login_prefixes` BEFORE INSERT ON `groups_login_prefixes` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups_login_prefixes`;
DROP TRIGGER IF EXISTS `before_update_groups_login_prefixes`;
DROP TRIGGER IF EXISTS `before_delete_groups_login_prefixes`;

ALTER TABLE `items` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_items`;
DROP TRIGGER IF EXISTS `before_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items` BEFORE INSERT ON `items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT platforms.id INTO @platformID FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1 ; SET NEW.platform_id=@platformID ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items` AFTER INSERT ON `items` FOR EACH ROW BEGIN INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items` BEFORE UPDATE ON `items` FOR EACH ROW BEGIN SELECT platforms.id INTO @platformID FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1 ; SET NEW.platform_id=@platformID ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_items`;

ALTER TABLE `items_ancestors` DROP COLUMN `version`;
DROP TABLE `history_items_ancestors`;
DROP TRIGGER IF EXISTS `before_insert_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_ancestors` BEFORE INSERT ON `items_ancestors` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_items_ancestors`;
DROP TRIGGER IF EXISTS `before_update_items_ancestors`;
DROP TRIGGER IF EXISTS `before_delete_items_ancestors`;

ALTER TABLE `items_items` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_items_items`;
DROP TRIGGER IF EXISTS `before_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_items` BEFORE INSERT ON `items_items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state) VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `groups_items_propagate`
    SELECT `id`, 'children' as `propagate_access` FROM `groups_items`
    WHERE `groups_items`.`item_id` = NEW.`parent_item_id`;
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_items_items`;
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
    END IF;
    IF (OLD.child_item_id != NEW.child_item_id OR OLD.parent_item_id != NEW.parent_item_id) THEN
        INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state)
        VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_items_items`;
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
    INSERT IGNORE INTO `groups_items_propagate`
    SELECT `id`, 'children' as `propagate_access` FROM `groups_items`
    WHERE `groups_items`.`item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd

ALTER TABLE `items_strings` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_items_strings`;
DROP TRIGGER IF EXISTS `before_insert_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_strings` BEFORE INSERT ON `items_strings` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_items_strings`;
DROP TRIGGER IF EXISTS `before_update_items_strings`;
DROP TRIGGER IF EXISTS `before_delete_items_strings`;

ALTER TABLE `languages` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_languages`;
DROP TRIGGER IF EXISTS `before_insert_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_languages` BEFORE INSERT ON `languages` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_languages`;
DROP TRIGGER IF EXISTS `before_update_languages`;
DROP TRIGGER IF EXISTS `before_delete_languages`;

ALTER TABLE `messages` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_messages`;
DROP TRIGGER IF EXISTS `before_insert_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_messages` BEFORE INSERT ON `messages` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_messages`;
DROP TRIGGER IF EXISTS `before_update_messages`;
DROP TRIGGER IF EXISTS `before_delete_messages`;

ALTER TABLE `threads` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_threads`;
DROP TRIGGER IF EXISTS `before_insert_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_threads` BEFORE INSERT ON `threads` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_threads`;
DROP TRIGGER IF EXISTS `before_update_threads`;
DROP TRIGGER IF EXISTS `before_delete_threads`;

ALTER TABLE `users` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_users`;
DROP TRIGGER IF EXISTS `before_insert_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users` BEFORE INSERT ON `users` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_users`;
DROP TRIGGER IF EXISTS `before_update_users`;
DROP TRIGGER IF EXISTS `before_delete_users`;

ALTER TABLE `users_items` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_users_items`;
DROP TRIGGER IF EXISTS `before_insert_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users_items` BEFORE INSERT ON `users_items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_users_items`;
DROP TRIGGER IF EXISTS `before_update_users_items`;
DROP TRIGGER IF EXISTS `before_delete_users_items`;

ALTER TABLE `users_threads` DROP INDEX `version`, DROP COLUMN `version`;
DROP TABLE `history_users_threads`;
DROP TRIGGER IF EXISTS `before_insert_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users_threads` BEFORE INSERT ON `users_threads` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_users_threads`;
DROP TRIGGER IF EXISTS `before_update_users_threads`;
DROP TRIGGER IF EXISTS `before_delete_users_threads`;

DROP TABLE IF EXISTS `schema_revision`;
DROP TABLE IF EXISTS `synchro_version`;

-- +migrate Down
ALTER TABLE `filters` ADD COLUMN `version` bigint(20) NOT NULL, ADD INDEX `version` (`version`);
CREATE TABLE `history_filters` (
  `history_id` int(11) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `name` varchar(45) NOT NULL DEFAULT '',
  `selected` tinyint(1) NOT NULL DEFAULT '0',
  `starred` tinyint(1) DEFAULT NULL,
  `start_date` datetime DEFAULT NULL,
  `end_date` datetime DEFAULT NULL,
  `archived` tinyint(1) DEFAULT NULL,
  `participated` tinyint(1) DEFAULT NULL,
  `unread` tinyint(1) DEFAULT NULL,
  `item_id` bigint(20) DEFAULT NULL,
  `group_id` bigint(20) DEFAULT NULL,
  `older_than` int(11) DEFAULT NULL,
  `newer_than` int(11) DEFAULT NULL,
  `users_search` varchar(200) DEFAULT NULL,
  `body_search` varchar(100) DEFAULT NULL,
  `important` tinyint(1) DEFAULT NULL,
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `user_idx` (`user_id`),
  KEY `version` (`version`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_filters` BEFORE INSERT ON `filters` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_filters` AFTER INSERT ON `filters` FOR EACH ROW BEGIN INSERT INTO `history_filters` (`id`,`version`,`user_id`,`name`,`selected`,`starred`,`start_date`,`end_date`,`archived`,`participated`,`unread`,`item_id`,`group_id`,`older_than`,`newer_than`,`users_search`,`body_search`,`important`) VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`name`,NEW.`selected`,NEW.`starred`,NEW.`start_date`,NEW.`end_date`,NEW.`archived`,NEW.`participated`,NEW.`unread`,NEW.`item_id`,NEW.`group_id`,NEW.`older_than`,NEW.`newer_than`,NEW.`users_search`,NEW.`body_search`,NEW.`important`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_filters` BEFORE UPDATE ON `filters` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`name` <=> NEW.`name` AND OLD.`starred` <=> NEW.`starred` AND OLD.`start_date` <=> NEW.`start_date` AND OLD.`end_date` <=> NEW.`end_date` AND OLD.`archived` <=> NEW.`archived` AND OLD.`participated` <=> NEW.`participated` AND OLD.`unread` <=> NEW.`unread` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`older_than` <=> NEW.`older_than` AND OLD.`newer_than` <=> NEW.`newer_than` AND OLD.`users_search` <=> NEW.`users_search` AND OLD.`body_search` <=> NEW.`body_search` AND OLD.`important` <=> NEW.`important`) THEN   SET NEW.version = @curVersion;   UPDATE `history_filters` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_filters` (`id`,`version`,`user_id`,`name`,`selected`,`starred`,`start_date`,`end_date`,`archived`,`participated`,`unread`,`item_id`,`group_id`,`older_than`,`newer_than`,`users_search`,`body_search`,`important`)       VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`name`,NEW.`selected`,NEW.`starred`,NEW.`start_date`,NEW.`end_date`,NEW.`archived`,NEW.`participated`,NEW.`unread`,NEW.`item_id`,NEW.`group_id`,NEW.`older_than`,NEW.`newer_than`,NEW.`users_search`,NEW.`body_search`,NEW.`important`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_filters` BEFORE DELETE ON `filters` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_filters` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_filters` (`id`,`version`,`user_id`,`name`,`selected`,`starred`,`start_date`,`end_date`,`archived`,`participated`,`unread`,`item_id`,`group_id`,`older_than`,`newer_than`,`users_search`,`body_search`,`important`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`user_id`,OLD.`name`,OLD.`selected`,OLD.`starred`,OLD.`start_date`,OLD.`end_date`,OLD.`archived`,OLD.`participated`,OLD.`unread`,OLD.`item_id`,OLD.`group_id`,OLD.`older_than`,OLD.`newer_than`,OLD.`users_search`,OLD.`body_search`,OLD.`important`, 1); END
-- +migrate StatementEnd

ALTER TABLE `groups` ADD COLUMN `version` bigint(20) NOT NULL AFTER `send_emails`, ADD INDEX `version` (`version`);
CREATE TABLE `history_groups` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `name` varchar(200) NOT NULL DEFAULT '',
  `grade` int(4) NOT NULL DEFAULT '-2',
  `grade_details` varchar(50) DEFAULT NULL,
  `description` text,
  `created_at` datetime DEFAULT NULL,
  `opened` tinyint(1) NOT NULL,
  `free_access` tinyint(1) NOT NULL,
  `team_item_id` bigint(20) DEFAULT NULL,
  `team_participating` tinyint(1) NOT NULL DEFAULT '0',
  `code` varchar(50) DEFAULT NULL,
  `code_lifetime` time DEFAULT NULL,
  `code_expires_at` datetime DEFAULT NULL,
  `redirect_path` text,
  `open_contest` tinyint(1) NOT NULL DEFAULT '0',
  `type` enum('Class','Team','Club','Friends','Other','UserSelf','UserAdmin','Base') NOT NULL,
  `send_emails` tinyint(1) NOT NULL,
  `ancestors_computed` tinyint(1) NOT NULL DEFAULT '0',
  `ancestors_computation_state` enum('done','processing','todo') NOT NULL DEFAULT 'todo',
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `lock_user_deletion_until` date DEFAULT NULL,
  PRIMARY KEY (`history_id`),
  KEY `version` (`version`),
  KEY `id` (`id`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups` BEFORE INSERT ON `groups` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_lifetime`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`) VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`created_at`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_lifetime`,NEW.`code_expires_at`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`); INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`name` <=> NEW.`name` AND OLD.`grade` <=> NEW.`grade` AND OLD.`grade_details` <=> NEW.`grade_details` AND OLD.`description` <=> NEW.`description` AND OLD.`created_at` <=> NEW.`created_at` AND OLD.`opened` <=> NEW.`opened` AND OLD.`free_access` <=> NEW.`free_access` AND OLD.`team_item_id` <=> NEW.`team_item_id` AND OLD.`team_participating` <=> NEW.`team_participating` AND OLD.`code` <=> NEW.`code` AND OLD.`code_lifetime` <=> NEW.`code_lifetime` AND OLD.`code_expires_at` <=> NEW.`code_expires_at` AND OLD.`redirect_path` <=> NEW.`redirect_path` AND OLD.`open_contest` <=> NEW.`open_contest` AND OLD.`type` <=> NEW.`type` AND OLD.`send_emails` <=> NEW.`send_emails`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_lifetime`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`)       VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`created_at`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_lifetime`,NEW.`code_expires_at`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups` BEFORE DELETE ON `groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`created_at`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_lifetime`,`code_expires_at`,`redirect_path`,`open_contest`,`type`,`send_emails`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`name`,OLD.`grade`,OLD.`grade_details`,OLD.`description`,OLD.`created_at`,OLD.`opened`,OLD.`free_access`,OLD.`team_item_id`,OLD.`team_participating`,OLD.`code`,OLD.`code_lifetime`,OLD.`code_expires_at`,OLD.`redirect_path`,OLD.`open_contest`,OLD.`type`,OLD.`send_emails`, 1); END
-- +migrate StatementEnd

ALTER TABLE `groups_ancestors` ADD COLUMN `version` bigint(20) NOT NULL;
CREATE TABLE `history_groups_ancestors` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `ancestor_group_id` bigint(20) NOT NULL,
  `child_group_id` bigint(20) NOT NULL,
  `is_self` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `version` (`version`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `ancestor_group_id` (`ancestor_group_id`,`child_group_id`),
  KEY `ancestor` (`ancestor_group_id`),
  KEY `descendant` (`child_group_id`),
  KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_ancestors` BEFORE INSERT ON `groups_ancestors` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_ancestors` AFTER INSERT ON `groups_ancestors` FOR EACH ROW BEGIN INSERT INTO `history_groups_ancestors` (`id`,`version`,`ancestor_group_id`,`child_group_id`,`is_self`) VALUES (NEW.`id`,@curVersion,NEW.`ancestor_group_id`,NEW.`child_group_id`,NEW.`is_self`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_ancestors` BEFORE UPDATE ON `groups_ancestors` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`ancestor_group_id` <=> NEW.`ancestor_group_id` AND OLD.`child_group_id` <=> NEW.`child_group_id` AND OLD.`is_self` <=> NEW.`is_self`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups_ancestors` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups_ancestors` (`id`,`version`,`ancestor_group_id`,`child_group_id`,`is_self`)       VALUES (NEW.`id`,@curVersion,NEW.`ancestor_group_id`,NEW.`child_group_id`,NEW.`is_self`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_ancestors` BEFORE DELETE ON `groups_ancestors` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_ancestors` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups_ancestors` (`id`,`version`,`ancestor_group_id`,`child_group_id`,`is_self`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`ancestor_group_id`,OLD.`child_group_id`,OLD.`is_self`, 1); END
-- +migrate StatementEnd

ALTER TABLE `groups_attempts` ADD COLUMN `version` bigint(20) NOT NULL AFTER `all_lang_prog`, ADD INDEX `version` (`version`);
CREATE TABLE `history_groups_attempts` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `group_id` bigint(20) NOT NULL,
  `item_id` bigint(20) NOT NULL,
  `creator_user_id` bigint(20) DEFAULT NULL,
  `order` int(11) NOT NULL,
  `score` float NOT NULL DEFAULT '0',
  `score_computed` float NOT NULL DEFAULT '0',
  `score_reeval` float DEFAULT '0',
  `score_diff_manual` float NOT NULL DEFAULT '0',
  `score_diff_comment` varchar(200) NOT NULL DEFAULT '',
  `submissions_attempts` int(11) NOT NULL DEFAULT '0',
  `tasks_tried` int(11) NOT NULL DEFAULT '0',
  `tasks_solved` int(11) NOT NULL DEFAULT '0',
  `children_validated` int(11) NOT NULL DEFAULT '0',
  `validated` tinyint(1) NOT NULL DEFAULT '0',
  `finished` tinyint(1) NOT NULL DEFAULT '0',
  `key_obtained` tinyint(1) NOT NULL DEFAULT '0',
  `tasks_with_help` int(11) NOT NULL DEFAULT '0',
  `hints_requested` mediumtext,
  `hints_cached` int(11) NOT NULL DEFAULT '0',
  `corrections_read` int(11) NOT NULL DEFAULT '0',
  `precision` int(11) NOT NULL DEFAULT '0',
  `autonomy` int(11) NOT NULL DEFAULT '0',
  `started_at` datetime DEFAULT NULL,
  `validated_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  `latest_activity_at` datetime DEFAULT NULL,
  `thread_started_at` datetime DEFAULT NULL,
  `best_answer_at` datetime DEFAULT NULL,
  `latest_answer_at` datetime DEFAULT NULL,
  `latest_hint_at` datetime DEFAULT NULL,
  `ranked` tinyint(1) NOT NULL DEFAULT '0',
  `all_lang_prog` varchar(200) DEFAULT NULL,
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `version` (`version`),
  KEY `item_id` (`item_id`),
  KEY `group_item` (`group_id`,`item_id`),
  KEY `group_id` (`group_id`),
  KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_attempts` BEFORE INSERT ON `groups_attempts` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; SET NEW.minus_score = -NEW.score; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_attempts` AFTER INSERT ON `groups_attempts` FOR EACH ROW BEGIN INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`ranked`,`all_lang_prog`) VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`order`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`started_at`,NEW.`validated_at`,NEW.`best_answer_at`,NEW.`latest_answer_at`,NEW.`thread_started_at`,NEW.`latest_hint_at`,NEW.`finished_at`,NEW.`latest_activity_at`,NEW.`ranked`,NEW.`all_lang_prog`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_attempts` BEFORE UPDATE ON `groups_attempts` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`order` <=> NEW.`order` AND OLD.`score` <=> NEW.`score` AND OLD.`score_computed` <=> NEW.`score_computed` AND OLD.`score_reeval` <=> NEW.`score_reeval` AND OLD.`score_diff_manual` <=> NEW.`score_diff_manual` AND OLD.`score_diff_comment` <=> NEW.`score_diff_comment` AND OLD.`submissions_attempts` <=> NEW.`submissions_attempts` AND OLD.`tasks_tried` <=> NEW.`tasks_tried` AND OLD.`children_validated` <=> NEW.`children_validated` AND OLD.`validated` <=> NEW.`validated` AND OLD.`finished` <=> NEW.`finished` AND OLD.`key_obtained` <=> NEW.`key_obtained` AND OLD.`tasks_with_help` <=> NEW.`tasks_with_help` AND OLD.`hints_requested` <=> NEW.`hints_requested` AND OLD.`hints_cached` <=> NEW.`hints_cached` AND OLD.`corrections_read` <=> NEW.`corrections_read` AND OLD.`precision` <=> NEW.`precision` AND OLD.`autonomy` <=> NEW.`autonomy` AND OLD.`started_at` <=> NEW.`started_at` AND OLD.`validated_at` <=> NEW.`validated_at` AND OLD.`best_answer_at` <=> NEW.`best_answer_at` AND OLD.`latest_answer_at` <=> NEW.`latest_answer_at` AND OLD.`thread_started_at` <=> NEW.`thread_started_at` AND OLD.`latest_hint_at` <=> NEW.`latest_hint_at` AND OLD.`finished_at` <=> NEW.`finished_at` AND OLD.`ranked` <=> NEW.`ranked` AND OLD.`all_lang_prog` <=> NEW.`all_lang_prog`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups_attempts` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`ranked`,`all_lang_prog`)       VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`order`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`started_at`,NEW.`validated_at`,NEW.`best_answer_at`,NEW.`latest_answer_at`,NEW.`thread_started_at`,NEW.`latest_hint_at`,NEW.`finished_at`,NEW.`latest_activity_at`,NEW.`ranked`,NEW.`all_lang_prog`) ; SET NEW.minus_score = -NEW.score; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_attempts` BEFORE DELETE ON `groups_attempts` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_attempts` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`ranked`,`all_lang_prog`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`group_id`,OLD.`item_id`,OLD.`creator_user_id`,OLD.`order`,OLD.`score`,OLD.`score_computed`,OLD.`score_reeval`,OLD.`score_diff_manual`,OLD.`score_diff_comment`,OLD.`submissions_attempts`,OLD.`tasks_tried`,OLD.`children_validated`,OLD.`validated`,OLD.`finished`,OLD.`key_obtained`,OLD.`tasks_with_help`,OLD.`hints_requested`,OLD.`hints_cached`,OLD.`corrections_read`,OLD.`precision`,OLD.`autonomy`,OLD.`started_at`,OLD.`validated_at`,OLD.`best_answer_at`,OLD.`latest_answer_at`,OLD.`thread_started_at`,OLD.`latest_hint_at`,OLD.`finished_at`,OLD.`latest_activity_at`,OLD.`ranked`,OLD.`all_lang_prog`, 1); END
-- +migrate StatementEnd

ALTER TABLE `groups_groups` ADD COLUMN `version` bigint(20) NOT NULL, ADD INDEX `version` (`version`);
CREATE TABLE `history_groups_groups` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `parent_group_id` bigint(20) NOT NULL,
  `child_group_id` bigint(20) NOT NULL,
  `child_order` int(11) NOT NULL,
  `type` enum('invitationSent','requestSent','invitationAccepted','requestAccepted','invitationRefused','requestRefused','removed','left','direct','joinedByCode') NOT NULL DEFAULT 'direct',
  `role` enum('manager','owner','member','observer') NOT NULL DEFAULT 'member',
  `inviting_user_id` int(11) DEFAULT NULL,
  `type_changed_at` datetime DEFAULT NULL,
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `version` (`version`),
  KEY `id` (`id`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `parent_group_id` (`parent_group_id`),
  KEY `child_group_id` (`child_group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN INSERT INTO `history_groups_groups` (`id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`type_changed_at`,`inviting_user_id`) VALUES (NEW.`id`,@curVersion,NEW.`parent_group_id`,NEW.`child_group_id`,NEW.`child_order`,NEW.`type`,NEW.`role`,NEW.`type_changed_at`,NEW.`inviting_user_id`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF NEW.version <> OLD.version THEN
        SET @curVersion = NEW.version;
    ELSE
        SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    END IF;
    IF NOT (OLD.`id` = NEW.`id` AND OLD.`parent_group_id` <=> NEW.`parent_group_id` AND
            OLD.`child_group_id` <=> NEW.`child_group_id` AND OLD.`child_order` <=> NEW.`child_order`AND
            OLD.`type` <=> NEW.`type` AND OLD.`role` <=> NEW.`role` AND OLD.`type_changed_at` <=> NEW.`type_changed_at` AND
            OLD.`inviting_user_id` <=> NEW.`inviting_user_id`) THEN
        SET NEW.version = @curVersion;
        UPDATE `history_groups_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
        INSERT INTO `history_groups_groups` (
            `id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`type_changed_at`,`inviting_user_id`
        ) VALUES (
                     NEW.`id`,@curVersion,NEW.`parent_group_id`,NEW.`child_group_id`,NEW.`child_order`,NEW.`type`,NEW.`role`,
                     NEW.`type_changed_at`,NEW.`inviting_user_id`
                 );
    END IF;
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
DROP TRIGGER IF EXISTS `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    UPDATE `history_groups_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_groups_groups` (
        `id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`type_changed_at`,`inviting_user_id`,`deleted`
    ) VALUES (
                 OLD.`id`,@curVersion,OLD.`parent_group_id`,OLD.`child_group_id`,OLD.`child_order`,OLD.`type`,OLD.`role`,
                 OLD.`type_changed_at`,OLD.`inviting_user_id`, 1
             );
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

ALTER TABLE `groups_items` ADD COLUMN `version` bigint(20) NOT NULL, ADD INDEX `version` (`version`);
CREATE TABLE `history_groups_items` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `group_id` bigint(20) NOT NULL,
  `item_id` bigint(20) NOT NULL,
  `creator_user_id` bigint(20) NOT NULL,
  `partial_access_since` datetime DEFAULT NULL,
  `access_reason` varchar(200) DEFAULT NULL,
  `full_access_since` datetime DEFAULT NULL,
  `solutions_access_since` datetime DEFAULT NULL,
  `owner_access` tinyint(1) NOT NULL DEFAULT '0',
  `manager_access` tinyint(1) NOT NULL DEFAULT '0',
  `cached_full_access_since` datetime DEFAULT NULL,
  `cached_partial_access_since` datetime DEFAULT NULL,
  `cached_solutions_access_since` datetime DEFAULT NULL,
  `cached_grayed_access_since` datetime DEFAULT NULL,
  `cached_access_reason` varchar(200) DEFAULT NULL,
  `cached_full_access` tinyint(1) NOT NULL DEFAULT '0',
  `cached_partial_access` tinyint(1) NOT NULL DEFAULT '0',
  `cached_access_solutions` tinyint(1) NOT NULL DEFAULT '0',
  `cached_grayed_access` tinyint(1) NOT NULL DEFAULT '0',
  `cached_manager_access` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `version` (`version`),
  KEY `id` (`id`),
  KEY `item_group` (`item_id`,`group_id`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `item_id` (`item_id`),
  KEY `group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_items` BEFORE INSERT ON `groups_items` FOR EACH ROW BEGIN
    IF (NEW.id IS NULL OR NEW.id = 0) THEN
        SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;
    END IF;
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    SET NEW.version = @curVersion;
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN
    INSERT INTO `history_groups_items` (
        `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_since`,`full_access_since`,`access_reason`,
        `solutions_access_since`,`owner_access`,`manager_access`,`cached_partial_access_since`,`cached_full_access_since`,
        `cached_solutions_access_since`,`cached_grayed_access_since`,`cached_full_access`,`cached_partial_access`,
        `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`
    ) VALUES (
                 NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_since`,
                 NEW.`full_access_since`,NEW.`access_reason`,NEW.`solutions_access_since`,NEW.`owner_access`,
                 NEW.`manager_access`,NEW.`cached_partial_access_since`,NEW.`cached_full_access_since`,
                 NEW.`cached_solutions_access_since`,NEW.`cached_grayed_access_since`,NEW.`cached_full_access`,
                 NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,NEW.`cached_manager_access`
             );
    INSERT INTO `groups_items_propagate` (`id`, `propagate_access`) VALUE (NEW.`id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_access` = 'self';
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_items` BEFORE UPDATE ON `groups_items` FOR EACH ROW BEGIN
    IF NEW.version <> OLD.version THEN
        SET @curVersion = NEW.version;
    ELSE
        SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    END IF;
    IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`item_id` <=> NEW.`item_id` AND
            OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`partial_access_since` <=> NEW.`partial_access_since` AND
            OLD.`full_access_since` <=> NEW.`full_access_since` AND OLD.`access_reason` <=> NEW.`access_reason` AND
            OLD.`solutions_access_since` <=> NEW.`solutions_access_since` AND OLD.`owner_access` <=> NEW.`owner_access` AND
            OLD.`manager_access` <=> NEW.`manager_access` AND OLD.`cached_partial_access_since` <=> NEW.`cached_partial_access_since` AND
            OLD.`cached_full_access_since` <=> NEW.`cached_full_access_since` AND OLD.`cached_solutions_access_since` <=> NEW.`cached_solutions_access_since` AND
            OLD.`cached_grayed_access_since` <=> NEW.`cached_grayed_access_since` AND OLD.`cached_full_access` <=> NEW.`cached_full_access` AND
            OLD.`cached_partial_access` <=> NEW.`cached_partial_access` AND OLD.`cached_access_solutions` <=> NEW.`cached_access_solutions` AND
            OLD.`cached_grayed_access` <=> NEW.`cached_grayed_access` AND OLD.`cached_manager_access` <=> NEW.`cached_manager_access`) THEN
        SET NEW.version = @curVersion;
        UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
        INSERT INTO `history_groups_items` (
            `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_since`,`full_access_since`,`access_reason`,
            `solutions_access_since`,`owner_access`,`manager_access`,`cached_partial_access_since`,`cached_full_access_since`,
            `cached_solutions_access_since`,`cached_grayed_access_since`,`cached_full_access`,`cached_partial_access`,
            `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`
        ) VALUES (
                     NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_since`,
                     NEW.`full_access_since`,NEW.`access_reason`,NEW.`solutions_access_since`,NEW.`owner_access`,
                     NEW.`manager_access`,NEW.`cached_partial_access_since`,NEW.`cached_full_access_since`,
                     NEW.`cached_solutions_access_since`,NEW.`cached_grayed_access_since`,NEW.`cached_full_access`,
                     NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,
                     NEW.`cached_manager_access`
                 );
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_items` BEFORE DELETE ON `groups_items` FOR EACH ROW BEGIN
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_groups_items` (
        `id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_since`,`full_access_since`,`access_reason`,
        `solutions_access_since`,`owner_access`,`manager_access`,`cached_partial_access_since`,`cached_full_access_since`,
        `cached_solutions_access_since`,`cached_grayed_access_since`,`cached_full_access`,`cached_partial_access`,
        `cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`deleted`
    ) VALUES (
                 OLD.`id`,@curVersion,OLD.`group_id`,OLD.`item_id`,OLD.`creator_user_id`,OLD.`partial_access_since`,
                 OLD.`full_access_since`,OLD.`access_reason`,OLD.`solutions_access_since`,OLD.`owner_access`,
                 OLD.`manager_access`,OLD.`cached_partial_access_since`,OLD.`cached_full_access_since`,
                 OLD.`cached_solutions_access_since`,OLD.`cached_grayed_access_since`,OLD.`cached_full_access`,
                 OLD.`cached_partial_access`,OLD.`cached_access_solutions`,OLD.`cached_grayed_access`,
                 OLD.`cached_manager_access`, 1);
END
-- +migrate StatementEnd


ALTER TABLE `groups_login_prefixes` ADD COLUMN `version` bigint(20) NOT NULL;
CREATE TABLE `history_groups_login_prefixes` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `group_id` bigint(20) NOT NULL,
  `prefix` varchar(100) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

DROP TRIGGER IF EXISTS `before_insert_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_login_prefixes` BEFORE INSERT ON `groups_login_prefixes` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_login_prefixes` AFTER INSERT ON `groups_login_prefixes` FOR EACH ROW BEGIN INSERT INTO `history_groups_login_prefixes` (`id`,`version`,`group_id`,`prefix`) VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`prefix`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_login_prefixes` BEFORE UPDATE ON `groups_login_prefixes` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`prefix` <=> NEW.`prefix`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups_login_prefixes` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups_login_prefixes` (`id`,`version`,`group_id`,`prefix`)       VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`prefix`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_login_prefixes` BEFORE DELETE ON `groups_login_prefixes` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_login_prefixes` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups_login_prefixes` (`id`,`version`,`group_id`,`prefix`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`group_id`,OLD.`prefix`, 1); END
-- +migrate StatementEnd

ALTER TABLE `items` ADD COLUMN `version` bigint(20) NOT NULL, ADD INDEX `version` (`version`);
CREATE TABLE `history_items` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `url` varchar(200) DEFAULT NULL,
  `platform_id` int(11) DEFAULT NULL,
  `text_id` varchar(200) DEFAULT NULL,
  `repository_path` text,
  `type` enum('Root','CustomProgressRoot','OfficialProgressRoot','CustomContestRoot','OfficialContestRoot','DomainRoot','Category','Level','Chapter','GenericChapter','StaticChapter','Section','Task','Course','ContestChapter','LimitedTimeChapter','Presentation') NOT NULL,
  `title_bar_visible` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `transparent_folder` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `display_details_in_parent` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT 'when true, display a large icon, the subtitle, and more within the parent chapter',
  `custom_chapter` tinyint(3) unsigned DEFAULT '0' COMMENT 'true if this is a chapter where users can add their own content. access to this chapter will not be propagated to its children',
  `display_children_as_tabs` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `uses_api` tinyint(1) NOT NULL DEFAULT '1',
  `read_only` tinyint(1) NOT NULL DEFAULT '0',
  `full_screen` enum('forceYes','forceNo','default','') NOT NULL DEFAULT 'default',
  `show_difficulty` tinyint(1) NOT NULL,
  `show_source` tinyint(1) NOT NULL,
  `hints_allowed` tinyint(1) NOT NULL,
  `fixed_ranks` tinyint(1) NOT NULL DEFAULT '0',
  `validation_type` enum('None','All','AllButOne','Categories','One','Manual') NOT NULL DEFAULT 'All',
  `validation_min` int(11) DEFAULT NULL,
  `preparation_state` enum('NotReady','Reviewing','Ready') NOT NULL DEFAULT 'NotReady',
  `unlocked_item_ids` text,
  `score_min_unlock` int(11) NOT NULL DEFAULT '100',
  `supported_lang_prog` varchar(200) DEFAULT NULL,
  `default_language_id` bigint(20) DEFAULT '1',
  `contest_entering_condition` enum('All','Half','One','None') NOT NULL DEFAULT 'None',
  `teams_editable` tinyint(1) NOT NULL,
  `qualified_group_id` bigint(20) DEFAULT NULL,
  `contest_max_team_size` int(11) NOT NULL DEFAULT '0',
  `has_attempts` tinyint(1) NOT NULL DEFAULT '0',
  `contest_opens_at` datetime DEFAULT NULL,
  `duration` time DEFAULT NULL,
  `contest_closes_at` datetime DEFAULT NULL,
  `show_user_infos` tinyint(1) NOT NULL DEFAULT '0',
  `contest_phase` enum('Running','Analysis','Closed') NOT NULL,
  `level` int(11) DEFAULT NULL,
  `no_score` tinyint(1) NOT NULL,
  `group_code_enter` tinyint(1) DEFAULT '0' COMMENT 'Offer users to enter through a group code',
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `id` (`id`),
  KEY `version` (`version`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items` BEFORE INSERT ON `items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; SELECT platforms.id INTO @platformID FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1 ; SET NEW.platform_id=@platformID ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items` AFTER INSERT ON `items` FOR EACH ROW BEGIN INSERT INTO `history_items` (`id`,`version`,`url`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`contest_entering_condition`,`teams_editable`,`qualified_group_id`,`contest_max_team_size`,`has_attempts`,`contest_opens_at`,`duration`,`contest_closes_at`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`) VALUES (NEW.`id`,@curVersion,NEW.`url`,NEW.`platform_id`,NEW.`text_id`,NEW.`repository_path`,NEW.`type`,NEW.`uses_api`,NEW.`read_only`,NEW.`full_screen`,NEW.`show_difficulty`,NEW.`show_source`,NEW.`hints_allowed`,NEW.`fixed_ranks`,NEW.`validation_type`,NEW.`validation_min`,NEW.`preparation_state`,NEW.`unlocked_item_ids`,NEW.`score_min_unlock`,NEW.`supported_lang_prog`,NEW.`default_language_id`,NEW.`contest_entering_condition`,NEW.`teams_editable`,NEW.`qualified_group_id`,NEW.`contest_max_team_size`,NEW.`has_attempts`,NEW.`contest_opens_at`,NEW.`duration`,NEW.`contest_closes_at`,NEW.`show_user_infos`,NEW.`contest_phase`,NEW.`level`,NEW.`no_score`,NEW.`title_bar_visible`,NEW.`transparent_folder`,NEW.`display_details_in_parent`,NEW.`display_children_as_tabs`,NEW.`custom_chapter`,NEW.`group_code_enter`); INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items` BEFORE UPDATE ON `items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`url` <=> NEW.`url` AND OLD.`platform_id` <=> NEW.`platform_id` AND OLD.`text_id` <=> NEW.`text_id` AND OLD.`repository_path` <=> NEW.`repository_path` AND OLD.`type` <=> NEW.`type` AND OLD.`uses_api` <=> NEW.`uses_api` AND OLD.`read_only` <=> NEW.`read_only` AND OLD.`full_screen` <=> NEW.`full_screen` AND OLD.`show_difficulty` <=> NEW.`show_difficulty` AND OLD.`show_source` <=> NEW.`show_source` AND OLD.`hints_allowed` <=> NEW.`hints_allowed` AND OLD.`fixed_ranks` <=> NEW.`fixed_ranks` AND OLD.`validation_type` <=> NEW.`validation_type` AND OLD.`validation_min` <=> NEW.`validation_min` AND OLD.`preparation_state` <=> NEW.`preparation_state` AND OLD.`unlocked_item_ids` <=> NEW.`unlocked_item_ids` AND OLD.`score_min_unlock` <=> NEW.`score_min_unlock` AND OLD.`supported_lang_prog` <=> NEW.`supported_lang_prog` AND OLD.`default_language_id` <=> NEW.`default_language_id` AND OLD.`contest_entering_condition` <=> NEW.`contest_entering_condition` AND OLD.`teams_editable` <=> NEW.`teams_editable` AND OLD.`qualified_group_id` <=> NEW.`qualified_group_id` AND OLD.`contest_max_team_size` <=> NEW.`contest_max_team_size` AND OLD.`has_attempts` <=> NEW.`has_attempts` AND OLD.`contest_opens_at` <=> NEW.`contest_opens_at` AND OLD.`duration` <=> NEW.`duration` AND OLD.`contest_closes_at` <=> NEW.`contest_closes_at` AND OLD.`show_user_infos` <=> NEW.`show_user_infos` AND OLD.`contest_phase` <=> NEW.`contest_phase` AND OLD.`level` <=> NEW.`level` AND OLD.`no_score` <=> NEW.`no_score` AND OLD.`title_bar_visible` <=> NEW.`title_bar_visible` AND OLD.`transparent_folder` <=> NEW.`transparent_folder` AND OLD.`display_details_in_parent` <=> NEW.`display_details_in_parent` AND OLD.`display_children_as_tabs` <=> NEW.`display_children_as_tabs` AND OLD.`custom_chapter` <=> NEW.`custom_chapter` AND OLD.`group_code_enter` <=> NEW.`group_code_enter`) THEN   SET NEW.version = @curVersion;   UPDATE `history_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_items` (`id`,`version`,`url`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`contest_entering_condition`,`teams_editable`,`qualified_group_id`,`contest_max_team_size`,`has_attempts`,`contest_opens_at`,`duration`,`contest_closes_at`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`)       VALUES (NEW.`id`,@curVersion,NEW.`url`,NEW.`platform_id`,NEW.`text_id`,NEW.`repository_path`,NEW.`type`,NEW.`uses_api`,NEW.`read_only`,NEW.`full_screen`,NEW.`show_difficulty`,NEW.`show_source`,NEW.`hints_allowed`,NEW.`fixed_ranks`,NEW.`validation_type`,NEW.`validation_min`,NEW.`preparation_state`,NEW.`unlocked_item_ids`,NEW.`score_min_unlock`,NEW.`supported_lang_prog`,NEW.`default_language_id`,NEW.`contest_entering_condition`,NEW.`teams_editable`,NEW.`qualified_group_id`,NEW.`contest_max_team_size`,NEW.`has_attempts`,NEW.`contest_opens_at`,NEW.`duration`,NEW.`contest_closes_at`,NEW.`show_user_infos`,NEW.`contest_phase`,NEW.`level`,NEW.`no_score`,NEW.`title_bar_visible`,NEW.`transparent_folder`,NEW.`display_details_in_parent`,NEW.`display_children_as_tabs`,NEW.`custom_chapter`,NEW.`group_code_enter`) ; END IF; SELECT platforms.id INTO @platformID FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1 ; SET NEW.platform_id=@platformID ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items` BEFORE DELETE ON `items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_items` (`id`,`version`,`url`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`contest_entering_condition`,`teams_editable`,`qualified_group_id`,`contest_max_team_size`,`has_attempts`,`contest_opens_at`,`duration`,`contest_closes_at`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`url`,OLD.`platform_id`,OLD.`text_id`,OLD.`repository_path`,OLD.`type`,OLD.`uses_api`,OLD.`read_only`,OLD.`full_screen`,OLD.`show_difficulty`,OLD.`show_source`,OLD.`hints_allowed`,OLD.`fixed_ranks`,OLD.`validation_type`,OLD.`validation_min`,OLD.`preparation_state`,OLD.`unlocked_item_ids`,OLD.`score_min_unlock`,OLD.`supported_lang_prog`,OLD.`default_language_id`,OLD.`contest_entering_condition`,OLD.`teams_editable`,OLD.`qualified_group_id`,OLD.`contest_max_team_size`,OLD.`has_attempts`,OLD.`contest_opens_at`,OLD.`duration`,OLD.`contest_closes_at`,OLD.`show_user_infos`,OLD.`contest_phase`,OLD.`level`,OLD.`no_score`,OLD.`title_bar_visible`,OLD.`transparent_folder`,OLD.`display_details_in_parent`,OLD.`display_children_as_tabs`,OLD.`custom_chapter`,OLD.`group_code_enter`, 1); END
-- +migrate StatementEnd

ALTER TABLE `items_ancestors` ADD COLUMN `version` bigint(20) NOT NULL;
CREATE TABLE `history_items_ancestors` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `ancestor_item_id` bigint(20) NOT NULL,
  `child_item_id` bigint(20) NOT NULL,
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `version` (`version`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `ancestor_item_id_child_item_id` (`ancestor_item_id`,`child_item_id`),
  KEY `ancestor_item_id` (`ancestor_item_id`),
  KEY `child_item_id` (`child_item_id`),
  KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_ancestors` BEFORE INSERT ON `items_ancestors` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_ancestors` AFTER INSERT ON `items_ancestors` FOR EACH ROW BEGIN INSERT INTO `history_items_ancestors` (`id`,`version`,`ancestor_item_id`,`child_item_id`) VALUES (NEW.`id`,@curVersion,NEW.`ancestor_item_id`,NEW.`child_item_id`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_ancestors` BEFORE UPDATE ON `items_ancestors` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`ancestor_item_id` <=> NEW.`ancestor_item_id` AND OLD.`child_item_id` <=> NEW.`child_item_id`) THEN   SET NEW.version = @curVersion;   UPDATE `history_items_ancestors` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_items_ancestors` (`id`,`version`,`ancestor_item_id`,`child_item_id`)       VALUES (NEW.`id`,@curVersion,NEW.`ancestor_item_id`,NEW.`child_item_id`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_ancestors` BEFORE DELETE ON `items_ancestors` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_ancestors` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_items_ancestors` (`id`,`version`,`ancestor_item_id`,`child_item_id`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`ancestor_item_id`,OLD.`child_item_id`, 1); END
-- +migrate StatementEnd

ALTER TABLE `items_items` ADD COLUMN `version` bigint(20) NOT NULL, ADD INDEX `version` (`version`);
CREATE TABLE `history_items_items` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `parent_item_id` bigint(20) NOT NULL,
  `child_item_id` bigint(20) NOT NULL,
  `child_order` int(11) NOT NULL,
  `category` enum('Undefined','Discovery','Application','Validation','Challenge') NOT NULL DEFAULT 'Undefined',
  `partial_access_propagation` enum('None','AsGrayed','AsPartial') NOT NULL DEFAULT 'None',
  `difficulty` int(11) NOT NULL DEFAULT '0',
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `id` (`id`),
  KEY `version` (`version`),
  KEY `parent_item_id` (`parent_item_id`),
  KEY `child_item_id` (`child_item_id`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `parent_child` (`parent_item_id`,`child_item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_items` BEFORE INSERT ON `items_items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state) VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                       `partial_access_propagation`,`difficulty`)
    VALUES (NEW.`id`,@curVersion,NEW.`parent_item_id`,NEW.`child_item_id`,NEW.`child_order`,NEW.`category`,
            NEW.`partial_access_propagation`,NEW.`difficulty`);
    INSERT IGNORE INTO `groups_items_propagate`
    SELECT `id`, 'children' as `propagate_access` FROM `groups_items`
    WHERE `groups_items`.`item_id` = NEW.`parent_item_id`;
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_items` BEFORE UPDATE ON `items_items` FOR EACH ROW BEGIN
    IF NEW.version <> OLD.version THEN
        SET @curVersion = NEW.version;
    ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    END IF;
    IF NOT (OLD.`id` = NEW.`id` AND OLD.`parent_item_id` <=> NEW.`parent_item_id` AND
            OLD.`child_item_id` <=> NEW.`child_item_id` AND OLD.`child_order` <=> NEW.`child_order` AND
            OLD.`category` <=> NEW.`category` AND OLD.`partial_access_propagation` <=> NEW.`partial_access_propagation` AND
            OLD.`difficulty` <=> NEW.`difficulty`) THEN
        SET NEW.version = @curVersion;
        UPDATE `history_items_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
        INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                           `partial_access_propagation`,`difficulty`)
        VALUES (NEW.`id`,@curVersion,NEW.`parent_item_id`,NEW.`child_item_id`,NEW.`child_order`,NEW.`category`,
                NEW.`partial_access_propagation`,NEW.`difficulty`);
    END IF;
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
    END IF;
    IF (OLD.child_item_id != NEW.child_item_id OR OLD.parent_item_id != NEW.parent_item_id) THEN
        INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state)
        VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
    UPDATE `history_items_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,
                                       `partial_access_propagation`,`difficulty`, `deleted`)
    VALUES (OLD.`id`,@curVersion,OLD.`parent_item_id`,OLD.`child_item_id`,OLD.`child_order`,OLD.`category`,
            OLD.`partial_access_propagation`,OLD.`difficulty`, 1);
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
    INSERT IGNORE INTO `groups_items_propagate`
    SELECT `id`, 'children' as `propagate_access` FROM `groups_items`
    WHERE `groups_items`.`item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd

ALTER TABLE `items_strings` ADD COLUMN `version` bigint(20) NOT NULL, ADD INDEX `version` (`version`);
CREATE TABLE `history_items_strings` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `item_id` bigint(20) NOT NULL,
  `language_id` bigint(20) NOT NULL,
  `translator` varchar(100) DEFAULT NULL,
  `title` varchar(200) DEFAULT NULL,
  `image_url` text,
  `subtitle` varchar(200) DEFAULT NULL,
  `description` text,
  `edu_comment` text,
  `ranking_comment` text,
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `id` (`id`),
  KEY `version` (`version`),
  KEY `item_language` (`item_id`,`language_id`),
  KEY `item_id` (`item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_strings` BEFORE INSERT ON `items_strings` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_strings` AFTER INSERT ON `items_strings` FOR EACH ROW BEGIN INSERT INTO `history_items_strings` (`id`,`version`,`item_id`,`language_id`,`translator`,`title`,`image_url`,`subtitle`,`description`,`edu_comment`,`ranking_comment`) VALUES (NEW.`id`,@curVersion,NEW.`item_id`,NEW.`language_id`,NEW.`translator`,NEW.`title`,NEW.`image_url`,NEW.`subtitle`,NEW.`description`,NEW.`edu_comment`,NEW.`ranking_comment`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_strings` BEFORE UPDATE ON `items_strings` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`language_id` <=> NEW.`language_id` AND OLD.`translator` <=> NEW.`translator` AND OLD.`title` <=> NEW.`title` AND OLD.`image_url` <=> NEW.`image_url` AND OLD.`subtitle` <=> NEW.`subtitle` AND OLD.`description` <=> NEW.`description` AND OLD.`edu_comment` <=> NEW.`edu_comment` AND OLD.`ranking_comment` <=> NEW.`ranking_comment`) THEN   SET NEW.version = @curVersion;   UPDATE `history_items_strings` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_items_strings` (`id`,`version`,`item_id`,`language_id`,`translator`,`title`,`image_url`,`subtitle`,`description`,`edu_comment`,`ranking_comment`)       VALUES (NEW.`id`,@curVersion,NEW.`item_id`,NEW.`language_id`,NEW.`translator`,NEW.`title`,NEW.`image_url`,NEW.`subtitle`,NEW.`description`,NEW.`edu_comment`,NEW.`ranking_comment`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_strings` BEFORE DELETE ON `items_strings` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_strings` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_items_strings` (`id`,`version`,`item_id`,`language_id`,`translator`,`title`,`image_url`,`subtitle`,`description`,`edu_comment`,`ranking_comment`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`item_id`,OLD.`language_id`,OLD.`translator`,OLD.`title`,OLD.`image_url`,OLD.`subtitle`,OLD.`description`,OLD.`edu_comment`,OLD.`ranking_comment`, 1); END
-- +migrate StatementEnd

ALTER TABLE `languages` ADD COLUMN `version` bigint(20) NOT NULL, ADD INDEX `version` (`version`);
CREATE TABLE `history_languages` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `name` varchar(100) NOT NULL DEFAULT '',
  `code` varchar(2) NOT NULL DEFAULT '',
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `id` (`id`),
  KEY `version` (`version`),
  KEY `code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_languages` BEFORE INSERT ON `languages` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_languages` AFTER INSERT ON `languages` FOR EACH ROW BEGIN INSERT INTO `history_languages` (`id`,`version`,`name`,`code`) VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`code`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_languages` BEFORE UPDATE ON `languages` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`name` <=> NEW.`name` AND OLD.`code` <=> NEW.`code`) THEN   SET NEW.version = @curVersion;   UPDATE `history_languages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_languages` (`id`,`version`,`name`,`code`)       VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`code`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_languages` BEFORE DELETE ON `languages` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_languages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_languages` (`id`,`version`,`name`,`code`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`name`,OLD.`code`, 1); END
-- +migrate StatementEnd

ALTER TABLE `messages` ADD COLUMN `version` bigint(20) NOT NULL, ADD INDEX `version` (`version`);
CREATE TABLE `history_messages` (
  `history_id` int(11) NOT NULL AUTO_INCREMENT,
  `id` int(11) NOT NULL,
  `thread_id` int(11) NOT NULL,
  `user_id` bigint(20) DEFAULT NULL,
  `submitted_at` datetime DEFAULT NULL,
  `published` tinyint(1) NOT NULL DEFAULT '1',
  `title` varchar(200) DEFAULT '',
  `body` varchar(2000) DEFAULT '',
  `trainers_only` tinyint(1) NOT NULL DEFAULT '0',
  `archived` tinyint(1) DEFAULT '0',
  `persistant` tinyint(1) DEFAULT NULL,
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `thread` (`thread_id`),
  KEY `version` (`version`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_messages` BEFORE INSERT ON `messages` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_messages` AFTER INSERT ON `messages` FOR EACH ROW BEGIN INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submitted_at`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`) VALUES (NEW.`id`,@curVersion,NEW.`thread_id`,NEW.`user_id`,NEW.`submitted_at`,NEW.`published`,NEW.`title`,NEW.`body`,NEW.`trainers_only`,NEW.`archived`,NEW.`persistant`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_messages` BEFORE UPDATE ON `messages` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`thread_id` <=> NEW.`thread_id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`submitted_at` <=> NEW.`submitted_at` AND OLD.`published` <=> NEW.`published` AND OLD.`title` <=> NEW.`title` AND OLD.`body` <=> NEW.`body` AND OLD.`trainers_only` <=> NEW.`trainers_only` AND OLD.`archived` <=> NEW.`archived` AND OLD.`persistant` <=> NEW.`persistant`) THEN   SET NEW.version = @curVersion;   UPDATE `history_messages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submitted_at`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`)       VALUES (NEW.`id`,@curVersion,NEW.`thread_id`,NEW.`user_id`,NEW.`submitted_at`,NEW.`published`,NEW.`title`,NEW.`body`,NEW.`trainers_only`,NEW.`archived`,NEW.`persistant`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_messages` BEFORE DELETE ON `messages` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_messages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submitted_at`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`thread_id`,OLD.`user_id`,OLD.`submitted_at`,OLD.`published`,OLD.`title`,OLD.`body`,OLD.`trainers_only`,OLD.`archived`,OLD.`persistant`, 1); END
-- +migrate StatementEnd

ALTER TABLE `threads` ADD COLUMN `version` bigint(20) NOT NULL, ADD INDEX `version` (`version`);
CREATE TABLE `history_threads` (
  `history_id` int(11) NOT NULL AUTO_INCREMENT,
  `id` int(11) NOT NULL,
  `type` enum('Help','Bug','General') NOT NULL,
  `creator_user_id` bigint(20) NOT NULL,
  `item_id` bigint(20) DEFAULT NULL,
  `latest_activity_at` datetime NOT NULL,
  `title` varchar(200) DEFAULT NULL,
  `admin_help_asked` tinyint(1) NOT NULL DEFAULT '0',
  `hidden` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `version` (`version`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_threads` BEFORE INSERT ON `threads` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_threads` AFTER INSERT ON `threads` FOR EACH ROW BEGIN INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`latest_activity_at`) VALUES (NEW.`id`,@curVersion,NEW.`type`,NEW.`creator_user_id`,NEW.`item_id`,NEW.`title`,NEW.`admin_help_asked`,NEW.`hidden`,NEW.`latest_activity_at`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_threads` BEFORE UPDATE ON `threads` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`type` <=> NEW.`type` AND OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`title` <=> NEW.`title` AND OLD.`admin_help_asked` <=> NEW.`admin_help_asked` AND OLD.`hidden` <=> NEW.`hidden` AND OLD.`latest_activity_at` <=> NEW.`latest_activity_at`) THEN   SET NEW.version = @curVersion;   UPDATE `history_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`latest_activity_at`)       VALUES (NEW.`id`,@curVersion,NEW.`type`,NEW.`creator_user_id`,NEW.`item_id`,NEW.`title`,NEW.`admin_help_asked`,NEW.`hidden`,NEW.`latest_activity_at`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_threads` BEFORE DELETE ON `threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`latest_activity_at`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`type`,OLD.`creator_user_id`,OLD.`item_id`,OLD.`title`,OLD.`admin_help_asked`,OLD.`hidden`,OLD.`latest_activity_at`, 1); END
-- +migrate StatementEnd

ALTER TABLE `users` ADD COLUMN `version` bigint(20) NOT NULL AFTER `notifications_read_at`, ADD INDEX `version` (`version`);
CREATE TABLE `history_users` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `login_id` bigint(20) DEFAULT NULL,
  `login` varchar(100) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `open_id_identity` varchar(255) DEFAULT NULL COMMENT 'User''s Open Id Identity',
  `password_md5` varchar(100) DEFAULT NULL,
  `salt` varchar(32) DEFAULT NULL,
  `recover` varchar(50) DEFAULT NULL,
  `registered_at` datetime DEFAULT NULL,
  `email` varchar(100) DEFAULT NULL,
  `email_verified` tinyint(1) NOT NULL DEFAULT '0',
  `first_name` varchar(100) DEFAULT NULL COMMENT 'User''s first name',
  `last_name` varchar(100) DEFAULT NULL COMMENT 'User''s last name',
  `student_id` text,
  `country_code` char(3) NOT NULL DEFAULT '',
  `time_zone` varchar(100) DEFAULT NULL,
  `birth_date` date DEFAULT NULL COMMENT 'User''s birth date',
  `graduation_year` int(11) NOT NULL DEFAULT '0' COMMENT 'User''s high school graduation year',
  `grade` int(11) DEFAULT NULL,
  `sex` enum('Male','Female') DEFAULT NULL,
  `address` mediumtext COMMENT 'User''s address',
  `zipcode` longtext COMMENT 'User''s postal code',
  `city` longtext COMMENT 'User''s city',
  `land_line_number` longtext COMMENT 'User''s phone number',
  `cell_phone_number` longtext COMMENT 'User''s mobil phone number',
  `default_language` char(3) NOT NULL DEFAULT 'fr' COMMENT 'User''s default language',
  `notify_news` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'TODO',
  `notify` enum('Never','Answers','Concerned') NOT NULL DEFAULT 'Answers',
  `public_first_name` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Publicly show user''s first name',
  `public_last_name` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Publicly show user''s first name',
  `free_text` mediumtext,
  `web_site` varchar(100) DEFAULT NULL,
  `photo_autoload` tinyint(1) NOT NULL DEFAULT '0',
  `lang_prog` varchar(30) DEFAULT 'Python',
  `latest_login_at` datetime DEFAULT NULL,
  `latest_activity_at` datetime DEFAULT NULL COMMENT 'User''s last activity time on the website',
  `last_ip` varchar(16) DEFAULT NULL,
  `basic_editor_mode` tinyint(4) NOT NULL DEFAULT '1',
  `spaces_for_tab` int(11) NOT NULL DEFAULT '3',
  `member_state` tinyint(4) NOT NULL DEFAULT '0',
  `godfather_user_id` int(11) DEFAULT NULL,
  `step_level_in_site` int(11) NOT NULL DEFAULT '0' COMMENT 'User''s level',
  `is_admin` tinyint(4) NOT NULL DEFAULT '0',
  `no_ranking` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'TODO',
  `help_given` int(11) NOT NULL DEFAULT '0' COMMENT 'TODO',
  `self_group_id` bigint(20) DEFAULT NULL,
  `owned_group_id` bigint(20) DEFAULT NULL,
  `access_group_id` bigint(20) DEFAULT NULL,
  `notifications_read_at` datetime DEFAULT NULL,
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `login_module_prefix` varchar(100) DEFAULT NULL,
  `creator_id` bigint(20) DEFAULT NULL COMMENT 'which user created a given login with the login generation tool',
  `allow_subgroups` tinyint(4) DEFAULT NULL COMMENT 'Allow to create subgroups',
  PRIMARY KEY (`history_id`),
  KEY `id` (`id`),
  KEY `version` (`version`),
  KEY `country_code` (`country_code`),
  KEY `godfather_user_id` (`godfather_user_id`),
  KEY `lang_prog` (`lang_prog`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `self_group_id` (`self_group_id`),
  KEY `owned_group_id` (`owned_group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users` BEFORE INSERT ON `users` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_users`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users` AFTER INSERT ON `users` FOR EACH ROW BEGIN INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registered_at`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`latest_login_at`,`latest_activity_at`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notifications_read_at`,`login_module_prefix`,`allow_subgroups`) VALUES (NEW.`id`,@curVersion,NEW.`login`,NEW.`open_id_identity`,NEW.`password_md5`,NEW.`salt`,NEW.`recover`,NEW.`registered_at`,NEW.`email`,NEW.`email_verified`,NEW.`first_name`,NEW.`last_name`,NEW.`country_code`,NEW.`time_zone`,NEW.`birth_date`,NEW.`graduation_year`,NEW.`grade`,NEW.`sex`,NEW.`student_id`,NEW.`address`,NEW.`zipcode`,NEW.`city`,NEW.`land_line_number`,NEW.`cell_phone_number`,NEW.`default_language`,NEW.`notify_news`,NEW.`notify`,NEW.`public_first_name`,NEW.`public_last_name`,NEW.`free_text`,NEW.`web_site`,NEW.`photo_autoload`,NEW.`lang_prog`,NEW.`latest_login_at`,NEW.`latest_activity_at`,NEW.`last_ip`,NEW.`basic_editor_mode`,NEW.`spaces_for_tab`,NEW.`member_state`,NEW.`godfather_user_id`,NEW.`step_level_in_site`,NEW.`is_admin`,NEW.`no_ranking`,NEW.`help_given`,NEW.`self_group_id`,NEW.`owned_group_id`,NEW.`access_group_id`,NEW.`notifications_read_at`,NEW.`login_module_prefix`,NEW.`allow_subgroups`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users` BEFORE UPDATE ON `users` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`login` <=> NEW.`login` AND OLD.`open_id_identity` <=> NEW.`open_id_identity` AND OLD.`password_md5` <=> NEW.`password_md5` AND OLD.`salt` <=> NEW.`salt` AND OLD.`recover` <=> NEW.`recover` AND OLD.`registered_at` <=> NEW.`registered_at` AND OLD.`email` <=> NEW.`email` AND OLD.`email_verified` <=> NEW.`email_verified` AND OLD.`first_name` <=> NEW.`first_name` AND OLD.`last_name` <=> NEW.`last_name` AND OLD.`country_code` <=> NEW.`country_code` AND OLD.`time_zone` <=> NEW.`time_zone` AND OLD.`birth_date` <=> NEW.`birth_date` AND OLD.`graduation_year` <=> NEW.`graduation_year` AND OLD.`grade` <=> NEW.`grade` AND OLD.`sex` <=> NEW.`sex` AND OLD.`student_id` <=> NEW.`student_id` AND OLD.`address` <=> NEW.`address` AND OLD.`zipcode` <=> NEW.`zipcode` AND OLD.`city` <=> NEW.`city` AND OLD.`land_line_number` <=> NEW.`land_line_number` AND OLD.`cell_phone_number` <=> NEW.`cell_phone_number` AND OLD.`default_language` <=> NEW.`default_language` AND OLD.`notify_news` <=> NEW.`notify_news` AND OLD.`notify` <=> NEW.`notify` AND OLD.`public_first_name` <=> NEW.`public_first_name` AND OLD.`public_last_name` <=> NEW.`public_last_name` AND OLD.`free_text` <=> NEW.`free_text` AND OLD.`web_site` <=> NEW.`web_site` AND OLD.`photo_autoload` <=> NEW.`photo_autoload` AND OLD.`lang_prog` <=> NEW.`lang_prog` AND OLD.`latest_login_at` <=> NEW.`latest_login_at` AND OLD.`latest_activity_at` <=> NEW.`latest_activity_at` AND OLD.`last_ip` <=> NEW.`last_ip` AND OLD.`basic_editor_mode` <=> NEW.`basic_editor_mode` AND OLD.`spaces_for_tab` <=> NEW.`spaces_for_tab` AND OLD.`member_state` <=> NEW.`member_state` AND OLD.`godfather_user_id` <=> NEW.`godfather_user_id` AND OLD.`step_level_in_site` <=> NEW.`step_level_in_site` AND OLD.`is_admin` <=> NEW.`is_admin` AND OLD.`no_ranking` <=> NEW.`no_ranking` AND OLD.`help_given` <=> NEW.`help_given` AND OLD.`self_group_id` <=> NEW.`self_group_id` AND OLD.`owned_group_id` <=> NEW.`owned_group_id` AND OLD.`access_group_id` <=> NEW.`access_group_id` AND OLD.`notifications_read_at` <=> NEW.`notifications_read_at` AND OLD.`login_module_prefix` <=> NEW.`login_module_prefix` AND OLD.`allow_subgroups` <=> NEW.`allow_subgroups`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registered_at`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`latest_login_at`,`latest_activity_at`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notifications_read_at`,`login_module_prefix`,`allow_subgroups`)       VALUES (NEW.`id`,@curVersion,NEW.`login`,NEW.`open_id_identity`,NEW.`password_md5`,NEW.`salt`,NEW.`recover`,NEW.`registered_at`,NEW.`email`,NEW.`email_verified`,NEW.`first_name`,NEW.`last_name`,NEW.`country_code`,NEW.`time_zone`,NEW.`birth_date`,NEW.`graduation_year`,NEW.`grade`,NEW.`sex`,NEW.`student_id`,NEW.`address`,NEW.`zipcode`,NEW.`city`,NEW.`land_line_number`,NEW.`cell_phone_number`,NEW.`default_language`,NEW.`notify_news`,NEW.`notify`,NEW.`public_first_name`,NEW.`public_last_name`,NEW.`free_text`,NEW.`web_site`,NEW.`photo_autoload`,NEW.`lang_prog`,NEW.`latest_login_at`,NEW.`latest_activity_at`,NEW.`last_ip`,NEW.`basic_editor_mode`,NEW.`spaces_for_tab`,NEW.`member_state`,NEW.`godfather_user_id`,NEW.`step_level_in_site`,NEW.`is_admin`,NEW.`no_ranking`,NEW.`help_given`,NEW.`self_group_id`,NEW.`owned_group_id`,NEW.`access_group_id`,NEW.`notifications_read_at`,NEW.`login_module_prefix`,NEW.`allow_subgroups`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users` BEFORE DELETE ON `users` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registered_at`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`latest_login_at`,`latest_activity_at`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notifications_read_at`,`login_module_prefix`,`allow_subgroups`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`login`,OLD.`open_id_identity`,OLD.`password_md5`,OLD.`salt`,OLD.`recover`,OLD.`registered_at`,OLD.`email`,OLD.`email_verified`,OLD.`first_name`,OLD.`last_name`,OLD.`country_code`,OLD.`time_zone`,OLD.`birth_date`,OLD.`graduation_year`,OLD.`grade`,OLD.`sex`,OLD.`student_id`,OLD.`address`,OLD.`zipcode`,OLD.`city`,OLD.`land_line_number`,OLD.`cell_phone_number`,OLD.`default_language`,OLD.`notify_news`,OLD.`notify`,OLD.`public_first_name`,OLD.`public_last_name`,OLD.`free_text`,OLD.`web_site`,OLD.`photo_autoload`,OLD.`lang_prog`,OLD.`latest_login_at`,OLD.`latest_activity_at`,OLD.`last_ip`,OLD.`basic_editor_mode`,OLD.`spaces_for_tab`,OLD.`member_state`,OLD.`godfather_user_id`,OLD.`step_level_in_site`,OLD.`is_admin`,OLD.`no_ranking`,OLD.`help_given`,OLD.`self_group_id`,OLD.`owned_group_id`,OLD.`access_group_id`,OLD.`notifications_read_at`,OLD.`login_module_prefix`,OLD.`allow_subgroups`, 1); END
-- +migrate StatementEnd

ALTER TABLE `users_items` ADD COLUMN `version` bigint(20) NOT NULL AFTER `all_lang_prog`, ADD INDEX `version` (`version`);
CREATE TABLE `history_users_items` (
  `history_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `item_id` bigint(20) NOT NULL,
  `active_attempt_id` bigint(20) DEFAULT NULL,
  `score` float NOT NULL DEFAULT '0',
  `score_computed` float NOT NULL DEFAULT '0',
  `score_reeval` float DEFAULT '0',
  `score_diff_manual` float NOT NULL DEFAULT '0',
  `score_diff_comment` varchar(200) DEFAULT NULL,
  `submissions_attempts` int(11) NOT NULL,
  `tasks_tried` int(11) NOT NULL,
  `tasks_solved` int(11) NOT NULL DEFAULT '0',
  `children_validated` int(11) NOT NULL,
  `validated` int(11) NOT NULL,
  `finished` int(11) NOT NULL,
  `key_obtained` tinyint(1) NOT NULL DEFAULT '0',
  `tasks_with_help` int(11) NOT NULL,
  `hints_requested` mediumtext,
  `hints_cached` int(11) NOT NULL,
  `corrections_read` int(11) NOT NULL,
  `precision` int(11) NOT NULL,
  `autonomy` int(11) NOT NULL,
  `started_at` datetime DEFAULT NULL,
  `validated_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  `latest_activity_at` datetime DEFAULT NULL,
  `thread_started_at` datetime DEFAULT NULL,
  `best_answer_at` datetime DEFAULT NULL,
  `latest_answer_at` datetime DEFAULT NULL,
  `latest_hint_at` datetime DEFAULT NULL,
  `contest_started_at` datetime DEFAULT NULL,
  `ranked` tinyint(1) NOT NULL,
  `all_lang_prog` varchar(200) DEFAULT NULL,
  `state` mediumtext,
  `answer` mediumtext,
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `platform_data_removed` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `id` (`id`),
  KEY `version` (`version`),
  KEY `item_user` (`item_id`,`user_id`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `item_id` (`item_id`),
  KEY `user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users_items` BEFORE INSERT ON `users_items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users_items` AFTER INSERT ON `users_items` FOR EACH ROW BEGIN INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`ranked`,`all_lang_prog`,`state`,`answer`) VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`item_id`,NEW.`active_attempt_id`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`started_at`,NEW.`validated_at`,NEW.`best_answer_at`,NEW.`latest_answer_at`,NEW.`thread_started_at`,NEW.`latest_hint_at`,NEW.`finished_at`,NEW.`latest_activity_at`,NEW.`ranked`,NEW.`all_lang_prog`,NEW.`state`,NEW.`answer`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users_items` BEFORE UPDATE ON `users_items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`active_attempt_id` <=> NEW.`active_attempt_id` AND OLD.`score` <=> NEW.`score` AND OLD.`score_computed` <=> NEW.`score_computed` AND OLD.`score_reeval` <=> NEW.`score_reeval` AND OLD.`score_diff_manual` <=> NEW.`score_diff_manual` AND OLD.`score_diff_comment` <=> NEW.`score_diff_comment` AND OLD.`tasks_tried` <=> NEW.`tasks_tried` AND OLD.`children_validated` <=> NEW.`children_validated` AND OLD.`validated` <=> NEW.`validated` AND OLD.`finished` <=> NEW.`finished` AND OLD.`key_obtained` <=> NEW.`key_obtained` AND OLD.`tasks_with_help` <=> NEW.`tasks_with_help` AND OLD.`hints_requested` <=> NEW.`hints_requested` AND OLD.`hints_cached` <=> NEW.`hints_cached` AND OLD.`corrections_read` <=> NEW.`corrections_read` AND OLD.`precision` <=> NEW.`precision` AND OLD.`autonomy` <=> NEW.`autonomy` AND OLD.`started_at` <=> NEW.`started_at` AND OLD.`validated_at` <=> NEW.`validated_at` AND OLD.`best_answer_at` <=> NEW.`best_answer_at` AND OLD.`latest_answer_at` <=> NEW.`latest_answer_at` AND OLD.`thread_started_at` <=> NEW.`thread_started_at` AND OLD.`latest_hint_at` <=> NEW.`latest_hint_at` AND OLD.`finished_at` <=> NEW.`finished_at` AND OLD.`ranked` <=> NEW.`ranked` AND OLD.`all_lang_prog` <=> NEW.`all_lang_prog` AND OLD.`state` <=> NEW.`state` AND OLD.`answer` <=> NEW.`answer`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`ranked`,`all_lang_prog`,`state`,`answer`)       VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`item_id`,NEW.`active_attempt_id`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`started_at`,NEW.`validated_at`,NEW.`best_answer_at`,NEW.`latest_answer_at`,NEW.`thread_started_at`,NEW.`latest_hint_at`,NEW.`finished_at`,NEW.`latest_activity_at`,NEW.`ranked`,NEW.`all_lang_prog`,NEW.`state`,NEW.`answer`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users_items` BEFORE DELETE ON `users_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`started_at`,`validated_at`,`best_answer_at`,`latest_answer_at`,`thread_started_at`,`latest_hint_at`,`finished_at`,`latest_activity_at`,`ranked`,`all_lang_prog`,`state`,`answer`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`user_id`,OLD.`item_id`,OLD.`active_attempt_id`,OLD.`score`,OLD.`score_computed`,OLD.`score_reeval`,OLD.`score_diff_manual`,OLD.`score_diff_comment`,OLD.`submissions_attempts`,OLD.`tasks_tried`,OLD.`children_validated`,OLD.`validated`,OLD.`finished`,OLD.`key_obtained`,OLD.`tasks_with_help`,OLD.`hints_requested`,OLD.`hints_cached`,OLD.`corrections_read`,OLD.`precision`,OLD.`autonomy`,OLD.`started_at`,OLD.`validated_at`,OLD.`best_answer_at`,OLD.`latest_answer_at`,OLD.`thread_started_at`,OLD.`latest_hint_at`,OLD.`finished_at`,OLD.`latest_activity_at`,OLD.`ranked`,OLD.`all_lang_prog`,OLD.`state`,OLD.`answer`, 1); END
-- +migrate StatementEnd

ALTER TABLE `users_threads` ADD COLUMN `version` bigint(20) NOT NULL, ADD INDEX `version` (`version`);
CREATE TABLE `history_users_threads` (
  `history_id` int(11) NOT NULL AUTO_INCREMENT,
  `id` bigint(20) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `thread_id` bigint(20) NOT NULL,
  `lately_viewed_at` datetime DEFAULT NULL,
  `participated` tinyint(1) NOT NULL DEFAULT '0',
  `lately_posted_at` datetime DEFAULT NULL,
  `starred` tinyint(1) DEFAULT NULL,
  `version` bigint(20) NOT NULL,
  `next_version` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_id`),
  KEY `user_thread` (`user_id`,`thread_id`),
  KEY `user` (`user_id`),
  KEY `version` (`version`),
  KEY `next_version` (`next_version`),
  KEY `deleted` (`deleted`),
  KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TRIGGER IF EXISTS `before_insert_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users_threads` BEFORE INSERT ON `users_threads` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `after_insert_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users_threads` AFTER INSERT ON `users_threads` FOR EACH ROW BEGIN INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`lately_viewed_at`,`lately_posted_at`,`starred`) VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`thread_id`,NEW.`lately_viewed_at`,NEW.`lately_posted_at`,NEW.`starred`); END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users_threads` BEFORE UPDATE ON `users_threads` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`thread_id` <=> NEW.`thread_id` AND OLD.`lately_viewed_at` <=> NEW.`lately_viewed_at` AND OLD.`lately_posted_at` <=> NEW.`lately_posted_at` AND OLD.`starred` <=> NEW.`starred`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`lately_viewed_at`,`lately_posted_at`,`starred`)       VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`thread_id`,NEW.`lately_viewed_at`,NEW.`lately_posted_at`,NEW.`starred`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_delete_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users_threads` BEFORE DELETE ON `users_threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`lately_viewed_at`,`lately_posted_at`,`starred`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`user_id`,OLD.`thread_id`,OLD.`lately_viewed_at`,OLD.`lately_posted_at`,OLD.`starred`, 1); END
-- +migrate StatementEnd

DROP TABLE IF EXISTS `schema_revision`;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8mb4;
CREATE TABLE `schema_revision` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
    `executed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `file` varchar(255) NOT NULL DEFAULT '',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

LOCK TABLES `schema_revision` WRITE;
ALTER TABLE `schema_revision` DISABLE KEYS;
INSERT INTO `schema_revision` VALUES (1,'2018-10-24 06:11:11','1.0/revision-002/synchro_versions.sql'),(2,'2018-10-24 06:11:11','1.0/revision-003/history_tables.sql'),(3,'2018-10-24 06:11:11','1.0/revision-004/historyID_autoincrement.sql'),(4,'2018-10-24 06:11:11','1.0/revision-008/root_category.sql'),(5,'2018-10-24 06:11:11','1.0/revision-016/fix_item_sType.sql'),(6,'2018-10-24 06:11:11','1.0/revision-016/groups_tables.sql'),(7,'2018-10-24 06:11:11','1.0/revision-017/missing_fields.sql'),(8,'2018-10-24 06:11:11','1.0/revision-018/fix_history_groups.sql'),(9,'2018-10-24 06:11:11','1.0/revision-020/groups_sType.sql'),(10,'2018-10-24 06:11:11','1.0/revision-028/access_modes.sql'),(11,'2018-10-24 06:11:11','1.0/revision-029/access_solutions.sql'),(12,'2018-10-24 06:11:11','1.0/revision-029/validationType.sql'),(13,'2018-10-24 06:11:11','1.0/revision-031/drop_group_owners.sql'),(14,'2018-10-24 06:11:11','1.0/revision-031/users.sql'),(15,'2018-10-24 06:11:11','1.0/revision-032/isAdmin_fix.sql'),(16,'2018-10-24 06:11:11','1.0/revision-032/sNotify.sql'),(17,'2018-10-24 06:11:11','1.0/revision-032/user_groups.sql'),(18,'2018-10-24 06:11:11','1.0/revision-044/users_items.sql'),(19,'2018-10-24 06:11:11','1.0/revision-045/ancestors.sql'),(20,'2018-10-24 06:11:11','1.0/revision-055/index_ancestors.sql'),(21,'2018-10-24 06:11:11','1.0/revision-057/groups_items_access.sql'),(22,'2018-10-24 06:11:11','1.0/revision-060/items_ancestors_access.sql'),(23,'2018-10-24 06:11:11','1.0/revision-061/access_dates.sql'),(24,'2018-10-24 06:11:11','1.0/revision-066/nextVersion_null.sql'),(25,'2018-10-24 06:11:11','1.0/revision-139/completing_items_table.sql'),(26,'2018-10-24 06:11:11','1.0/revision-161/history_items.sql'),(27,'2018-10-24 06:11:11','1.0/revision-161/history_items_ancestors.sql'),(28,'2018-10-24 06:11:11','1.0/revision-212/user_item_index.sql'),(29,'2018-10-24 06:11:11','1.0/revision-227/groups_items_propagation.sql'),(30,'2018-10-24 06:11:11','1.0/revision-246/0-users_items_computations.sql'),(31,'2018-10-24 06:11:11','1.0/revision-246/items_groups_ancestors.sql'),(32,'2018-10-24 06:11:11','1.0/revision-246/optimizations.sql'),(33,'2018-10-24 06:11:11','1.0/revision-268/users_and_items.sql'),(34,'2018-10-24 06:11:11','1.0/revision-277/forum.sql'),(35,'2018-10-24 06:11:11','1.0/revision-321/manual_validation.sql'),(36,'2018-10-24 06:11:11','1.0/revision-321/platforms.sql'),(37,'2018-10-24 06:11:11','1.0/revision-321/users_answers.sql'),(38,'2018-10-24 06:11:11','1.0/revision-532/bugfixes.sql'),(39,'2018-10-24 06:11:11','1.0/revision-533/user_index.sql'),(40,'2018-10-24 06:11:11','1.0/revision-534/groups_items_index.sql'),(41,'2018-10-24 06:11:11','1.0/revision-536/index_item_item.sql'),(42,'2018-10-24 06:11:11','1.0/revision-537/default_propagation_users_items.sql'),(43,'2018-10-24 06:11:11','1.0/revision-538/task_url_index.sql'),(44,'2018-10-24 06:11:11','1.0/revision-539/item.bFullScreen.sql'),(45,'2018-10-24 06:11:11','1.0/revision-540/groups.sql'),(46,'2018-10-24 06:11:11','1.0/revision-541/items.sIconUrl.sql'),(47,'2018-10-24 06:11:11','1.0/revision-543/users.sRegistrationDate.sql'),(48,'2018-10-24 06:11:11','1.0/revision-544/limittedTimeContests.sql'),(49,'2018-10-24 06:11:11','1.0/revision-545/contest_adjustment.sql'),(50,'2018-10-24 06:11:11','1.0/revision-546/more_groups_groups_indexes.sql'),(51,'2018-10-24 06:11:11','1.0/revision-546/more_indexes.sql'),(52,'2018-10-24 06:11:11','1.0/revision-547/groups_defaults.sql'),(53,'2018-10-24 06:11:11','1.0/revision-548/items_bShowUserInfos.sql'),(54,'2018-10-24 06:11:11','1.0/revision-549/users_defaults.sql'),(55,'2018-10-24 06:11:32','1.1/revision-001/bugfix.sql'),(56,'2018-10-24 06:11:32','1.1/revision-002/platforms.sql'),(57,'2018-10-24 06:11:32','1.1/revision-003/drop_tmp.sql'),(58,'2018-10-24 06:11:33','1.1/revision-004/groups_stextid.sql'),(59,'2018-10-24 06:11:33','1.1/revision-004/unlocks.sql'),(60,'2018-10-24 06:11:33','1.1/revision-005/iScoreReeval.sql'),(61,'2018-10-24 06:11:33','1.1/revision-006/items_bReadOnly.sql'),(62,'2018-10-24 06:11:34','1.1/revision-007/fix_default_values.sql'),(63,'2018-10-24 06:11:34','1.1/revision-007/fix_history_bdeleted.sql'),(64,'2018-10-24 06:11:34','1.1/revision-007/fix_history_groups.sql'),(65,'2018-10-24 06:11:34','1.1/revision-007/fix_history_users.sql'),(66,'2018-10-24 06:11:34','1.1/revision-008/schema_revision.sql'),(67,'2018-10-24 06:11:34','1.1/revision-009/bFixedRanks.sql'),(68,'2018-10-24 06:11:34','1.1/revision-010/error_log.sql'),(69,'2018-10-24 06:11:35','1.1/revision-011/reducing_item_stype.sql'),(70,'2018-10-24 06:11:35','1.1/revision-012/fix_items_sDuration.sql'),(71,'2018-10-24 06:11:35','1.1/revision-012/users_items_sAnswer.sql'),(72,'2018-10-24 06:11:35','1.1/revision-013/items_idDefaultLanguage.sql'),(73,'2018-10-24 06:11:35','1.1/revision-014/default_values.sql'),(74,'2018-10-24 06:11:35','1.1/revision-014/default_values_bugfix.sql'),(75,'2018-10-24 06:11:35','1.1/revision-014/groups_fields.sql'),(76,'2018-10-24 06:11:35','1.1/revision-014/groups_login_prefixes.sql'),(77,'2018-10-24 06:11:36','1.1/revision-014/lm_prefix.sql'),(78,'2018-10-24 06:11:36','1.1/revision-015/groups_null_fileds.sql'),(79,'2018-10-24 06:11:36','1.1/revision-015/nulls.sql'),(80,'2018-10-24 06:11:36','1.1/revision-015/update_root_groups_textid.sql'),(81,'2018-10-24 06:11:36','1.1/revision-015/users.allowSubgroups.sql'),(82,'2018-10-24 06:11:36','1.1/revision-016/groupCodeEnter.sql'),(83,'2018-10-24 06:11:36','1.1/revision-016/items.sql'),(84,'2018-10-24 06:11:36','1.1/revision-016/text_varchar_bugfix.sql'),(85,'2018-10-24 06:11:37','1.1/revision-017/sHintsRequested.sql'),(86,'2018-10-24 06:11:37','1.1/revision-018/groups_add_teams.sql'),(87,'2018-10-24 06:11:37','1.1/revision-019/history_groups_add_teams.sql'),(88,'2018-10-24 06:11:37','1.1/revision-020/groups_add_teams.sql'),(89,'2018-10-24 06:11:37','1.1/revision-021/graduation_grade.sql'),(90,'2018-10-24 06:11:37','1.1/revision-021/groups_login_prefixes.sql'),(91,'2018-10-24 06:11:37','1.1/revision-021/user_creator_id.sql'),(92,'2018-10-24 06:11:38','1.1/revision-022/attempts.sql'),(93,'2018-10-24 06:11:38','1.1/revision-022/history_group_login_prefixes_index.sql'),(94,'2018-10-24 06:11:38','1.1/revision-023/indexes.sql'),(95,'2018-10-24 06:11:38','1.1/revision-024/indexes.sql'),(96,'2018-10-24 06:11:38','1.1/revision-024/remove_history_users_answers.sql'),(97,'2018-10-24 06:11:38','1.1/revision-025/bTeamsEditable.sql'),(98,'2018-10-24 06:11:38','1.1/revision-026/sBestAnswerDate.sql'),(99,'2018-10-24 06:11:38','1.1/revision-027/badges.sql'),(100,'2018-10-24 06:11:38','1.1/revision-028/lockUserDeletionDate.sql'),(101,'2018-10-24 06:11:39','1.1/revision-028/nulls.sql'),(102,'2018-10-24 06:11:39','1.1/revision-029/items_sRepositoryPath.sql'),(103,'2018-10-24 06:11:39','1.1/revision-029/platform_baseUrl.sql'),(104,'2018-10-24 06:11:39','1.1/revision-029/user_items_platform_data.sql'),(105,'2019-09-24 22:10:44','1.1/revision-030/items_strings_sImageUrl.sql'),(106,'2019-09-24 22:10:44','1.1/revision-031/items_strings_sImageUrl.sql'),(107,'2019-09-24 22:10:44','1.1/revision-031/users_items_sAdditionalTime.sql'),(108,'2019-09-24 22:10:44','1.1/revision-032/history_groups_attempts_key_ID.sql'),(110,'2019-09-24 22:10:44','1.1/revision-033/fields_fix.sql'),(111,'2019-09-24 22:10:44','1.1/revision-033/fields_fix_type.sql'),(112,'2019-09-24 22:10:44','1.1/revision-033/keys.sql'),(113,'2019-09-24 22:10:44','1.1/revision-033/remove_history_fields.sql');
ALTER TABLE `schema_revision` ENABLE KEYS;
UNLOCK TABLES;

DROP TABLE IF EXISTS `synchro_version`;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8mb4;
CREATE TABLE `synchro_version` (
    `id` tinyint(1) NOT NULL,
    `version` int(11) NOT NULL,
    `last_server_version` int(11) NOT NULL,
    `last_client_version` int(11) NOT NULL,
    PRIMARY KEY (`id`),
    KEY `version` (`version`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

LOCK TABLES `synchro_version` WRITE;
ALTER TABLE `synchro_version` DISABLE KEYS;
INSERT INTO `synchro_version` VALUES (0,0,0,0);
ALTER TABLE `synchro_version` ENABLE KEYS;
UNLOCK TABLES;
