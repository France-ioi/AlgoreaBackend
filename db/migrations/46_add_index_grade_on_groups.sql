-- +migrate Up
ALTER TABLE `groups` ADD INDEX  `grade` (`grade`);

-- +migrate Down
ALTER TABLE `groups` DROP INDEX `grade`;
