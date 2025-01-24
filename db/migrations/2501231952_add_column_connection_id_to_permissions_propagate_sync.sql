-- +migrate Up
ALTER TABLE `permissions_propagate_sync`
  ADD COLUMN `connection_id` BIGINT UNSIGNED NOT NULL FIRST,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (`connection_id`, `group_id`, `item_id`),
  DROP INDEX `propagate_to_group_id_item_id`,
  ADD INDEX `connection_id_propagate_to_group_id_item_id` (`connection_id`, `propagate_to`, `group_id`, `item_id`);

-- +migrate Down
ALTER TABLE `permissions_propagate_sync`
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (`group_id`, `item_id`),
  DROP INDEX `connection_id_propagate_to_group_id_item_id`,
  ADD INDEX `propagate_to_group_id_item_id` (`propagate_to`, `group_id`, `item_id`),
  DROP COLUMN `connection_id`;
