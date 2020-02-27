-- +migrate Up
ALTER TABLE `groups_contest_items`
    ADD CONSTRAINT `fk_groups_contest_items_group_id_groups_id`
        FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    ADD CONSTRAINT `fk_groups_contest_items_item_id_items_id`
        FOREIGN KEY (`item_id`) REFERENCES `items`(`id`) ON DELETE CASCADE;

DELETE `groups_propagate` FROM `groups_propagate`
    LEFT JOIN `groups` ON `groups`.`id` = `groups_propagate`.`id`
WHERE `groups`.`id` IS NULL;

ALTER TABLE `groups_propagate`
    ADD CONSTRAINT `fk_groups_propagate_id_groups_id`
        FOREIGN KEY (`id`) REFERENCES `groups`(`id`) ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE `groups_contest_items`
    DROP FOREIGN KEY `fk_groups_contest_items_group_id_groups_id`,
    DROP FOREIGN KEY `fk_groups_contest_items_item_id_items_id`;

ALTER TABLE `groups_propagate`
    DROP FOREIGN KEY `fk_groups_propagate_id_groups_id`;
