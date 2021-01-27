-- +migrate Up
ALTER TABLE `results`
    MODIFY COLUMN `result_propagation_state` ENUM('done','to_be_propagated','to_be_recomputed') NOT NULL
        DEFAULT 'done'
        COMMENT 'Used by the algorithm that computes results for items that have children and unlocks items if needed ("to_be_propagated" means that ancestors should be recomputed).';

-- +migrate Down
ALTER TABLE `results`
    MODIFY COLUMN `result_propagation_state` ENUM('done','processing','to_be_propagated','to_be_recomputed') NOT NULL
        DEFAULT 'done'
        COMMENT 'Used by the algorithm that computes results for items that have children and unlocks items if needed ("to_be_propagated" means that ancestors should be recomputed).';
