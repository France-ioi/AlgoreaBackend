-- +migrate Up
RENAME TABLE `results_propagate_items` TO `results_recompute_for_items`;

-- +migrate Down
RENAME TABLE `results_recompute_for_items` TO `results_propagate_items`;
