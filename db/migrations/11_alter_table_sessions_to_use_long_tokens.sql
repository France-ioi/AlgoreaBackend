-- +migrate Up
ALTER TABLE `sessions` DROP PRIMARY KEY;
ALTER TABLE `sessions` MODIFY COLUMN `sAccessToken` VARBINARY(2000);
ALTER TABLE `sessions` ADD INDEX `sAccessTokenPrefix` (`sAccessToken`(767));

-- +migrate Down
ALTER TABLE `sessions` DROP INDEX `sAccessTokenPrefix`;
ALTER TABLE `sessions` MODIFY COLUMN `sAccessToken` VARBINARY(64);
ALTER TABLE `sessions` ADD PRIMARY KEY (`sAccessToken`);
