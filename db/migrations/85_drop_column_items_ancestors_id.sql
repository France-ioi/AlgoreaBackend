-- +migrate Up
ALTER TABLE `items_ancestors`
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`ancestor_item_id`, `child_item_id`),
    DROP INDEX `ancestor_item_id_child_item_id`,
    DROP COLUMN `id`,
    ADD CONSTRAINT `fk_items_ancestors_ancestor_item_id_items_id` FOREIGN KEY (`ancestor_item_id`) REFERENCES `items`(`id`),
    ADD CONSTRAINT `fk_items_ancestors_child_item_id_items_id` FOREIGN KEY (`child_item_id`) REFERENCES `items`(`id`);

DROP TRIGGER `before_insert_items_ancestors`;

-- +migrate Down
ALTER TABLE `items_ancestors`
    DROP PRIMARY KEY,
    ADD COLUMN `id` BIGINT(20) FIRST;

UPDATE `items_ancestors` SET `id` = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;

ALTER TABLE `items_ancestors`
    MODIFY COLUMN `id` BIGINT(20) NOT NULL AUTO_INCREMENT,
    ADD PRIMARY KEY (`id`),
    ADD UNIQUE KEY `ancestor_item_id_child_item_id` (`ancestor_item_id`,`child_item_id`),
    DROP FOREIGN KEY `fk_items_ancestors_ancestor_item_id_items_id`,
    DROP FOREIGN KEY `fk_items_ancestors_child_item_id_items_id`;

-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_ancestors` BEFORE INSERT ON `items_ancestors` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
