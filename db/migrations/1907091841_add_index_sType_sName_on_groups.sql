-- +migrate Up
ALTER TABLE `groups` ADD INDEX `TypeName` (`sType`, `sName`);


-- +migrate Down
ALTER TABLE `groups` DROP INDEX `TypeName`;

