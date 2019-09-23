-- +migrate Up
ALTER TABLE `users_answers` ADD INDEX  `attempt_id` (`attempt_id`);

-- +migrate Down
ALTER TABLE `users_answers` DROP INDEX `attempt_id`;
