-- +migrate Up
ALTER TABLE `results_propagate_items`
  ADD COLUMN `is_being_processed` TINYINT(1) NOT NULL DEFAULT 0 AFTER `item_id`,
  ADD INDEX `is_being_processed` (`is_being_processed`);

-- +migrate Down
ALTER TABLE `results_propagate_items`
  DROP INDEX `is_being_processed`,
  DROP COLUMN `is_being_processed`;
