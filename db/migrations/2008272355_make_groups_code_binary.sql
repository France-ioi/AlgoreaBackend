-- +migrate Up
ALTER TABLE `groups`
    MODIFY COLUMN `code` VARBINARY(50) DEFAULT NULL COMMENT 'Code that can be used to join the group (if it is opened)';

-- +migrate Down
ALTER TABLE `groups`
    MODIFY COLUMN `code` VARCHAR(50) DEFAULT NULL COMMENT 'Code that can be used to join the group (if it is opened)';
