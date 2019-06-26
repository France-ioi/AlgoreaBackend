-- +migrate Up
ALTER TABLE `groups` ADD INDEX IDType (ID, sType);


-- +migrate Down
ALTER TABLE `groups` DROP INDEX IDType;

