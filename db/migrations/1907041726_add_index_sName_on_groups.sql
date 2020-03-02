-- +migrate Up
ALTER TABLE `groups` ADD INDEX `sName` (`sName`);


-- +migrate Down
ALTER TABLE `groups` DROP INDEX `sName`;

