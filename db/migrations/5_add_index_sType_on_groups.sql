-- +migrate Up
ALTER TABLE `groups` ADD INDEX `sType` (`sType`);


-- +migrate Down
ALTER TABLE `groups` DROP INDEX `sType`;

