-- +migrate Up
ALTER TABLE `users` MODIFY COLUMN `tempUser` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether it is a temporary user. If so, the user will be deleted soon.';

-- +migrate Down
ALTER TABLE `users` MODIFY COLUMN `tempUser` tinyint(1) NOT NULL COMMENT 'Whether it is a temporary user. If so, the user will be deleted soon.';
