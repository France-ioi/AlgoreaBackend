-- +migrate Up
CREATE TABLE `permissions_granted` (
    `group_id` BIGINT(20) NOT NULL,
    `item_id` BIGINT(20) NOT NULL,
    `source_group_id` BIGINT(20) NOT NULL,
    `origin` ENUM('group_membership','item_unlocking','self','other') NOT NULL,
    `latest_update_on` DATETIME NOT NULL DEFAULT NOW()
        COMMENT 'Last time one of the attributes has been modified',
    `can_view` ENUM('none','info','content','content_with_descendants','solution') NOT NULL DEFAULT 'none'
        COMMENT 'The level of visibility the group has on the item',
    `can_grant_view` ENUM('none','content','content_with_descendants','solution','transfer') NOT NULL DEFAULT 'none'
        COMMENT 'The level of visibility that the group can give on this item to other groups on which it has the right to',
    `can_watch` ENUM('none','result','answer','transfer') NOT NULL DEFAULT 'none'
        COMMENT 'The level of observation a group has for an item, on the activity of the users he can watch',
    `can_edit` ENUM('none','children','all','transfer') NOT NULL DEFAULT 'none'
        COMMENT 'The level of edition permissions a group has on an item',
    `is_owner` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'Whether the group is the owner of this item. Implies the maximum level in all of the above permissions. Can delete the item.',
    `can_view_value` TINYINT(3) UNSIGNED AS (`can_view` + 0) NOT NULL
        COMMENT 'can_view as an integer (to use comparison operators)',
    `can_grant_view_value` TINYINT(3) UNSIGNED AS (`can_grant_view` + 0) NOT NULL
        COMMENT 'can_grant_view as an integer (to use comparison operators)',
    `can_watch_value` TINYINT(3) UNSIGNED AS (`can_watch` + 0) NOT NULL
        COMMENT 'can_watch as an integer (to use comparison operators)',
    `can_edit_value` TINYINT(3) UNSIGNED AS (`can_edit` + 0) NOT NULL
        COMMENT 'can_edit as an integer (to use comparison operators)',
    PRIMARY KEY (`group_id`,`item_id`,`source_group_id`,`origin`),
    INDEX `group_id_item_id` (`group_id`, `item_id`),
    CONSTRAINT `fk_permissions_granted_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_permissions_granted_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items`(`id`) ON DELETE CASCADE
) COMMENT 'Raw permissions given to a group on an item' ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `permissions_generated` (
    `group_id` BIGINT(20) NOT NULL,
    `item_id` BIGINT(20) NOT NULL,
    `can_view_generated` ENUM('none','info','content','content_with_descendants','solution') NOT NULL DEFAULT 'none'
        COMMENT 'The aggregated level of visibility the group has on the item',
    `can_grant_view_generated` ENUM('none','content','content_with_descendants','solution','transfer') NOT NULL DEFAULT 'none'
        COMMENT 'The aggregated level of visibility that the group can give on this item to other groups on which it has the right to',
    `can_watch_generated` ENUM('none','result','answer','transfer') NOT NULL DEFAULT 'none'
        COMMENT 'The aggregated level of observation a group has for an item, on the activity of the users he can watch',
    `can_edit_generated` ENUM('none','children','all','transfer') NOT NULL DEFAULT 'none'
        COMMENT 'The aggregated level of edition permissions a group has on an item',
    `is_owner_generated` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'Whether the group is the owner of this item. Implies the maximum level in all of the above permissions. Can delete the item.',
    `can_view_generated_value` TINYINT(3) UNSIGNED AS (`can_view_generated` + 0) NOT NULL
        COMMENT 'can_view_generated as an integer (to use comparison operators)',
    `can_grant_view_generated_value` TINYINT(3) UNSIGNED AS (`can_grant_view_generated` + 0) NOT NULL
        COMMENT 'can_grant_view_generated as an integer (to use comparison operators)',
    `can_watch_generated_value` TINYINT(3) UNSIGNED AS (`can_watch_generated` + 0) NOT NULL
        COMMENT 'can_watch_generated as an integer (to use comparison operators)',
    `can_edit_generated_value` TINYINT(3) UNSIGNED AS (`can_edit_generated` + 0) NOT NULL
        COMMENT 'can_edit_generated as an integer (to use comparison operators)',
    PRIMARY KEY (`group_id`,`item_id`),
    INDEX `item_id` (`item_id`),
    CONSTRAINT `fk_permissions_generated_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_permissions_generated_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items`(`id`) ON DELETE CASCADE
) COMMENT 'Actual permissions that the group has, considering the aggregation and the propagation' ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `permissions_propagate` (
     `group_id` BIGINT(20) NOT NULL,
     `item_id` BIGINT(20) NOT NULL,
     `propagate_to` enum('self', 'children') NOT NULL
         COMMENT 'Which permissions should be recomputed for the group-item pair on the next iteration, either for the pair or for its children (through item hierarchy)',
     PRIMARY KEY (`group_id`,`item_id`),
     INDEX `propagate_to` (`propagate_to`),
     CONSTRAINT `fk_permissions_propagate_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
     CONSTRAINT `fk_permissions_propagate_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items`(`id`) ON DELETE CASCADE
) COMMENT 'Used by the access rights propagation algorithm to keep track of the status of the propagation' ENGINE=InnoDB DEFAULT CHARSET=utf8;

DELETE `items_items` FROM `items_items` LEFT JOIN `items` ON `items`.`id` = `items_items`.`parent_item_id` WHERE `items`.`id` IS NULL;
DELETE `items_items` FROM `items_items` LEFT JOIN `items` ON `items`.`id` = `items_items`.`child_item_id` WHERE `items`.`id` IS NULL;

-- +migrate StatementBegin
CREATE TRIGGER `after_insert_permissions_granted` AFTER INSERT ON `permissions_granted` FOR EACH ROW BEGIN
    INSERT INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    VALUE (NEW.`group_id`, NEW.`item_id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_to` = 'self';
END
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_granted` AFTER UPDATE ON `permissions_granted` FOR EACH ROW BEGIN
    IF NOT (NEW.`can_view` <=> OLD.`can_view` AND NEW.`can_grant_view` <=> OLD.`can_grant_view` AND
            NEW.`can_watch` <=> OLD.`can_watch` AND NEW.`can_edit` <=> OLD.`can_edit` AND
            NEW.`is_owner` <=> OLD.`is_owner`) THEN
        INSERT INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`) VALUE (NEW.`group_id`, NEW.`item_id`, 'self')
        ON DUPLICATE KEY UPDATE `propagate_to` = 'self';
    END IF;
END
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TRIGGER `after_delete_permissions_granted` AFTER DELETE ON `permissions_granted` FOR EACH ROW BEGIN
    INSERT INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`) VALUE (OLD.`group_id`, OLD.`item_id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_to` = 'self';
END
-- +migrate StatementEnd

ALTER TABLE `items_items`
    ADD COLUMN `content_view_propagation` ENUM('none', 'as_info', 'as_content') NOT NULL DEFAULT 'none'
        COMMENT 'Defines how a can_view=”content” permission propagates' AFTER `partial_access_propagation`,
    ADD COLUMN `upper_view_levels_propagation` ENUM('use_content_view_propagation', 'as_content_with_descendants', 'as_is')
        NOT NULL DEFAULT 'use_content_view_propagation'
        COMMENT 'Defines how can_view="content_with_descendants"|"solution" permissions propagate',
    ADD COLUMN `grant_view_propagation` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'Whether can_grant_view propagates (as the same value, with “solution” as the upper limit)',
    ADD COLUMN `watch_propagation` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'Whether can_watch propagates (as the same value, with “answer” as the upper limit)',
    ADD COLUMN `edit_propagation` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'Whether can_edit propagates (as the same value, with “all” as the upper limit)',
    ADD COLUMN `content_view_propagation_value` TINYINT(3) UNSIGNED AS (`content_view_propagation`) NOT NULL
        COMMENT 'content_view_propagation as an integer (to use comparison operators)',
    ADD COLUMN `upper_view_levels_propagation_value` TINYINT(3) UNSIGNED
        AS (`upper_view_levels_propagation`) NOT NULL
        COMMENT 'upper_view_levels_propagation as an integer (to use comparison operators)';

DROP TRIGGER IF EXISTS `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
    FROM `permissions_generated`
    WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id`;
END
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS `after_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_items_items` AFTER UPDATE ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
    FROM `permissions_generated`
    WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id` OR `permissions_generated`.`item_id` = OLD.`parent_item_id`;
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

    INSERT IGNORE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
        SELECT `permissions_generated`.`group_id`, `permissions_generated`.`item_id`, 'children' as `propagate_to`
        FROM `permissions_generated`
        WHERE `permissions_generated`.`item_id` = OLD.`parent_item_id`;
END
-- +migrate StatementEnd

DELETE `groups_items` FROM `groups_items` LEFT JOIN `groups` ON `groups`.`id` = `groups_items`.`group_id` WHERE `groups`.`id` IS NULL;
DELETE `groups_items` FROM `groups_items` LEFT JOIN `items` ON `items`.`id` = `groups_items`.`item_id` WHERE `items`.`id` IS NULL;

INSERT INTO `permissions_granted` (group_id, item_id, source_group_id, latest_update_on, can_view, is_owner)
SELECT `groups_items`.`group_id`,
       `groups_items`.`item_id`,
       IFNULL(`groups_items`.`creator_id`, IF(`groups_items`.`creator_user_id` = 0, -1, -3)) AS `source_group_id`,
       IFNULL(NULLIF(GREATEST(
                             IFNULL(`groups_items`.`partial_access_since`, '1000-01-01 00:00:00'),
                             IFNULL(`groups_items`.`full_access_since`, '1000-01-01 00:00:00'),
                             IFNULL(`groups_items`.`solutions_access_since`, '1000-01-01 00:00:00')
                         ), '1000-01-01 00:00:00'), NOW()) AS `latest_update_on`,
       CASE
           WHEN `groups_items`.`solutions_access_since` IS NOT NULL THEN 'solution'
           WHEN `groups_items`.`full_access_since` IS NOT NULL THEN 'content_with_descendants'
           WHEN `groups_items`.`partial_access_since` IS NOT NULL THEN 'content'
           ELSE 'none'
       END AS 'can_view',
       `groups_items`.`owner_access` AS `is_owner`
FROM `groups_items`
WHERE `groups_items`.`partial_access_since` IS NOT NULL OR
      `groups_items`.`full_access_since` IS NOT NULL OR
      `groups_items`.`solutions_access_since` IS NOT NULL OR
      `groups_items`.`owner_access`;

INSERT INTO `permissions_granted` (group_id, item_id, source_group_id, latest_update_on, can_view, can_grant_view, can_watch, can_edit, is_owner)
SELECT `groups_items`.`group_id`,
       `groups_items`.`item_id`,
       IFNULL(`groups_items`.`creator_id`, IF(`groups_items`.`creator_user_id` = 0, -1, -3)) AS `source_group_id`,
       NOW() AS `latest_update_on`,
       'solution' AS `can_view`,
       'solution' AS `can_grant_view`,
       'answer' AS `can_watch`,
       'all' AS `can_edit`,
       0 AS `is_owner`
FROM `groups_items`
WHERE `groups_items`.`manager_access`
ON DUPLICATE KEY UPDATE
    `permissions_granted`.`latest_update_on` = LEAST(`permissions_granted`.`latest_update_on`, NOW()),
    `permissions_granted`.`can_view` = 'solution',
    `permissions_granted`.`can_grant_view` = 'solution',
    `permissions_granted`.`can_watch` = 'answer',
    `permissions_granted`.`can_edit` = 'all';

UPDATE `items_items` SET `content_view_propagation` = `partial_access_propagation`+0,
                         `upper_view_levels_propagation` = 'as_is',
                         `grant_view_propagation` = 1,
                         `watch_propagation` = 1,
                         `edit_propagation` = 1;

ALTER TABLE `items_items` DROP COLUMN `partial_access_propagation`;

DROP TRIGGER `before_insert_groups_items`;
DROP TRIGGER `after_insert_groups_items`;
DROP TRIGGER `after_update_groups_items`;
DROP TRIGGER `after_delete_groups_items`;

DROP TABLE `groups_items`;
DROP TABLE `groups_items_propagate`;

-- +migrate Down
DROP TRIGGER `after_insert_permissions_granted`;
DROP TRIGGER `after_update_permissions_granted`;

CREATE TABLE `groups_items` (
  `id` bigint(20) NOT NULL,
  `group_id` bigint(20) NOT NULL,
  `item_id` bigint(20) NOT NULL,
  `creator_user_id` bigint(20) NOT NULL COMMENT 'User who created the entry',
  `creator_id` bigint(20) DEFAULT NULL COMMENT 'User who created the entry',
  `partial_access_since` datetime DEFAULT NULL COMMENT 'At what date the group obtains partial access to the item',
  `access_reason` varchar(200) DEFAULT NULL COMMENT 'Manual comment about why the current access was given',
  `full_access_since` datetime DEFAULT NULL COMMENT 'At what date the group obtains full access to the item',
  `solutions_access_since` datetime DEFAULT NULL COMMENT 'At what date the group obtains solution access to the item',
  `owner_access` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the group has owner access to this item',
  `manager_access` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the group has manager access to this item (not inherited)',
  `cached_full_access_since` datetime DEFAULT NULL COMMENT 'At what date the user obtains full access, taking into account access inherited from the ancestors',
  `cached_partial_access_since` datetime DEFAULT NULL COMMENT 'At what date the user obtains partial access, taking into account access inherited from the ancestors',
  `cached_solutions_access_since` datetime DEFAULT NULL COMMENT 'At what date the user obtains solution access, taking into account access inherited from the ancestors',
  `cached_grayed_access_since` datetime DEFAULT NULL COMMENT 'At what date the user obtains grayed access (can see the title but not the content), taking into account access inherited from the ancestors',
  `cached_access_reason` varchar(200) DEFAULT NULL COMMENT 'How was the current access obtained (generated from reasons given manually to ancestor items)',
  `cached_full_access` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the group currently has full access to this item',
  `cached_partial_access` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the group currently has partial access to this item',
  `cached_access_solutions` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the group currently has solution access to this item',
  `cached_grayed_access` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the group currently has grayed access to this item',
  `cached_manager_access` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the group currently has manager access to this item',
  PRIMARY KEY (`id`),
  UNIQUE KEY `item_id` (`item_id`,`group_id`),
  KEY `group_id` (`group_id`) COMMENT 'idGroup',
  KEY `itemtem_id` (`item_id`),
  KEY `full_access` (`cached_full_access`,`cached_full_access_since`),
  KEY `access_solutions` (`cached_access_solutions`,`cached_solutions_access_since`),
  KEY `partial_access` (`cached_partial_access`,`cached_partial_access_since`),
  CONSTRAINT `fk_groups_items_creator_id_users_group_id` FOREIGN KEY (`creator_id`) REFERENCES `users`(`group_id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Access given to a group on a specific item, both directly and inherited from ancestor items.';

CREATE TABLE `groups_items_propagate` (
  `id` bigint(20) NOT NULL,
  `propagate_access` enum('self','children') NOT NULL,
  PRIMARY KEY (`id`),
  KEY `propagate_access` (`propagate_access`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Used by the access rights propagation algorithm to keep track of the status of the propagation.';

-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_items` BEFORE INSERT ON `groups_items` FOR EACH ROW BEGIN
    IF (NEW.id IS NULL OR NEW.id = 0) THEN
        SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;
    END IF;
END
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN
    INSERT INTO `groups_items_propagate` (`id`, `propagate_access`) VALUE (NEW.`id`, 'self')
    ON DUPLICATE KEY UPDATE `propagate_access` = 'self';
END
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TRIGGER `after_update_groups_items` AFTER UPDATE ON `groups_items` FOR EACH ROW BEGIN
    # As a date change may result in access change for descendants of the item, mark the entry as to be recomputed
    IF NOT (NEW.`full_access_since` <=> OLD.`full_access_since`AND NEW.`partial_access_since` <=> OLD.`partial_access_since`AND
            NEW.`solutions_access_since` <=> OLD.`solutions_access_since`AND NEW.`manager_access` <=> OLD.`manager_access`AND
            NEW.`access_reason` <=> OLD.`access_reason`) THEN
        INSERT INTO `groups_items_propagate` (`id`, `propagate_access`) VALUE (NEW.`id`, 'self')
        ON DUPLICATE KEY UPDATE `propagate_access` = 'self';
    END IF;
END
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TRIGGER `after_delete_groups_items` AFTER DELETE ON `groups_items` FOR EACH ROW BEGIN DELETE FROM groups_items_propagate where id = OLD.id ; END
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `groups_items_propagate`
    SELECT `id`, 'children' as `propagate_access` FROM `groups_items`
    WHERE `groups_items`.`item_id` = NEW.`parent_item_id`;
END
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS `after_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_items_items` AFTER UPDATE ON `items_items` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `groups_items_propagate`
        SELECT `id`, 'children' as `propagate_access`
        FROM `groups_items`
        WHERE `groups_items`.`item_id` = NEW.`parent_item_id` OR `groups_items`.`item_id` = OLD.`parent_item_id`;
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

ALTER TABLE `items_items`
    ADD COLUMN `partial_access_propagation` enum('None','AsGrayed','AsPartial') NOT NULL DEFAULT 'None'
        COMMENT 'Specifies how to propagate partial access to the child item' AFTER `category`;
UPDATE `items_items` SET `partial_access_propagation` = `content_view_propagation` + 0;

INSERT INTO `groups_items` (
    `group_id`, `item_id`, `creator_id`, `creator_user_id`, `solutions_access_since`,
    `full_access_since`, `partial_access_since`, `owner_access`, `manager_access`)
SELECT `permissions_granted`.`group_id`,
       `permissions_granted`.`item_id`,
       IF(`permissions_granted`.`source_group_id` <= 0, NULL, `permissions_granted`.`source_group_id`) AS `creator_id`,
       IF(`permissions_granted`.`source_group_id` < 0, 0, `permissions_granted`.`source_group_id`) AS `creator_user_id`,
       IF(`permissions_granted`.`can_view` = 'solution',
           `permissions_granted`.`latest_update_on`, NULL) AS `solutions_access_since`,
       IF(`permissions_granted`.`can_view` IN ('solution', 'content_with_descendants'),
          `permissions_granted`.`latest_update_on`, NULL) AS `full_access_since`,
       IF(`permissions_granted`.`can_view` IN ('solution', 'content_with_descendants', 'content'),
          `permissions_granted`.`latest_update_on`, NULL) AS `partial_access_since`,
       `permissions_granted`.`is_owner` AS `owner_access`,
       `permissions_granted`.`can_edit` = 'all' AS `manager_access`
FROM `permissions_granted`
ON DUPLICATE KEY UPDATE
    `groups_items`.`creator_id` = IFNULL(`groups_items`.`creator_id`, VALUES(`creator_id`)),
    `groups_items`.`creator_user_id` = IFNULL(`groups_items`.`creator_user_id`, VALUES(`creator_user_id`)),
    `groups_items`.`solutions_access_since` = IFNULL(`groups_items`.`solutions_access_since`, VALUES(`solutions_access_since`)),
    `groups_items`.`full_access_since` = IFNULL(`groups_items`.`full_access_since`, VALUES(`full_access_since`)),
    `groups_items`.`partial_access_since` = IFNULL(`groups_items`.`partial_access_since`, VALUES(`partial_access_since`)),
    `groups_items`.`owner_access` = GREATEST(`groups_items`.`owner_access`, VALUES(`owner_access`)),
    `groups_items`.`manager_access` = GREATEST(`groups_items`.`manager_access`, VALUES(`manager_access`));

DROP TABLE `permissions_generated`;
DROP TABLE `permissions_granted`;
DROP TABLE `permissions_propagate`;

ALTER TABLE `items_items`
    DROP COLUMN `content_view_propagation`,
    DROP COLUMN `upper_view_levels_propagation`,
    DROP COLUMN `grant_view_propagation`,
    DROP COLUMN `watch_propagation`,
    DROP COLUMN `edit_propagation`,
    DROP COLUMN `content_view_propagation_value`,
    DROP COLUMN `upper_view_levels_propagation_value`;
