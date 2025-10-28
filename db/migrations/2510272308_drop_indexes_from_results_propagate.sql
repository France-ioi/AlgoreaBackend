-- +goose Up
ALTER TABLE `results_propagate`
  DROP INDEX `state`,
  DROP PRIMARY KEY,
  DROP CONSTRAINT `fk_results_propagate_to_results`;

-- +goose Down
ALTER TABLE `results_propagate`
  ADD PRIMARY KEY (`participant_id`, `attempt_id`, `item_id`),
  ADD INDEX `state` (`state`),
  ADD CONSTRAINT `fk_results_propagate_to_results`
    FOREIGN KEY (`participant_id`, `attempt_id`, `item_id`)
      REFERENCES `results` (`participant_id`, `attempt_id`, `item_id`) ON DELETE CASCADE;
