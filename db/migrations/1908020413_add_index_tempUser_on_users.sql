-- +migrate Up
ALTER TABLE `users` ADD INDEX  `tempUser` (`tempUser`);

-- +migrate Down
ALTER TABLE `users` DROP INDEX `tempUser`;
