-- +migrate Up
ALTER TABLE `groups` ADD UNIQUE INDEX `sPassword` (`sPassword`);

-- +migrate Down
ALTER TABLE `groups` DROP INDEX `sPassword`;
