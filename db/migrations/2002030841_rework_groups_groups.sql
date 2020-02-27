-- +migrate Up
ALTER TABLE `groups_groups`
    DROP PRIMARY KEY,
    DROP COLUMN `id`,
    DROP KEY `parent_group_id`,
    DROP KEY `child_group_id`,
    ADD CONSTRAINT `fk_groups_groups_parent_group_id_groups_id`
        FOREIGN KEY (`parent_group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    ADD CONSTRAINT `fk_groups_groups_child_group_id_groups_id`
        FOREIGN KEY (`child_group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    DROP KEY `parentchild`,
    ADD PRIMARY KEY (`parent_group_id`, `child_group_id`);

DROP TRIGGER IF EXISTS `before_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd

DROP VIEW IF EXISTS groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

-- +migrate Down
ALTER TABLE `groups_groups`
    ADD COLUMN `id` BIGINT(20) FIRST;

UPDATE `groups_groups`
SET `id` = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;

ALTER TABLE `groups_groups`
    ADD UNIQUE KEY `parentchild` (`parent_group_id`,`child_group_id`),
    ADD KEY `parent_group_id` (`parent_group_id`),
    ADD KEY `child_group_id` (`child_group_id`),
    DROP FOREIGN KEY `fk_groups_groups_parent_group_id_groups_id`,
    DROP FOREIGN KEY `fk_groups_groups_child_group_id_groups_id`,
    DROP PRIMARY KEY,
    MODIFY COLUMN `id` BIGINT(20) NOT NULL AUTO_INCREMENT PRIMARY KEY;

DROP TRIGGER IF EXISTS `before_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd

DROP VIEW IF EXISTS groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;
