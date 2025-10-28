-- +goose Up
CREATE TABLE `results_propagate_internal` (
  `participant_id` bigint NOT NULL,
  `attempt_id` bigint NOT NULL DEFAULT '0',
  `item_id` bigint NOT NULL,
  `state` enum('to_be_propagated','to_be_recomputed','propagating','recomputing') NOT NULL COMMENT '"to_be_propagated" means that ancestors should be recomputed',
  PRIMARY KEY (`participant_id`,`attempt_id`,`item_id`),
  KEY `state` (`state`),
  CONSTRAINT `fk_results_propagate_internal_to_results` FOREIGN KEY (`participant_id`, `attempt_id`, `item_id`) REFERENCES `results` (`participant_id`, `attempt_id`, `item_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
  COMMENT='Internally used by the algorithm that computes results for items that have children and unlocks items if needed. Do not insert into this table directly, use results_propagate instead.';

-- +goose Down
INSERT INTO `results_propagate` (participant_id, attempt_id, item_id, state)
  SELECT participant_id, attempt_id, item_id, state
  FROM `results_propagate_internal`
ON DUPLICATE KEY UPDATE results_propagate.state = IF(
  IF(VALUES(state) = 'propagating', 'to_be_propagated', VALUES(state)) = 'to_be_recomputed',
  'to_be_recomputed',
  IF(results_propagate.state = 'propagating', 'to_be_propagated', results_propagate.state)
);
DROP TABLE `results_propagate_internal`;
