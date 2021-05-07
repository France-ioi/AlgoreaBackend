-- +migrate Up
ALTER TABLE `items_ancestors`
    DROP FOREIGN KEY `fk_items_ancestors_child_item_id_items_id`;
ALTER TABLE `items_ancestors`
    ADD CONSTRAINT `fk_items_ancestors_child_item_id_items_id` FOREIGN KEY (`child_item_id`) REFERENCES `items` (`id`)
        ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE `items_ancestors`
    DROP FOREIGN KEY `fk_items_ancestors_child_item_id_items_id`;
ALTER TABLE `items_ancestors`
    ADD CONSTRAINT `fk_items_ancestors_child_item_id_items_id` FOREIGN KEY (`child_item_id`) REFERENCES `items` (`id`);
