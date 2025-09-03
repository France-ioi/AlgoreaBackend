-- +migrate Up
ALTER TABLE `results_propagate` MODIFY `state` ENUM('to_be_propagated','to_be_recomputed','propagating', 'recomputing') NOT NULL COMMENT '"to_be_propagated" means that ancestors should be recomputed';

-- +migrate Down
ALTER TABLE `results_propagate` MODIFY `state` ENUM('to_be_propagated','to_be_recomputed','propagating') NOT NULL COMMENT '"to_be_propagated" means that ancestors should be recomputed';
