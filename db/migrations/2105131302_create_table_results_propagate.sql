-- +migrate Up
CREATE TABLE `results_propagate` (
   `participant_id` bigint NOT NULL,
   `attempt_id` bigint NOT NULL DEFAULT '0',
   `item_id` bigint NOT NULL,
   `state` enum('to_be_propagated','to_be_recomputed') NOT NULL COMMENT '"to_be_propagated" means that ancestors should be recomputed',
   PRIMARY KEY (`participant_id`,`attempt_id`,`item_id`),
   KEY `state` (`state`),
   CONSTRAINT `fk_results_propagate_to_results` FOREIGN KEY (`participant_id`, `attempt_id`, `item_id`) REFERENCES `results` (`participant_id`, `attempt_id`, `item_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Used by the algorithm that computes results for items that have children and unlocks items if needed.';

-- +migrate Down
DROP TABLE `results_propagate`;
