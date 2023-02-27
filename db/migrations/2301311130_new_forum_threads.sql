-- +migrate Up
CREATE TABLE `threads` (
  `participant_id` BIGINT NOT NULL,
  CONSTRAINT `fk_threads_participant_id_groups_id` FOREIGN KEY (`participant_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
  `item_id` BIGINT NOT NULL,
  CONSTRAINT `fk_threads_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items`(`id`) ON DELETE CASCADE,
  `status` ENUM('waiting_for_participant','waiting_for_trainer', 'closed') NOT NULL,
  `helper_group_id` BIGINT NOT NULL,
  `latest_update_at` DATETIME NOT NULL DEFAULT NOW() COMMENT 'Last time a message was posted or the status was updated.',
  `message_count` INT NOT NULL DEFAULT 0 COMMENT 'Approximation of the number of message sent on the thread.',
  PRIMARY KEY (`participant_id`,`item_id`),
  CONSTRAINT `fk_threads_helper_group_id_groups_id` FOREIGN KEY (`helper_group_id`) REFERENCES `groups` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Discussion thread related to participant-item pair.';

-- +migrate Down
DROP TABLE `threads`;
