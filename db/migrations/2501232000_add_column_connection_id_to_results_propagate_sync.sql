-- +migrate Up
ALTER TABLE `results_propagate_sync`
  ADD COLUMN `connection_id` BIGINT UNSIGNED NOT NULL FIRST,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (`connection_id`,`participant_id`,`attempt_id`,`item_id`),
  DROP INDEX `state`,
  ADD INDEX `connection_id_state` (`connection_id`, `state`);

-- +migrate Down
ALTER TABLE `results_propagate_sync`
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (`participant_id`,`attempt_id`,`item_id`),
  DROP INDEX `connection_id_state`,
  ADD INDEX `state` (`state`),
  DROP COLUMN `connection_id`;
