-- +migrate Up
CREATE TABLE `results_propagate_items` (
  `item_id` BIGINT(19) NOT NULL,
  PRIMARY KEY (`item_id`),
  CONSTRAINT `fk_results_propagate_items_to_items` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE
)
  COMMENT='Used by the algorithm that computes results. All results for the item_id have to be recomputed when the item_id is in this table.'
  COLLATE='utf8_general_ci'
  ENGINE=InnoDB
;

-- +migrate Down
DROP TABLE `results_propagate_items`;
