-- +migrate Up
ALTER TABLE `permissions_propagate`
  ADD INDEX `propagate_to_group_id_item_id`
    (`propagate_to`,`group_id`,`item_id`)
;

ALTER TABLE `permissions_propagate`
  DROP INDEX `propagate_to`;

-- +migrate Down
ALTER TABLE `permissions_propagate`
  DROP INDEX `propagate_to_group_id_item_id`;

ALTER TABLE `permissions_propagate`
  ADD INDEX `propagate_to`
    (`propagate_to`)
;
