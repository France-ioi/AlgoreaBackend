-- +migrate Up
# 13 rows
DELETE `items_items` FROM `items_items`
  JOIN `items_items` AS `orig` USING(`parent_item_id`, `child_item_id`)
WHERE `items_items`.`child_order` > `orig`.`child_order` OR
      `items_items`.`child_order` = `orig`.`child_order` AND `items_items`.`id` > `orig`.`id`;

ALTER TABLE `items_items`
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`parent_item_id`, `child_item_id`),
    DROP INDEX `parent_child`,
    DROP INDEX `parent_version`,
    DROP COLUMN `id`,
    ADD CONSTRAINT `fk_items_items_parent_item_id_items_id` FOREIGN KEY (`parent_item_id`) REFERENCES `items`(`id`),
    ADD CONSTRAINT `fk_items_items_child_item_id_items_id` FOREIGN KEY (`child_item_id`) REFERENCES `items`(`id`);

DROP TRIGGER `before_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_items` BEFORE INSERT ON `items_items` FOR EACH ROW BEGIN INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state) VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd

-- +migrate Down
ALTER TABLE `items_items`
    DROP PRIMARY KEY,
    ADD COLUMN `id` BIGINT(20) FIRST;

UPDATE `items_items` SET `id` = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;

ALTER TABLE `items_items`
    MODIFY COLUMN `id` BIGINT(20) NOT NULL,
    ADD PRIMARY KEY (`id`),
    DROP FOREIGN KEY `fk_items_items_parent_item_id_items_id`,
    DROP FOREIGN KEY `fk_items_items_child_item_id_items_id`,
    ADD INDEX `parent_child` (`parent_item_id`, `child_item_id`),
    ADD INDEX `parent_version` (`parent_item_id`);

DROP TRIGGER `before_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_items` BEFORE INSERT ON `items_items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state) VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd
