-- +goose Up
ALTER TABLE `results_propagate`
  MODIFY COLUMN `state` enum('to_be_propagated','to_be_recomputed') NOT NULL
    COMMENT '"to_be_propagated" means that ancestors should be recomputed';

-- +goose Down
ALTER TABLE `results_propagate`
  MODIFY COLUMN `state` enum('to_be_propagated','to_be_recomputed','propagating','recomputing') NOT NULL
    COMMENT '"to_be_propagated" means that ancestors should be recomputed';
