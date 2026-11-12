-- +goose Up
-- Nothing in the platform reads these columns. is_admin and access_group_id are loaded
-- into the session user and then ignored. The other six are never read or written
-- except by a dump of every column.
ALTER TABLE `users`
  DROP COLUMN `open_id_identity`,
  DROP COLUMN `password_md5`,
  DROP COLUMN `salt`,
  DROP COLUMN `recover`,
  DROP COLUMN `member_state`,
  DROP COLUMN `help_given`,
  DROP COLUMN `is_admin`,
  DROP COLUMN `access_group_id`;

-- +goose Down
ALTER TABLE `users`
  ADD COLUMN `open_id_identity` varchar(255) DEFAULT NULL COMMENT 'User''s Open Id Identity',
  ADD COLUMN `password_md5` varchar(100) DEFAULT NULL,
  ADD COLUMN `salt` varchar(32) DEFAULT NULL,
  ADD COLUMN `recover` varchar(50) DEFAULT NULL,
  ADD COLUMN `member_state` tinyint NOT NULL DEFAULT '0' COMMENT 'On old website, indicates if the user is a member of France-ioi',
  ADD COLUMN `help_given` int NOT NULL DEFAULT '0' COMMENT 'How many times did the user help others (# of discussions)',
  ADD COLUMN `is_admin` tinyint NOT NULL DEFAULT '0' COMMENT 'Is the user an admin? Not used?',
  ADD COLUMN `access_group_id` bigint DEFAULT NULL;
