-- +migrate Up
DELETE `items_propagate` FROM `items_propagate` LEFT JOIN `items` USING (`id`) WHERE `items`.`id` IS NULL;

ALTER TABLE `items_propagate`
    ADD CONSTRAINT `fk_id` FOREIGN KEY (`id`) REFERENCES `items`(`id`) ON DELETE CASCADE;

DROP TRIGGER `after_delete_items`;

-- +migrate Down
ALTER TABLE `items_propagate` DROP FOREIGN KEY `fk_id`;

-- +migrate StatementBegin
CREATE TRIGGER `after_delete_items` AFTER DELETE ON `items` FOR EACH ROW BEGIN DELETE FROM items_propagate where id = OLD.id ; END
-- +migrate StatementEnd
