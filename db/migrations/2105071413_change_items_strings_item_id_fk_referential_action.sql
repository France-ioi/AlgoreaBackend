-- +migrate Up
ALTER TABLE `items_strings`
    DROP FOREIGN KEY `fk_items_strings_item_id_items_id`;
ALTER TABLE `items_strings`
    ADD CONSTRAINT `fk_items_strings_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`)
        ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE `items_strings`
    DROP FOREIGN KEY `fk_items_strings_item_id_items_id`;
ALTER TABLE `items_strings`
    ADD CONSTRAINT `fk_items_strings_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`);
