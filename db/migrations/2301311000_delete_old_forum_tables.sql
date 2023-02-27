-- +migrate Up
DROP TABLE `messages`;
DROP TABLE `users_threads`;
DROP TABLE `threads`;

-- +migrate Down
CREATE TABLE `messages` (
  `id` BIGINT(19) NOT NULL,
  `thread_id` BIGINT(19) NULL DEFAULT NULL,
  `user_id` BIGINT(19) NULL DEFAULT NULL,
  `submitted_at` DATETIME NULL DEFAULT NULL,
  `published` TINYINT(1) NOT NULL DEFAULT '1',
  `title` VARCHAR(200) NULL DEFAULT '' COLLATE 'utf8_general_ci',
  `body` VARCHAR(2000) NULL DEFAULT '' COLLATE 'utf8_general_ci',
  `trainers_only` TINYINT(1) NOT NULL DEFAULT '0',
  `archived` TINYINT(1) NULL DEFAULT '0',
  `persistant` TINYINT(1) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `thread_id` (`thread_id`) USING BTREE,
  INDEX `fk_messages_user_id_users_group_id` (`user_id`) USING BTREE,
  CONSTRAINT `fk_messages_user_id_users_group_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`group_id`) ON UPDATE NO ACTION ON DELETE SET NULL
)
  COLLATE='utf8_general_ci'
  ENGINE=InnoDB
;

CREATE TABLE `users_threads` (
   `id` BIGINT(19) NOT NULL,
   `user_id` BIGINT(19) NOT NULL,
   `thread_id` BIGINT(19) NOT NULL,
   `lately_viewed_at` DATETIME NULL DEFAULT NULL,
   `participated` TINYINT(1) NOT NULL DEFAULT '0',
   `lately_posted_at` DATETIME NULL DEFAULT NULL,
   `starred` TINYINT(1) NULL DEFAULT NULL,
   PRIMARY KEY (`id`) USING BTREE,
   UNIQUE INDEX `user_id_thread_id` (`user_id`, `thread_id`) USING BTREE,
   INDEX `user_id` (`user_id`) USING BTREE
)
  COLLATE='utf8_general_ci'
  ENGINE=InnoDB
;

CREATE TABLE `threads` (
   `id` BIGINT(19) NOT NULL,
   `type` ENUM('Help','Bug','General') NOT NULL COLLATE 'utf8_general_ci',
   `latest_activity_at` DATETIME NULL DEFAULT NULL,
   `creator_id` BIGINT(19) NOT NULL,
   `item_id` BIGINT(19) NULL DEFAULT NULL,
   `title` VARCHAR(200) NULL DEFAULT NULL COLLATE 'utf8_general_ci',
   `admin_help_asked` TINYINT(1) NOT NULL DEFAULT '0',
   `hidden` TINYINT(1) NOT NULL DEFAULT '0',
   PRIMARY KEY (`id`) USING BTREE
)
  COLLATE='utf8_general_ci'
  ENGINE=InnoDB
;
