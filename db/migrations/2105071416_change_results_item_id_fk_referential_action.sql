-- +migrate Up
ALTER TABLE `results`
    DROP FOREIGN KEY `fk_results_item_id_items_id`;
ALTER TABLE `results`
    ADD CONSTRAINT `fk_results_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`)
        ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE `results`
    DROP FOREIGN KEY `fk_results_item_id_items_id`;
ALTER TABLE `results`
    ADD CONSTRAINT `fk_results_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`);
