-- +migrate Up
DROP TABLE `users_items`;

-- +migrate Down
CREATE TABLE `users_items` (
   `user_id` BIGINT(20) NOT NULL,
   `item_id` BIGINT(20) NOT NULL,
   `active_attempt_id` BIGINT(20) NOT NULL COMMENT 'Current attempt selected by this user.',
   PRIMARY KEY (`user_id`,`item_id`),
   KEY `item_id` (`item_id`),
   KEY `user_id` (`user_id`),
   KEY `fk_users_items_active_attempt_id_attempts_id` (`active_attempt_id`),
   CONSTRAINT `fk_users_items_active_attempt_id_attempts_id` FOREIGN KEY (`active_attempt_id`) REFERENCES `attempts` (`id`) ON DELETE CASCADE,
   CONSTRAINT `fk_users_items_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE,
   CONSTRAINT `fk_users_items_user_id_users_group_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`group_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Information about the activity of users on items';
