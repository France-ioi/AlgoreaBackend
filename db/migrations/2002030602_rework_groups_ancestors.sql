-- +migrate Up
DELETE `groups_ancestors` FROM `groups_ancestors`
    LEFT JOIN `groups` ON `groups`.`id` = `groups_ancestors`.`ancestor_group_id`
WHERE `groups`.`id` IS NULL;

DELETE `groups_ancestors` FROM `groups_ancestors`
    LEFT JOIN `groups` ON `groups`.`id` = `groups_ancestors`.`child_group_id`
WHERE `groups`.`id` IS NULL;

ALTER TABLE `groups_ancestors`
    DROP PRIMARY KEY,
    DROP COLUMN `id`,
    DROP KEY `ancestor`,
    DROP COLUMN `is_self`,
    ADD COLUMN `is_self` TINYINT(1) GENERATED ALWAYS AS (`ancestor_group_id` = `child_group_id`) VIRTUAL
        COMMENT 'Whether ancestor_group_id = child_group_id (auto-generated)' AFTER `child_group_id`,
    ADD CONSTRAINT `fk_groups_ancestors_ancestor_group_id_groups_id`
        FOREIGN KEY (`ancestor_group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    ADD CONSTRAINT `fk_groups_ancestors_child_group_id_groups_id`
        FOREIGN KEY (`child_group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    DROP KEY `ancestor_group_id`,
    ADD PRIMARY KEY (`ancestor_group_id`, `child_group_id`);

DROP TRIGGER IF EXISTS `before_insert_groups_ancestors`;

DROP VIEW IF EXISTS groups_ancestors_active;
CREATE VIEW groups_ancestors_active AS SELECT * FROM groups_ancestors WHERE NOW() < expires_at;

-- +migrate Down
ALTER TABLE `groups_ancestors`
    ADD COLUMN `id` BIGINT(20) FIRST,
    DROP COLUMN `is_self`,
    ADD COLUMN `is_self` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether ancestor_group_id = child_group_id.' AFTER `child_group_id`;

UPDATE `groups_ancestors`
SET `id` = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000,
    `is_self` = (`ancestor_group_id` = `child_group_id`);

ALTER TABLE `groups_ancestors`
    ADD UNIQUE KEY `ancestor_group_id` (`ancestor_group_id`,`child_group_id`),
    ADD KEY `ancestor` (`ancestor_group_id`),
    DROP FOREIGN KEY `fk_groups_ancestors_ancestor_group_id_groups_id`,
    DROP FOREIGN KEY `fk_groups_ancestors_child_group_id_groups_id`,
    DROP PRIMARY KEY,
    MODIFY COLUMN `id` BIGINT(20) NOT NULL AUTO_INCREMENT PRIMARY KEY;

-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_ancestors` BEFORE INSERT ON `groups_ancestors` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd

DROP VIEW IF EXISTS groups_ancestors_active;
CREATE VIEW groups_ancestors_active AS SELECT * FROM groups_ancestors WHERE NOW() < expires_at;
