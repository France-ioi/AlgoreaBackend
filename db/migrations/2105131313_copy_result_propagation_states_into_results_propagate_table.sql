-- +migrate Up
INSERT INTO `results_propagate`
SELECT `participant_id`, `attempt_id`, `item_id`, `result_propagation_state` AS `state`
FROM `results`
WHERE `result_propagation_state` = 'to_be_propagated' OR `result_propagation_state` = 'to_be_recomputed';

-- +migrate Down
INSERT INTO `results` (`participant_id`, `attempt_id`, `item_id`, `result_propagation_state`)
SELECT `participant_id`, `attempt_id`, `item_id`, `state` AS `result_propagation_state`
FROM `results_propagate`
ON DUPLICATE KEY UPDATE `result_propagation_state` = VALUES(`result_propagation_state`);
